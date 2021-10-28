package main

import (
	"os"
	"time"
	"strings"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/amo13/anarchy-droid/logger"
	"github.com/amo13/anarchy-droid/helpers"

	"github.com/getsentry/sentry-go"
)

const AppName = "Anarchy-Droid"
var AppVersion string	// AppVersion infected from FyneApp.toml during build
var BuildDate string	// Build date injected during build
// use: go build -ldflags "-X main.BuildDate=`date +%Y-%m-%d` -X main.AppVersion=`awk -F'[ ="]+' '$1 == "Version" { print $2 }' FyneApp.toml`" .
// or use sed to insert the values in place after checkout and before compilation:
// sed -i "s/.*var AppVersion string.*/var AppVersion string = \"`awk -F'[ ="]+' '$1 == "Version" { print $2 }' FyneApp.toml`\"/" main.go
// sed -i "s/.*var BuildDate string.*/var BuildDate string = \"`date +%Y-%m-%d`\"/" main.go

var a fyne.App
var w fyne.Window
var active_screen string

func main() {
	// Set AppVersion to "DEVELOPMENT" if it was left empty
	if AppVersion == "" {
		AppVersion = "DEVELOPMENT"
	}

	// Propagate AppVersion and BuildDate to the logger package
	logger.Consent = true
	logger.AppName = AppName
	logger.AppVersion = AppVersion
	logger.BuildDate = BuildDate

	err := sentry.Init(sentry.ClientOptions{
		// Either set your DSN here or set the SENTRY_DSN environment variable.
		Dsn: "https://26d9d7416f0e45ac8ab1733c8d691f1d@o1013551.ingest.sentry.io/5978898",
		// Either set environment and release here or set the SENTRY_ENVIRONMENT
		// and SENTRY_RELEASE environment variables.
		// Environment: "",
		Release: AppName + "@" + AppVersion,
		// Enable printing of SDK debug messages.
		// Useful when getting started or trying to figure something out.
		Debug: false,
	})
	if err != nil {
		logger.Log("sentry.Init: " + err.Error())
	}
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{ID: logger.Sessionmap["id"]})
	})
	// Flush buffered events before the program terminates.
	// Set the timeout to the maximum duration the program can afford to wait.
	defer sentry.Flush(5 * time.Second)

	a = app.NewWithID("com.anarchy-droid")
	a.SetIcon(resourceIconPng)

	w = a.NewWindow(AppName)
	w.SetMaster()

	w.Resize(fyne.NewSize(562, 226))
	// w.SetFixedSize(true)

	active_screen = "initScreen"
	w.SetContent(initScreen())

	// On MacOS, move into the user's Downloads folder
	// to prevent being either inside the read-only
	// application package or inside a jail folder
	if runtime.GOOS == "darwin" {
		u, _ := helpers.Cmd("whoami")
		u = strings.ReplaceAll(u, "\n", "")
		os.Chdir("/Users/" + u + "/Downloads")
	}

	// Set working directory to a subdir named like the app
	_, err = os.Stat(AppName)
	if os.IsNotExist(err) {
		err = os.Mkdir(AppName, 0755)
	    if err != nil {
	        logger.LogError("Error setting working directory:", err)
	    }
	}
	os.Chdir(AppName)

	// Create log directory if it does not exist
	_, err = os.Stat("log")
	if os.IsNotExist(err) {
		err = os.Mkdir("log", 0755)
	    if err != nil {
	        logger.LogError("Unable to create log directory:", err)
	    }
	}

	go func() {
		go logger.Report(map[string]string{"progress":"Setup App"})
		_, err := initApp()
		// logger.Log(success)
		if err != nil {
			logger.LogError("Setup failed:", err)
		}
	}()

	w.ShowAndRun()
}