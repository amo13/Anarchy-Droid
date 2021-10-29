package main

import(
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"

	"github.com/amo13/anarchy-droid/helpers"
	"github.com/amo13/anarchy-droid/device"
	"github.com/amo13/anarchy-droid/logger"
	"github.com/amo13/anarchy-droid/get"

	"strings"
	"fmt"
)

// Left side
var Lbl_android_version *widget.Label
var Select_rom *widget.Select
var Radio_rom_source *widget.RadioGroup
var Chk_user_rom *widget.Check
var Lbl_user_rom *widget.Label

func changedRom(rom string) {
	if rom == "" || rom == Select_rom.PlaceHolder {
		return
	}

	registerRomChoice()
}

func registerRomChoice() {
	if Chk_user_rom.Checked {

	} else {
		if Radio_rom_source.Selected == "Official Releases" {
			if Select_gapps.Selected == "MicroG" && Select_rom.Selected == "LineageOS" && get.A1.Upstream.Rom["LineageOSMicroG"] != nil {
				get.A1.User.Rom = get.A1.Upstream.Rom["LineageOSMicroG"]
			} else if get.A1.Upstream.Rom[Select_rom.Selected] != nil {
				get.A1.User.Rom = get.A1.Upstream.Rom[Select_rom.Selected]
			} else {
				logger.Log("Cannot register rom choice: " + Select_rom.Selected + " not in get.A1.Upstream.Rom")
			}
		} else if Radio_rom_source.Selected == "Archive" {
			if get.A1.Archive.Rom[Select_rom.Selected] != nil {
				get.A1.User.Rom = get.A1.Archive.Rom[Select_rom.Selected]
			} else {
				logger.Log("Cannot register rom choice: " + Select_rom.Selected + " not in get.A1.Archive.Rom")
			}
		} else {
			logger.Log("invalid rom source selection: Radio_rom_source.Selected is " + Radio_rom_source.Selected)
		}
	}

	displayFileNameAndAndroidVersion()
	selectOpenGappsVersion()
	setGappsSelectability()	
}

// Prevent changing Gapps selection for roms providing Gapps or MicroG
// Except for LineageOSMicroG if LineageOS is also available
func setGappsSelectability() {
	if get.A1.User.Rom.Name == "CalyxOS" || get.A1.User.Rom.Name == "e-OS" {
		Select_gapps.SetSelected("MicroG")
		Select_gapps.Disable()
	} else if get.A1.User.Rom.Name == "LineageOSMicroG" {
		if Chk_user_rom.Checked {
			Select_gapps.SetSelected("MicroG")
			Select_gapps.Disable()
		} else {
			if Radio_rom_source.Selected == "Official Releases" {
				if get.A1.Upstream.Rom["LineageOS"] != nil {
					Select_gapps.Enable()
				} else {
					Select_gapps.SetSelected("MicroG")
					Select_gapps.Disable()
				}
			} else if Radio_rom_source.Selected == "Archive" {
				if helpers.IsStringInSlice("LineageOS", get.A1.Archive.Romlist) {
					Select_gapps.Enable()
				} else {
					Select_gapps.SetSelected("MicroG")
					Select_gapps.Disable()
				}
			} else {
				Select_gapps.Enable()
			}
		}
	} else if get.A1.User.Rom.Name == "PixelExperience" {
		Select_gapps.SetSelected("OpenGapps")
		Select_gapps.Disable()
	} else {
		Select_gapps.Enable()
	}
}

func displayFileNameAndAndroidVersion() {
	if get.A1.User.Rom.Filename != "" {		// Display file name and android version of selected rom
		Lbl_user_rom.SetText(get.A1.User.Rom.Filename)
		if get.A1.User.Rom.Android_version != "" {
			Lbl_android_version.SetText("Android " + get.A1.User.Rom.Android_version)			
		} else {
			Lbl_android_version.SetText("")
			Select_opengapps_version.ClearSelected()
		}
	} else {
		Lbl_user_rom.SetText("No rom available")
	}
}

func changedRomSource(source string) {
	if source == "" {
		return
	}
	Select_rom.Options = compileSelectableRomList()
	if len(Select_rom.Options) > 0 {
		Select_rom.Enable()
		selectDefaultRomForGivenSource()
	} else {
		Lbl_user_rom.SetText("No rom available")
	}
}

