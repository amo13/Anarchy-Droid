package adb

import (
	"strings"
	"strconv"
	"runtime"
	"regexp"
	"fmt"

	"github.com/amo13/anarchy-droid/helpers"
	"github.com/amo13/anarchy-droid/logger"
)

var Sudopw string = ""
var Nosudo bool = false

func adb_command() string {
	switch runtime.GOOS {
	case "windows":
		return "bin\\platform-tools\\adb.exe"
	case "darwin":
		return "bin/platform-tools/adb"
	default:	// linux
		if Nosudo {
			return "bin/platform-tools/adb"
		} else if Sudopw == "" {
			return "sudo bin/platform-tools/adb"
		} else {
			return "echo " + Sudopw + " | sudo -S bin/platform-tools/adb"
		}
	}
}

// Returns trimmed stdout of a given adb command
func Cmd(command string) (stdout string, err error) {
	stdout, stderr := helpers.Cmd(adb_command() + " " + command)
	if stderr != "" {
		if strings.Contains(stderr, "no devices/emulators found") {
			return "", fmt.Errorf("disconnected")
		} else if strings.Contains(stderr, "device offline") {
			return "", fmt.Errorf("disconnected")
		} else if strings.Contains(stderr, "device unauthorized") {
			return "", fmt.Errorf("unauthorized")
		} else if strings.Contains(stderr, "device still authorizing") {
			return "", fmt.Errorf("unauthorized")
		} else if command == "kill-server" && (strings.Contains(stderr, "Connection refused") || strings.Contains(stderr, "cannot connect to daemon")) {
			return "", fmt.Errorf("connection refused")
		} else if strings.Contains(stderr, "daemon not running; starting now") {
			return stdout, nil
		} else if strings.Contains(stderr, "[sudo]") {
			logger.Log("Stderr contains [sudo]")
			if strings.Contains(stderr, "Connection refused") && command == "kill-server" {
				return "", fmt.Errorf("connection refused")
			} else {
				logger.Log("Bug: sudo password prompt instead of command output. Retrying " + command)
				return "", KillServer()
			}
		} else if strings.Contains(stderr, "adb: failed to read command: Success") || strings.Contains(stderr, "adb: failed to read command: No error") {
			return stdout, nil
		}
		
		logger.LogError("ADB command " + command + " gave an unexpected error:", fmt.Errorf("stdout: " + stdout + "; stderr: " + stderr))
	}

	return strings.Trim(strings.Trim(stdout, "\n"), " "), nil
}

// Check for disconnection error or suddenly unauthorized error
func unavailable(err error) bool {
	if err != nil {
		if err.Error() == "disconnected" || err.Error() == "unauthorized" {
			return true
		} else {
			logger.LogError("Unknown ADB error:", err)
		}
	}

	return false
}

func StartServer() error {
	_, err := Cmd("start-server")
	return err
}

func KillServer() error {
	_, err := Cmd("kill-server")
	return err
}

func State() string {
	// Call helpers.Cmd because we need stdout and stderr
	stdout, stderr := helpers.Cmd(adb_command() + " get-state")

	if strings.HasPrefix(stderr, "error: no device") {
		return "disconnected"
	} else if strings.HasPrefix(stderr, "error: device offline") {
		return "disconnected"
	} else if strings.HasPrefix(stderr, "error: insufficient permissions") {
		return "unauthorized"
	} else if strings.HasPrefix(stderr, "error: device unauthorized") {
		return "unauthorized"
	} else if strings.HasPrefix(stderr, "error: device still authorizing") {
		return "unauthorized"
	} else if strings.HasPrefix(stdout, "device") {
		booting, _ := IsBooting()
		if booting {
			return "booting"
		} else {
			return "android"
		}
	} else if strings.HasPrefix(stdout, "sideload") {
		return "sideload"
	} else if strings.HasPrefix(stdout, "recovery") {
		return "recovery"
	} else if strings.Contains(stderr, "daemon not running; starting now") {
		return State()
	} else {
		logger.Log("unknown state:\nStdOUT:" + stdout + "\nStdERR:" + stderr)
		return("unknown")
	}
}

