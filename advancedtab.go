package main

import(
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/theme"

	"github.com/amo13/anarchy-droid/helpers"
	"github.com/amo13/anarchy-droid/logger"
	"github.com/amo13/anarchy-droid/get"
)

// Left side

var Chk_magisk *widget.Check
var Chk_sigspoof *widget.Check
var Chk_swype *widget.Check
var Chk_gsync *widget.Check
var Chk_copypartitions *widget.Check

func openWebBrowserMagisk() {
	OpenWebBrowser("https://github.com/topjohnwu/Magisk")
}

func openWebBrowserSigspoof() {
	OpenWebBrowser("https://gitlab.com/Nanolx/NanoDroid/-/blob/master/doc/microGsetup.md")
}

func openWebBrowserCopyPartitions() {
	OpenWebBrowser("https://wiki.lineageos.org/devices/pioneer/install#pre-install-instructions")
}


// Right side

var Chk_skipunlock *widget.Check
var Chk_skipwipedata *widget.Check
var Chk_skipflashtwrp *widget.Check
var Chk_user_twrp *widget.Check
var Lbl_user_twrp *widget.Label

func chkSkipUnlockChanged(checked bool) {
	if checked {
		Chk_skipwipedata.Enable()
	} else {
		Chk_skipwipedata.Checked = false
		Chk_skipwipedata.Disable()
		Chk_skipwipedata.Refresh()
	}
}

func chkSkipFlashTwrpChanged(checked bool) {

}

func chkUserTwrpChanged(checked bool) {
	if checked {
		Dialog_user_twrp := dialog.NewFileOpen(userTwrpSelected, w)
		Dialog_user_twrp.SetFilter(storage.NewExtensionFileFilter([]string{".img"}))
		Dialog_user_twrp.Show()
	} else {
		// Use Upstream/Archive TWRP
		selectDefaultTwrp()
	}
}

func selectDefaultTwrp() {
	if get.A1.Archive.Override_twrp.Img.Href != "" {
		get.A1.User.Twrp = get.A1.Archive.Override_twrp
	} else if get.A1.Upstream.Twrp.Img.Href != "" {
		get.A1.User.Twrp = get.A1.Upstream.Twrp
	} else if get.A1.Archive.Twrp.Img.Href != "" {
		get.A1.User.Twrp = get.A1.Archive.Twrp
	} else {
		Lbl_user_twrp.SetText("No TWRP available")
		return
	}

	Lbl_user_twrp.SetText(get.A1.User.Twrp.Img.Filename)
}

func userTwrpSelected(urc fyne.URIReadCloser, err error) {
	if err != nil {
		logger.LogError("error on user TWRP file selection: " + urc.URI().Scheme() + urc.URI().Name() + urc.URI().Extension() + ":", err)
		selectDefaultTwrp()
		return
	}
	if urc == nil || urc.URI() == nil {
		// User hits cancel
		selectDefaultTwrp()
		return
	}

	get.A1.User.Twrp.Img = &get.Item{}		// Reset the twrp item
	Lbl_user_twrp.SetText(helpers.ExtractFileNameFromHref(urc.URI().String()))
	get.A1.User.Twrp.Img.Href = urc.URI().String()
	get.A1.User.Twrp.Img.Filename = helpers.ExtractFileNameFromHref(urc.URI().String())
}


func initAdvancedtabWidgets() {
	// Left side
	Chk_magisk = widget.NewCheck("Install Magisk (coming soon)", func(bool) {})
	Chk_magisk.Disable()
	Chk_sigspoof = widget.NewCheck("Signature Spoofing Patch", func(bool) {})
	Chk_gsync = widget.NewCheck("Install Google Sync Adapters", func(bool) {})
	Chk_swype = widget.NewCheck("Install Google Swype Libraries", func(bool) {})
	Chk_copypartitions = widget.NewCheck("Flash copy-partitions.zip", func(bool) {})

	// Right side
	Chk_skipunlock = widget.NewCheck("Assume bootloader already unlocked", chkSkipUnlockChanged)
	Chk_skipwipedata = widget.NewCheck("Do not wipe the data partition", func(bool) {})
	Chk_skipwipedata.Disable()
	Chk_skipflashtwrp = widget.NewCheck("Assume TWRP already installed", chkSkipFlashTwrpChanged)
	Chk_user_twrp = widget.NewCheck("Provide your own TWRP image", chkUserTwrpChanged)
	Lbl_user_twrp = widget.NewLabel("")
	Lbl_user_twrp.Wrapping = fyne.TextTruncate
	Lbl_user_twrp.Alignment = fyne.TextAlignCenter
}

func setDefaultsAdvancedtab() {
	Chk_sigspoof.SetChecked(true)	
	Chk_copypartitions.SetChecked(true)
}

func advancedtab() fyne.CanvasObject {
	// Left side
	magisk_info_icon := newTappableIcon(theme.InfoIcon())
	magisk_info_icon.OnTapped = openWebBrowserMagisk
	box_magisk := container.NewBorder(nil, nil, nil, magisk_info_icon, Chk_magisk)

	sigspoof_info_icon := newTappableIcon(theme.InfoIcon())
	sigspoof_info_icon.OnTapped = openWebBrowserSigspoof
	box_sigspoof := container.NewBorder(nil, nil, nil, sigspoof_info_icon, Chk_sigspoof)

	copypartitions_info_icon := newTappableIcon(theme.InfoIcon())
	copypartitions_info_icon.OnTapped = openWebBrowserCopyPartitions
	box_copypartitions := container.NewBorder(nil, nil, nil, copypartitions_info_icon, Chk_copypartitions)

	leftside := container.NewVBox(box_magisk, box_sigspoof, Chk_gsync, Chk_swype, box_copypartitions)
	leftcard := widget.NewCard("", "", leftside)

	// Right side
	rightside := container.NewVBox(Chk_skipunlock, Chk_skipwipedata, Chk_skipflashtwrp, Chk_user_twrp, Lbl_user_twrp)
	rightcard := widget.NewCard("", "", rightside)

	grid := container.New(layout.NewGridLayout(2), leftcard, rightcard)
	return container.NewVBox(layout.NewSpacer(), grid, layout.NewSpacer())
}