func compileSelectableRomList() []string {
	if Radio_rom_source.Selected == "Official Releases" {
		return helpers.RemoveStringFromSlice(get.A1.Upstream.Romlist, "LineageOSMicroG")
	} else if Radio_rom_source.Selected == "Archive" {
		return helpers.RemoveStringFromSlice(get.A1.Archive.Romlist, "LineageOSMicroG")
	} else {
		logger.Log("unable to compile the list of selectable roms")
		logger.Log("Radio_rom_source.Selected is neither \"Official Releases\" nor \"Archive\"")
		return []string{}
	}
}



func chkUserRomChanged(checked bool) {
	if checked {
		Select_rom.Disable()
		Radio_rom_source.Disable()
		Lbl_android_version.SetText("")
		Dialog_user_rom := dialog.NewFileOpen(userRomSelected, w)
		Dialog_user_rom.SetFilter(storage.NewExtensionFileFilter([]string{".zip"}))
		Dialog_user_rom.Show()
	} else {
		Select_rom.Enable()
		Radio_rom_source.Enable()
		Lbl_user_rom.SetText("")
		registerRomChoice()
	}
}

func userRomSelected(urc fyne.URIReadCloser, err error) {
	if err != nil {
		logger.LogError("error on user rom file selection " + urc.URI().Scheme() + urc.URI().Name() + urc.URI().Extension() + ":", err)
		return
	}
	if urc == nil || urc.URI() == nil {
		// User hits cancel
		Chk_user_rom.SetChecked(false)
		return
	}

	get.A1.User.Rom = &get.Item{}	// Reset the rom item to remove previous item content
	Lbl_user_rom.SetText(helpers.ExtractFileNameFromHref(urc.URI().String()))
	get.A1.User.Rom.Href = urc.URI().String()
	get.A1.User.Rom.Filename = helpers.ExtractFileNameFromHref(urc.URI().String())
	romname, androidversion, err := get.GuessRomNameAndAndroidVersion(get.A1.User.Rom.Filename)
	if err != nil {
		logger.LogError("Unable to guess rom name and android version of " + get.A1.User.Rom.Filename + ":", err)
		// return here or move on?
	} else {
		get.A1.User.Rom.Name = romname
		get.A1.User.Rom.Android_version = androidversion
	}

	registerRomChoice()
}

func openWebBrowserRom() {
	switch strings.ToLower(Select_rom.Selected) {
	case "lineageos":
		if get.A1.User.Rom.Name == "LineageOSMicroG" {
			OpenWebBrowser("https://lineage.microg.org/")
		} else {
			OpenWebBrowser("https://lineageos.org/")
		}
	case "resurrectionremix":
		OpenWebBrowser("https://resurrectionremix.com/")
	case "aospextended", "aex":
		OpenWebBrowser("https://aospextended.com/")
	case "carbonrom":
		OpenWebBrowser("https://carbonrom.org/")
	case "omnirom":
		OpenWebBrowser("https://omnirom.org/")
	case "crdroid":
		OpenWebBrowser("https://crdroid.net/")
	case "e", "e-os", "eos", "/e/", "/e/os", "e os":
		OpenWebBrowser("https://e.foundation/e-os/")
	case "evolutionx":
		OpenWebBrowser("https://evolution-x.org/")
	case "calyxos":
		OpenWebBrowser("https://calyxos.org/")
	case "grapheneos":
		OpenWebBrowser("https://grapheneos.org/")
	case "pixelexperience":
		OpenWebBrowser("https://download.pixelexperience.org/")
	case "mokee":
		OpenWebBrowser("https://www.mokeedev.com/en/")
	default:
	}
}

func ReloadRoms() {
	if device.D1.Codename == "" {
		return
	}

	logger.Log("Reloading available roms...")
	Lbl_user_rom.SetText("Reloading available roms...")

	Select_rom.Disable()

	get.A1 = get.NewAvailable()
	err := get.A1.Populate(device.D1.Codename)
	if err != nil {
		logger.LogError("unable to populate the list of available roms:", err)
	}

	logger.Log("Found Upstream: " + strings.Join(get.A1.Upstream.Romlist[:], " "))
	logger.Log("Found Archived: " + strings.Join(get.A1.Archive.Romlist[:], " "))

	// Reset rom and twrp selections
	Lbl_user_rom.SetText("")
	Chk_user_rom.SetChecked(false)
	Lbl_user_twrp.SetText("")
	Chk_user_twrp.SetChecked(false)
	selectDefaultSource()
	selectDefaultTwrp()
	registerRomChoice()
}