func IsConnected() bool {
	if helpers.IsStringInSlice(State(), []string{"android","recovery","unauthorized","sideload","booting"}) {
		return true
	} else {
		return false
	}
}

func IsReady() bool {
	switch State() {
	case "recovery":
		return true
	case "android":
		booting, err := IsBooting()
		if unavailable(err) {
			return false
		}
		if booting {
			return false
		} else {
			return true
		}
	default:
		return false
	}
}

func IsBootComplete() (bool, error) {
	// Do not query the full props map before booting is completed
	bootcomplete, stderr := helpers.Cmd(adb_command() + " shell getprop dev.bootcomplete")
	if stderr != "" {
		logger.Log("Error executing adb shell getprop dev.bootcomplete: " + stderr)
		return true, fmt.Errorf(stderr)
	}

	if strings.HasPrefix(bootcomplete, "1") {
		return true, nil
	} else {
		// Do not query the full props map before booting is completed
		boot_completed, stderr := helpers.Cmd(adb_command() + " shell getprop sys.boot_completed")
		if stderr != "" {
			logger.Log("Error executing adb shell getprop sys.boot_completed: " + stderr)
			return true, fmt.Errorf(stderr)
		}

		if strings.HasPrefix(boot_completed, "1") {
			return true, nil
		} else {
			return false, nil
		}
	}
}

func IsBooting() (bool, error) {
	complete, err := IsBootComplete()
	if unavailable(err) {
		return false, err
	}

	stdout, _ := helpers.Cmd(adb_command() + " get-state")
	if strings.HasPrefix(stdout, "device") && !complete {
		return true, nil
	} else {
		return false, nil
	}
}

func Reboot(target string) (err error) {
	switch strings.ToLower(target) {
	case "fastboot":
		_, err = Cmd("reboot " + strings.ToLower("bootloader"))
	case "heimdall":
		_, err = Cmd("reboot " + strings.ToLower("download"))
	case "bootloader":
		b, err := Brand()
		if err != nil {
			return err
		}
		if b == "samsung" {
			return Reboot("heimdall")
		} else {
			return Reboot("fastboot")
		}
	case "recovery","sideload","sideload-auto-reboot","download":
		_, err = Cmd("reboot " + strings.ToLower(target))
	default:
		_, err = Cmd("reboot")
	}
	if unavailable(err) {
		return err
	}

	return nil
}

func WhoAmI() (user string, err error) {
	user, err = Cmd("shell whoami")
	if unavailable(err) {
		return "", err
	}

	return user, nil
}

func GetPropMap() (map[string]string, error) {
	stdout, err := Cmd("shell getprop")
	if unavailable(err) {
		return make(map[string]string), err
	}

	var m map[string]string
	var lines []string

	// lines = strings.Split(strings.Trim(stdout, "\n"), "\n")
	lines = helpers.StringToLinesSlice(stdout)
	m = make(map[string]string)
	for _, pair := range lines {
		// Remove trailing carriage return if found
		if strings.HasSuffix(pair, string(13)) {
			pair = pair[:len(pair)-1]
		}
		// drop malformed prop lines (e.g. containing line breaks)
		if !strings.HasPrefix(pair, "[") || !strings.HasSuffix(pair, "]") {
			logger.Log("Dropped line from ADB getprop: " + pair)
			continue
		}
	    z := strings.Split(pair, ": ")
	    m[strings.Trim(strings.Trim(z[0], "["), "]")] = strings.Trim(strings.Trim(z[1], "["), "]")
	}

	return m, err
}

func GetProp(prop string) (string, error) {
	props, err := GetPropMap()
	if unavailable(err) {
		return "", err
	}

	return props[prop], nil
}

