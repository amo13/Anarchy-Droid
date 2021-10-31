package device

import (
	"time"
	"errors"
	"strings"

	"github.com/amo13/anarchy-droid/logger"
	"github.com/amo13/anarchy-droid/lookup"
	"github.com/amo13/anarchy-droid/helpers"
	"github.com/amo13/anarchy-droid/device/adb"
	"github.com/amo13/anarchy-droid/device/twrp"
	"github.com/amo13/anarchy-droid/device/fastboot"
)

func (d *Device) Observe() {
	go d.startObserver()
}

func (d *Device) startObserver() {
	stopped_observing := false
	new_state := "unknown"
	last_state := "unknown"
	for {
		time.Sleep(1 * time.Second)
		if d.ObserveMe {
			if stopped_observing {
				stopped_observing = false
				logger.Log("observing device connection again")
			}
			new_state = d.doObservation(last_state)
		} else {
			if !stopped_observing {
				stopped_observing = true
				logger.Log("stopped observing device connection")
			}
		}
		last_state = new_state
	}
}

func (d *Device) doObservation(last_state string) string {
	new_state := d.GetState()
	if new_state != last_state {
		last_state = new_state
		d.changeDetected(new_state)
	}

	// Clear request if requested state reached
	if d.State_request != "" {
		if d.State_request == "bootloader" {
			if d.State == "fastboot" || d.State == "heimdall" {
				logger.Log("Reached requested state", d.State_request)
				d.State_request = ""
				d.State_reached <- d.State
			}
		} else if d.State_request == d.State {
			logger.Log("Reached requested state", d.State_request)
			d.State_request = ""
			d.State_reached <- d.State
		}
	}

	if d.State_request != "" && d.State != "disconnected" {
		logger.Log("Device state requested:", d.State_request)
		d.HandleStateRequest(d.State_request)
	}

	return new_state
}

func (d *Device) changeDetected(new_state string) {
	logger.Log("Device connection update:", new_state)

	// Clear device info and start anew if a new device has been connected
	if !d.Flashing && !d.isSameDevice(new_state) && helpers.IsStringInSlice(new_state, []string{"android", "recovery", "fastboot"}) {
		logger.Log("New device detected, clearing and starting anew.")
		d.StartOver()
	}

	// Prepend new state to States_history and save current state
	d.States_history = append([]string{new_state}, d.States_history...)
	d.State = new_state

	// Read ADB props and fastboot vars if not done yet
	if helpers.IsStringInSlice(new_state, []string{"android", "recovery", "fastboot"}) {
		d.ReadMissingProps()
	}
}

func (d *Device) StartOver() {
	d.ObserveMe = false	// Stop observing old device object
	D1 = NewDevice()

	// Read ADB props and fastboot vars if not done yet
	if helpers.IsStringInSlice(D1.GetState(), []string{"android", "recovery", "fastboot"}) {
		d.ReadMissingProps()
		if d.Model != "" || d.Codename != "" {
			logger.Report(map[string]string{"progress":"Device connected: " + d.Model + " (" + d.Codename + ")"})
		}
	}
}

