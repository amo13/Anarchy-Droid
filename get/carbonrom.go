package get

import (
	"github.com/gocolly/colly"

	"github.com/amo13/anarchy-droid/helpers"
	"github.com/amo13/anarchy-droid/logger"

	"strings"
	"sort"
	"fmt"
)

// Download latest Carbonrom zip into flash folder and return the file name
func Carbonrom(codename string) (string, error) {
	dl_url, err := CarbonromLatestAvailableHref(codename)
	if err != nil {
		return "", err
	}

	// Parse the file name from the href
	file_name := helpers.ExtractFileNameFromHref(dl_url)

	err = DownloadFile("flash/" + file_name, dl_url, ".md5sum")
	if err != nil {
		return "", err
	}

	return file_name, nil
}

// Returns file name of the latest available Carbonrom zip
func CarbonromLatestAvailableFileName(codename string) (string, error) {
	href, err := CarbonromLatestAvailableHref(codename)
	if err != nil {
		return "", err
	}

	return helpers.ExtractFileNameFromHref(href), nil
}

// Returns download link of the latest available Carbonrom zip
func CarbonromLatestAvailableHref(codename string) (string, error) {
	if A1.Upstream.Rom["Carbonrom"] != nil {
		if A1.Upstream.Rom["Carbonrom"].Href != "" {
			return A1.Upstream.Rom["Carbonrom"].Href, nil
		}
	}

	url := "https://get.carbonrom.org/device-" + codename + ".html"

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
		logger.LogError("Carbonrom:", err)
	})

	c.OnHTML("td dl dd a", func(e *colly.HTMLElement) {
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
	A1.Upstream.Rom["Carbonrom"] = &Item{}
	A1.Upstream.Rom["Carbonrom"].Name = "Carbonrom"
	A1.Upstream.Rom["Carbonrom"].Href = latest_available
	A1.Upstream.Rom["Carbonrom"].Checksum_url_suffix = ".md5sum"
	A1.Upstream.Rom["Carbonrom"].Filename = helpers.ExtractFileNameFromHref(latest_available)
	v, err := CarbonromParseVersion(A1.Upstream.Rom["Carbonrom"].Filename)
	if err != nil {
		return latest_available, fmt.Errorf("unable to parse Carbonrom version in %s", A1.Upstream.Rom["Carbonrom"].Filename)
	}
	A1.Upstream.Rom["Carbonrom"].Version = v
	av, err := CarbonromParseAndroidVersion(v)
	if err != nil {
		return latest_available, fmt.Errorf("unable to parse Carbonrom Android version with %s", v)
	}
	A1.Upstream.Rom["Carbonrom"].Android_version = av

	return latest_available, nil
}

func CarbonromParseVersion(filename string) (string, error) {
	return helpers.GenericParseVersion(filename), nil
}

func CarbonromParseAndroidVersion(romversion string) (string, error) {
	return ResurrectionRemixParseAndroidVersion(romversion)
}