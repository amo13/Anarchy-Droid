package get

import (
	"io"
	"os"
	"fmt"
	"strings"
	"net/http"
	"path/filepath"

	"github.com/amo13/anarchy-droid/logger"
	"github.com/amo13/anarchy-droid/helpers"

	"github.com/codingsince1985/checksum"
)

func StatusCode(url string) (code string, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	return resp.Status, nil
}

// Assume we have the correct file if it already exists
func DownloadFile(file_path string, url string, checksum_url_suffix string) (err error) {
	_, err = os.Stat(file_path)
	if os.IsNotExist(err) {
		return DownloadAndOverwriteFile(file_path, url, checksum_url_suffix)
	} else {
		if err != nil {
			return err
		} else {
			return nil
		}
	}
}

// Does not redownload if the checksum (still) matches with upstream
func DownloadAndOverwriteFile(file_path string, url string, checksum_url_suffix string) (err error) {
	// Create parent dir
	err = os.Mkdir(filepath.Dir(file_path), 0755)
	if err != nil {
		if os.IsExist(err) {
			// ignore and overwrite
		} else {
			return err
		}
	}

	// Don't redownload and overwrite if file exists
	// and the checksum (still) matches with upstream
	_, err = os.Stat(file_path)
	if err == nil && checksum_url_suffix != "" && !strings.HasSuffix(url, checksum_url_suffix) {
		integrity_passed, err := VerifyIntegrity(file_path, url, checksum_url_suffix)
		if err != nil {
			return err
		}
		if integrity_passed == true {
			return nil
		} else {
			os.Remove(file_path)
		}
	}

	// Create the file
	out, err := os.Create(file_path)
	if err != nil  {
		return err
	}
	defer out.Close()

	// Create the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	// Set referer to same url (dl.twrp.me requirement)
	req.Header.Set("Referer", url)
	client := &http.Client{}

	// Get the data
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil  {
		return err
	}

	// Try to verify integrity
	if checksum_url_suffix == "" || strings.HasSuffix(url, checksum_url_suffix) {
		// Skip if no suffix provided and prevent infinite recursion
	} else {
		// Skip integrity check if the checksum file is not downloadable
		// No error is returned in that case and we simply assume the file is fine
		status_code, err := StatusCode(url + checksum_url_suffix)
		if err != nil {
			return err
		}
		if status_code != "200 OK" {
			logger.Log("Could not verify the integrity of", file_path, "because", url + checksum_url_suffix, "returns status code", status_code)
			return nil
		}

		integrity_passed, err := VerifyIntegrity(file_path, url, checksum_url_suffix)
		if err != nil {
			return err
		}
		if integrity_passed == false {
			return fmt.Errorf("Integrity verification failed for %s", file_path)
		}
	}

	return nil
}

// Return false only on explicit verification failure
// Return true if the checksum file could not be downloaded
func VerifyIntegrity(file_path string, url string, suffix string) (isCorrect bool, err error) {
	// Try to download a checksum file
	err = DownloadFile(file_path + ".checksum", url + suffix, suffix)
	if err != nil {
		os.Remove(file_path)
		os.Remove(file_path + ".checksum")
		// If unable to download the checksum file (not available?), assume verified
		return true, err
	}

	// Compute checksum of the file to be verified
	cs := ""
	switch suffix {
	case ".md5", "?md5", ".md5sum", "?md5sum": 
		cs, err = checksum.MD5sum(file_path)
		if err != nil {
			os.Remove(file_path)
			os.Remove(file_path + ".checksum")
			return false, err
		}
	case ".sha256", "?sha256", ".sha256sum", "?sha256sum": 
		cs, err = checksum.SHA256sum(file_path)
		if err != nil {
			os.Remove(file_path)
			os.Remove(file_path + ".checksum")
			return false, err
		}
	default: return false, fmt.Errorf("Cannot compute checksum for %s: not implemented", suffix)
	}

	// Read checksum from the downloaded checksum file
	cs_upstream, err := helpers.FirstWordInFile(file_path + ".checksum")
	if err != nil {
		os.Remove(file_path)
		os.Remove(file_path + ".checksum")
		return false, err
	}

	if cs != cs_upstream {
		os.Remove(file_path)
		os.Remove(file_path + ".checksum")
		return false, fmt.Errorf("Checksum verification failed")
	}

	os.Remove(file_path + ".checksum")

	return true, nil
}