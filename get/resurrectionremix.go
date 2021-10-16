package get

import (
	"github.com/gocolly/colly"

	"github.com/amo13/anarchy-droid/helpers"
	"github.com/amo13/anarchy-droid/logger"

	"net/http"
	"strconv"
	"strings"
	"sort"
	"fmt"
)

// Download latest ResurrectionRemix zip into flash folder and return the file name
func ResurrectionRemix(codename string) (string, error) {
	dl_url, err := ResurrectionRemixLatestAvailableHref(codename)
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

// Returns file name of the latest available ResurrectionRemix zip
func ResurrectionRemixLatestAvailableFileName(codename string) (string, error) {
	href, err := ResurrectionRemixLatestAvailableHref(codename)
	if err != nil {
		return "", err
	}

	return helpers.ExtractFileNameFromHref(href), nil
}

// Returns download link of the latest available ResurrectionRemix zip
func ResurrectionRemixLatestAvailableHref(codename string) (string, error) {
	if A1.Upstream.Rom["ResurrectionRemix"] != nil {
		if A1.Upstream.Rom["ResurrectionRemix"].Href != "" {
			return A1.Upstream.Rom["ResurrectionRemix"].Href, nil
		}
	}

	var redirect_url string

	// Follow redirects to the actual file URL from a mirror
	resp, err := http.Head("https://get.resurrectionremix.com/")
	if err != nil {
		logger.LogError("Error requesting (following redirect from) https://get.resurrectionremix.com/:", err)
	} else {
		redirect_url = resp.Request.URL.String()
	}

	if redirect_url == "" {
		return "", fmt.Errorf("Unable to get sourceforge redirection link for Resurrection Remix: redirection url empty")
	}

	device_url := redirect_url + codename + "/"

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
		logger.LogError("ResurrectionRemix:", err)
	})

	c.OnHTML("tr th a", func(e *colly.HTMLElement) {
		versions_available = append(versions_available, e.Attr("href"))
	})

	c.Visit(device_url)

	filter := func(s string) bool { return strings.HasSuffix(s, ".zip/download") }
	versions_available_filtered := helpers.FilterStringSlice(versions_available, filter)

	sort.Strings(versions_available_filtered)

	latest_available := ""
	if len(versions_available_filtered) > 0 {
		latest_available = versions_available_filtered[len(versions_available_filtered)-1]
	} else {
		return "", nil
	}

	var dl_url string

	// Sometimes a sourceforge mirror is irresponsive. Retry in that case up to 3 times.
	for tries := 1; tries < 5; tries++ {
		// Follow redirects to the actual file URL from a mirror
		resp, err := http.Head(latest_available)
		if err != nil {
			logger.Log("Retrying to follow ResurrectionRemix Sourceforge redirects to download mirror...")
		} else {
			dl_url = resp.Request.URL.String()
			break
		}
	}

	if dl_url == "" || !strings.HasSuffix(dl_url, ".zip") {
		logger.Log("Unable to follow ResurrectionRemix sourceforge redirects to mirror")
		return "", fmt.Errorf("unable to follow sourceforge redirects to mirror")
	}

	// Populate the A1 structs of availables
	A1.Mutex.Lock()
	defer A1.Mutex.Unlock()
	A1.Upstream.Rom["ResurrectionRemix"] = &Item{}
	A1.Upstream.Rom["ResurrectionRemix"].Name = "ResurrectionRemix"
	A1.Upstream.Rom["ResurrectionRemix"].Href = dl_url
	A1.Upstream.Rom["ResurrectionRemix"].Checksum_url_suffix = ""
	A1.Upstream.Rom["ResurrectionRemix"].Filename = helpers.ExtractFileNameFromHref(dl_url)
	v, err := ResurrectionRemixParseVersion(A1.Upstream.Rom["ResurrectionRemix"].Filename)
	if err != nil {
		return latest_available, fmt.Errorf("unable to parse ResurrectionRemix version in %s", A1.Upstream.Rom["ResurrectionRemix"].Filename)
	}
	A1.Upstream.Rom["ResurrectionRemix"].Version = v
	av, err := ResurrectionRemixParseAndroidVersion(v)
	if err != nil {
		return latest_available, fmt.Errorf("unable to parse ResurrectionRemix Android version with %s: %s", v, err.Error())
	}
	A1.Upstream.Rom["ResurrectionRemix"].Android_version = av

	return latest_available, nil
}

func ResurrectionRemixParseVersion(filename string) (string, error) {
	return helpers.GenericParseVersion(filename), nil
}

func ResurrectionRemixParseAndroidVersion(romversion string) (string, error) {
	if romversion != "" {
		maj := strings.Split(romversion, ".")
		maj_int, err := strconv.Atoi(maj[0])
		if err != nil {
			return "", err
		}

		v := strconv.Itoa(maj_int + 2)

		if v == "8" {
			v = "8.1"
		} else if v == "7" {
			v = "7.1"
		}

		return v, nil
	} else {
		return "", fmt.Errorf("Unable to tell android version of Resurrection Remix or CarbonRom: %s", romversion)
	}
}