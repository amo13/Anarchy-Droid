package twrp

import (
	"os"
	"fmt"
	"time"
	"regexp"
	"strings"
	"io/ioutil"

	"github.com/amo13/anarchy-droid/logger"
	"github.com/amo13/anarchy-droid/helpers"
	"github.com/amo13/anarchy-droid/device/adb"
)

const Logpath = "log/"

func Cmd(args ...string) (stdout string, err error) {
	return adb.Cmd(append([]string{"shell", "twrp"}, args...)...)
}

// Check for disconnection error or suddenly unauthorized error
func unavailable(err error) bool {
	if err != nil {
		if err.Error() == "disconnected" || err.Error() == "unauthorized" {
			return true
		} else {
			logger.LogError("Unknown ADB error:", err)
		}
	}

	return false
}

func IsConnected() bool {
	return adb.State() == "recovery"
}

func VersionConnected() (string, error) {
	result, err := Cmd("version")
	if unavailable(err) {
		return "", err
	}

	re := regexp.MustCompile(`(?:(\d+\.[.\d]*\d+))`)
	return strings.Join(re.FindAllString(result, -1), ""), nil
}

func wipe(partition string) error {
	if adb.State() == "recovery" {
		_, err := adb.Cmd("shell", "twrp", "wipe", partition)
		if err != nil {
			return err
		}
	} else {
		logger.Log("Device not in recovery mode, cannot wipe partitions")
		return fmt.Errorf("Recovery not connected")
	}

	return nil
}

func WipeDirty() error {
	time.Sleep(1 * time.Second)
	err := wipe("cache")
	if err != nil {
		return err
	}

	time.Sleep(1 * time.Second)
	err = wipe("dalvik")
	if err != nil {
		return err
	}

	return nil
}

func WipeClean() error {
	logger.Log("Formating the data partition...")
	time.Sleep(1 * time.Second)
	err := FormatData()
	if err != nil {
		return err
	}

	logger.Log("Wiping the device caches...")
	time.Sleep(1 * time.Second)
	err = WipeDirty()
	if err != nil {
		return err
	}

	logger.Log("Wiping the data partition...")
	time.Sleep(1 * time.Second)
	err = wipe("data")
	if err != nil {
		return err
	}

	return nil
}

func FormatData() error {
	err := UnmountData()
	if err != nil {
		return err
	}

	if adb.State() == "recovery" {
		err = formatDataORS()
		if err != nil {
			err = formatDataOldschool()
			if err != nil {
				return err
			}
		}
	}

	err = MountData()
	if err != nil {
		return err
	}

	return nil
}

func formatDataORS() error {
	if adb.State() == "recovery" {
		stdout, err := adb.Cmd("shell", "twrp", "format", "data")
		if err != nil {
			return err
		}

		if strings.Contains(stdout, "Unrecognized script command") {
			return fmt.Errorf("Unrecognized script command: adb shell twrp format data")
		} else {
			return nil
		}
	} else {
		logger.Log("Device not in recovery mode, cannot format data")
		return fmt.Errorf("Recovery not connected")
	}

	return nil
}

func formatDataOldschool() error {
	data_path_candidates, err := findDataPartitionPathCandidates()
	if err != nil {
		return err
	}

	data_fs_candidates, err := findDataPartitionFilesystemCandidates()
	if err != nil {
		return err
	}

	for _, data_fs := range data_fs_candidates {
		for _, data_path := range data_path_candidates {
			logger.Log("Attempting to format", data_path, "as", data_fs)
			if data_fs == "f2fs" {
				r, err := adb.Cmd("shell", "mkfs.f2fs", "-t", "0", data_path)
				if err != nil {
					logger.Log("Format error:", err.Error())
				}

				if strings.Contains(strings.ToLower(r), "format successful") {
					return nil
				} else {
					logger.Log("Did not seem to work:\n", r)
				}
			} else if data_fs == "ext4" {
				r, err := adb.Cmd("shell", "make_ext4fs ", data_path)
				if err != nil {
					logger.Log("Format error:", err.Error())
				}

				if strings.Contains(strings.ToLower(r), "created filesystem") {
					return nil
				} else {
					logger.Log("Did not seem to work:\n", r)
				}
			} else if data_fs == "" {
				return fmt.Errorf("Unknown data partition filesystem")
			} else {
				return fmt.Errorf("Unable to format data to " + data_fs)
			}
		}
	}

	return fmt.Errorf("Failed to format data")
}

