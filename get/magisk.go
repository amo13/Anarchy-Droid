package get

import (
	"github.com/gocolly/colly"
	"anarchy-droid/helpers"
	"anarchy-droid/logger"
	"strings"
	"sort"
	"fmt"
)

// Download latest Magisk zip into flash folder and return the file name
func Magisk() (string, error) {
	dl_url, err := MagiskLatestAvailableHref()
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

// Returns file name of the latest available Magisk zip
func MagiskLatestAvailableFileName() (string, error) {
	href, err := MagiskLatestAvailableHref()
	if err != nil {
		return "", err
	}

	return helpers.ExtractFileNameFromHref(href), nil
}

// Returns download link of the latest available Magisk zip
func MagiskLatestAvailableHref() (string, error) {
	if A1.Upstream.Magisk.Href != "" {
		return A1.Upstream.Magisk.Href, nil
	}

	url := "https://github.com/topjohnwu/Magisk/releases"

	status_code, err := StatusCode(url)
	if err != nil {
		return "", err
	}
	if status_code == "404" {
		return "", fmt.Errorf("not available")
	}

	versions_available := make([]string, 0)

	c := colly.NewCollector()

	c.OnError(func(_ *colly.Response, err error) {
		logger.LogError("Magisk:", err)
	})

	c.OnHTML("a", func(e *colly.HTMLElement) {
		versions_available = append(versions_available, e.Attr("href"))
	})

	c.Visit(url)

	filter := func(s string) bool { return strings.HasSuffix(s, ".apk") && strings.Contains(helpers.ExtractFileNameFromHref(s), "Magisk") && !strings.Contains(helpers.ExtractFileNameFromHref(s), "Manager") }
	versions_available_filtered := helpers.FilterStringSlice(versions_available, filter)

	sort.Strings(versions_available_filtered)

	latest_available := versions_available_filtered[len(versions_available_filtered)-1]

	dl_url := "https://github.com" + latest_available

	A1.Mutex.Lock()
	defer A1.Mutex.Unlock()
	A1.Upstream.Magisk.Name = "Magisk"
	A1.Upstream.Magisk.Href = dl_url
	A1.Upstream.Magisk.Checksum_url_suffix = ""
	A1.Upstream.Magisk.Filename = helpers.ExtractFileNameFromHref(dl_url)
	v, err := MagiskParseVersion(A1.Upstream.Magisk.Filename)
	if err != nil {
		return latest_available, fmt.Errorf("unable to parse Magisk version in %s", A1.Upstream.Magisk.Filename)
	}
	A1.Upstream.Magisk.Version = v

	return dl_url, nil
}

func MagiskParseVersion(filename string) (string, error) {
	version := helpers.GenericParseVersion(filename)
	if version == "" {
		return version, fmt.Errorf("unable to parse version")
	} else {
		return version, nil
	}
}

