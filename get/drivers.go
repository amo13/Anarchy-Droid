package get

import (
	"os"
	"anarchy-droid/helpers"
)

func Zadig() error {
	url := "https://github.com/pbatard/libwdi/releases/download/b730/zadig-2.5.exe"
	fallback_url := "https://stuff.free-droid.com/zadig-2.5.exe"

	err := DownloadFile("bin/zadig-2.5.exe", url, "")
	if err != nil {
		err = DownloadFile("bin/zadig-2.5.exe", fallback_url, ".sha256")
		if err != nil {
			return err
		}
	}

	return nil
}

func AdbDriver() error {
	url := "https://cdn.universaladbdriver.com/wp-content/uploads/universaladbdriver_v6.0.zip"
	fallback_url := "https://stuff.free-droid.com/universaladbdriver_v6.0.zip"

	err := DownloadFile("universaladbdriver_v6.0.zip", url, "")
	if err != nil {
		err = DownloadFile("universaladbdriver_v6.0.zip", fallback_url, ".sha256")
		if err != nil {
			return err
		}
	}

	err = helpers.Unzip("universaladbdriver_v6.0.zip", "bin")
	if err != nil {
		return err
	}

	err = os.Remove("universaladbdriver_v6.0.zip")
	if err != nil {
		return err
	}

	return nil
}