func selectOpenGappsVersion() {
	if device.D1.Arch != "" {
		// Populate selectable options for device cpu architecture
		versionkeys := make([]string, 0, len(get.A1.Upstream.OpenGapps[device.D1.Arch]))
	    for key := range get.A1.Upstream.OpenGapps[device.D1.Arch] {
	        versionkeys = append(versionkeys, key)
	    }
		Select_opengapps_version.Options = versionkeys

		// Select proper OpenGapps version if available
		if get.A1.User.Rom.Android_version != "" {
			version, err := formatToOpenGappsAndroidVersion(get.A1.User.Rom.Android_version)
			if err != nil {
				logger.LogError("Unable to format " + get.A1.User.Rom.Android_version + " to OpenGapps android version format:", err)
				Select_opengapps_version.ClearSelected()
			} else {
				if helpers.IsStringInSlice(version, Select_opengapps_version.Options) {
					Select_opengapps_version.SetSelected(version)
				} else {
					Select_opengapps_version.ClearSelected()
				}
			}
		} else {
			Select_opengapps_version.ClearSelected()
		}
	}
}

func formatToOpenGappsAndroidVersion(v string) (string, error) {
	if strings.HasPrefix(v, "4.4") {
		return "4.4", nil
	} else if strings.HasPrefix(v, "5.0") {
		return "5.0", nil
	} else if strings.HasPrefix(v, "5") {
		return "5.1", nil
	} else if strings.HasPrefix(v, "6") {
		return "6.0", nil
	} else if strings.HasPrefix(v, "7.0") {
		return "7.0", nil
	} else if strings.HasPrefix(v, "7") {
		return "7.1", nil
	} else if strings.HasPrefix(v, "8.0") {
		return "8.0", nil
	} else if strings.HasPrefix(v, "8") {
		return "8.1", nil
	} else if strings.HasPrefix(v, "9") {
		return "9.0", nil
	} else if strings.HasPrefix(v, "10") {
		return "10.0", nil
	} else if strings.HasPrefix(v, "11") {
		return "11.0", nil
	} else {
		return "", fmt.Errorf("Unknown android version: " + v)
	}
}

func openWebBrowserRomSource() {
	if Radio_rom_source.Selected == "Archive" {
		OpenWebBrowser("https://archive.free-droid.com/")
	} else {
		openWebBrowserRom()
	}
}

// Right side
var Chk_fdroid *widget.Check
var Chk_aurora *widget.Check
var Chk_playstore *widget.Check
var Select_gapps *widget.Select
var Select_opengapps_variant *widget.Select
var Select_opengapps_version *widget.Select

func selectGappsChanged(value string) {
	switch value {
	case "OpenGapps":
		Chk_playstore.SetChecked(true)
		Chk_playstore.Disable()
		Select_opengapps_variant.Enable()
		Select_opengapps_version.Enable()
		Chk_sigspoof.SetChecked(false)
		Chk_swype.SetChecked(true)
		Chk_swype.Disable()
		Chk_gsync.SetChecked(true)
		Chk_gsync.Disable()

		if !Chk_user_rom.Checked {
			// Prefer LineageOS over LineageOSMicroG if OpenGapps chosen and both available
			if get.A1.User.Rom.Name == "LineageOSMicroG" {
				if Radio_rom_source.Selected == "Official Releases" {
					if helpers.IsStringInSlice("LineageOS", get.A1.Upstream.Romlist) {
						get.A1.User.Rom = get.A1.Upstream.Rom["LineageOS"]
					}
				} else if Radio_rom_source.Selected == "Archive" {
					if helpers.IsStringInSlice("LineageOS", get.A1.Archive.Romlist) {
						get.A1.User.Rom = get.A1.Archive.Rom["LineageOS"]
					}
				}
			}
		}
	case "MicroG":
		Chk_playstore.SetChecked(false)
		Chk_playstore.Enable()
		Select_opengapps_variant.Disable()
		Select_opengapps_version.Disable()
		Chk_sigspoof.SetChecked(true)
		Chk_swype.SetChecked(false)
		Chk_swype.Enable()
		Chk_gsync.SetChecked(false)
		Chk_gsync.Enable()

		if !Chk_user_rom.Checked {
			// Prefer LineageOSMicroG over LineageOS if MicroG chosen and both available
			if get.A1.User.Rom.Name == "LineageOS" {
				if Radio_rom_source.Selected == "Official Releases" {
					if get.A1.Upstream.Rom["LineageOSMicroG"] != nil {
						get.A1.User.Rom = get.A1.Upstream.Rom["LineageOSMicroG"]
					}
				} else if Radio_rom_source.Selected == "Archive" {
					if get.A1.Archive.Rom["LineageOSMicroG"] != nil {
						get.A1.User.Rom = get.A1.Archive.Rom["LineageOSMicroG"]
					}
				}
			}
		}
	case "Nothing":
		Chk_playstore.SetChecked(false)
		Chk_playstore.Disable()
		Select_opengapps_variant.Disable()
		Select_opengapps_version.Disable()
		Chk_sigspoof.SetChecked(false)
		Chk_swype.SetChecked(false)
		Chk_swype.Enable()
		Chk_gsync.SetChecked(false)
		Chk_gsync.Enable()

		if !Chk_user_rom.Checked {
			// Prefer LineageOS over LineageOSMicroG if Nothing chosen and both available
			if get.A1.User.Rom.Name == "LineageOSMicroG" {
				if Radio_rom_source.Selected == "Official Releases" {
					if helpers.IsStringInSlice("LineageOS", get.A1.Upstream.Romlist) {
						get.A1.User.Rom = get.A1.Upstream.Rom["LineageOS"]
					}
				} else if Radio_rom_source.Selected == "Archive" {
					if helpers.IsStringInSlice("LineageOS", get.A1.Archive.Romlist) {
						get.A1.User.Rom = get.A1.Archive.Rom["LineageOS"]
					}
				}
			}
		}
	case "":	// This case is for .ClearSelected()
	default:
		logger.Log("unexpected value in Gapps selection widget:", value)
	}
	Lbl_user_rom.SetText(get.A1.User.Rom.Filename)
}

