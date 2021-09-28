package main

import(
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"net/url"

	"anarchy-droid/logger"
)

func abouttab() fyne.CanvasObject {
	info1 := widget.NewLabel("\nThis application is only possible thanks to all the people behind the roms, TWRP, MicroG, F-Droid, Aurora, Magisk, Nanodroid and Heimdall. Additionally, we should also be thankful to all the individuals who contributed essential code here and there!")
	info1.Wrapping = fyne.TextWrapWord
	info1.Alignment = fyne.TextAlignCenter

	chk_reports := widget.NewCheck("Help improve this application by reporting device name and installation success", func(b bool) {logger.Consent = b})
	chk_reports.SetChecked(true)
	info2 := widget.NewLabel("The collected data is publicly available: ")
	info2.Alignment = fyne.TextAlignTrailing

	href := "https://stats.free-droid.com/index.php?module=Widgetize&action=iframe&secondaryDimension=eventAction&disableLink=0&widget=1&moduleToWidgetize=Events&actionToWidgetize=getCategory&idSite=3&period=range&date=2019-07-01,today&disableLink=1&widget=1"
	u, err := url.Parse(href)
	if err != nil {
		logger.LogError("unable to parse " + href + " as URL:", err)
	}
	reports_hyperlink := widget.NewHyperlink("click here to take a look.", u)

	info3 := widget.NewLabel("Author: Amaury Bodet  -  Version: " + AppVersion + "  -  Build: " + BuildDate + "  -  License: GPL-3.0")
	info3.Alignment = fyne.TextAlignCenter

	return container.NewVBox(info1,
		container.NewCenter(chk_reports),
		container.NewCenter(container.NewHBox(info2, reports_hyperlink)),
		info3)
}