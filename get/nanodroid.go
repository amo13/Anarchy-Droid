package get

import (
	"github.com/gocolly/colly"
	"anarchy-droid/helpers"
	"anarchy-droid/logger"
	"strings"
	"regexp"
	"sort"
	"fmt"
)

// Download latest NanoDroid zip into flash folder and return the file name
// which can be one of "full", "fdroid", "microg" or "patcher"
func NanoDroid(which string) (string, error) {
	dl_url, err := NanoDroidLatestAvailableHref(which)
	if err != nil {
		return "", err
	}

	// Parse the file name from the href
	file_name := helpers.ExtractFileNameFromHref(dl_url)

	err = DownloadFile("flash/" + file_name, dl_url, ".sha256")
	if err != nil {
		return "", err
	}

	return file_name, nil
}

// Returns file name of the latest available NanoDroid zip
func NanoDroidLatestAvailableFileName(which string) (string, error) {
	href, err := NanoDroidLatestAvailableHref(which)
	if err != nil {
		return "", err
	}

	return helpers.ExtractFileNameFromHref(href), nil
}

// Returns download link of the latest available NanoDroid zip
func NanoDroidLatestAvailableHref(which string) (string, error) {
	if A1.Upstream.NanoDroid[strings.Title(which)] != nil {
		if A1.Upstream.NanoDroid[strings.Title(which)].Href != "" {
			return A1.Upstream.NanoDroid[strings.Title(which)].Href, nil
		}
	}

	things_available, err := nanoDroidAllAvailable()
	if err != nil {
		return "", err
	}

	switch strings.ToLower(which) {
	case "", "full":
		which = "full"
	case "fdroid", "f-droid":
		which = "fdroid"
	case "microg":
		which = "microG"
	case "patcher":
		which = "patcher"
	case "google":
		which = "Google"
	default:
		return "", fmt.Errorf("unclear which nanodroid package to get")
	}

	filter := func(s string) bool {return true}
	if which == "full" {
		r := regexp.MustCompile(`NanoDroid-\d`)
		filter = func(s string) bool { return strings.HasSuffix(s, ".zip") && strings.HasPrefix(s, "NanoDroid-") && r.MatchString(s) }
	} else {
		filter = func(s string) bool { return strings.HasSuffix(s, ".zip") && strings.HasPrefix(s, "NanoDroid-" + which) }
	}
	things_available_filtered := helpers.FilterStringSlice(things_available, filter)

	sort.Strings(things_available_filtered)

	latest_available := things_available_filtered[len(things_available_filtered)-1]

	dl_url := "https://downloads.nanolx.org/NanoDroid/Stable/" + latest_available

	// Populate the A1 structs of availables
	A1.Mutex.Lock()
	defer A1.Mutex.Unlock()
	A1.Upstream.NanoDroid[strings.Title(which)] = &Item{}
	A1.Upstream.NanoDroid[strings.Title(which)].Href = dl_url
	A1.Upstream.NanoDroid[strings.Title(which)].Checksum_url_suffix = ".sha256"
	A1.Upstream.NanoDroid[strings.Title(which)].Filename = helpers.ExtractFileNameFromHref(dl_url)
	v, err := NanoDroidParseVersion(A1.Upstream.NanoDroid[strings.Title(which)].Filename)
	if err != nil {
		return dl_url, fmt.Errorf("unable to parse NanoDroid version in %s", A1.Upstream.NanoDroid[strings.Title(which)].Filename)
	}
	A1.Upstream.NanoDroid[strings.Title(which)].Version = v

	return dl_url, nil
}

// Helper function for NanoDroidLatestAvailableHref(which string)
func nanoDroidAllAvailable() ([]string, error) {
	url := "https://downloads.nanolx.org/NanoDroid/Stable"

	status_code, err := StatusCode(url)
	if err != nil {
		return []string{}, err
	}
	if status_code == "404" {
		return []string{}, fmt.Errorf("not available")
	}

	things_available := make([]string, 0)

	c := colly.NewCollector()

	c.OnError(func(_ *colly.Response, err error) {
		logger.LogError("NanoDroid:", err)
	})

	c.OnHTML("table a", func(e *colly.HTMLElement) {
		things_available = append(things_available, helpers.ExtractFileNameFromHref(e.Attr("href")))
	})

	c.Visit(url)

	return things_available, nil
}

func NanoDroidParseVersion(filename string) (string, error) {
	version := helpers.GenericParseVersion(filename)
	if version == "" {
		return version, fmt.Errorf("unable to parse version")
	} else {
		return version, nil
	}
}

