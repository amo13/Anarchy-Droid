package device

import (
	"os"
	"fmt"
	"strings"

	"github.com/amo13/anarchy-droid/logger"
	"github.com/amo13/anarchy-droid/lookup"
	"github.com/amo13/anarchy-droid/helpers"
	"github.com/amo13/anarchy-droid/device/adb"
	"github.com/amo13/anarchy-droid/device/twrp"
	"github.com/amo13/anarchy-droid/device/fastboot"
	"github.com/amo13/anarchy-droid/device/heimdall"
)

var D1 = NewDevice()
var UnlockableBrands = []string{"sony", "motorola", "samsung", "nvidia", "oneplus"}

func NewDevice() Device {
	return Device{
		ObserveMe: true,
		State: "disconnected",
		State_request: "",
		State_reached: make(chan string),
		States_history: []string{},
		Flashing: false,
		Scanning: false,
		Model: "",
		Codename: "",
		Brand: "",
		IsBrandUnlockable: false,
		Name: "",
		Arch: "",
		Imei: "",
		IsAB: false,
		IsAB_checked: false,
		IsUnlocked: false,
		IsSupported: true,
		IsSupported_checked: false,
		TwrpVersionConnected: "",
		AdbProps: map[string]string{},
		FastbootVars: map[string]string{},
	}
}

type Device struct {
	ObserveMe bool
	State string
	State_request string
	State_reached chan string
	States_history []string
	Flashing bool
	Scanning bool
	Model string
	Codename string
	Brand string
	IsBrandUnlockable bool
	Name string
	Arch string
	Imei string
	IsAB bool
	IsAB_checked bool
	IsUnlocked bool
	IsSupported bool
	IsSupported_checked bool
	TwrpVersionConnected string
	AdbProps map[string]string
	FastbootVars map[string]string
}

func (d *Device) Test(model string) {
	d.ObserveMe = false
	d.State = "simulation"
	adb.Simulation = true
	d.Model = model
	d.Arch = "simulation"
	d.Imei = "simulation"
	d.IsAB_checked = true
	d.IsUnlocked = true

	d.ReadMissingProps()
}

func (d *Device) GetState() string {
	adb_state := adb.State()
	if helpers.IsStringInSlice(adb_state, []string{"android","recovery","unauthorized","sideload","booting"}) {
		return adb_state
	} else if adb_state == "disconnected" {
		fastboot_state := fastboot.State()
		if fastboot_state == "connected" {
			return "fastboot"
		} else if fastboot_state == "disconnected" {
			heimdall_state := heimdall.State()
			if heimdall_state == "connected" {
				return "heimdall"
			} else if heimdall_state == "disconnected" {
				return "disconnected"
			} else {
				logger.LogError("Cannot determine heimdall connection state", fmt.Errorf("unknown heimdall state"))
			}
		} else {
			logger.LogError("Cannot determine fastboot connection state", fmt.Errorf("unknown fastboot state"))
		}
	} else {
		logger.LogError("Cannot determine ADB connection state", fmt.Errorf("unknown adb state: " + adb_state))
	}

	return "unknown"
}

func (d *Device) Reboot(target string) (err error) {
	switch d.State {
	case "android", "recovery":
		err = adb.Reboot(target)
	case "fastboot":
		err = fastboot.Reboot(target)
	case "heimdall":
		err = heimdall.Reboot()
	default:
		err = fmt.Errorf("Cannot reboot device right now")
	}

	return err
}

func (d *Device) HandleStateRequest(req_state string) {
	if !d.Flashing {
		logger.Log("Skipped device state request to " + req_state + " because the user cancelled flashing")
		return
	}
	
	if req_state == "sideload" && d.State == "recovery" {
		err := twrp.OpenSideload()
		if err != nil {
			logger.LogError("Unable to open sideload:", err)
		}
	} else if req_state == "recovery" && d.State == "sideload" {
		// Do nothing: simply wait for sideload to finish
	} else {
		err := d.Reboot(req_state)
		if err != nil && err.Error() != "Cannot reboot device right now" {
			logger.LogError("Unable to reboot device to " + req_state + ":", err)
		}
	}
}

func (d *Device) Unlock() error {
	if !d.Flashing {
		logger.Log("User cancelled flashing")
		return fmt.Errorf("cancelled")
	}

	switch strings.ToLower(d.Brand) {
	case "":
		return fmt.Errorf("Unknown brand")
	case "samsung":
		return nil	// No need on samsung devices
	default:
		unlock_data, err := d.GetUnlockData()
		if err != nil && err.Error() != "No unlock data needed" {
			return err
		}

		err = d.DoUnlock(unlock_data)
		if err != nil {
			return err
		}

		d.IsUnlocked = true
		return nil
	}
}

func (d *Device) DoUnlock(unlock_data string) error {
	if !d.Flashing {
		logger.Log("User cancelled flashing")
		return fmt.Errorf("cancelled")
	}
	
	switch strings.ToLower(d.Brand) {
	case "":
		return fmt.Errorf("Unknown brand")
	case "samsung":
		return fmt.Errorf("No unlock needed on Samsung")
	case "motorola":
		return d.UnlockMotorola(unlock_data)
	case "sony":
		return d.UnlockSony(unlock_data)
	default:
		return fastboot.UnlockGeneric()
	}
}

