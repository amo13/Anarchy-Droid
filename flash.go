package main

import(
	"github.com/amo13/anarchy-droid/logger"
	"github.com/amo13/anarchy-droid/helpers"
	"github.com/amo13/anarchy-droid/device"
	"github.com/amo13/anarchy-droid/device/twrp"
	"github.com/amo13/anarchy-droid/get"

	"os"
	"fmt"
	"sync"
	"time"
	"strings"
)

var Files map[string]string

func prepareFlash() error {
	w.SetContent(flashingScreen())
	active_screen = "flashingScreen"

	go logger.Report(map[string]string{"progress":"Start"})
	logger.Log("Starting flashing procedure.")

	// Download files
	logger.Log("Downloading files...")
	Lbl_progressbar.SetText("Downloading files...")
	Progressbar.Start()

	err := fmt.Errorf("")
	Files, err = downloadFiles()
	if err != nil {
		Progressbar.Stop()
		logger.LogError("Error downloading files:", err)
		Lbl_flashing_instructions.SetText("Failed to download the necessary files:\n" + err.Error())
		return err
	}

	Progressbar.Stop()
	logger.Log("Files downloaded successfully:", helpers.MapToString(Files))
	Lbl_progressbar.SetText("Files downloaded successfully!")

	device.D1.Flashing = true

	// Try to unlock the device if needed
	if !device.D1.IsUnlocked && !Chk_skipunlock.Checked {
		go logger.Report(map[string]string{"progress":"Unlock"})
		logger.Log("Trying to unlock the bootloader...")
		Lbl_progressbar.SetText("Trying to unlock the bootloader...")

		switch strings.ToLower(device.D1.Brand) {
		case "sony":
			w.SetContent(sonyUnlockScreen())
		case "motorola":
			w.SetContent(motorolaUnlockScreen())
		default:
			err := device.D1.Unlock()
			if err != nil {
				logger.LogError("Unlocking the device seems to have failed:", err)
				Lbl_flashing_instructions.SetText("Unlocking the device seems to have failed:\n" + err.Error())
				return err
			}
		}

		logger.Log("Bootloader unlocked successfully!")
		Lbl_progressbar.SetText("Bootloader unlocked successfully!")
		go logger.Report(map[string]string{"progress":"Unlock successful"})
	} else {
		err := bootTwrpStep()
		if err != nil {
			if err.Error() == "manually booting recovery failed" {
				logger.Log("manually booting recovery failed")
				Lbl_flashing_instructions.SetText("Manually booting TWRP failed.\n\nPlease restart and try again.")
			} else {
				logger.LogError("Error booting TWRP:", err)
				Lbl_flashing_instructions.SetText("Error booting TWRP:\n" + err.Error())
			}
		}
	}

	return nil
}

func unlockStep(unlock_code string) {
	// Start goroutine to prevent blocking the UI calling this function with a button
	go func() {
		w.SetContent(flashingScreen())

		Progressbar.Start()

		err := device.D1.DoUnlock(unlock_code)
		if err != nil {
			logger.LogError("DoUnlock failed: ", err)
			Progressbar.Stop()
			if err.Error() == "not allowed" {
				Lbl_flashing_instructions.SetText("Unlocking the bootloader not allowed. OEM unlock has apparently not been enabled. Please enable it in your device settings and restart the application.")
			} else {
				Lbl_flashing_instructions.SetText("Unlocking the bootloader failed:\n" + err.Error())
			}
			return
		}

		Progressbar.Stop()

		// Observe if the device reboots
		// If yes, simply notify the user about the factory reset
		// and ask him to activate usb debugging in the settings again
		time.Sleep(5 * time.Second)
		if device.D1.State == "disconnected" {
			Lbl_flashing_instructions.SetText("Your device has been wiped and is now rebooting. This means unlocking the bootloader was probably successful!\nPlease reactivate USB Debugging in the system settings to continue: In Settings > About Phone: Tap 7 times on Build Number. Then in Settings > Developer Options: Activate USB Debugging.")
		}

		err = bootTwrpStep()
		if err != nil {
			if err.Error() == "manually booting recovery failed" {
				logger.Log("manually booting recovery failed")
				Lbl_flashing_instructions.SetText("Manually booting TWRP failed.\n\nPlease restart and try again.")
			} else {
				logger.LogError("Error booting TWRP:", err)
				Lbl_flashing_instructions.SetText("Error booting TWRP:\n" + err.Error())
			}
		}
	}()
}

