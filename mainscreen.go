package main

import(
	"time"
	"net/url"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/container"

	"github.com/amo13/anarchy-droid/get"
	"github.com/amo13/anarchy-droid/device"
	"github.com/amo13/anarchy-droid/logger"
)

var last_codename string	// Used in IsNewDevice() to help call ReloadRoms() when a new device (codename) is connected

func mainScreen() fyne.CanvasObject {
	initAllWidgets()
	setDefaults()
	launchGuiUpdateLoop()

	tabs := container.NewAppTabs(
		container.NewTabItem("        Start        ", starttab()),
		container.NewTabItem("       Settings      ", settingstab()),
		container.NewTabItem("       Advanced      ", advancedtab()),
		container.NewTabItem("         Help        ", helptab()),
		container.NewTabItem("         About        ", abouttab()),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	tabContainer := container.NewHBox(layout.NewSpacer(), tabs, layout.NewSpacer())
	return tabContainer
}

func initAllWidgets() {
	initStarttabWidgets()
	initSettingstabWidgets()
	initAdvancedtabWidgets()
}

func setDefaults() {
	setDefaultsStarttab()
	setDefaultsSettingstab()
	setDefaultsAdvancedtab()
}

// Helper function to open a web browser at given url
func OpenWebBrowser(href string) {
	u, err := url.Parse(href)
	if err != nil {
		logger.LogError("unable to parse " + href + " as URL.\n", err)
		return
	}
	err = fyne.CurrentApp().OpenURL(u)
	if err != nil {
		logger.LogError("Error opening " + u.String() + " in web browser:", err)
	}
}

func launchGuiUpdateLoop() {
	go func () {
		for {
			time.Sleep(250 * time.Millisecond)
			if active_screen == "mainScreen" {
				updateMainScreen()
			} else if active_screen == "flashingScreen" {
				updateFlashingScreen()
			}
		}
	}()
}

func IsNewDevice() bool {
	if last_codename != device.D1.Codename {
		last_codename = device.D1.Codename
		return true
	} else {
		return false
	}
}

// Periodically called
func updateMainScreen() {
	// Wait if device is currently scanning
	if device.D1.Scanning {
		Lbl_instructions.SetText("Scanning the device...")
		return
	}

	// Checkbox "Assume bootloader unlocked"
	if device.D1.IsUnlocked {
		Chk_skipunlock.SetChecked(true)
		Chk_skipunlock.Disable()
	} else {
		Chk_skipunlock.Enable()
	}

	if device.D1.IsAB_checked && !device.D1.IsAB {
		Chk_copypartitions.SetChecked(false)
		Chk_copypartitions.Disable()
	} else {
		Chk_copypartitions.Enable()
	}

	// Wait if reloading roms
	if get.A1.Reloading {
		Lbl_instructions.SetText("Loading available roms...")
		return
	}

	if device.D1.State != "disconnected" {
		if device.D1.Model != "" {
			Lbl_device_detection.SetText(device.D1.Model + " connected!")
		} else {
			Lbl_device_detection.SetText("Device connected!")
		}

		brand_codename_string := ""
		if device.D1.Brand != "" {
			brand_codename_string = "Brand: " + device.D1.Brand
		}
		if device.D1.Codename != "" {
			if device.D1.Brand != "" {
				brand_codename_string = brand_codename_string + " - "
			}
			brand_codename_string = brand_codename_string + "Codename: " + device.D1.Codename
		}
		Lbl_brand_codename.SetText(brand_codename_string)

		if IsNewDevice() {
			go func() {
				ReloadRoms()

				// Tick Chk_skipflashtwrp if the correct version of TWRP is alrady connected
				if device.D1.State == "recovery" {
					if device.D1.TwrpVersionConnected == strings.Split(get.A1.User.Twrp.Img.Version, "_")[0] {
						Chk_skipflashtwrp.SetChecked(true)
					}
				}
			}()

			return
		}
	} else {
		Lbl_device_detection.SetText("No device connected")
	}

	switch device.D1.State {
	case "unauthorized":
		Lbl_instructions.SetText("Device unauthorized!\n\nPlease ALLOW and hit OK on your device screen.")
	case "disconnected":
		Lbl_instructions.SetText(initial_instructions)
	case "booting":
		Lbl_instructions.SetText("Device booting...")
	case "sideload":
		Lbl_instructions.SetText("Device in sideload mode.\n\nPlease wait for it to finish.")
	case "heimdall", "fastboot":
		Lbl_instructions.SetText("Please reboot your device to Android.")
	case "recovery", "android", "simulation":
		deviceRecognized()
	default:
		Lbl_instructions.SetText("Unknown device connection state.")
	}

	Chk_skipunlock.Refresh()
	Lbl_device_detection.Refresh()
	Lbl_instructions.Refresh()
}

// Update the start button and the instructions on "Start" tab
// Called periodically if a device is connected via ADB
func deviceRecognized() {
	if device.D1.Codename == "" || device.D1.Model == "" || device.D1.Brand == "" {
		Lbl_instructions.SetText("Trying to recognize the device...")
		return
	}

	if device.D1.IsSupported {
		if !device.D1.IsUnlocked && (device.D1.IsBrandUnlockable || Chk_skipunlock.Checked) {	// unlock needed and feasible
			if get.A1.User.Twrp.Img.Href != "" {	// got TWRP image
				if get.A1.User.Rom.Href != "" {	// got TWRP image and rom
					// If OpenGapps is selected, make sure a version is also selected
					if Select_gapps.Selected != "OpenGapps" ||
					(Select_opengapps_version.Selected != "" &&
					Select_opengapps_version.Selected != Select_opengapps_version.PlaceHolder &&
					Select_opengapps_variant.Selected != "" &&
					Select_opengapps_variant.Selected != Select_opengapps_variant.PlaceHolder) {
						if Chk_gotbackups.Checked {
							Btn_start.Enable()
						} else {
							Btn_start.Disable()
						}
						Lbl_instructions.SetText("Ready!")
					} else {	// otherwise prompt user to select the correct OpenGapps version
						Btn_start.Disable()
						Lbl_instructions.SetText("Please set the OpenGapps version to the Android version of the rom you wish to install.")
					}
				} else {	// got TWRP image but missing rom
					Btn_start.Disable()
					Lbl_instructions.SetText("Missing rom.\nIf you've got one, you can select it in the settings tab.")
				}
			} else {		// missing TWRP image
				if get.A1.User.Rom.Href != "" {	// missing TWRP image but got rom
					Btn_start.Disable()
					Lbl_instructions.SetText("Missing TWRP image.\nIf you've got one, you can select it in the advanced tab. If it is already installed, reboot to TWRP.")
				} else {	// missing TWRP image and rom
					Btn_start.Disable()
					Lbl_instructions.SetText("Missing rom and TWRP image.\nIf you've got those, you can select them in the settings and advanced tabs.")
				}
			}
		} else if !device.D1.IsUnlocked && !device.D1.IsBrandUnlockable {	// unlock needed but not feasible
			Btn_start.Disable()
			Lbl_instructions.SetText("Unfortunately, " + AppName + " does not support your device.\n\nYou can still try to install TWRP on your device by yourself and connect it again.")
		} else if device.D1.IsUnlocked {	// already unlocked
			if get.A1.User.Twrp.Img.Href != "" {	// got TWRP image
				if get.A1.User.Rom.Href != "" {	// got TWRP image and rom
					// if Chk_gotbackups.Checked {
					// 	Btn_start.Enable()
					// } else {
					// 	Btn_start.Disable()
					// }
					// Lbl_instructions.SetText("Ready!")
					// If OpenGapps is selected, make sure a version is also selected
					if Select_gapps.Selected != "OpenGapps" ||
					(Select_opengapps_version.Selected != "" &&
					Select_opengapps_version.Selected != Select_opengapps_version.PlaceHolder &&
					Select_opengapps_variant.Selected != "" &&
					Select_opengapps_variant.Selected != Select_opengapps_variant.PlaceHolder) {
						if Chk_gotbackups.Checked {
							Btn_start.Enable()
						} else {
							Btn_start.Disable()
						}
						Lbl_instructions.SetText("Ready!")
					} else {	// otherwise prompt user to select the correct OpenGapps version
						Btn_start.Disable()
						Lbl_instructions.SetText("Please set the OpenGapps version to the Android version of the rom you wish to install.")
					}
				} else {	// got TWRP image but missing rom
					Btn_start.Disable()
					Lbl_instructions.SetText("Missing rom.\nIf you've got one, you can select it in the settings tab.")
				}
			} else {		// missing TWRP image
				if get.A1.User.Rom.Href != "" {	// missing TWRP image but got rom
					Btn_start.Disable()
					Lbl_instructions.SetText("Missing TWRP image.\nIf you've got one, you can select it in the advanced tab.")
					// but prompt user to reboot to twrp if twrp is installed
				} else {	// missing TWRP image and rom
					Btn_start.Disable()
					Lbl_instructions.SetText("Missing rom and TWRP image.\nIf you've got those, you can select them in the settings and advanced tabs.")
				}
			}
		}
	} else {
		Lbl_instructions.SetText("Unfortunately, " + AppName + " does not support your device.")
	}
}