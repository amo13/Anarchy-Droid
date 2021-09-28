package main

import(
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"anarchy-droid/device"
	"anarchy-droid/logger"
)

var Lbl_flashing_title *widget.Label
var Lbl_flashing_instructions *widget.Label
var Lbl_boot_states *widget.Label
var Lbl_progressbar *widget.Label
var Progressbar *widget.ProgressBarInfinite

func btnCancelClicked() {
	logger.Log("User clicked Cancel")
	device.D1.Flashing = false
	Progressbar.Stop()
	Lbl_flashing_instructions.SetText("You cancelled.\n\nPlease restart the application.")
	logger.Report(map[string]string{"progress":"Cancelled"})
}

func flashingScreen() fyne.CanvasObject {
	Lbl_flashing_title := widget.NewLabelWithStyle("Installation", fyne.TextAlignCenter ,fyne.TextStyle{Bold: true})
	Lbl_flashing_instructions = widget.NewLabel("Please wait.")
	Lbl_flashing_instructions.Wrapping = fyne.TextWrapWord
	Lbl_flashing_instructions.Alignment = fyne.TextAlignCenter
	Lbl_boot_states = widget.NewLabel("")
	Lbl_progressbar = widget.NewLabel("Some info about the progress here...")
	Btn_cancel := widget.NewButton("Cancel", btnCancelClicked)
	Progressbar = widget.NewProgressBarInfinite()
	Progressbar.Stop()

	Center_flashing_box := container.NewVBox(Lbl_flashing_title, Lbl_flashing_instructions)
	progress_text_and_cancel := container.NewHBox(Lbl_progressbar, layout.NewSpacer(), Btn_cancel)
	box := container.NewVBox(Center_flashing_box, layout.NewSpacer(), Lbl_boot_states, progress_text_and_cancel, Progressbar)
	return box
}

func updateFlashingScreen() {
	// Display requested and current device states
	if device.D1.State_request != "" {
		Lbl_boot_states.SetText("Trying to get the device into " + device.D1.State_request + ". Current state is " + device.D1.State + ".")
	} else {
		Lbl_boot_states.SetText("")
	}
}