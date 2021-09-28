package main

import(
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"net/url"
	"anarchy-droid/logger"
)

func helptab() fyne.CanvasObject {
	tryfirst := widget.NewLabel("If your device is not detected, restart your computer and try another USB cable, port or hub.\nIf you are on Windows, try to install the official drivers for your device.")
	tryfirst.Wrapping = fyne.TextWrapWord

	href := "https://developer.android.com/studio/run/oem-usb#Drivers"
	u, err := url.Parse(href)
	if err != nil {
		logger.LogError("unable to parse " + href + " as URL:", err)
	}
	link_to_drivers := widget.NewHyperlinkWithStyle("Download drivers", u, fyne.TextAlignCenter, fyne.TextStyle{})

	leftside := container.NewVBox(tryfirst, widget.NewLabel(""), link_to_drivers)
	leftcard := widget.NewCard("", "", leftside)

	rightside := container.NewVBox()
	rightcard := widget.NewCard("", "", rightside)

	grid := container.New(layout.NewGridLayout(2), leftcard, rightcard)
	return container.NewVBox(layout.NewSpacer(), grid, layout.NewSpacer())
}