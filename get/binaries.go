package get

import (
	"os"
	"fmt"
	"runtime"

	"github.com/amo13/anarchy-droid/logger"
	"github.com/amo13/anarchy-droid/helpers"
)

func Binaries() error {
	if !gotPlatformTools() {
		err := dlPlatformTools()
		if err != nil {
			return err
		}
	}

	if !gotHeimdall() {
		err := dlHeimdall()
		if err != nil {
			return err
		}
	}

	if !gotPlatformTools() {
		return fmt.Errorf("unable to download and extract the platform-tools binaries")
	}

	if !gotHeimdall() {
		return fmt.Errorf("unable to download and extract the heimdall binary")
	}

	return nil
}

func dlPlatformTools() error {
	url := "https://dl.google.com/android/repository"
	url = url + "/platform-tools-latest-" + runtime.GOOS + ".zip"

	err := DownloadFile("platform-tools.zip", url, "")
	if err != nil {
		return err
	}

	logger.Log("Done downloading platform tools")

	err = helpers.Unzip("platform-tools.zip", "bin")
	if err != nil {
		return err
	}

	logger.Log("Done extracting platform tools")

	err = os.Remove("platform-tools.zip")
	if err != nil {
		return err
	}

	return nil
}

func dlHeimdall() error {
	url := "https://github.com/amo13/Heimdall/releases/download/v1.4.2/"
	thisOS := runtime.GOOS
	if thisOS == "darwin" {
		thisOS = "macos"
	}
	url = url + "/heimdall-" + thisOS
	if thisOS == "windows" {
		url = url + ".exe"
	}

	if thisOS == "windows" {
		err := DownloadFile("bin/heimdall/heimdall.exe", url, "")
		if err != nil {
			return err
		}
	} else {
		err := DownloadFile("bin/heimdall/heimdall", url, "")
		if err != nil {
			return err
		}
	}

	logger.Log("Done downloading heimdall")
	
	return nil
}

func gotPlatformTools() bool {
	suffix := ""
	if runtime.GOOS == "windows" {
		suffix = ".exe"
	}

	_, err := os.Stat("bin/platform-tools/adb" + suffix)
	if os.IsNotExist(err) {
		return false
	}
	_, err = os.Stat("bin/platform-tools/fastboot" + suffix)
	if os.IsNotExist(err) {
		return false
	}
	
	return true
}

func gotHeimdall() bool {
	suffix := ""
	if runtime.GOOS == "windows" {
		suffix = ".exe"
	}

	_, err := os.Stat("bin/heimdall/heimdall" + suffix)
	if os.IsNotExist(err) {
		return false
	}

	return true
}