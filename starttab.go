package main

import(
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/amo13/anarchy-droid/logger"
)

var ReadyToStart bool


// Left side

var Btn_start *widget.Button
var Chk_gotbackups *widget.Check
var Lbl_device_detection *widget.Label
var Lbl_brand_codename *widget.Label

func btnStartClicked() {
	go func() {
		err := prepareFlash()
		if err != nil {
			logger.LogError("prepareFlash() failed:", err)
		}
	}()
}

func chkGotbackupsChanged(value bool) {
	updateMainScreen()
}


// Right side

var Lbl_instructions *widget.Label
var initial_instructions = "Welcome to " + AppName + "!\n\nOn your connected Android device:\n1. In Settings > About Phone:\nTap 7 times on Build Number\n2. In Settings > Developer Options:\nActivate Android/USB Debugging\nand OEM Unlock (if you've got that)"


func initStarttabWidgets() {
	Btn_start = widget.NewButton("Start", btnStartClicked)
	Chk_gotbackups = widget.NewCheck("I've got backups of all I need", chkGotbackupsChanged)
	Lbl_device_detection = widget.NewLabel("")
	Lbl_brand_codename = widget.NewLabel("")
	Lbl_instructions = widget.NewLabel("")
	Lbl_instructions.Wrapping = fyne.TextWrapWord
}

func setDefaultsStarttab() {
	Lbl_device_detection.SetText("No device detected")
	Lbl_device_detection.Alignment = fyne.TextAlignCenter
	Lbl_brand_codename.SetText("")
	Lbl_brand_codename.Alignment = fyne.TextAlignCenter
	Lbl_instructions.SetText(initial_instructions)
	Btn_start.Disable()
}

func starttab() fyne.CanvasObject {
	// Left side
	empty := widget.NewLabel("")
	leftside := container.NewVBox(Btn_start, Chk_gotbackups, empty, Lbl_device_detection, Lbl_brand_codename)
	leftcard := widget.NewCard("", "", leftside)

	// Right side
	rightside := container.NewVBox(Lbl_instructions)
	rightcard := widget.NewCard("", "", rightside)

	grid := container.New(layout.NewGridLayout(2), leftcard, rightcard)
	return container.NewVBox(layout.NewSpacer(), grid, layout.NewSpacer())
}