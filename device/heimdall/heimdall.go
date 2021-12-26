package heimdall

import (
	"github.com/amo13/anarchy-droid/helpers"
	"github.com/amo13/anarchy-droid/logger"

	"runtime"
	"strings"
	"fmt"
)

var Sudopw string = ""
var Nosudo bool = false

func heimdall_command() string {
	switch runtime.GOOS {
	case "windows":
		return "bin\\heimdall\\heimdall.exe"
	case "darwin":
		return "bin/heimdall/heimdall"
	default:
		if Nosudo {
			return "bin/heimdall/heimdall"
		} else if Sudopw == "" {
			return "sudo bin/heimdall/heimdall"
		} else {
			return "printf " + Sudopw + " | sudo -S bin/heimdall/heimdall"
		}
	}
}

// Returns the non-empty or longer of stdout and stderr for a given fastboot command
func Cmd(args ...string) (stdout string, err error) {
	if !available() {
		return "", fmt.Errorf("disconnected")
	}

	stdout, stderr := helpers.Cmd(heimdall_command(), args...)
	if stdout != "" && stderr == "" {
		return strings.Trim(strings.Trim(stdout, "\n"), " "), nil
	} else if stdout == "" && stderr != "" {
		return strings.Trim(strings.Trim(stderr, "\n"), " "), nil
	} else if stdout != "" && stderr != "" {
		if len(stdout) >= len(stderr) {
			return strings.Trim(strings.Trim(stdout, "\n"), " "), nil
		} else {
			return strings.Trim(strings.Trim(stderr, "\n"), " "), nil
		}
	} else {
		return "", fmt.Errorf("disconnected")
	}
}

// Check for disconnection error
func unavailable(err error) bool {
	if err != nil {
		if err.Error() == "disconnected" {
			return true
		} else {
			logger.LogError("Unknown heimdall error:", err)
		}
	}

	return false
}

func available() bool {
	if State() == "connected" {
		return true
	} else {
		return false
	}
}

func State() string {
	stdout, _ := helpers.Cmd(heimdall_command(), "detect")

	if stdout == "" {
		return "disconnected"
	} else {
		return "connected"
	}
}

func FlashRecovery(img_file string, partition string) error {
	result, err := Cmd("flash", "--" + partition, img_file, "--no-reboot")
	if unavailable(err) {
		return err
	}

	// Log the result for reference
	logger.Log("------------ heimdall flash ... ------------")
	logger.Log(result)
	logger.Log("--------------------------------------------")

	if strings.Contains(strings.ToLower(result), "upload successful") {
		return nil
	} else if strings.Contains(strings.ToLower(result), "failed to access device") {
		logger.LogError("heimdall failed to access device", fmt.Errorf(result))
		return fmt.Errorf("heimdall failed to access device")
	} else if strings.Contains(strings.ToLower(result), "upload failed") {
		logger.Log("heimdall failed to flash recovery")
		return fmt.Errorf("heimdall failed to flash recovery")
	} else {
		logger.LogError("unknown heimdall response:", fmt.Errorf(result))
		return fmt.Errorf("unknown heimdall response: " + result)
	}
}

func Reboot() error {
	logger.Log("Rebooting device...")

	_, err := Cmd("print-pit")
	if unavailable(err) {
		return err
	}

	return err
}