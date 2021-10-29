package get

import (
	"github.com/gocolly/colly"

	"github.com/amo13/anarchy-droid/helpers"
	"github.com/amo13/anarchy-droid/logger"

	"net/http"
	"strings"
	"sort"
	"fmt"
)

// Download latest CrDroid zip into flash folder and return the file name
func CrDroid(codename string) (string, error) {
	dl_url, err := CrDroidLatestAvailableHref(codename)
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

// Returns file name of the latest available CrDroid zip
func CrDroidLatestAvailableFileName(codename string) (string, error) {
	href, err := CrDroidLatestAvailableHref(codename)
	if err != nil {
		return "", err
	}

	return helpers.ExtractFileNameFromHref(href), nil
}

// Returns download link of the latest available CrDroid zip
func CrDroidLatestAvailableHref(codename string) (string, error) {
	if A1.Upstream.Rom["crDroid"] != nil {
		if A1.Upstream.Rom["crDroid"].Href != "" {
			return A1.Upstream.Rom["crDroid"].Href, nil
		}
	}

	device_url := "https://sourceforge.net/projects/crdroid/files/" + codename + "/"

	status_code, err := StatusCode(device_url)
	if err != nil {
		return "", err
	}
	if status_code != "200 OK" {
		return "", fmt.Errorf("not available")
	}


	// crDroid 4.x, 5.x, 6.x, 7.x
	versions_available := make([]string, 0)

	c := colly.NewCollector()

	c.OnError(func(_ *colly.Response, err error) {
		logger.LogError("CrDroid:", err)
	})

	c.OnHTML("tr th a", func(e *colly.HTMLElement) {
		versions_available = append(versions_available, e.Attr("href"))
	})

	c.Visit(device_url)

	filter := func(s string) bool { return strings.HasSuffix(s, ".x/") }
	versions_available_filtered := helpers.FilterStringSlice(versions_available, filter)

	sort.Strings(versions_available_filtered)

	latest_available := ""
	if len(versions_available_filtered) > 0 {
		latest_available = versions_available_filtered[len(versions_available_filtered)-1]
	} else {
		return "", nil
	}


	// Actual rom files for latest available crDroid version
	files_available := make([]string, 0)

	d := colly.NewCollector()

	d.OnError(func(_ *colly.Response, err error) {
		logger.LogError("crDroid:", err)
	})

	d.OnHTML("tr th a", func(e *colly.HTMLElement) {
		files_available = append(files_available, e.Attr("href"))
	})

	d.Visit("https://sourceforge.net" + latest_available)

	filter2 := func(s string) bool { return strings.HasSuffix(s, ".zip/download") }
	files_available_filtered := helpers.FilterStringSlice(files_available, filter2)

	sort.Strings(files_available_filtered)

	latest_file_available := ""
	if len(files_available_filtered) > 0 {
		latest_file_available = files_available_filtered[len(files_available_filtered)-1]
	} else {
		return "", nil
	}


	var dl_url string

	// Sometimes a sourceforge mirror is irresponsive. Retry in that case up to 3 times.
	for tries := 1; tries < 5; tries++ {
		// Follow redirects to the actual file URL from a mirror
		resp, err := http.Head(latest_file_available)
		if err != nil {
			logger.Log("Retrying to follow CrDroid Sourceforge redirects to download mirror...", err.Error())
		} else {
			dl_url = resp.Request.URL.String()
			break
		}
	}

	if dl_url == "" || (!strings.HasSuffix(dl_url, ".zip") && !strings.Contains(dl_url, ".zip?")) {
		logger.Log("Unable to follow CrDroid sourceforge redirects to mirror")
		return "", fmt.Errorf("unable to follow sourceforge redirects to mirror")
	}

	// Populate the A1 structs of availables
	A1.Mutex.Lock()
	defer A1.Mutex.Unlock()
	A1.Upstream.Rom["crDroid"] = &Item{}
	A1.Upstream.Rom["crDroid"].Name = "crDroid"
	A1.Upstream.Rom["crDroid"].Href = dl_url
	A1.Upstream.Rom["crDroid"].Checksum_url_suffix = ""
	A1.Upstream.Rom["crDroid"].Filename = helpers.ExtractFileNameFromHref(dl_url)
	v, err := CrDroidParseVersion(A1.Upstream.Rom["crDroid"].Filename)
	if err != nil {
		return latest_available, fmt.Errorf("unable to parse CrDroid version in %s", A1.Upstream.Rom["crDroid"].Filename)
	}
	A1.Upstream.Rom["crDroid"].Version = v
	av, err := CrDroidParseAndroidVersion(A1.Upstream.Rom["crDroid"].Filename)
	if err != nil {
		return latest_available, fmt.Errorf("unable to parse CrDroid Android version with %s: %s", v, err.Error())
	}
	A1.Upstream.Rom["crDroid"].Android_version = av

	return dl_url, nil
}

func CrDroidParseAndroidVersion(filename string) (string, error) {
	parts := strings.Split(filename, "-")
	if len(parts) >= 2 {
		return parts[1], nil
	} else {
		return "", fmt.Errorf("unable to parse crDroid version in", filename)
	}
}

func CrDroidParseVersion(filename string) (string, error) {
	parts := strings.Split(filename, "-")
	if len(parts) >= 4 {
		v := parts[len(parts)-1]
		if strings.HasPrefix(v, "v") {
			v = v[1:]
		}
		if strings.HasSuffix(v, ".zip") {
			v = v[:len(v)-4]
		}
		return v, nil
	} else {
		return "", fmt.Errorf("unable to parse crDroid android version in", filename)
	}
}