package main

import(
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"net/url"

	"github.com/amo13/anarchy-droid/logger"
)

func helptab() fyne.CanvasObject {
	// Left side
	tryfirst := widget.NewLabel("If your device is not detected, restart your computer and try another USB cable, port or hub.\nIf you are on Windows, try to install drivers for your device.")
	tryfirst.Wrapping = fyne.TextWrapWord

	href1 := "https://developer.android.com/studio/run/oem-usb#Drivers"
	u1, err := url.Parse(href1)
	if err != nil {
		logger.LogError("unable to parse " + href1 + " as URL:", err)
	}
	link_to_official_drivers := widget.NewHyperlinkWithStyle("Official drivers", u1, fyne.TextAlignCenter, fyne.TextStyle{})

	href2 := "https://stuff.free-droid.com/universaladbdriver_v6.0.zip"
	u2, err := url.Parse(href2)
	if err != nil {
		logger.LogError("unable to parse " + href2 + " as URL:", err)
	}
	link_to_universal_drivers := widget.NewHyperlinkWithStyle("Universal drivers", u2, fyne.TextAlignCenter, fyne.TextStyle{})


	leftside := container.NewVBox(tryfirst, widget.NewLabel(""), container.NewCenter(container.NewHBox(link_to_official_drivers, link_to_universal_drivers)))
	leftcard := widget.NewCard("", "", leftside)

	rightside := container.NewVBox()
	rightcard := widget.NewCard("", "", rightside)

	grid := container.New(layout.NewGridLayout(2), leftcard, rightcard)
	return container.NewVBox(layout.NewSpacer(), grid, layout.NewSpacer())
}