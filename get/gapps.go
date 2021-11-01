package get

import (
	"github.com/amo13/anarchy-droid/helpers"
	"github.com/amo13/anarchy-droid/logger"

	"encoding/json"
	"net/http"
	"strings"
	"fmt"
)

// Download latest OpenGapps zip into flash folder and return the file name
func OpenGapps(arch string, android_version string, variant string) (string, error) {
	dl_url, err := OpenGappsLatestAvailableHref(arch, android_version, variant)
	if err != nil {
		return "", err
	}

	// Parse the file name from the href
	file_name := helpers.ExtractFileNameFromHref(dl_url)

	err = DownloadFile("flash/" + file_name, dl_url, "")
	if err != nil {
		return "", err
	}

	return file_name, nil
}

// Returns file name of the latest available OpenGapps zip
func OpenGappsLatestAvailableFileName(arch string, android_version string, variant string) (string, error) {
	href, err := OpenGappsLatestAvailableHref(arch, android_version, variant)
	if err != nil {
		return "", err
	}

	return helpers.ExtractFileNameFromHref(href), nil
}

// Structs for parsing the OpenGapps JSON
type OpenGappsJson struct {
	Arch map[string]Arch `json:"archs"`
	Ready bool
}

type Arch struct {
	Api map[string]Api `json:"apis"`
	Date string `json:"date"`
	Human_date string `json:"human_date"`
}

type Api struct {
	Variants []Variant `json:"variants"`
}

type Variant struct {
	Name string `json:"name"`
	Zip string `json:"zip"`
	Zip_size string `json:"zip_size"`
	Md5 string `json:"md5"`
	Version_info string `json:"version_info"`
	Source_report string `json:"source_report"`
}

var ParsedOpenGappsJson OpenGappsJson

func OpenGappsParseJson() (OpenGappsJson, error) {
	if !ParsedOpenGappsJson.Ready {
		content, err := helpers.ReadFromURL("https://api.opengapps.org/list")
		if err != nil {
			logger.LogError("Unable to read list from OpenGapps API URL:", err)
		}

	    json.Unmarshal([]byte(content), &ParsedOpenGappsJson)
	    ParsedOpenGappsJson.Ready = true
	}

    return ParsedOpenGappsJson, nil
}


func OpenGappsAvailableForAndroidVersions(arch string) ([]string, error) {
	data, err := OpenGappsParseJson()
	if err != nil {
		return []string{}, err
	}

	versions := []string{}

	for version, _ := range data.Arch[arch].Api {
		versions = append(versions, version)
	}

	return versions, nil
}

func OpenGappsAvailableVariants(arch string, android_version string) ([]string, error) {
	data, err := OpenGappsParseJson()
	if err != nil {
		return []string{}, err
	}

	variants := []string{}

	for _, variant := range data.Arch[arch].Api[android_version].Variants {
		variants = append(variants, variant.Name)
	}

	return variants, nil
}

func OpenGappsIndexOfVariant(arch string, android_version string, variant_wish string) (int, error) {
	data, err := OpenGappsParseJson()
	if err != nil {
		return 9999, err
	}

	index := 9999
	for i, variant := range data.Arch[arch].Api[android_version].Variants {
		if variant.Name == variant_wish {
			index = i
			break
		}
	}

	if index == 9999 {
		return 0, fmt.Errorf("variant %s does not exist for arch %s and android version %s", variant_wish, arch, android_version)
	} else {
		return index, nil
	}
}

// Returns download link of the latest available OpenGapps zip
// for the given architecture and Android version
// arch must be one of "arm", "arm64", "x86" and "x86_64"
// android_version must be in the form of "X.Y"
func OpenGappsLatestAvailableHref(arch string, android_version string, variant string) (string, error) {
	variant_index, err := OpenGappsIndexOfVariant(arch, android_version, variant)
	if err != nil {
		return "", err
	}

	data, err := OpenGappsParseJson()
	if err != nil {
		return "", err
	}

	sourceforge_link := data.Arch[arch].Api[android_version].Variants[variant_index].Zip

	if sourceforge_link == "" {
		logger.Log("Unable to traverse OpenGapps JSON to Sourceforge download link.")
		return "", fmt.Errorf("unable to parse opengapps json")
	}

	var dl_url string

	// Sometimes a sourceforge mirror is irresponsive. Retry in that case up to 3 times.
	for tries := 1; tries < 5; tries++ {
		// Follow redirects to the actual file URL from a mirror
		resp, err := http.Head(sourceforge_link)
		if err != nil {
			logger.Log("Retrying to follow OpenGapps Sourceforge redirects to download mirror...")
		} else {
			dl_url = resp.Request.URL.String()
			break
		}
	}

	if dl_url == "" || (!strings.HasSuffix(dl_url, ".zip") && !strings.Contains(dl_url, ".zip?")) {
		logger.Log("Unable to follow OpenGapps sourceforge redirects to mirror")
		return "", fmt.Errorf("unable to follow sourceforge redirects to mirror")
	}

	return dl_url, nil
}

func OpenGappsPopulateAvailablesStruct() error {
	var slver []string
	var slvar []string
	var variants []string
	var err error

	A1.Mutex.Lock()
	defer A1.Mutex.Unlock()
	for _, arch := range []string{"arm", "arm64", "x86", "x86_64"} {
		slver, err = OpenGappsAvailableForAndroidVersions(arch)
		if err != nil {
			return err
		}

		for _, android_version := range slver {
			slvar, err = OpenGappsAvailableVariants(arch, android_version)
			if err != nil {
				return err
			}

			variants = []string{}

			for _, variant := range slvar {
				variants = append(variants, variant)
			}

			if A1.Upstream.OpenGapps[arch] == nil {
				A1.Upstream.OpenGapps[arch] = make(map[string][]string)
			}

			A1.Upstream.OpenGapps[arch][android_version] = variants
		}
	}

	return nil
}