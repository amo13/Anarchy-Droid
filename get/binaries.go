package get

import (
	"os"
	"fmt"
	"runtime"
	"anarchy-droid/logger"
	"anarchy-droid/helpers"
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
	url := "https://gitlab.com/free-droid/free-droid/-/raw/master/heimdall"
	thisOS := runtime.GOOS
	if thisOS == "windows" {
		thisOS = "win"
	}
	url = url + "/heimdall-" + thisOS + ".zip"

	err := DownloadFile("heimdall.zip", url, "")
	if err != nil {
		return err
	}

	logger.Log("Done downloading heimdall")

	err = helpers.Unzip("heimdall.zip", "bin/heimdall/")
	if err != nil {
		return err
	}

	logger.Log("Done extracting heimdall")

	err = os.Remove("heimdall.zip")
	if err != nil {
		return err
	}

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