func (d *Device) ReadMissingProps() {
	d.Scanning = true 	// So the gui can wait for this to complete

	err := errors.New("")
	conn_state := d.State
	if conn_state == "android" || conn_state == "recovery" {
		if len(d.AdbProps) == 0 {
			d.AdbProps, err = adb.GetPropMap()
			if err != nil {
				logger.LogError("Unable to get ADB props map:", err)
			}
		}
	} else if conn_state == "fastboot" {
		if len(d.FastbootVars) == 0 {
			d.FastbootVars, err = fastboot.GetVarMap()
			if err != nil {
				logger.LogError("Unable to get fastboot vars map:", err)
			}
		}
	}
	if d.Model == "" {
		if len(d.AdbProps) > 0 {
			d.Model = adb.ModelFromPropMap(d.AdbProps)
		} else if len(d.FastbootVars) > 0 {
			d.Model = fastboot.ModelFromVarMap(d.FastbootVars)
		}

		// Propagate the codename to the logger package
		logger.Device_model = d.Model
	}
	if d.Codename == "" {
		if d.Model != "" {
			d.Codename, err = lookup.ModelToCodename(d.Model)
			if err != nil {
				if strings.Contains(err.Error(), "ambiguous") {
					d.Codename_ambiguous = true
				} else {
					logger.LogError("Unable to lookup model to codename:", err)
				}
			} else {
				d.Codename_ambiguous = false
			}
		} else if len(d.AdbProps) > 0 {
			d.Codename = adb.CodenameFromPropMap(d.AdbProps)
		}

		// Propagate the codename to the logger package
		logger.Device_codename = d.Codename
	}
	if d.Brand == "" {
		if d.Codename != "" {
			d.Brand, err = lookup.CodenameToBrand(d.Codename)
			if err != nil {
				switch err.Error() {
				case "not found":
				case "ambiguous":
				default:
					logger.LogError("Unable to lookup codename to brand:", err)
				}
			}
		}
		if d.Codename == "" || d.Brand == "" {
			if len(d.AdbProps) > 0 {
				d.Brand = adb.BrandFromPropMap(d.AdbProps)
			}
		}
	}
	if d.IsBrandUnlockable == false && d.Brand != "" {
		d.IsBrandUnlockable = helpers.IsStringInSlice(strings.ToLower(d.Brand), UnlockableBrands)
	}
	if d.Name == "" {
		if d.Codename != "" {
			d.Name, err = lookup.CodenameToNameCsv(d.Codename)
			if err != nil {
				switch err.Error() {
				case "not found":
				case "ambiguous":
				default:
					logger.LogError("Unable to lookup codename to name from CSV:", err)
				}
			}
		}
	}
	if d.Arch == "" {
		if len(d.AdbProps) > 0 {
			d.Arch, err = adb.CpuArchFromPropMap(d.AdbProps)
			if err != nil {
				logger.LogError("Unable to read cpu arch from adb props map:", err)
			}
		}
	}
	if d.Imei == "" {
		if d.State == "android" {
			d.Imei, err = adb.Imei()
			if err != nil {
				logger.Log(err.Error())
			}
		} else if d.State == "fastboot" && len(d.FastbootVars) > 0 {
			d.Imei = fastboot.ImeiFromVarMap(d.FastbootVars)
		}
	}
	if d.IsAB_checked == false {
		if len(d.AdbProps) > 0 {
			d.IsAB = adb.IsABFromPropMap(d.AdbProps)
			d.IsAB_checked = true
		} else if len(d.FastbootVars) > 0 {
			d.IsAB = fastboot.IsABFromVarMap(d.FastbootVars)
			d.IsAB_checked = true
		}
	}
	// Recheck if the device is unlocked until it is
	if !d.IsUnlocked {
		if len(d.FastbootVars) > 0 {
			d.IsUnlocked = fastboot.IsUnlockedFromVarMap(d.FastbootVars)
		}

		if !d.IsUnlocked && len(d.AdbProps) > 0 && adb.IsCustomRomFromMap(d.AdbProps) {
			d.IsUnlocked = true
		}

		if !d.IsUnlocked && d.State == "recovery" {
			d.IsUnlocked = true
		}
	}
	if !d.IsSupported_checked {
		if d.Codename != "" {
			d.IsSupported, err = lookup.IsSupported(d.Codename)
			if err != nil {
				logger.LogError("Unable to lookup if this codename is supported:", err)
			}
			d.IsSupported_checked = true
		}
	}
	if d.State == "recovery" && d.TwrpVersionConnected == "" {
		twrp_v, err := twrp.VersionConnected()
		if err != nil {
			logger.LogError("Unable to determine version of connected TWRP:", err)
		} else {
			d.TwrpVersionConnected = twrp_v
		}
	} else {
		d.TwrpVersionConnected = ""
	}

	d.Scanning = false 	// So the gui can wait for this to complete
}

// Checks whether the same or a new device is connected
// In doubt, assume it is the same device
func (d *Device) isSameDevice(state string) bool {
	if helpers.IsStringInSlice(state, []string{"android","recovery","booting"}) {
		model, err := adb.Model()
		if err != nil {
			logger.LogError("Unable to read model from adb:", err)
			return true
		}

		if model != d.Model {
			codename, err := lookup.ModelToCodename(model)
			if err != nil {
				if err.Error() != "ambiguous" {
					logger.LogError("Unable to lookup model to codename:", err)
				}

				return true
			}

			if codename != "" && codename != d.Codename {
				return false
			}
		}

		// If model was not recognized, try reading codename from ADB
		if model == "" {
			codename, err := adb.Codename()
			if err != nil {
				logger.LogError("Unable to read codename from adb:", err)
				return true
			}

			if d.Codename == codename {
				return true
			} else {
				return false
			}
		}
	} else if state == "fastboot" {
		model, err := fastboot.Model()
		if err != nil {
			logger.LogError("Unable to read model from fastboot:", err)
			return true
		}

		if model != d.Model {
			codename, err := lookup.ModelToCodename(model)
			if err.Error() != "ambiguous" {
				logger.LogError("Unable to lookup model to codename:", err)
			} else {
				return true
			}

			if codename != "" && codename != d.Codename {
				return false
			}
		}
	}

	return true
}