func Imei() (string, error) {
	maj, err := MajorAndroidVersion()
	if unavailable(err) {
		return "", err
	}

	imei := "not found"

	if maj >= 5 {
		s, err := Cmd("shell service call iphonesubinfo 1")
		if unavailable(err) {
			return "", err
		}

		re1 := regexp.MustCompile(`\d\.`)
		re2 := regexp.MustCompile(`\d`)
		imei = strings.Join(re2.FindAllString(strings.Join(re1.FindAllString(s, -1), ""), -1), "")
	} else {
		s, err := Cmd("shell dumpsys iphonesubinfo")
		if unavailable(err) {
			return "", err
		}
		
		re := regexp.MustCompile(`\d{15,}`)
		imei = strings.Join(re.FindAllString(s, -1), "")
	}

	return imei, nil
}

func ShowImeiOnDeviceScreen() error {
	_, err := Cmd("shell am start -n com.android.settings/com.android.settings.deviceinfo.ImeiInformation")
	if unavailable(err) {
		return err
	}

	return nil
}

func AndroidVersion() (string, error) {
	s, err := GetProp("ro.build.version.release")
	if unavailable(err) {
		return "0", err
	}

	return s, nil
}

func MajorAndroidVersion() (int, error) {
	v, err := AndroidVersion()
	if unavailable(err) {
		return 0, err
	}

	i, err := strconv.Atoi(strings.Split(v, ".")[0])
	if err != nil {
		logger.LogError("", fmt.Errorf("Unable to retrieve the major Android version: String \"%s\" to Int conversion failed.\n", v))
		return 0, err
	}

	return i, nil
}

// Read codename from adb props. Unreliable, therefore relied on
// only if lookup from model to codename is unsuccessful
func Codename() (string, error) {
	props, err := GetPropMap()
	if unavailable(err) {
		return "", err
	}

	return CodenameFromPropMap(props), nil
}

func CodenameFromPropMap(props map[string]string) string {
	codename := props["ro.build.product"]
	if codename == "" {
		codename = props["ro.product.device"]
	}
	if codename == "" {
		codename = props["ro.product.name"]
	}

	return codename
}

func Model() (string, error) {
	props, err := GetPropMap()
	if unavailable(err) {
		return "", err
	}

	return ModelFromPropMap(props), nil
}

func ModelFromPropMap(props map[string]string) string {
	model := props["ro.product.model"]
	if model == "" {
		return props["ro.omni.device"]
	}

	return model
}

func Brand() (string, error) {
	props, err := GetPropMap()
	if unavailable(err) {
		return "", err
	}

	return BrandFromPropMap(props), nil
}

func BrandFromPropMap(props map[string]string) string {
	brand := props["ro.product.brand"]
	if brand == "" {
		return props["ro.product.manufacturer"]
	}

	return brand
}

func IsAB() (bool, error) {
	props, err := GetPropMap()
	if unavailable(err) {
		return false, err
	}

	return IsABFromPropMap(props), nil
}

func IsABFromPropMap(props map[string]string) bool {
	prop := props["ro.build.ab_update"]

	return strings.ToLower(prop) == "true"
}

func CpuArch() (string, error) {
	props, err := GetPropMap()
	if unavailable(err) {
		return "", err
	}
	
	return CpuArchFromPropMap(props)
}

func CpuArchFromPropMap(props map[string]string) (string, error) {
	prop := props["ro.product.cpu.abi"]

	switch strings.ToLower(prop) {
	case "armeabi-v7a":
		return "arm", nil
	case "arm64-v8a":
		return "arm64", nil
	case "x86", "x86_64":
		return strings.ToLower(prop), nil
	case "":
		return "", nil
	default:
		return "", fmt.Errorf("unknown cpu architecture: %s", prop)
	}
}

func IsCustomRomFromMap(props map[string]string) bool {
	if props["ro.build.type"] == "userdebug" || 
	strings.Contains(props["ro.build.flavor"], "userdebug") ||
	strings.Contains(props["ro.build.display.id"], "userdebug") {
		return true
	} else {
		return false
	}
}

func Push(local string, remote string) error {
	_, err := Cmd("push " + local + " " + remote)
	return err
}

func Pull(remote string, local string) error {
	_, err := Cmd("pull " + remote + " " + local)
	return err
}

func Remount() error {
	_, err := Cmd("remount")
	return err
}

func Root() error {
	_, err := Cmd("root")
	return err
}

func Unroot() error {
	_, err := Cmd("unroot")
	return err
}