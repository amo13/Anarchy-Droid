package main

import(
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/container"

	"github.com/amo13/anarchy-droid/get"
	"github.com/amo13/anarchy-droid/device"
	"github.com/amo13/anarchy-droid/logger"
	"github.com/amo13/anarchy-droid/helpers"
	"github.com/amo13/anarchy-droid/device/adb"
	"github.com/amo13/anarchy-droid/device/fastboot"
	"github.com/amo13/anarchy-droid/device/heimdall"

	"github.com/creativeprojects/go-selfupdate"

	"os"
	"fmt"
	"strings"
	"runtime"
	"net/url"
)

var Icon_internet *widget.Icon
var Icon_binaries *widget.Icon
var Icon_uptodate *widget.Icon
var Icon_adbserver *widget.Icon
var Lbl_init_infotext *widget.Label

func initScreen() fyne.CanvasObject {
	logo := canvas.NewImageFromResource(resourceIconPng)
	logo.FillMode = canvas.ImageFillOriginal
	// logo.SetMinSize(fyne.NewSize(30, 30))

	Lbl_init_text1 := widget.NewLabel("Initializing " + AppName + " " + AppVersion)
	Lbl_init_text1.Alignment = fyne.TextAlignCenter

	Lbl_init_infotext = widget.NewLabel(AppName + " " + AppVersion)
	Lbl_init_infotext.Alignment = fyne.TextAlignCenter

	Lbl_init_internet := widget.NewLabel("Internet connection")
	Icon_internet = widget.NewIcon(nil)
	Container_init_internet := container.NewHBox(layout.NewSpacer(), Icon_internet, Lbl_init_internet, layout.NewSpacer())

	Lbl_init_binaries := widget.NewLabel("Get the binaries")
	Icon_binaries = widget.NewIcon(nil)
	Container_init_binaries := container.NewHBox(layout.NewSpacer(), Icon_binaries, Lbl_init_binaries, layout.NewSpacer())

	Lbl_init_uptodate := widget.NewLabel("Application up to date")
	Icon_uptodate = widget.NewIcon(nil)
	Container_init_uptodate := container.NewHBox(layout.NewSpacer(), Icon_uptodate, Lbl_init_uptodate, layout.NewSpacer())

	Lbl_init_adbserver := widget.NewLabel("Listen for device connections")
	Icon_adbserver = widget.NewIcon(nil)
	Container_init_adbserver := container.NewHBox(layout.NewSpacer(), Icon_adbserver, Lbl_init_adbserver, layout.NewSpacer())

	// Left side
	leftside := container.NewVBox(layout.NewSpacer(), logo, layout.NewSpacer())

	// Right side
	rightside := container.NewVBox(
		layout.NewSpacer(),
		Container_init_internet, Container_init_binaries,
		Container_init_uptodate, Container_init_adbserver,
		layout.NewSpacer(),
	)

	grid := container.New(layout.NewGridLayout(2), leftside, rightside)
	return container.NewVBox(Lbl_init_text1, layout.NewSpacer(), grid, layout.NewSpacer(), Lbl_init_infotext)
}