func checkManualRecoveryBoot(reboot_instructions string) error {
	// Some devices can't boot TWRP directly from the bootloader
	// but need TWRP to be flashed to the recovery partition first
	// and then the user needs to hold a combination of hardware keys.
	// In that case, display the instructions to the user and
	// wait (block here) until the device is connected in recovery.
	if reboot_instructions != "" {
		Lbl_flashing_instructions.SetText(reboot_instructions)
		go logger.Report(map[string]string{"progress":"Manually boot recovery"})

		// In that case, TWRP has been actually flashed/installed
		// and not only temporarily booted
		Chk_skipflashtwrp.SetChecked(true)

		for helpers.IsStringInSlice(device.D1.State, []string{"fastboot", "heimdall", "disconnected"}) {
			time.Sleep(1 * time.Second)
		}

		if device.D1.State != "recovery" {
			go logger.Report(map[string]string{"progress":"Manually booting recovery failed"})
			return fmt.Errorf("manually booting recovery failed")
		} else {
			go logger.Report(map[string]string{"progress":"Manually booting recovery succeeded"})
		}
	}

	return nil
}

func bootTwrpStep() error {
	logger.Log("Arrived at TWRP booting step")

	if !Chk_skipflashtwrp.Checked {
		go logger.Report(map[string]string{"progress":"Boot TWRP"})
		logger.Log("Trying to boot/flash TWRP...")

		if Files["twrp_img"] == "" {
			return fmt.Errorf("Cannot boot TWRP: missing image file")
		}

		reboot_instructions, err := device.D1.BootRecovery(Files["twrp_img"])
		if err != nil {
			if err.Error() == "heimdall failed to access device" {
				err = installDriversWithZadig()
				if err != nil {
					logger.LogError("Failed to download zadig", err)
					return err
				}
				// Retry
				reboot_instructions, err = device.D1.BootRecovery(Files["twrp_img"])
				if err != nil {
					logger.LogError("TWRP boot attempt returns the following error:", err)
					return err
				}
			} else {
				logger.LogError("TWRP boot attempt returns the following error:", err)
				return err
			}
		}

		// Displays instructions and waits if needed
		err = checkManualRecoveryBoot(reboot_instructions)
		if err != nil {
			// Logging handled one step up the call stack
			return err
		}

		time.Sleep(5 * time.Second)
	}

	device.D1.State_request = "recovery"
	<-device.D1.State_reached	// blocks until recovery is reached

	if device.D1.IsAB {
		err := installOnAB()
		if err != nil {
			logger.LogError("Error during AB installation:", err)
			Lbl_flashing_instructions.SetText("Error during installation:\n" + err.Error())
		}
	} else {
		err := installOnAOnly()
		if err != nil {
			logger.LogError("Error during A-only installation:", err)
			Lbl_flashing_instructions.SetText("Error during installation:\n" + err.Error())
		}
	}

	return nil
}

func installDriversWithZadig() error {
	err := get.Zadig()
	if err != nil {
		return err
	}

	err = writeZadigConfig()
	if err != nil {
		return err
	}

	Lbl_flashing_instructions.SetText("Please install/replace the drivers for your device...\nSelect from the list what could be your device and press the button. (Sometimes it can be names like 05c6:9008, SGH-T959V or Generic Serial.)")

	os.Chdir("bin")
	defer os.Chdir("..")

	stdout, stderr := helpers.Cmd(`Powershell -Command "& { Start-Process \"zadig.exe\" -Verb RunAs }`)
	logger.Log("Zadig stdout:", stdout)
	logger.LogError("Zadig stderr:", fmt.Errorf(stderr))

	return nil
}

