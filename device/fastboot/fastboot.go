package fastboot

import (
	"github.com/amo13/anarchy-droid/logger"
	"github.com/amo13/anarchy-droid/helpers"
	"github.com/amo13/anarchy-droid/device/adb"

	"runtime"
	"strings"
	"fmt"
)

var Sudopw string = ""
var Nosudo bool = false

func fastboot_command() string {
	switch runtime.GOOS {
	case "windows":
		return "bin\\platform-tools\\fastboot.exe"
	case "darwin":
		return "bin/platform-tools/fastboot"
	default:	// linux
		if Nosudo {
			return "bin/platform-tools/fastboot"
		} else if Sudopw == "" {
			return "sudo bin/platform-tools/fastboot"
		} else {
			return "echo " + Sudopw + " | sudo -S bin/platform-tools/fastboot"
		}
	}
}

// Returns the non-empty or longer one of stdout and stderr for a given fastboot command
func Cmd(command string) (stdout string, err error) {
	if !available() {
		return "", fmt.Errorf("disconnected")
	}

	stdout, stderr := helpers.Cmd(fastboot_command() + " " + command)
	if stdout != "" && stderr == "" {
		return strings.Trim(strings.Trim(stdout, "\n"), " "), nil
	} else if stdout == "" && stderr != "" {
		return strings.Trim(strings.Trim(stderr, "\n"), " "), nil
	} else if stdout != "" && stderr != "" {
		if len(stdout) >= len(stderr) {
			return strings.Trim(strings.Trim(stdout, "\n"), " "), nil
		} else {
			return strings.Trim(strings.Trim(stderr, "\n"), " "), nil
		}
	} else {
		return "", fmt.Errorf("disconnected")
	}
}

// Check for disconnection error
func unavailable(err error) bool {
	if err != nil {
		if err.Error() == "disconnected" {
			return true
		} else {
			logger.LogError("Unknown fastboot error:", err)
		}
	}

	return false
}

func available() bool {
	if State() == "connected" {
		return true
	} else {
		return false
	}
}

func State() string {
	stdout, _ := helpers.Cmd(fastboot_command() + " devices")

	if stdout == "" {
		return "disconnected"
	} else {
		return "connected"
	}
}

func Reboot(target string) error {
	if !available() {
		return fmt.Errorf("disconnected")
	}

	if target == "bootloader" {
		_, err := Cmd("reboot bootloader")
		return err
	} else {
		_, err := Cmd("reboot")
		return err
	}
}

func GetVarMap() (map[string]string, error) {
	if !available() {
		return make(map[string]string), fmt.Errorf("disconnected")
	}

	stdout, err := Cmd("getvar all")
	if unavailable(err) {
		return make(map[string]string), err
	}

	var m map[string]string
	var lines []string

	lines = strings.Split(strings.Trim(stdout, "\n"), "\n")
	m = make(map[string]string)
	for _, line := range lines {
		// drop lines not starting with (bootloader) or not containing ": "
		if !strings.HasPrefix(line, "(bootloader) ") || !strings.Contains(line, ": ") { continue }
		z := strings.Split(line, ": ")
	    m[z[0][13:]] = strings.Trim(z[1], "\n")
	}

	return m, err
}

func GetVar(v string) (string, error) {
	m, err := GetVarMap()
	if unavailable(err) {
		return "", err
	}

	return m[v], nil
}

func Model() (string, error) {
	m, err := GetVarMap()
	if unavailable(err) {
		return "", err
	}

	return ModelFromVarMap(m), nil
}

func ModelFromVarMap(m map[string]string) string {
	v := m["product"]

	if v != "" {
		return v
	} else {
		return m["sku"]
	}
}

func Imei() (string, error) {
	m, err := GetVarMap()
	if unavailable(err) {
		return "", err
	}

	return ImeiFromVarMap(m), nil
}

func ImeiFromVarMap(m map[string]string) string {
	return m["imei"]
}

