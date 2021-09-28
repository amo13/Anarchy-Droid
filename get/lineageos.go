package get

import (
	"github.com/gocolly/colly"
	"anarchy-droid/helpers"
	"anarchy-droid/logger"
	"strings"
	"sort"
	"fmt"
)

// Download latest LineageOS zip into flash folder and return the file name
func Lineageos(codename string) (string, error) {
	dl_url, err := LineageosLatestAvailableHref(codename)
	if err != nil {
		return "", err
	}

	// Parse the file name from the href
	file_name := helpers.ExtractFileNameFromHref(dl_url)

	err = DownloadFile("flash/" + file_name, dl_url, "?sha256")
	if err != nil {
		return "", err
	}

	return file_name, nil
}

// Returns file name of the latest available LineageOS zip
func LineageosLatestAvailableFileName(codename string) (string, error) {
	href, err := LineageosLatestAvailableHref(codename)
	if err != nil {
		return "", err
	}

	return helpers.ExtractFileNameFromHref(href), nil
}

// Returns download link of the latest available LineageOS zip
func LineageosLatestAvailableHref(codename string) (string, error) {
	if A1.Upstream.Rom["LineageOS"] != nil {
		if A1.Upstream.Rom["LineageOS"].Href != "" {
			return A1.Upstream.Rom["LineageOS"].Href, nil
		}
	}

	url := "https://download.lineageos.org/" + codename

	status_code, err := StatusCode(url)
	if err != nil {
		return "", err
	}
	if status_code == "404 Not Found" {
		return "", fmt.Errorf("not available")
	}

	versions_available := make([]string, 0)

	c := colly.NewCollector()

	c.OnError(func(_ *colly.Response, err error) {
		logger.LogError("LineageOS:", err)
	})

	c.OnHTML("td a", func(e *colly.HTMLElement) {
		versions_available = append(versions_available, e.Attr("href"))
	})

	c.Visit(url)

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
	A1.Upstream.Rom["LineageOS"] = &Item{}
	A1.Upstream.Rom["LineageOS"].Name = "LineageOS"
	A1.Upstream.Rom["LineageOS"].Href = latest_available
	A1.Upstream.Rom["LineageOS"].Checksum_url_suffix = "?sha256"
	A1.Upstream.Rom["LineageOS"].Filename = helpers.ExtractFileNameFromHref(latest_available)
	v, err := LineageosParseVersion(A1.Upstream.Rom["LineageOS"].Filename)
	if err != nil {
		return latest_available, fmt.Errorf("unable to parse LineageOS version in %s", A1.Upstream.Rom["LineageOS"].Filename)
	}
	A1.Upstream.Rom["LineageOS"].Version = v
	av, err := LineageosParseAndroidVersion(v)
	if err != nil {
		return latest_available, fmt.Errorf("unable to parse LineageOS Android version with %s", v)
	}
	A1.Upstream.Rom["LineageOS"].Android_version = av

	return latest_available, nil
}

func LineageosParseVersion(filename string) (string, error) {
	parts := strings.Split(filename, "-")
	if len(parts) >= 2 {
		return parts[1], nil
	} else {
		return "", fmt.Errorf("unable to parse version")
	}
}

func LineageosParseAndroidVersion(losversion string) (string, error) {
	switch losversion {
	case "13", "13.0":
		return "6", nil
	case "14.1":
		return "7", nil
	case "15", "15.0":
		return "8", nil
	case "15.1":
		return "8.1", nil
	case "16", "16.0":
		return "9", nil
	case "17.1":
		return "10", nil
	case "18.1":
		return "11", nil
	default:
		return "", fmt.Errorf("Unable to tell android version of LineageOS %s", losversion)
	}
}

// Download latest LineageOS for MicroG zip into flash folder and return the file name
func LineageosMicrog(codename string) (string, error) {
	dl_url, err := LineageosMicrogLatestAvailableHref(codename)
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

// Returns file name of the latest available LineageOS for MicroG zip
func LineageosMicrogLatestAvailableFileName(codename string) (string, error) {
	href, err := LineageosMicrogLatestAvailableHref(codename)
	if err != nil {
		return "", err
	}

	return helpers.ExtractFileNameFromHref(href), nil
}

// Returns download link of the latest available LineageOS for MicroG zip
func LineageosMicrogLatestAvailableHref(codename string) (string, error) {
	if A1.Upstream.Rom["LineageOSMicroG"] != nil {
		if A1.Upstream.Rom["LineageOSMicroG"].Href != "" {
			return A1.Upstream.Rom["LineageOSMicroG"].Href, nil
		}
	}

	base_url := "https://download.lineage.microg.org"
	url := base_url + "/" + codename

	status_code, err := StatusCode(url)
	if err != nil {
		return "", err
	}
	if status_code == "404 Not Found" {
		return "", fmt.Errorf("not available")
	}

	versions_available := make([]string, 0)

	c := colly.NewCollector()

	c.OnError(func(_ *colly.Response, err error) {
		logger.LogError("LineageOSMicroG", err)
	})

	c.OnHTML("table a", func(e *colly.HTMLElement) {
		versions_available = append(versions_available, e.Attr("href"))
	})

	c.Visit(url)

	filter := func(s string) bool { return strings.HasSuffix(s, ".zip") }
	versions_available_filtered := helpers.FilterStringSlice(versions_available, filter)

	sort.Strings(versions_available_filtered)

	latest_available := ""
	if len(versions_available_filtered) > 0 {
		latest_available = versions_available_filtered[len(versions_available_filtered)-1]
	} else {
		return "", nil
	}

	dl_url := base_url + latest_available

	// Populate the A1 structs of availables
	A1.Mutex.Lock()
	defer A1.Mutex.Unlock()
	A1.Upstream.Rom["LineageOSMicroG"] = &Item{}
	A1.Upstream.Rom["LineageOSMicroG"].Name = "LineageOSMicroG"
	A1.Upstream.Rom["LineageOSMicroG"].Href = dl_url
	A1.Upstream.Rom["LineageOSMicroG"].Checksum_url_suffix = ".sha256sum"
	A1.Upstream.Rom["LineageOSMicroG"].Filename = helpers.ExtractFileNameFromHref(dl_url)
	v, err := LineageosMicrogParseVersion(A1.Upstream.Rom["LineageOSMicroG"].Filename)
	if err != nil {
		return dl_url, fmt.Errorf("unable to parse LineageOSMicroG version in %s", A1.Upstream.Rom["LineageOSMicroG"].Filename)
	}
	A1.Upstream.Rom["LineageOSMicroG"].Version = v
	av, err := LineageosMicrogParseAndroidVersion(v)
	if err != nil {
		return latest_available, fmt.Errorf("unable to parse LineageOSMicroG Android version with %s", v)
	}
	A1.Upstream.Rom["LineageOSMicroG"].Android_version = av

	return dl_url, nil
}

func LineageosMicrogParseVersion(filename string) (string, error) {
	return LineageosParseVersion(filename)
}

func LineageosMicrogParseAndroidVersion(losversion string) (string, error) {
	return LineageosParseAndroidVersion(losversion)
}