func (d *Device) GetUnlockData() (string, error) {
	if !d.Flashing {
		logger.Log("User cancelled flashing")
		return "", fmt.Errorf("cancelled")
	}
	
	switch strings.ToLower(d.Brand) {
	case "":
		return "", fmt.Errorf("Unknown brand")
	case "samsung", "oneplus", "nvidia":
		return "", fmt.Errorf("No unlock data needed")
	case "sony":
		if d.Imei != "" {
			return d.Imei, nil
		} else {
			// Also able to return Imei while ADB connected
			return fastboot.GetUnlockData(d.Brand)
		}
	default:
		d.State_request = "fastboot"
		<-d.State_reached	// blocks until fastboot is reached
		return fastboot.GetUnlockData(d.Brand)
	}
}

func (d *Device) UnlockMotorola(unlock_code string) error {
	if unlock_code == "" {
		return fmt.Errorf("No unlock code provided")
	}

	d.State_request = "fastboot"
	<-d.State_reached	// blocks until fastboot is reached

	return fastboot.UnlockMotorola(unlock_code)
}

func (d *Device) UnlockSony(unlock_code string) error {
	if unlock_code == "" {
		return fmt.Errorf("No unlock code provided")
	}

	d.State_request = "fastboot"
	<-d.State_reached	// blocks until fastboot is reached

	return fastboot.UnlockSony(unlock_code)
}

// Boot a given recovery image.
// If a partition name other than "boot" can be looked up,
// try to flash the image to the looked up partition
// Returns user instructions to boot recovery after flash (key combination)
func (d *Device) BootRecovery(img_file string) (string, error) {
	if !d.Flashing {
		logger.Log("User cancelled flashing")
		return "", fmt.Errorf("cancelled")
	}
	
	_, err := os.Stat(img_file)
	if os.IsNotExist(err) {
	  return "", fmt.Errorf("%s does not exist, can't flash or boot it", img_file)
	}

	if !helpers.IsStringInSlice(d.State, []string{"fastboot", "heimdall"}) {
		d.State_request = "bootloader"
		<-d.State_reached	// Blocks until bootloader is connected
	}

	user_instructions, err := lookup.RecoveryKeyCombination(d.Codename)
	if err != nil {
		logger.LogError("Unable to lookup recovery key combination:", err)
		return "", err
	}
	if user_instructions == "" {
		user_instructions, err = lookup.RecoveryKeyCombination(d.Brand)
		if err != nil {
			logger.LogError("Unable to lookup recovery key combination:", err)
			return "", err
		}
		// Set default instructions if none were found
		if user_instructions == "" {
			user_instructions = "Please reboot directly into recovery without booting Android in between. Unfortunately, no instructions have been found on how to do this with your device, sorry.\nHint: Usually, you can achieve this by holding a combination of hardware buttons on your device."
		}
	}

	partition, err := lookup.RecoveryPartition(d.Codename)
	if err != nil {
		logger.Log("Unable to lookup the recovery partition name for", d.Codename)
		return "", err
	}

	if d.State == "fastboot" {
		if partition == "" || strings.ToLower(partition) == "boot" {
			return "", fastboot.BootRecovery(d.Brand, img_file)
		} else {
			return user_instructions, fastboot.FlashRecovery(d.Brand, img_file, partition)
		}
	} else if d.State == "heimdall" {
		if partition == "" {
			return user_instructions, heimdall.FlashRecovery(img_file, "RECOVERY")
		} else {
			return user_instructions, heimdall.FlashRecovery(img_file, partition)
		}
	} else {
		return "", fmt.Errorf("Cannot flash or boot recovery: device bootloader not connected")
	}
}

// Flashes a rom zip file using TWRP and adb sideload.
// "Clean flash" (formating data) if wipe == "clean"
func (d *Device) FlashRom(zip_file string, wipe string) error {
	if !d.Flashing {
		logger.Log("User cancelled flashing")
		return fmt.Errorf("cancelled")
	}
	
	_, err := os.Stat(zip_file)
	if os.IsNotExist(err) {
	  return fmt.Errorf("%s does not exist, can't flash it", zip_file)
	}

	if d.State != "recovery" {
		d.State_request = "recovery"
		<-d.State_reached	// Blocks until recovery is connected
	}

	if d.State == "recovery" {
		// Wipe caches (and format data if "clean")
		if wipe == "clean" {
			logger.Log("Clean-Wiping the device...")
			err = twrp.WipeClean()
			if err != nil {
				return err
			}
		} else {
			logger.Log("Dirty-Wiping the device...")
			err = twrp.WipeDirty()
			if err != nil {
				return err
			}
		}

		d.State_request = "sideload"
		<-d.State_reached	// Blocks until sideload is connected

		// Flash the zip
		logger.Log("Sideloading the rom zip...")
		err = twrp.Sideload(zip_file)
		if err != nil {
			return err
		}

		return nil
	} else {
		return fmt.Errorf("Cannot flash the rom: TWRP not connected")
	}
}

func (d *Device) FlashZip(zip_file string) error {
	if !d.Flashing {
		logger.Log("User cancelled flashing")
		return fmt.Errorf("cancelled")
	}
	
	_, err := os.Stat(zip_file)
	if os.IsNotExist(err) {
	  return fmt.Errorf("%s does not exist, can't flash it", zip_file)
	}

	if d.State != "recovery" {
		d.State_request = "recovery"
		<-d.State_reached	// Blocks until recovery is connected
	}

	if d.State == "recovery" {
		d.State_request = "sideload"
		<-d.State_reached	// Blocks until sideload is connected

		// Flash the zip
		err = twrp.Sideload(zip_file)
		if err != nil {
			return err
		}

		return nil
	} else {
		return fmt.Errorf("Cannot flash the zip: TWRP not connected")
	}
}

