package main

import(
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/dialog"

	"strings"
	"net/url"

	"github.com/amo13/anarchy-droid/get"
	"github.com/amo13/anarchy-droid/logger"
	"github.com/amo13/anarchy-droid/lookup"
	"github.com/amo13/anarchy-droid/helpers"
)

var Btn_bootloop_help *widget.Button
var Lbl_bootloop_entry *widget.Label
var Lbl_bootloop_info *widget.Label
var Entry_bootloop_model *widget.Entry
var Btn_bootloop_start_rescue *widget.Button


func btnBootloopHelpClicked() {
	Lbl_bootloop_entry.Show()
	Lbl_bootloop_info.Show()
	Entry_bootloop_model.Show()
	Btn_bootloop_start_rescue.Show()
}

func btnBootloopStartRescueClicked() {
	if Entry_bootloop_model.Text == "" { return	}
	
	bootloop_codename, err := lookup.ModelToCodename(Entry_bootloop_model.Text)
	if err != nil {
		if strings.Contains(err.Error(), "ambiguous") {
			cc, err := lookup.ModelToCodenameCandidates(Entry_bootloop_model.Text)
			if err != nil {
				logger.LogError("Error retrieving codename candidates from model " + Entry_bootloop_model.Text, err)
				return
			}

			mc := []string{}

			// Populate model candidates slice
			// for _, codename := range cc {
				models, err := lookup.CodenamesToModels(cc)
				if err != nil {
					logger.LogError("Failed to convert codenames to models.", err)
				}

				for _, model := range models {
					mc = append(mc, model)
				}
			// }

			Candidates.Options = helpers.UniqueNonEmptyElementsOfSlice(mc)

			candidates_dialog := dialog.NewCustom("Select your device model", "OK", Candidates, w)
			candidates_dialog.SetOnClosed(func() {
				bootloop_codename, err = lookup.ModelToCodename(Candidates.Selected)
				if err != nil {
					logger.LogError("Unable to lookup model to codename:", err)
				}
			})
			candidates_dialog.Show()
		} else {
			logger.LogError("Unable to lookup model to codename:", err)
		}
	}

	if bootloop_codename != "" {
		Lbl_bootloop_info.SetText("Device: " + bootloop_codename)
		Btn_bootloop_start_rescue.Disable()

		// Check if TWRP is available
		if !Chk_user_twrp.Checked {
			// Force new queries
			get.A1 = get.NewAvailable()
			err := get.A1.Populate(bootloop_codename)
			if err != nil {
				logger.LogError("unable to populate the list of available roms:", err)
			}
			selectDefaultTwrp()

			if get.A1.User.Twrp.Img.Href == "" {
				Lbl_bootloop_info.SetText("Sorry, no TWRP available.")
				Btn_bootloop_start_rescue.Enable()
				return
			}
		}

		err := bootloopRescue(bootloop_codename)
		if err != nil {
			logger.LogError("Failed to rescue from bootloop.", err)
		}
	} else {
		Lbl_bootloop_info.SetText("Sorry, unknown device.")
	}
	
	Btn_bootloop_start_rescue.Enable()
}

func initHelptabWidgets() {
	Btn_bootloop_help = widget.NewButton("My device is not booting any more", btnBootloopHelpClicked)
	Lbl_bootloop_entry = widget.NewLabel("Enter your device model:")
	Lbl_bootloop_info = widget.NewLabel("")
	Entry_bootloop_model = widget.NewEntry()
	Btn_bootloop_start_rescue = widget.NewButton("Rescue", btnBootloopStartRescueClicked)
}

func setDefaultsHelptab() {
	Lbl_bootloop_entry.Hide()
	Lbl_bootloop_info.Hide()
	Entry_bootloop_model.Hide()
	Entry_bootloop_model.PlaceHolder = "For example: SM-G900F"
	Btn_bootloop_start_rescue.Hide()
}

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

	href2 := "https://stuff.anarchy-droid.com/universaladbdriver_v6.0.zip"
	u2, err := url.Parse(href2)
	if err != nil {
		logger.LogError("unable to parse " + href2 + " as URL:", err)
	}
	link_to_universal_drivers := widget.NewHyperlinkWithStyle("Universal drivers", u2, fyne.TextAlignCenter, fyne.TextStyle{})


	leftside := container.NewVBox(tryfirst, widget.NewLabel(""), container.NewCenter(container.NewHBox(link_to_official_drivers, link_to_universal_drivers)))
	leftcard := widget.NewCard("", "", leftside)

	rightside := container.NewVBox(Btn_bootloop_help, Lbl_bootloop_entry, Entry_bootloop_model, Btn_bootloop_start_rescue, Lbl_bootloop_info)
	rightcard := widget.NewCard("", "", rightside)

	grid := container.New(layout.NewGridLayout(2), leftcard, rightcard)
	return container.NewVBox(layout.NewSpacer(), grid, layout.NewSpacer())
}