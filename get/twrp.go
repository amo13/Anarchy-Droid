package get

import (
	"github.com/gocolly/colly"
	"anarchy-droid/helpers"
	"anarchy-droid/logger"
	"strings"
	"sort"
	"fmt"
)

// Download latest TWRP image into flash folder and return the file name
func TwrpImg(codename string) (string, error) {
	dl_url, err := TwrpImgLatestAvailableHref(codename)
	if err != nil {
		return "", err
	}

	file_name := helpers.ExtractFileNameFromHref(dl_url)

	err = DownloadFile("flash/" + file_name, dl_url, ".sha256")
	if err != nil {
		return "", err
	}

	return file_name, nil
}

// Download latest TWRP zip into flash folder and return the file name
func TwrpZip(codename string) (string, error) {
	dl_url, err := TwrpZipLatestAvailableHref(codename)
	if err != nil {
		return "", err
	}

	file_name := helpers.ExtractFileNameFromHref(dl_url)

	err = DownloadFile("flash/" + file_name, dl_url, ".sha256")
	if err != nil {
		return "", err
	}

	return file_name, nil
}

// Returns file name of the latest available TWRP image
func TwrpImgLatestAvailableHref(codename string) (string, error) {
	// Only retrieve upstream data if not done yet
	if A1.Upstream.Twrp.Img.Href != "" {
		return A1.Upstream.Twrp.Img.Href, nil
	}

	url := "https://dl.twrp.me/" + codename

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
		logger.LogError("TwrpImg:", err)
	})

	c.OnHTML("table a", func(e *colly.HTMLElement) {
		versions_available = append(versions_available, e.Text)
	})

	c.Visit(url)

	filter := func(s string) bool { return strings.HasSuffix(s, ".img") }
	versions_available_filtered := helpers.FilterStringSlice(versions_available, filter)

	sort.Strings(versions_available_filtered)

	latest_available := ""
	if len(versions_available_filtered) > 0 {
		latest_available = versions_available_filtered[len(versions_available_filtered)-1]
	} else {
		return "", nil
	}

	dl_url := url + "/" + latest_available

	// Populate the A1 structs of availables
	A1.Mutex.Lock()
	defer A1.Mutex.Unlock()
	A1.Upstream.Twrp.Img.Name = "TWRP"
	A1.Upstream.Twrp.Img.Href = dl_url
	A1.Upstream.Twrp.Img.Checksum_url_suffix = ".sha256"
	A1.Upstream.Twrp.Img.Filename = helpers.ExtractFileNameFromHref(dl_url)
	v, err := TwrpImgParseVersion(A1.Upstream.Twrp.Img.Filename)
	if err != nil {
		return dl_url, fmt.Errorf("unable to parse TWRP version in %s", A1.Upstream.Twrp.Img.Filename)
	}
	A1.Upstream.Twrp.Img.Version = v

	return dl_url, nil
}

// Returns file name of the latest available TWRP zip
func TwrpZipLatestAvailableHref(codename string) (string, error) {
	// Only retrieve upstream data if not done yet
	if A1.Upstream.Twrp.Zip.Href != "" {
		return A1.Upstream.Twrp.Zip.Href, nil
	}

	url := "https://dl.twrp.me/" + codename

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
		logger.LogError("TwrpZip:", err)
	})

	c.OnHTML("table a", func(e *colly.HTMLElement) {
		versions_available = append(versions_available, e.Text)
	})

	c.Visit(url)

	filter := func(s string) bool { return strings.HasSuffix(s, codename + ".zip") }
	versions_available_filtered := helpers.FilterStringSlice(versions_available, filter)

	sort.Strings(versions_available_filtered)

	latest_available := ""
	if len(versions_available_filtered) > 0 {
		latest_available = versions_available_filtered[len(versions_available_filtered)-1]
	} else {
		return "", nil
	}

	dl_url := url + "/" + latest_available

	// Populate the A1 structs of availables
	A1.Mutex.Lock()
	defer A1.Mutex.Unlock()
	A1.Upstream.Twrp.Zip.Name = "TWRP"
	A1.Upstream.Twrp.Zip.Href = dl_url
	A1.Upstream.Twrp.Zip.Checksum_url_suffix = ".sha256"
	A1.Upstream.Twrp.Zip.Filename = helpers.ExtractFileNameFromHref(dl_url)
	v, err := TwrpZipParseVersion(A1.Upstream.Twrp.Zip.Filename)
	if err != nil {
		return dl_url, fmt.Errorf("unable to parse TWRP version in %s", A1.Upstream.Twrp.Zip.Filename)
	}
	A1.Upstream.Twrp.Zip.Version = v

	return dl_url, nil
}

func TwrpImgParseVersion(filename string) (string, error) {
	parts := strings.Split(filename, "-")
	if len(parts) >= 2 {
		return parts[1], nil
	} else {
		return "", fmt.Errorf("unable to parse version")
	}
}

func TwrpZipParseVersion(filename string) (string, error) {
	parts := strings.Split(filename, "-")
	if len(parts) >= 3 {
		return parts[2], nil
	} else {
		return "", fmt.Errorf("unable to parse TWRP version")
	}
}