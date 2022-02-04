package main

import(
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/amo13/anarchy-droid/logger"
	"github.com/amo13/anarchy-droid/device"
	"github.com/amo13/anarchy-droid/device/adb"
)

var Center_flashing_box *fyne.Container
var Unlock_sony_box *fyne.Container
var Lbl_unlock_info *widget.Label
var Lbl_unlock_data *widget.Label
var Btn_first_unlock_step *widget.Button

func sonyUnlockScreen() fyne.CanvasObject {
	Lbl_unlocking_title := widget.NewLabelWithStyle("Unlock the device", fyne.TextAlignCenter ,fyne.TextStyle{Bold: true})
	Lbl_unlock_info = widget.NewLabel("On the Sony website, select your device, enter your IMEI, check the two boxes below and submit to get your unlock code.\n")
	Lbl_unlock_info.Wrapping = fyne.TextWrapWord
	Lbl_unlock_info.Alignment = fyne.TextAlignCenter
	Lbl_unlock_data = widget.NewLabel("Your IMEI: " + device.D1.Imei)
	if device.D1.Imei == "" {
		err := adb.ShowImeiOnDeviceScreen()
		if err != nil {
			logger.LogError("Error trying to show IMEI on device screen:", err)
			// What now?
		}
		Lbl_unlock_data.SetText("Your IMEI should be showing on your device screen.")
	}
	btn_open_sony_website := widget.NewButton("Open Sony website", func() {OpenWebBrowser("https://developer.sony.com/develop/open-devices/get-started/unlock-bootloader/#unlock-code")})
	input := widget.NewEntry()
	input.PlaceHolder = "Enter the unlock code here"

	Unlock_sony_box = container.NewVBox(Lbl_unlocking_title, Lbl_unlock_info,
		container.NewGridWithColumns(2,
			Lbl_unlock_data,
			btn_open_sony_website,
			input,
			widget.NewButton("Unlock and continue", func() {
				if input.Text != "" {
					unlockStep(input.Text)
				}
			}),
		),
	)

	return Unlock_sony_box
}

func motorolaUnlockScreen() fyne.CanvasObject {
	Lbl_unlocking_title := widget.NewLabelWithStyle("Unlock the device", fyne.TextAlignCenter ,fyne.TextStyle{Bold: true})
	Lbl_unlock_info = widget.NewLabel("Click the first button to reboot your device into the bootloader and read the needed unlock data. If the unlock data can be read out successfully, a website will open guiding you through the process of obtaining the unlock code for your device. Once you have the unlock code, enter it in the box below and click the second button to continue.")
	Lbl_unlock_info.Wrapping = fyne.TextWrapWord

	Btn_first_unlock_step = widget.NewButton("Open unlock guide", func() { go func() {
		Btn_first_unlock_step.SetText("Please wait...")
		Btn_first_unlock_step.Disable()
		defer Btn_first_unlock_step.Enable()
		defer Btn_first_unlock_step.SetText("Open unlock guide")
		unlock_data, err := device.D1.GetUnlockData()
		if err != nil && err.Error() != "unlocked" {
			logger.LogError("Error during retrieval of unlock data:", err)
			// What now?
		} else if err.Error() != "unlocked" {
			logger.Log("Unlock data is:", unlock_data)
			OpenWebBrowser("https://help.free-droid.com/unlock-motorola/?code=" + unlock_data)
		} else {
			logger.Log("Bootloader is already unlocked")
			w.SetContent(flashingScreen())
			Lbl_flashing_instructions.SetText("Your bootloader is already unlocked.")
			bootTwrpStep()
		}
		}()
	})

	input := widget.NewEntry()
	input.PlaceHolder = "Enter the unlock code here"

	Unlock_motorola_box := container.NewVBox(Lbl_unlocking_title, Lbl_unlock_info,
		container.NewCenter(Btn_first_unlock_step),
		container.NewGridWithColumns(2, input,
		widget.NewButton("Unlock and continue", func() {
			if input.Text != "" {
				unlockStep(input.Text)
			}
		}),
	))

	return Unlock_motorola_box
}

func fairphoneUnlockScreen() fyne.CanvasObject {
	Lbl_unlocking_title := widget.NewLabelWithStyle("Unlock the device", fyne.TextAlignCenter ,fyne.TextStyle{Bold: true})
	Lbl_unlock_info = widget.NewLabel("Follow the instructions:")
	Lbl_unlock_info.Wrapping = fyne.TextWrapWord
	Lbl_unlock_info.Alignment = fyne.TextAlignCenter
	Lbl_unlock_data = widget.NewLabel("")
	if device.D1.Imei != "" && device.D1.SerialNumber != "" {
		Lbl_unlock_data.SetText("Your IMEI: " + device.D1.Imei + " - Your Serial Number: " + device.D1.SerialNumber + "\n")
	}
	btn_open_fairphone_website := widget.NewButton("Open Fairphone website", func() {OpenWebBrowser("https://www.fairphone.com/en/bootloader-unlocking-code-for-fairphone-3/")})

	Unlock_fairphone_box := container.NewVBox(Lbl_unlocking_title, Lbl_unlock_data,
		container.NewGridWithColumns(2,
			Lbl_unlock_info,
			btn_open_fairphone_website,
			widget.NewLabelWithStyle("Once you are done:", fyne.TextAlignCenter ,fyne.TextStyle{}),
			widget.NewButton("Continue", func() {
				unlockStep("")
			}),
		),
	)

	return Unlock_fairphone_box
}