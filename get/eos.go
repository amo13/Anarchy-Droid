package get

import (
	"github.com/gocolly/colly"

	"github.com/amo13/anarchy-droid/helpers"
	"github.com/amo13/anarchy-droid/logger"

	"strings"
	"sort"
	"fmt"
)

// Download latest EOS zip into flash folder and return the file name
func EOS(codename string) (string, error) {
	dl_url, err := EOSLatestAvailableHref(codename)
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

// Returns file name of the latest available EOS zip
func EOSLatestAvailableFileName(codename string) (string, error) {
	href, err := EOSLatestAvailableHref(codename)
	if err != nil {
		return "", err
	}

	return helpers.ExtractFileNameFromHref(href), nil
}

// Returns download link of the latest available EOS zip
func EOSLatestAvailableHref(codename string) (string, error) {
	if A1.Upstream.Rom["e-OS"] != nil {
		if A1.Upstream.Rom["e-OS"].Href != "" {
			return A1.Upstream.Rom["e-OS"].Href, nil
		}
	}

	device_url := "https://images.ecloud.global/dev/" + codename + "/"

	status_code, err := StatusCode(device_url)
	if err != nil {
		return "", err
	}
	if status_code != "200 OK" {
		return "", fmt.Errorf("not available")
	}

	versions_available := make([]string, 0)

	c := colly.NewCollector()

	c.OnError(func(_ *colly.Response, err error) {
		logger.LogError("EOS:", err)
	})

	c.OnHTML("a", func(e *colly.HTMLElement) {
		versions_available = append(versions_available, e.Attr("href"))
	})

	c.Visit(device_url)

	filter := func(s string) bool { return strings.HasSuffix(s, ".zip") }
	versions_available_filtered := helpers.FilterStringSlice(versions_available, filter)

	sort.Strings(versions_available_filtered)

	latest_available := ""
	if len(versions_available_filtered) > 0 {
		latest_available = versions_available_filtered[len(versions_available_filtered)-1]
	} else {
		return "", nil
	}

	// Populate the A1 structs of availables
	A1.Mutex.Lock()
	defer A1.Mutex.Unlock()
	A1.Upstream.Rom["e-OS"] = &Item{}
	A1.Upstream.Rom["e-OS"].Name = "e-OS"
	A1.Upstream.Rom["e-OS"].Href = device_url + latest_available
	A1.Upstream.Rom["e-OS"].Checksum_url_suffix = ".sha256sum"
	A1.Upstream.Rom["e-OS"].Filename = helpers.ExtractFileNameFromHref(latest_available)
	v, err := EOSParseVersion(A1.Upstream.Rom["e-OS"].Filename)
	if err != nil {
		return latest_available, fmt.Errorf("unable to parse e-OS version in %s", A1.Upstream.Rom["e-OS"].Filename)
	}
	A1.Upstream.Rom["e-OS"].Version = v
	av, err := EOSParseAndroidVersion(v)
	if err != nil {
		return latest_available, fmt.Errorf("unable to parse e-OS Android version with %s: %s", v, err.Error())
	}
	A1.Upstream.Rom["e-OS"].Android_version = av

	return latest_available, nil
}

func EOSParseVersion(filename string) (string, error) {
	parts := strings.Split(filename, "-")
	if len(parts) >= 3 {
		return parts[2], nil
	} else {
		return "", fmt.Errorf("unable to parse e-OS version in", filename)
	}
}

func EOSParseAndroidVersion(romversion string) (string, error) {
	if romversion != "" {
		switch strings.ToLower(romversion) {
		case "n", "nougat":
			return "7.1", nil
		case "o", "oreo":
			return "8.1", nil
		case "p", "pie":
			return "9", nil
		case "q":
			return "10", nil
		case "r":
			return "11", nil
		default:
			return "", fmt.Errorf("Unknown e-OS android version: %s", romversion)
		}
	} else {
		return "", fmt.Errorf("Unable to tell android version of e-OS: %s", romversion)
	}
}