func initApp() (bool, error) {
	Lbl_init_infotext.SetText("Checking internet connection...")
	status_code, err := get.StatusCode("https://gitlab.com/free-droid/free-droid/raw/master/lookups/codenames.yml")
	if err != nil {
		return false, err
	}
	if status_code == "200 OK" {
		Icon_internet.SetResource(theme.ConfirmIcon())
	} else {
		Icon_internet.SetResource(theme.CancelIcon())
		return false, nil
	}

	Lbl_init_infotext.SetText("Downloading binaries...")
	err = get.Binaries()
	if err != nil {
		Icon_binaries.SetResource(theme.CancelIcon())
		return false, err
	} else {
		Icon_binaries.SetResource(theme.ConfirmIcon())
	}

	Lbl_init_infotext.SetText("Updating application...")
	if AppVersion != "DEVELOPMENT" {
		err = selfUpdate(AppVersion)
		if err != nil {
			if err.Error() == "Update successful, please restart the application" {
				Icon_uptodate.SetResource(theme.MediaReplayIcon())
				info_dialog := dialog.NewInformation("Update successful", AppName + " has been updated. Please restart the application.", w)
				info_dialog.SetOnClosed(func() { a.Quit() } )
				info_dialog.Show()
				return true, nil
			} else if strings.HasPrefix(err.Error(), "error occurred while detecting version") {
				logger.LogError("Unable to check for update:", err)
				logger.Log("Continuing anyway...")
				Icon_uptodate.SetResource(theme.ConfirmIcon())
			} else {
				Icon_uptodate.SetResource(theme.CancelIcon())
				info_dialog := dialog.NewInformation("A new version has been released!", "Please download the latest version of" + AppName + ". The download page will open once you close the application.", w)
				info_dialog.SetOnClosed(func() { OpenWebBrowser("https://github.com/amo13/Anarchy-Droid/releases/latest"); a.Quit() } )
				info_dialog.Show()
				return false, err
			}
		} else {
			Icon_uptodate.SetResource(theme.ConfirmIcon())
		}
	} else {
		Icon_uptodate.SetResource(theme.ConfirmIcon())
	}

	if runtime.GOOS == "linux" {
		password := widget.NewPasswordEntry()
		href := "https://github.com/amo13/Anarchy-Droid#linux"
		u, err := url.Parse(href)
		if err != nil {
			logger.LogError("unable to parse " + href + " as URL:", err)
		}
		info := widget.NewHyperlink("Show more info on this", u)
		udev := false
		items := []*widget.FormItem{
			widget.NewFormItem("Sudo password", password),
			widget.NewFormItem("Why does this app need sudo?", info),
			widget.NewFormItem("I have setup udev rules for my device", widget.NewCheck("", func(checked bool) {
				udev = checked
				if checked {
					password.Disable()
				} else {
					password.Enable()
				}
			})),
		}

		dialog.ShowForm(AppName + " needs either sudo password or udev rules", "Continue", "Exit", items, func(b bool) {
			if b {
				Lbl_init_infotext.Text = "Restarting ADB server..."
				adb.Nosudo = udev
				adb.Sudopw = password.Text
				fastboot.Nosudo = udev
				fastboot.Sudopw = password.Text
				heimdall.Nosudo = udev
				heimdall.Sudopw = password.Text

				finishInitApp()
			} else {
				a.Quit()
			}
		}, w)

		w.Canvas().Focus(password)

		return true, nil
	} else {
		Lbl_init_infotext.Text = "Restarting ADB server..."
		return finishInitApp()
	}
}

func finishInitApp() (bool, error) {
	// Detect if the given sudo password is wrong and exit in that case
	if adb.Sudopw != "" {
		_, stderr := helpers.Cmd("printf " + adb.Sudopw + " | sudo -S ls")
		if strings.Contains(stderr, "incorrect password") {
			logger.Log("sudo password seems to be wrong")
			info_dialog := dialog.NewError(fmt.Errorf("The sudo password seems to be wrong."), w)
			info_dialog.SetOnClosed(func() {
				logger.Log("Exiting...")
				a.Quit()
			} )
			info_dialog.Show()
			return false, fmt.Errorf("wrong sudo password")
		}
	}

	// Restart the ADB server (as root on linux)
	err := adb.KillServer()
	if err != nil && err.Error() != "connection refused" {
		Icon_adbserver.SetResource(theme.CancelIcon())
		return false, err
	}
	Lbl_init_infotext.SetText("Restarting ADB server...")
	err = adb.StartServer()
	if err != nil {
		Icon_adbserver.SetResource(theme.CancelIcon())
		return false, err
	}
	Icon_adbserver.SetResource(theme.ConfirmIcon())

	// Kill adb server if the app is closed
	// but only if it has been started
	w.SetOnClosed(func() { adb.KillServer() })

	// Start watching for device connections
	device.D1.Observe()

	if Icon_internet.Resource == theme.ConfirmIcon() &&
		Icon_binaries.Resource == theme.ConfirmIcon() &&
		Icon_adbserver.Resource == theme.ConfirmIcon() {
		go logger.Report(map[string]string{"progress":"Setup Successful"})
		active_screen = "mainScreen"
		w.SetContent(mainScreen())
		return true, nil
	} else {
		go logger.Report(map[string]string{"progress":"Setup Failed"})
		return false, nil
	}
}

func selfUpdate(version string) error {
	latest, found, err := selfupdate.DetectLatest("amo13/Anarchy-Droid")
	if err != nil {
		return fmt.Errorf("error occurred while detecting version: %v", err)
	}
	if !found {
		return fmt.Errorf("latest version for %s/%s could not be found from github repository", runtime.GOOS, runtime.GOARCH)
	}

	if latest.LessOrEqual(version) {
		logger.Log("Current version", version, "is the latest")
		return nil
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not locate executable path: %v", err)
	}
	if err := selfupdate.UpdateTo(latest.AssetURL, latest.AssetName, exe); err != nil {
		return fmt.Errorf("error occurred while updating binary: %v", err)
	}
	logger.Log("Successfully updated to version", latest.Version())
	return fmt.Errorf("Update successful, please restart the application")
}