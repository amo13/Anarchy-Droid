package get

func Zadig() error {
	url := "https://github.com/pbatard/libwdi/releases/download/b755/zadig-2.6.exe"
	fallback_url := "https://stuff.free-droid.com/zadig-2.6.exe"

	err := DownloadFile("bin/zadig.exe", url, "")
	if err != nil {
		err = DownloadFile("bin/zadig.exe", fallback_url, ".sha256")
		if err != nil {
			return err
		}
	}

	return nil
}

func AdbDriver() error {
	url := "https://github.com/koush/adb.clockworkmod.com/releases/download/v1.0.0/UniversalAdbDriverSetup.msi"
	fallback_url := "https://stuff.free-droid.com/UniversalAdbDriverSetup.msi"

	err := DownloadFile("bin/UniversalAdbDriverSetup.msi", url, "")
	if err != nil {
		err = DownloadFile("bin/UniversalAdbDriverSetup.msi", fallback_url, "")
		if err != nil {
			return err
		}
	}

	return nil
}