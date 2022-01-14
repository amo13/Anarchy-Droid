package get

import (
	"github.com/amo13/anarchy-droid/helpers"
	"github.com/amo13/anarchy-droid/logger"

	"encoding/json"
	"strings"
	"strconv"
	"fmt"
)

type AospExtendedApiResponse struct {
	Error bool `json:"error"`
	Filename string `json:"filename"`
	Md5 string `json:"md5"`
	Url string `json:"url"`
}

var ParsedAospExtendedApiResponse AospExtendedApiResponse

// Download latest AospExtended zip into flash folder and return the file name
func AospExtended(codename string) (string, error) {
	dl_url, err := AospExtendedLatestAvailableHref(codename)
	if err != nil {
		return "", err
	}

	// Parse the file name from the href
	file_name := helpers.ExtractFileNameFromHref(dl_url)

	err = DownloadFile("flash/" + file_name, dl_url, ".sha256sum")
	if err != nil {
		return "", err
	}

	return file_name, nil
}

// Returns file name of the latest available AospExtended zip
func AospExtendedLatestAvailableFileName(codename string) (string, error) {
	href, err := AospExtendedLatestAvailableHref(codename)
	if err != nil {
		return "", err
	}

	return helpers.ExtractFileNameFromHref(href), nil
}

func AospExtendedParseApiResponse(codename string) (AospExtendedApiResponse, error) {
	for _, version := range []string{"r", "q"} {
		content, err := helpers.ReadFromURL("https://api.aospextended.com/ota_v2/" + codename + "/" + version)
		if err != nil {
			logger.Log(err.Error())
			continue
		}

	    json.Unmarshal([]byte(content), &ParsedAospExtendedApiResponse)
	    if !ParsedAospExtendedApiResponse.Error {
	    	break
	    }
	}
	
	return ParsedAospExtendedApiResponse, nil
}


// Returns download link of the latest available AospExtended zip
func AospExtendedLatestAvailableHref(codename string) (string, error) {
	if A1.Upstream.Rom["AospExtended"] != nil {
		if A1.Upstream.Rom["AospExtended"].Href != "" {
			return A1.Upstream.Rom["AospExtended"].Href, nil
		}
	}

	data, err := AospExtendedParseApiResponse(codename)
	if err != nil {
		return "", err
	}

	if data.Error || data.Url == "" {
		return "", fmt.Errorf("not available")
	}

	// Populate the A1 structs of availables
	A1.Mutex.Lock()
	defer A1.Mutex.Unlock()
	A1.Upstream.Rom["AospExtended"] = &Item{}
	A1.Upstream.Rom["AospExtended"].Name = "AospExtended"
	A1.Upstream.Rom["AospExtended"].Href = data.Url
	A1.Upstream.Rom["AospExtended"].Checksum_url_suffix = ""
	A1.Upstream.Rom["AospExtended"].Filename = data.Filename
	A1.Upstream.Rom["AospExtended"].Version = helpers.GenericParseVersion(A1.Upstream.Rom["AospExtended"].Filename)
	av, err := AospExtendedParseAndroidVersion(A1.Upstream.Rom["AospExtended"].Version)
	if err != nil {
		return data.Url, fmt.Errorf("unable to parse AospExtended Android version with %s: %s", A1.Upstream.Rom["AospExtended"].Version, err.Error())
	}
	A1.Upstream.Rom["AospExtended"].Android_version = av

	return data.Url, nil
}

func AospExtendedParseAndroidVersion(romversion string) (string, error) {
	if romversion != "" {
		maj := strings.Split(romversion, ".")[0]
		maj_int, err := strconv.Atoi(maj)
		if err != nil {
			return "", err
		}

		v := strconv.Itoa(maj_int + 3)

		if v == "8" {
			v = "8.1"
		} else if v == "7" {
			v = "7.1"
		}

		return v, nil
	} else {
		return "", fmt.Errorf("Unable to tell android version of AospExtended: %s", romversion)
	}
}