func IsAB() (bool, error) {
	m, err := GetVarMap()
	if unavailable(err) {
		return false, err
	}

	return IsABFromVarMap(m), nil
}

func IsABFromVarMap(m map[string]string) bool {
	v := m["slot-slot"]

	return v == "2"
}

func ActiveSlot() (string, error) {
	m, err := GetVarMap()
	if unavailable(err) {
		return "", err
	}

	return ActiveSlotFromVarMap(m), nil
}

func ActiveSlotFromVarMap(m map[string]string) string {
	current := m["current-slot"]
	running := m["running-slot"]

	if current != "" {
		return strings.Trim(strings.ToLower(current), "_")
	} else {
		return strings.Trim(strings.ToLower(running), "_")
	}
}

// In doubt, returns false
func IsUnlocked() (bool, error) {
	m, err := GetVarMap()
	if unavailable(err) {
		return false, err
	}

	return IsUnlockedFromVarMap(m), nil
}

func IsUnlockedFromVarMap(m map[string]string) bool {
	unlocked := m["unlocked"]
	securestate := m["securestate"]

	if strings.ToLower(unlocked) == "yes" || strings.ToLower(unlocked) == "true" || strings.ToLower(securestate) == "unlocked" {
		return true
	}

	return false
}

// Retrieves the needed data to unlock the bootloader
// returns an "unlocked" error if already unlocked
func GetUnlockData(brand string) (string, error) {
	unlocked, err := IsUnlocked()
	if err != nil {
		return "", err
	}
	if unlocked {
		return "", fmt.Errorf("unlocked")
	}
	switch strings.ToLower(brand) {
	case "motorola":
		return GetUnlockDataMotorola()
	case "sony":
		return GetUnlockDataSony()
	default:
		return "", fmt.Errorf("not implemented")
	}
}

func GetUnlockDataMotorola() (string, error) {
	data, err := Cmd("oem get_unlock_data")
	if unavailable(err) {
		return "", err
	}

	// Log the result for reference
	logger.Log("------- fastboot oem get_unlock_data -------")
	logger.Log(data)
	logger.Log("--------------------------------------------")

	// parse and scrub the data
	lines := strings.Split(data, "\n")
	result := ""
	if strings.HasPrefix(lines[0], "(bootloader) ") {
		if IsUnlockDataMotorolaParsable(data) {
			for _, line := range lines {
				if strings.HasPrefix(line, "(bootloader) ") {
					result = result + strings.Split(line, " ")[1]
				} else {
					continue
				}
			}
		}
	} else if strings.HasPrefix(lines[0], "INFO") {
		if IsUnlockDataMotorolaParsable(data) {
			for _, line := range lines {
				if strings.HasPrefix(line, "INFO") {
					result = result + strings.Trim(line[4:], " ")
				} else {
					continue
				}
			}
		}
	}

	if result != "" {
		return result, nil
	} else {
		return "", fmt.Errorf("unable to parse unlock data")
	}
}

func IsUnlockDataMotorolaParsable(data string) bool {
	return !strings.Contains(data, "slot") && !strings.Contains(data, "not found") && !strings.Contains(data, "nlock")
}

func GetUnlockDataSony() (string, error) {
	if available() {
		return Imei()
	} else {
		if helpers.IsStringInSlice(adb.State(), []string{"android","recovery"}) {
			return adb.Imei()
		} else {
			return "", fmt.Errorf("disconnected")
		}
	}
}

func Unlock(brand string, unlock_code string) error {
	switch strings.ToLower(brand) {
	case "motorola":
		return UnlockMotorola(unlock_code)
	case "sony":
		return UnlockSony(unlock_code)
	case "oneplus":
		return UnlockOneplus()
	case "nvidia":
		return UnlockNvidia()
	case "generic":
		return UnlockGeneric()
	default:
		return fmt.Errorf("not implemented")
	}
}