func selectOpengappsVariantChanged(value string) {

}

func selectOpengappsVersionChanged(value string) {
	// Update available variants for selected version
	if value != "" && value != Select_opengapps_version.PlaceHolder {
		Select_opengapps_variant.Options = get.A1.Upstream.OpenGapps[device.D1.Arch][value]
	}
}

func openWebBrowserGapps() {
	switch Select_gapps.Selected {
	case "OpenGapps":
		OpenWebBrowser("https://opengapps.org/")
	case "MicroG":
		OpenWebBrowser("https://microg.org/")
	default:
	}
}

func openWebBrowserAurora() {
	OpenWebBrowser("https://gitlab.com/AuroraOSS/AuroraStore")
}

func openWebBrowserOpengappsVariants() {
	OpenWebBrowser("https://github.com/opengapps/opengapps/wiki/Package-Comparison")
}

func openWebBrowserFdroid() {
	OpenWebBrowser("https://f-droid.org")
}


// Prefer upstream official release over roms from archive
// Disable Radio group if the Archive has no rom
func selectDefaultSource() {
	Radio_rom_source.SetSelected("")

	// Prefer upstream official releases over archive
	if len(get.A1.Upstream.Romlist) > 0 && Radio_rom_source.Selected != "Official Releases" {
		Radio_rom_source.SetSelected("Official Releases")
	} else {
		if len(get.A1.Archive.Romlist) > 0 && Radio_rom_source.Selected != "Archive" {
			Radio_rom_source.SetSelected("Archive")
		} else {
			// If no rom is found at all
			Radio_rom_source.Selected = "Official Releases"		// Change selection
			Radio_rom_source.Refresh()							// without callback
		}
	}

	// Disable Radio group if Upstream or the Archive has no rom
	if len(get.A1.Upstream.Romlist) == 0 || len(get.A1.Archive.Romlist) == 0 {
		Radio_rom_source.Disable()
	}

	// Notify user that no rom is available if both Upstream and Archive have no rom
	if len(get.A1.Upstream.Romlist) == 0 && len(get.A1.Archive.Romlist) == 0 {
		Lbl_user_rom.SetText("No rom available")
	}
}

// Prefer LineageOS over other roms
// Select first in list if LineageOS is not available
func selectDefaultRomForGivenSource() {
	Select_rom.ClearSelected()
	if len(Select_rom.Options) > 0 {
		// Prefer LineageOS over other roms
		for _, romname := range Select_rom.Options {
			if romname == "LineageOS" && Select_rom.Selected != "LineageOS" {
				Select_rom.SetSelected("LineageOS")
			}
		}

		// Select first in list if LineageOS is not available
		if Select_rom.Selected != "LineageOSMicroG" && Select_rom.Selected != "LineageOS" {
			if Select_rom.Selected != Select_rom.Options[0] {
				Select_rom.SetSelected(Select_rom.Options[0])
			}
		}
	}
}