// find paths under which the partition mounted as /data is accessible
func findDataPartitionPathCandidates() ([]string, error) {
	if adb.State() != "recovery" {
		logger.Log("Device not in recovery mode, cannot open sideload")
		return []string{}, fmt.Errorf("Recovery not connected")
	}

	candidates := []string{}

	// first possibility
	r, err := adb.Cmd("shell", "cat", "/etc/fstab")
	if err != nil {
		return []string{}, err
	}
	for _, line := range helpers.StringToLinesSlice(r) {
		if strings.Contains(line, "/data") && len(strings.Split(line, " ")) > 0 {
			candidates = append(candidates, strings.Split(line, " ")[0])
			break
		}
	}

	// second possibility
	r, err = adb.Cmd("shell", "cat", "/etc/recovery.fstab")
	if err != nil {
		return []string{}, err
	}
	for _, line := range helpers.StringToLinesSlice(r) {
		if strings.Contains(line, "/data") && len(strings.Split(line, " ")) > 2 {
			candidates = append(candidates, strings.Split(line, " ")[2])
			break
		}
	}

	// third possibility
	log, err := GetAndReadLog()
	if err != nil {
		return []string{}, err
	}
	loglines := helpers.StringToLinesSlice(log)
	for _, line := range loglines {
		if strings.HasPrefix(line, "/data | /dev") {
			for _, line_part := range strings.Split(line, " ") {
				if strings.Contains(line_part, "/dev/") {
					candidates = append(candidates, line_part)
					break
				}
			}
			break
		}
	}

	return helpers.UniqueNonEmptyElementsOfSlice(candidates), nil
}

func findDataPartitionFilesystemCandidates() ([]string, error) {
	candidates := []string{}

	if adb.State() != "recovery" {
		logger.Log("Device not in recovery mode, cannot open sideload")
		return candidates, fmt.Errorf("Recovery not connected")
	}

	// first candidate
	r, err := adb.Cmd("shell", "cat", "/etc/fstab")
	if err != nil {
		return candidates, err
	}
	for _, line := range helpers.StringToLinesSlice(r) {
		if strings.Contains(line, "/data") {
			if len(strings.Split(line, " ")) > 2 {
				candidates = append(candidates, strings.Split(line, " ")[2])
			}
		}
	}

	// second candidate
	r, err = adb.Cmd("shell", "cat", "/etc/recovery.fstab")
	if err != nil {
		return candidates, err
	}
	for _, line := range helpers.StringToLinesSlice(r) {
		if strings.Contains(line, "/data") && len(strings.Split(line, " ")) > 1 {
			candidates = append(candidates, strings.Split(line, " ")[1])
		}
	}

	if len(candidates) > 0 {
		return candidates, nil
	} else {
		return candidates, fmt.Errorf("Unable to determine the data partition filesystem")
	}
}

func OpenSideload() error {
	if adb.State() == "recovery" {
		_, err := adb.Cmd("shell", "twrp", "sideload")
		if err != nil {
			return err
		}
	} else {
		logger.Log("Device not in recovery mode, cannot open sideload")
		return fmt.Errorf("Recovery not connected")
	}

	return nil
}

func Sideload(file_path string) error {
	_, err := os.Stat(file_path)
	if os.IsNotExist(err) {
		return err
	}

	if adb.State() == "sideload" {
		_, err = adb.Cmd("sideload", file_path)
		if err != nil {
			return err
		}
	} else {
		logger.Log("Device not in sideload mode, cannot sideload", helpers.ExtractFileNameFromHref(file_path))
		return fmt.Errorf("Sideload not connected")
	}

	return nil
}

// True if recovery.log is retrievable and contains "Set page: 'main"
func IsReady() (bool, error) {
	log, err := GetAndReadLog()
	if err != nil {
		return false, err
	}

	loglines := helpers.StringToLinesSlice(log)
	for _, line := range loglines {
		if strings.Contains(line, "Set page: 'main") {
			return true, nil
		}
	}

	return false, nil
}

