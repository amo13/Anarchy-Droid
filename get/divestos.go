package get

import (
	"github.com/amo13/anarchy-droid/helpers"
	"github.com/amo13/anarchy-droid/logger"

	"encoding/json"
	"strings"
	"fmt"
)

type DivestosApiResponse struct {
	Filename string `json:"filename"`
	Version string `json:"version"`
	Url string `json:"url"`
}

var ParsedDivestosApiResponse DivestosApiResponse

// Download latest Divestos zip into flash folder and return the file name
func Divestos(codename string) (string, error) {
	dl_url, err := DivestosLatestAvailableHref(codename)
	if err != nil {
		return "", err
	}

	// Get the file name
	file_name, err := DivestosLatestAvailableFileName(codename)
	if err != nil {
		return "", err
	}

	err = DownloadFile("flash/" + file_name, dl_url, "")
	if err != nil {
		return "", err
	}

	return file_name, nil
}

// Returns file name of the latest available Divestos zip
func DivestosLatestAvailableFileName(codename string) (string, error) {
	if A1.Upstream.Rom["DivestOS"] != nil {
		if A1.Upstream.Rom["DivestOS"].Filename != "" {
			return A1.Upstream.Rom["DivestOS"].Filename, nil
		}
	}

	_, err := DivestosLatestAvailableHref(codename)
	if err != nil {
		return "", err
	}

	if A1.Upstream.Rom["DivestOS"].Filename == "" {
		href := A1.Upstream.Rom["DivestOS"].Href
		if href == "" {
			return "", fmt.Errorf("not found")
		} else {
			return helpers.ExtractFileNameFromHref(href), nil
		}
	}

	return A1.Upstream.Rom["DivestOS"].Filename, nil
}

func DivestosParseApiResponse(codename string) (DivestosApiResponse, error) {
	url := "https://divestos.org/updater.php?base=LineageOS&device=" + codename

	status_code, err := StatusCode(url)
	if err != nil {
		return ParsedDivestosApiResponse, err
	}
	if status_code == "404 Not Found" {
		return ParsedDivestosApiResponse, fmt.Errorf("not available")
	}

	content, err := helpers.ReadFromURL(url)
	if err != nil {
		logger.Log("DivestosParseApiResponse:", err.Error())
		return ParsedDivestosApiResponse, err
	}

	if strings.Trim(string(content), " ") == "Unknown base/device" {
		return ParsedDivestosApiResponse, fmt.Errorf("not available")
	}

	var ApiResponseMap map[string][]DivestosApiResponse

    err = json.Unmarshal([]byte(content), &ApiResponseMap)
    if err != nil {
    	logger.LogError("Unable to unmarshal response from " + url, err)
    }

    if len(ApiResponseMap["response"]) == 0 {
    	return ParsedDivestosApiResponse, fmt.Errorf("DivestOS gave an unexpected JSON response: " + string(content))
    }
	
	// If there are more than one zip available,
	// assume that the first one in the list is the latest.
	return ApiResponseMap["response"][0], nil
}


// Returns download link of the latest available Divestos zip
func DivestosLatestAvailableHref(codename string) (string, error) {
	if A1.Upstream.Rom["DivestOS"] != nil {
		if A1.Upstream.Rom["DivestOS"].Href != "" {
			return A1.Upstream.Rom["DivestOS"].Href, nil
		}
	}

	data, err := DivestosParseApiResponse(codename)
	if err != nil {
		return "", err
	}

	if data.Url == "" {
		return "", fmt.Errorf("not available")
	}

	// Populate the A1 structs of availables
	A1.Mutex.Lock()
	defer A1.Mutex.Unlock()
	A1.Upstream.Rom["DivestOS"] = &Item{}
	A1.Upstream.Rom["DivestOS"].Name = "DivestOS"
	A1.Upstream.Rom["DivestOS"].Href = data.Url
	A1.Upstream.Rom["DivestOS"].Checksum_url_suffix = ""
	A1.Upstream.Rom["DivestOS"].Filename = data.Filename
	A1.Upstream.Rom["DivestOS"].Version = data.Version
	av, err := DivestosParseAndroidVersion(A1.Upstream.Rom["DivestOS"].Version)
	if err != nil {
		return data.Url, fmt.Errorf("unable to parse DivestOS Android version with %s: %s", A1.Upstream.Rom["DivestOS"].Version, err.Error())
	}
	A1.Upstream.Rom["DivestOS"].Android_version = av

	return data.Url, nil
}

func DivestosParseAndroidVersion(romversion string) (string, error) {
	return LineageosParseAndroidVersion(romversion)
}