func writeZadigConfig() error {
	file, err := os.OpenFile("bin/zadig.ini", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
    if err != nil {
        return err
    }

    // Write the file
    fmt.Fprintln(file, "[general]")
    fmt.Fprintln(file, "advanced_mode = false")
	fmt.Fprintln(file, "exit_on_success = true")
	fmt.Fprintln(file, "log_level = 0")
	fmt.Fprintln(file, "  ")
	fmt.Fprintln(file, "[device]")
	fmt.Fprintln(file, "list_all = true")
	fmt.Fprintln(file, "include_hubs = false")
	fmt.Fprintln(file, "trim_whitespaces = true")
	fmt.Fprintln(file, "  ")
	fmt.Fprintln(file, "[driver]  ")
	fmt.Fprintln(file, "default_driver = 0")
	fmt.Fprintln(file, "extract_only = false")
	fmt.Fprintln(file, "  ")
	fmt.Fprintln(file, "[security]")

    err = file.Close()
    if err != nil {
        return err
    }

    return nil
}

func installOnAOnly() error {
	device.D1.State_request = "recovery"
	<-device.D1.State_reached	// blocks until recovery is reached

	logger.Log("Begin installOnAOnly()")

	// Wait for TWRP to be ready
	// User might need to unlock the data partition with a pattern
	waitForTwrpReady()

	Lbl_flashing_instructions.SetText("Great! Now relax and watch the magic happen!")
	Progressbar.Start()

	time.Sleep(1 * time.Second)

	if Files["rom"] != "" {
		logger.Log("Start rom installation...")
		Lbl_progressbar.SetText("Installing the operating system rom...")
		go logger.Report(map[string]string{"progress":"Flash rom"})
		if !Chk_skipwipedata.Checked {
			err := device.D1.FlashRom(Files["rom"], "clean")
			if err != nil {
				logger.LogError("Error clean-wiping device or flashing rom " + Files["rom"] + ":", err)
				return err
			}
		} else {
			err := device.D1.FlashRom(Files["rom"], "dirty")
			if err != nil {
				logger.LogError("Error dirty-wiping device or flashing rom " + Files["rom"] + ":", err)
				return err
			}
		}

		time.Sleep(1 * time.Second)
	}

	return finishInstallation()
}

func installOnAB() error {
	device.D1.State_request = "recovery"
	<-device.D1.State_reached	// blocks until recovery is reached

	logger.Log("Begin installOnAB()")

	// Wait for TWRP to be ready
	// User might need to unlock the data partition with a pattern
	waitForTwrpReady()

	Lbl_flashing_instructions.SetText("Great! Now relax and watch the magic happen!")
	Progressbar.Start()

	time.Sleep(1 * time.Second)

	if Chk_copypartitions.Checked {
		logger.Log("Sideloading copy-partitions.zip...")
		go logger.Report(map[string]string{"progress":"Copy Partitions"})

		Lbl_progressbar.SetText("Sideloading copy-partitions.zip...")
		err := device.D1.FlashZip(Files["copypartitions"])
		if err != nil {
			logger.LogError("Error flashing " + Files["copypartitions"] + ":", err)
			logger.Log("Proceeding anyway...")
		}

		time.Sleep(1 * time.Second)

		// Reboot TWRP afterwards
		if err != nil {
			go logger.Report(map[string]string{"progress":"Reboot TWRP"})
			logger.Log("Trying to boot/flash TWRP again...")

			if Files["twrp_img"] == "" {
				return fmt.Errorf("Cannot boot TWRP: missing image file")
			}

			if Chk_skipflashtwrp.Checked {
				device.D1.Reboot("recovery")
			} else {
				reboot_instructions, err := device.D1.BootRecovery(Files["twrp_img"])
				if err != nil {
					logger.LogError("TWRP boot attempt returns the following error:", err)
					return err
				}

				// Displays instructions and waits if needed
				err = checkManualRecoveryBoot(reboot_instructions)
				if err != nil {
					// Logging handled one step up the call stack
					return err
				}
			}

			time.Sleep(5 * time.Second)

			waitForTwrpReady()
		}
	}

	if Files["rom"] != "" {
		logger.Log("Start rom installation...")
		Lbl_progressbar.SetText("Installing the operating system rom...")
		go logger.Report(map[string]string{"progress":"Flash rom"})
		if !Chk_skipwipedata.Checked {
			err := device.D1.FlashRom(Files["rom"], "clean")
			if err != nil {
				logger.LogError("Error clean-wiping device or flashing rom " + Files["rom"] + ":", err)
				return err
			}
		} else {
			err := device.D1.FlashRom(Files["rom"], "dirty")
			if err != nil {
				logger.LogError("Error dirty-wiping device or flashing rom " + Files["rom"] + ":", err)
				return err
			}
		}

		// Flashing an AB rom replaces the recovery (at least LineageOS does so)
		Chk_skipflashtwrp.SetChecked(false)
	}

	time.Sleep(1 * time.Second)

	// Reboot to TWRP so the active slot switches
	go logger.Report(map[string]string{"progress":"Reboot TWRP"})
	logger.Log("Trying to boot/flash TWRP again...")

	if Files["twrp_img"] == "" {
		return fmt.Errorf("Cannot boot TWRP: missing image file")
	}

	reboot_instructions, err := device.D1.BootRecovery(Files["twrp_img"])
	if err != nil {
		logger.LogError("TWRP boot attempt returns the following error:", err)
		return err
	}

	// Displays instructions and waits if needed
	err = checkManualRecoveryBoot(reboot_instructions)
	if err != nil {
		// Logging handled one step up the call stack
		return err
	}

	time.Sleep(5 * time.Second)

	waitForTwrpReady()

	return finishInstallation()
}

func finishInstallation() error {
	// Send NanoDroid config file
	logger.Log("Sending the NanoDroid setup file...")
	time.Sleep(1 * time.Second)
	err := twrp.SendNanodroidSetup(createNanoDroidSetup())
	if err != nil {
		logger.LogError("Error sending the NanoDroid setup file:", err)
	}

	if Files["gapps"] != "" {
		logger.Log("Start gapps installation...")
		go logger.Report(map[string]string{"progress":"Flash Gapps or MicroG"})
		if Select_gapps.Selected == "MicroG" {
			Lbl_progressbar.SetText("Installing MicroG...")
		} else {
			Lbl_progressbar.SetText("Installing Google framework and apps...")
		}
		err := device.D1.FlashZip(Files["gapps"])
		if err != nil {
			logger.LogError("Error flashing " + Files["gapps"] + ":", err)
			logger.Log("Proceeding anyway...")
		}

		time.Sleep(1 * time.Second)
	}

	if Files["aurora"] != "" {
		logger.Log("Start aurora installation...")
		go logger.Report(map[string]string{"progress":"Flash Aurora Store"})
		Lbl_progressbar.SetText("Installing Aurora Store...")
		err := device.D1.FlashZip(Files["aurora"])
		if err != nil {
			logger.LogError("Error flashing " + Files["aurora"] + ":", err)
			logger.Log("Proceeding anyway...")
		}

		time.Sleep(1 * time.Second)
	}

	if Files["fdroid"] != "" {
		logger.Log("Start F-Droid installation...")
		go logger.Report(map[string]string{"progress":"Flash F-Droid"})
		Lbl_progressbar.SetText("Installing F-Droid...")
		err := device.D1.FlashZip(Files["fdroid"])
		if err != nil {
			logger.LogError("Error flashing " + Files["fdroid"] + ":", err)
			logger.Log("Proceeding anyway...")
		}

		time.Sleep(1 * time.Second)
	}

	if Files["gsyncswype"] != "" {
		logger.Log("Start Gsync/Swype installation...")
		go logger.Report(map[string]string{"progress":"Flash Gsync or swype"})
		Lbl_progressbar.SetText("Installing Google sync adapters and/or Swype libraries...")
		err := device.D1.FlashZip(Files["gsyncswype"])
		if err != nil {
			logger.LogError("Error flashing " + Files["gsyncswype"] + ":", err)
			logger.Log("Proceeding anyway...")
		}

		time.Sleep(1 * time.Second)
	}

	if Files["patcher"] != "" {
		logger.Log("Installing rom patcher...")
		go logger.Report(map[string]string{"progress":"Flash patcher"})
		Lbl_progressbar.SetText("Patching the system for signature spoofing...")
		err := device.D1.FlashZip(Files["patcher"])
		if err != nil {
			logger.LogError("Error flashing " + Files["patcher"] + ":", err)
			logger.Log("Proceeding anyway...")
		}

		time.Sleep(1 * time.Second)
	}

	logger.Log("Finished.")
	go logger.Report(map[string]string{"progress":"Finished successfully"})
	Lbl_progressbar.SetText("")
	Progressbar.Stop()
	Lbl_flashing_instructions.SetText("Installation finished!\n\nNotice: The first boot will take longer.")

	if !device.D1.Flashing {
		logger.Log("User cancelled flashing")
		return fmt.Errorf("cancelled")
	}

	device.D1.State_request = "recovery"
	<-device.D1.State_reached	// blocks until recovery is reached
	device.D1.Reboot("android")

	time.Sleep(20 * time.Second)

	// Reset everything
	device.D1.StartOver()
	get.A1 = get.NewAvailable()
	w.SetContent(mainScreen())

	return nil
}

func waitForTwrpReady() {
	ready, err := twrp.IsReady()
	if err != nil {
		logger.LogError("Unable to check if TWRP is ready:", err)
	}
	for !ready {
		Lbl_flashing_instructions.SetText("Waiting for TWRP to be ready...\n\nIf you can, please unlock TWRP on your device screen.")
		time.Sleep(1 * time.Second)
		ready, err = twrp.IsReady()
		if err != nil {
			logger.LogError("Unable to check if TWRP is ready:", err)
		}
	}
}

type RetrievalError struct {
	What string
	Href string
	Error error
}

// Downloads everything needed in parallel
func downloadFiles() (map[string]string, error) {
	Files = make(map[string]string)
	var wg sync.WaitGroup
	errs := make(chan RetrievalError)

	rom_path := ""
	if Chk_user_rom.Checked {
		rom_path = get.A1.User.Rom.Href
	} else {
		rom_path = "flash/" + get.A1.User.Rom.Filename
		wg.Add(1)
		go func() {
			defer wg.Done()

			err := get.DownloadFile(rom_path, get.A1.User.Rom.Href, get.A1.User.Rom.Checksum_url_suffix)
			if err != nil {
				errs <- RetrievalError{get.A1.User.Rom.Name, get.A1.User.Rom.Href, err}
			}
		}()
	}
	if rom_path != "" {
		Files["rom"] = rom_path
	}

	twrp_img_path := ""
	if Chk_user_twrp.Checked {
		twrp_img_path = get.A1.User.Twrp.Img.Href
	} else {
		twrp_img_path = "flash/" + get.A1.User.Twrp.Img.Filename
		wg.Add(1)
		go func() {
			defer wg.Done()

			err := get.DownloadFile(twrp_img_path, get.A1.User.Twrp.Img.Href, get.A1.User.Twrp.Img.Checksum_url_suffix)
			if err != nil {
				errs <- RetrievalError{"TWRP image", get.A1.User.Twrp.Img.Href, err}
			}
		}()
	}
	if twrp_img_path != "" {
		Files["twrp_img"] = twrp_img_path
	}

	twrp_zip_path := ""
	if !Chk_user_twrp.Checked && get.A1.User.Twrp.Zip.Version == get.A1.User.Twrp.Img.Version && get.A1.User.Twrp.Img.Version != "" {
		twrp_zip_path = "flash/" + get.A1.User.Twrp.Zip.Filename
		wg.Add(1)
		go func() {
			defer wg.Done()

			err := get.DownloadFile(twrp_zip_path, get.A1.User.Twrp.Zip.Href, get.A1.User.Twrp.Zip.Checksum_url_suffix)
			if err != nil {
				errs <- RetrievalError{"TWRP zip", get.A1.User.Twrp.Zip.Href, err}
			}
		}()
	}
	if twrp_zip_path != "" {
		Files["twrp_zip"] = twrp_zip_path
	}

	magisk_path := ""
	if Chk_magisk.Checked {
		magisk_path = "flash/" + get.A1.Upstream.Magisk.Filename
		wg.Add(1)
		go func() {
			defer wg.Done()

			err := get.DownloadFile(magisk_path, get.A1.Upstream.Magisk.Href, get.A1.Upstream.Magisk.Checksum_url_suffix)
			if err != nil {
				errs <- RetrievalError{"Magisk", get.A1.Upstream.Magisk.Href, err}
			}
		}()
	}
	if magisk_path != "" {
		Files["magisk"] = magisk_path
	}

	gapps_path := ""
	if Select_gapps.Selected == "OpenGapps" {
		// PixelExperience rom has Gapps preinstalled
		if get.A1.User.Rom.Name != "PixelExperience" {
			wg.Add(1)
			go func() {
				defer wg.Done()

				gapps_filename, err := get.OpenGapps(device.D1.Arch, Select_opengapps_version.Selected, Select_opengapps_variant.Selected)
				if err != nil {
					errs <- RetrievalError{"OpenGapps", "Download link returned by API", err}
				}
				gapps_path = "flash/" + gapps_filename
			}()
		}
	} else if Select_gapps.Selected == "MicroG" {
		// Following roms already include MicroG according to https://github.com/microg/GmsCore/wiki/Signature-Spoofing (30.08.2021)
		if Chk_aurora.Checked || Chk_playstore.Checked || !helpers.IsStringInSlice(get.A1.User.Rom.Name, []string{"LineageOSMicroG", "CalyxOS", "eOS"}) {
			gapps_path = "flash/" + get.A1.Upstream.NanoDroid["MicroG"].Filename
			wg.Add(1)
			go func() {
				defer wg.Done()

				err := get.DownloadFile(gapps_path, get.A1.Upstream.NanoDroid["MicroG"].Href, get.A1.Upstream.NanoDroid["MicroG"].Checksum_url_suffix)
				if err != nil {
					errs <- RetrievalError{"NanoDroid-MicroG", get.A1.Upstream.NanoDroid["MicroG"].Href, err}
				}
			}()
		}
	}
	if gapps_path != "" {
		Files["gapps"] = gapps_path
	}

	aurora_path := ""
	// Only if we don't want MicroG but still want Aurora Store.
	// If we want the Play Store with MicroG, we already have the MicroG zip containing the Play Store
	if (Chk_aurora.Checked && Select_gapps.Selected != "MicroG") {
		aurora_path = "flash/" + get.A1.Upstream.NanoDroid["MicroG"].Filename
		wg.Add(1)
		go func() {
			defer wg.Done()

			err := get.DownloadFile(aurora_path, get.A1.Upstream.NanoDroid["MicroG"].Href, get.A1.Upstream.NanoDroid["MicroG"].Checksum_url_suffix)
			if err != nil {
				errs <- RetrievalError{"NanoDroid-MicroG", get.A1.Upstream.NanoDroid["MicroG"].Href, err}
			}
		}()
	}
	if aurora_path != "" {
		Files["aurora"] = aurora_path
	}

	fdroid_path := ""
	if Chk_fdroid.Checked {
		// LineageOSMicrog has F-Droid preinstalled
		if get.A1.User.Rom.Name != "LineageOSMicroG" {
			fdroid_path = "flash/" + get.A1.Upstream.NanoDroid["Fdroid"].Filename
			wg.Add(1)
			go func() {
				defer wg.Done()

				err := get.DownloadFile(fdroid_path, get.A1.Upstream.NanoDroid["Fdroid"].Href, get.A1.Upstream.NanoDroid["Fdroid"].Checksum_url_suffix)
				if err != nil {
					errs <- RetrievalError{"NanoDroid-Fdroid", get.A1.Upstream.NanoDroid["Fdroid"].Href, err}
				}
			}()
		}
	}
	if fdroid_path != "" {
		Files["fdroid"] = fdroid_path
	}

	patcher_path := ""
	if Chk_sigspoof.Checked {
		// Following roms have native signature spoofing according to https://github.com/microg/GmsCore/wiki/Signature-Spoofing (30.08.2021)
		if !helpers.IsStringInSlice(get.A1.User.Rom.Name, []string{"LineageOSMicroG", "CalyxOS", "e-OS", "AospExtended", "ArrowOS", "CarbonRom", "crDroid", "Omnirom", "Marshrom", "ResurrectionRemix"}) {
			patcher_path = "flash/" + get.A1.Upstream.NanoDroid["Patcher"].Filename
			wg.Add(1)
			go func() {
				defer wg.Done()

				err := get.DownloadFile(patcher_path, get.A1.Upstream.NanoDroid["Patcher"].Href, get.A1.Upstream.NanoDroid["Patcher"].Checksum_url_suffix)
				if err != nil {
					errs <- RetrievalError{"NanoDroid-Patcher", get.A1.Upstream.NanoDroid["Patcher"].Href, err}
				}
			}()
		}
	}
	if patcher_path!= "" {
		Files["patcher"] = patcher_path
	}

	gsyncswype_path := ""
	if Chk_gsync.Checked || Chk_swype.Checked {
		gsyncswype_path = "flash/" + get.A1.Upstream.NanoDroid["Google"].Filename
		wg.Add(1)
		go func() {
			defer wg.Done()

			err := get.DownloadFile(gsyncswype_path, get.A1.Upstream.NanoDroid["Google"].Href, get.A1.Upstream.NanoDroid["Google"].Checksum_url_suffix)
			if err != nil {
				errs <- RetrievalError{"Gsync/Swype", get.A1.Upstream.NanoDroid["Google"].Href, err}
			}
		}()
	}
	if gsyncswype_path != "" {
		Files["gsyncswype"] = gsyncswype_path
	}

	copypartitions_path := ""
	copypartitions_path = "flash/" + get.A1.Upstream.CopyPartitions.Filename
	wg.Add(1)
	go func() {
		defer wg.Done()

		err := get.DownloadFile(copypartitions_path, get.A1.Upstream.CopyPartitions.Href, get.A1.Upstream.CopyPartitions.Checksum_url_suffix)
		if err != nil {
			errs <- RetrievalError{"copy-partitions.zip", get.A1.Upstream.CopyPartitions.Href, err}
		}
	}()
	if copypartitions_path != "" {
		Files["copypartitions"] = copypartitions_path
	}


	go func() {
		wg.Wait()
		close(errs)
	}()

	for e := range errs {
		if e.Error != nil {
			logger.LogError("Error retrieving " + e.What + " from " + e.Href + " :", e.Error)
			return map[string]string{}, fmt.Errorf(e.What + ": " + e.Error.Error())
		}
	}

	return Files, nil
}

func createNanoDroidSetup() map[string]string {
	setup := make(map[string]string)

	// For the values, refer to the NanoDroid documentation:
    // https://gitlab.com/Nanolx/NanoDroid/-/blob/master/doc/AlterInstallation.md
	if Select_gapps.Selected == "MicroG" { setup["microg"] = "1" }
    if Select_gapps.Selected == "MicroG" { setup["mapsv1"] = "1" }
    if Chk_fdroid.Checked { setup["fdroid"] = "1" }
    if Chk_gsync.Checked { setup["gsync"] = "1" }
    if Chk_swype.Checked { setup["swipe"] = "1" }
    if Chk_playstore.Checked && Chk_aurora.Checked {
    	setup["play"] = "30"
    } else if !Chk_playstore.Checked && Chk_aurora.Checked {
    	setup["play"] = "21"
    } else if Chk_playstore.Checked && !Chk_aurora.Checked {
    	setup["play"] = "10"
    } else {
    	setup["play"] = "01"
    }

    return setup
}