func UnlockMotorola(unlock_code string) error {
	result, err := Cmd("oem unlock " + unlock_code)
	if unavailable(err) {
		return err
	}

	// Log the result for reference
	logger.Log("--------- fastboot oem unlock ... ---------")
	logger.Log(result)
	logger.Log("-------------------------------------------")

	if strings.Contains(strings.ToLower(result), "allow oem unlock") {
		logger.Log("OEM unlock has apparently not been enabled...")
		return fmt.Errorf("not allowed")
	} else if strings.Contains(strings.ToLower(result), "re-run this command") {
		logger.Log("Re-running the unlock command to confirm unlock request...")
		return UnlockMotorola(unlock_code)
	} else if strings.Contains(strings.ToLower(result), "failed") {
		logger.Log("bootloader unlock failed")
		return fmt.Errorf("failed")
	} else if strings.Contains(strings.ToLower(result), "already unlocked") {
		logger.Log("bootloader already unlocked")
		return nil
	} else if strings.Contains(strings.ToLower(result), "is unlocked") ||
	strings.Contains(strings.ToLower(result), "succe") ||
	strings.Contains(strings.ToLower(result), "okay") ||
	strings.Contains(strings.ToLower(result), "complete") {
		logger.Log("bootloader successfully unlocked")
		return nil
	} else {
		logger.Log("unknown response")
		return fmt.Errorf("unknown response")
	}
}

func UnlockSony(unlock_code string) error {
	result, err := Cmd("oem unlock 0x" + unlock_code)
	if unavailable(err) {
		return err
	}

	// Log the result for reference
	logger.Log("--------- fastboot oem unlock ... ---------")
	logger.Log(result)
	logger.Log("-------------------------------------------")

	if strings.Contains(strings.ToLower(result), "not allowed") {
		logger.Log("OEM unlock has apparently not been enabled...")
		return fmt.Errorf("not allowed")
	} else if strings.Contains(strings.ToLower(result), "already") {
		logger.Log("bootloader already unlocked")
		return nil
	} else if strings.Contains(strings.ToLower(result), "failed") {
		logger.Log("bootloader unlock failed")
		return fmt.Errorf("failed")
	} else if strings.Contains(strings.ToLower(result), "re-run this command") {
		logger.Log("Re-running the unlock command to confirm unlock request...")
		return UnlockSony(unlock_code)
	} else if strings.Contains(strings.ToLower(result), "is unlocked") ||
	strings.Contains(strings.ToLower(result), "succe") ||
	strings.Contains(strings.ToLower(result), "okay") {
		logger.Log("bootloader successfully unlocked")
		return nil
	} else {
		logger.Log("unknown response")
		return fmt.Errorf("unknown response")
	}
}

func UnlockOneplus() error {
	return UnlockGeneric()
}

func UnlockNvidia() error {
	return UnlockGeneric()
}

func UnlockGeneric() error {
	result, err := Cmd("oem unlock")
	if unavailable(err) {
		return err
	}

	// Log the result for reference
	logger.Log("----------- fastboot oem unlock -----------")
	logger.Log(result)
	logger.Log("-------------------------------------------")

	if strings.Contains(strings.ToLower(result), "allow oem unlock") {
		logger.Log("OEM unlock has apparently not been enabled...")
		return fmt.Errorf("not allowed")
	} else if strings.Contains(strings.ToLower(result), "failed") {
		logger.Log("bootloader unlock failed")
		return fmt.Errorf("failed")
	} else if strings.Contains(strings.ToLower(result), "Total time: 0.000s") {
		logger.Log("bootloader already unlocked")
		return nil
	} else if strings.Contains(strings.ToLower(result), "re-run this command") {
		logger.Log("Re-running the unlock command to confirm unlock request...")
		return UnlockGeneric()
	} else if strings.Contains(strings.ToLower(result), "is unlocked") ||
	strings.Contains(strings.ToLower(result), "succe") ||
	strings.Contains(strings.ToLower(result), "okay") {
		logger.Log("bootloader successfully unlocked")
		return nil
	} else {
		logger.Log("unknown response")
		return fmt.Errorf("unknown response")
	}
}