func initSettingstabWidgets() {
	// Left side
	Lbl_android_version = widget.NewLabel("")
	Select_rom = widget.NewSelect([]string{}, changedRom)
	Radio_rom_source = widget.NewRadioGroup([]string{"Archive", "Official Releases"}, changedRomSource)
	Radio_rom_source.Horizontal = true
	Chk_user_rom = widget.NewCheck("Provide your own rom zip file", chkUserRomChanged)
	Lbl_user_rom = widget.NewLabel("")
	Lbl_user_rom.Wrapping = fyne.TextTruncate
	Lbl_user_rom.Alignment = fyne.TextAlignCenter

	// Right side
	Chk_fdroid = widget.NewCheck("Install F-Droid", func(bool) {})
	Chk_aurora = widget.NewCheck("Install Aurora Store", func(bool) {})
	Chk_playstore = widget.NewCheck("Install Google Play Store", func(bool) {})
	Select_gapps = widget.NewSelect([]string{"MicroG", "OpenGapps", "Nothing"}, selectGappsChanged)
	Select_opengapps_variant = widget.NewSelect([]string{"pico", "nano"}, selectOpengappsVariantChanged)
	Select_opengapps_version = widget.NewSelect([]string{}, selectOpengappsVersionChanged)
}

func setDefaultsSettingstab() {
	Radio_rom_source.SetSelected("Official Releases")
	Select_rom.PlaceHolder = "Select rom"
	Select_rom.Disable()
	Chk_fdroid.SetChecked(true)
	Chk_aurora.SetChecked(true)
	Select_opengapps_version.PlaceHolder = "Select version"
	Select_opengapps_version.Disable()
	Select_opengapps_variant.SetSelected("pico")
	Select_opengapps_variant.Disable()
	Select_gapps.SetSelected("MicroG")
}

func settingstab() fyne.CanvasObject {
	// Left side
	box_labels := container.NewHBox(widget.NewLabel("Choose your rom:"), layout.NewSpacer(), Lbl_android_version)

	rom_info_icon := newTappableIcon(theme.InfoIcon())
	rom_info_icon.OnTapped = openWebBrowserRom
	rom_reload_icon := newTappableIcon(theme.ViewRefreshIcon())
	rom_reload_icon.OnTapped = func() {go ReloadRoms()}
	box_rom := container.NewBorder(nil, nil, nil, container.NewHBox(rom_info_icon, rom_reload_icon), Select_rom)

	box_source := Radio_rom_source

	leftside := container.NewVBox(box_labels, box_rom, box_source, Chk_user_rom, Lbl_user_rom)
	leftcard := widget.NewCard("", "", leftside)

	// Right side
	gapps_info_icon := newTappableIcon(theme.InfoIcon())
	gapps_info_icon.OnTapped = openWebBrowserGapps
	box_gapps := container.NewBorder(nil, nil, nil, gapps_info_icon, Select_gapps)

	opengappsvariants_info_icon := newTappableIcon(theme.InfoIcon())
	opengappsvariants_info_icon.OnTapped = openWebBrowserOpengappsVariants
	box_opengappsvariants := container.NewBorder(nil, nil, nil, opengappsvariants_info_icon, Select_opengapps_variant)

	box_opengappsvervar := container.NewHBox(Select_opengapps_version, layout.NewSpacer(), box_opengappsvariants)

	fdroid_info_icon := newTappableIcon(theme.InfoIcon())
	fdroid_info_icon.OnTapped = openWebBrowserFdroid
	box_fdroid := container.NewBorder(nil, nil, nil, fdroid_info_icon, Chk_fdroid)

	aurora_info_icon := newTappableIcon(theme.InfoIcon())
	aurora_info_icon.OnTapped = openWebBrowserAurora
	box_aurora := container.NewBorder(nil, nil, nil, aurora_info_icon, Chk_aurora)

	rightside := container.NewVBox(box_gapps, box_opengappsvervar, box_fdroid, box_aurora, Chk_playstore)
	rightcard := widget.NewCard("", "", rightside)

	grid := container.New(layout.NewGridLayout(2), leftcard, rightcard)
	return container.NewVBox(layout.NewSpacer(), grid, layout.NewSpacer())
}