func SendNanodroidSetup(setup map[string]string) error {
	file, err := os.OpenFile(".nanodroid-setup", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
    if err != nil {
        return err
    }

    // Use the following defaults for values not provided
    // For the values, refer to the NanoDroid documentation:
    // https://gitlab.com/Nanolx/NanoDroid/-/blob/master/doc/AlterInstallation.md
    if setup["microg"] == "" { setup["microg"] = "1" }
    if setup["fdroid"] == "" { setup["fdroid"] = "1" }
    if setup["apps"] == "" { setup["apps"] = "0" }
    if setup["play"] == "" { setup["play"] = "21" }
    if setup["overlay"] == "" { setup["overlay"] = "0" }
    if setup["zelda"] == "" { setup["zelda"] = "0" }
    if setup["mapsv1"] == "" { setup["mapsv1"] = "1" }
    if setup["init"] == "" { setup["init"] = "0" }
    if setup["gsync"] == "" { setup["gsync"] = "0" }
    if setup["swipe"] == "" { setup["swipe"] = "0" }
    if setup["forcesystem"] == "" { setup["forcesystem"] = "1" }
    if setup["nlpbackend"] == "" { setup["nlpbackend"] = "1100" }
    if setup["nano"] == "" { setup["nano"] = "1" }
    if setup["bash"] == "" { setup["bash"] = "1" }
    if setup["utils"] == "" { setup["utils"] = "1" }
    if setup["fonts"] == "" { setup["fonts"] = "0" }

    // Write the file
    file.WriteString("nanodroid_microg=" + setup["microg"] + "\n")
    file.WriteString("nanodroid_fdroid=" + setup["fdroid"] + "\n")
    file.WriteString("nanodroid_apps=" + setup["apps"] + "\n")
    file.WriteString("nanodroid_play=" + setup["play"] + "\n")
    file.WriteString("nanodroid_overlay=" + setup["overlay"] + "\n")
    file.WriteString("nanodroid_zelda=" + setup["zelda"] + "\n")
    file.WriteString("nanodroid_mapsv1=" + setup["mapsv1"] + "\n")
    file.WriteString("nanodroid_init=" + setup["init"] + "\n")
    file.WriteString("nanodroid_gsync=" + setup["gsync"] + "\n")
    file.WriteString("nanodroid_swipe=" + setup["swipe"] + "\n")
    file.WriteString("nanodroid_forcesystem=" + setup["forcesystem"] + "\n")
    file.WriteString("nanodroid_nlpbackend=" + setup["nlpbackend"] + "\n")
    file.WriteString("nanodroid_nano=" + setup["nano"] + "\n")
    file.WriteString("nanodroid_bash=" + setup["bash"] + "\n")
    file.WriteString("nanodroid_utils=" + setup["utils"] + "\n")
    file.WriteString("nanodroid_fonts=" + setup["fonts"] + "\n")

    err = file.Close()
    if err != nil {
        return err
    }

    err = adb.Push(".nanodroid-setup", "/data/media/0/")
    if err != nil {
        return err
    }

    stdout, err := adb.Cmd("shell", "ls", "/data/media/0/.nanodroid-setup")
    if err != nil {
        logger.LogError("Error checking if .nanodroid-setup was sent successfully:", err)
        return err
    }
    if strings.Contains(stdout, "No such file or directory") {
    	logger.Log("Failed to send .nanodroid-setup!")
    }

    return nil
}

func getLog() (string, error) {
	err := adb.Pull("/tmp/recovery.log", Logpath)
	if err != nil {
		return "", err
	}
	_, err = os.Stat(Logpath + "recovery.log")
	if os.IsNotExist(err) {
		return "", err
	}

	return Logpath + "recovery.log", nil
}

func ReadLog() (string, error) {
	_, err := os.Stat(Logpath + "recovery.log")
	if os.IsNotExist(err) {
		return "", err
	}

	content, err := ioutil.ReadFile(Logpath + "recovery.log")
    if err != nil {
        return "", err
    }

    return string(content), nil
}

func GetAndReadLog() (string, error) {
	_, err := getLog()
	if err != nil {
		return "", err
	}

	return ReadLog()
}

func IsDataMounted() (bool, error) {
	mounts, err := adb.Cmd("shell", "cat", "/proc/mounts")
	if unavailable(err) {
		return false, err
	}

	lines := helpers.StringToLinesSlice(mounts)
	filter := func(s string) bool { return strings.Contains(s, "/data") }
	matches := helpers.FilterStringSlice(lines, filter)

	if len(matches) > 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func MountData() error {
	mounted, err := IsDataMounted()
	if unavailable(err) {
		return err
	}

	if !mounted {
		_, err := adb.Cmd("shell", "mount", "/data")
		if unavailable(err) {
			return err
		}
	}

	return nil
}

func UnmountData() error {
	mounted, err := IsDataMounted()
	if unavailable(err) {
		return err
	}

	if mounted {
		_, err := adb.Cmd("shell", "umount", "/data")
		if unavailable(err) {
			return err
		}
	}

	return nil
}

func IsDataMountable() (bool, error) {
	mounted, err := IsDataMounted()
	if unavailable(err) {
		return false, err
	}

	if mounted {
		return true, nil
	} else {
		err = MountData()
		if unavailable(err) {
			return false, err
		}

		mounted, err = IsDataMounted()
		if unavailable(err) {
			return false, err
		}

		if mounted {
			err = MountData()
			if unavailable(err) {
				return false, err
			}

			return true, nil
		} else {
			return false, nil
		}
	}
}

// Partition size does not report 0 MB
func IsDataUsable() (bool, error) {
	log, err := GetAndReadLog()
	if err != nil {
		return false, err
	}

	loglines := helpers.StringToLinesSlice(log)
	filter := func(s string) bool { return strings.HasPrefix(s, "/data | /dev") }
	matches := helpers.FilterStringSlice(loglines, filter)
	if len(matches) > 0 {
		// Do not confuse "backup size: 0mb" in the same line! Therefore "| size: 0mb"
		if strings.Contains(matches[0], "| size: 0mb") {
			return false, nil
		} else {
			return true, nil
		}
	} else {
		mountable, err := IsDataMountable()
		if unavailable(err) {
			return false, err
		}

		return mountable, nil
	}
}

func WasLastSideloadSuccesful() (bool, error) {
	// Count the number of lines in the recovery.log before pulling a file anew
	log, err := ReadLog()
	if err != nil {
		if !os.IsNotExist(err) {
			log = ""
		} else {
			return false, err
		}
	}

	loglines := helpers.StringToLinesSlice(log)
	last_lines_count := len(loglines)

	log, err = GetAndReadLog()
	if err != nil {
		return false, err
	}

	new_lines_count := len(loglines)
	if last_lines_count >= new_lines_count {
		last_lines_count = 0
	}

	// reverse loglines order
	for i, j := 0, len(loglines)-1; i < j; i, j = i+1, j-1 {
        loglines[i], loglines[j] = loglines[j], loglines[i]
    }

    last := ""

    // Find the last occurence of "Updater process ended with"
	for _, line := range loglines[last_lines_count:] {
		if strings.Contains(line, "Updater process ended with") {
			re := regexp.MustCompile(`Updater process ended with (.*)`)
			last = re.FindAllStringSubmatch(line, -1)[0][1]
			break
		}
	}

	switch last {
	case "RC=0":
		return true, nil
	case "ERROR: 1":
		return false, nil
	case "":
		return false, fmt.Errorf("Unable to parse last sideload success")
	default:
		return false, fmt.Errorf("Unknown sideload result in the TWRP log")
	}
}

func IsNanodroidMissingSpace() (bool, error) {
	log, err := GetAndReadLog()
	if err != nil {
		return false, err
	}

	loglines := helpers.StringToLinesSlice(log)

	// I know, this is not very reliable, but I had no better idea
	// Improvements are welcome!
	if len(loglines) >= 100 {
		for _, line := range loglines[len(loglines):] {
			if strings.Contains(line, "Less than 512 MB free space availabe from TWRP") ||
			strings.Contains(line, "No space left on device") ||
			strings.Contains(line, "not enough space available!") ||
			strings.Contains(line, "unzip: failed to extract /dev/tmp/") {
				return true, nil
			}
		}
	}

	return false, nil
}

func RomHasNativeSigspoof() (bool, error) {
	log, err := GetAndReadLog()
	if err != nil {
		return false, err
	}

	loglines := helpers.StringToLinesSlice(log)

	// reverse loglines order
	for i, j := 0, len(loglines)-1; i < j; i, j = i+1, j-1 {
        loglines[i], loglines[j] = loglines[j], loglines[i]
    }

    // Search for the last occurence of "Framework Patcher"
    last_index := 0
    for i, line := range loglines {
    	if strings.Contains(line, "Framework Patcher") {
    		last_index = i
    		break
    	}
    }

    if last_index != 0 {
	    for _, line := range loglines[:last_index] {
	    	if strings.Contains(line, "ROM has native signature spoofing already") {
	    		return true, nil
	    	}
	    }
    } else {
    	logger.Log("Unable to find the last logs of the Framework Patcher")
    }

    return false, nil
}