func FlashStartupLogo(logo_file string) error {
	result, err := Cmd("flash logo " + logo_file)
	if unavailable(err) {
		return err
	}

	// Log the result for reference
	logger.Log("--------- fastboot flash logo ... ---------")
	logger.Log(result)
	logger.Log("-------------------------------------------")

	if strings.Contains(result, "Sending") && strings.Contains(result, "Writing") && strings.Contains(result, "OKAY") {
		return nil
	} else {
		logger.Log("unknown response")
		return fmt.Errorf("unknown response")
	}
}

func BootRecovery(brand string, img_file string) error {
	switch strings.ToLower(brand) {
	case "motorola":
		return bootRecoveryMotorola(img_file)
	case "sony":
		return bootRecoverySony(img_file)
	case "oneplus":
		return bootRecoveryOneplus(img_file)
	case "nvidia":
		return bootRecoveryNvidia(img_file)
	case "generic":
		return bootRecoveryGeneric(img_file)
	default:
		return fmt.Errorf("not implemented")
	}
}

func bootRecoveryMotorola(img_file string) error {
	return bootRecoveryGeneric(img_file)
}

func bootRecoverySony(img_file string) error {
	return bootRecoveryGeneric(img_file)
}

func bootRecoveryOneplus(img_file string) error {
	return bootRecoveryGeneric(img_file)
}

func bootRecoveryNvidia(img_file string) error {
	return bootRecoveryGeneric(img_file)
}

func bootRecoveryGeneric(img_file string) error {
	result, err := Cmd("boot " + img_file)
	if unavailable(err) {
		return err
	}

	// Log the result for reference
	logger.Log("------------ fastboot boot ... ------------")
	logger.Log(result)
	logger.Log("-------------------------------------------")

	if strings.Contains(result, "Sending") && strings.Contains(result, "Booting") && strings.Contains(result, "OKAY") {
		return nil
	} else {
		logger.Log("unknown response")
		logger.Log("is the recovery booting?")
		return fmt.Errorf("unknown response")
	}
}

func FlashRecovery(brand string, img_file string, partition string) error {
	switch strings.ToLower(brand) {
	case "motorola":
		return flashRecoveryMotorola(img_file, partition)
	case "sony":
		return flashRecoverySony(img_file, partition)
	case "oneplus":
		return flashRecoveryOneplus(img_file, partition)
	case "nvidia":
		return flashRecoveryNvidia(img_file, partition)
	case "generic":
		return flashRecoveryGeneric(img_file, partition)
	default:
		return fmt.Errorf("not implemented")
	}
}

func flashRecoveryMotorola(img_file string, partition string) error {
	return flashRecoveryGeneric(img_file, partition)
}

func flashRecoverySony(img_file string, partition string) error {
	return flashRecoveryGeneric(img_file, partition)
}

func flashRecoveryOneplus(img_file string, partition string) error {
	return flashRecoveryGeneric(img_file, partition)
}

func flashRecoveryNvidia(img_file string, partition string) error {
	return flashRecoveryGeneric(img_file, partition)
}

func flashRecoveryGeneric(img_file string, partition string) error {
	result, err := Cmd("flash " + partition + " " + img_file)
	if unavailable(err) {
		return err
	}

	// Log the result for reference
	logger.Log("----------- fastboot flash ... ------------")
	logger.Log(result)
	logger.Log("-------------------------------------------")

	if strings.Contains(result, "no such partition") || strings.Contains(result, "invalid partition") {
		logger.Log("unknown partition")
		return fmt.Errorf("unknown partition")
	} else {
		if strings.Contains(result, "Sending") && strings.Contains(result, "Writing") && strings.Contains(result, "OKAY") {
			return nil
		} else {
			logger.Log("unknown response")
			return fmt.Errorf("unknown response")
		}
	}
}