package get

import (
	"github.com/amo13/anarchy-droid/helpers"
	"github.com/amo13/anarchy-droid/logger"

	"github.com/google/go-github/github"

	"github.com/gocolly/colly"

	"net/http"
	"context"
	"strings"
	"sort"
	"time"
	"fmt"
)

// Download latest Micro5k unofficial microG installer zip into flash folder and return the file name
func Micro5kMicroG(which string) (string, error) {
	dl_url, err := Micro5kMicroGLatestAvailableHref(which)
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

func Micro5kMicroGLatestAvailableOSS() (map[string]string, error) {
	client := github.NewClient(&http.Client{
		Timeout: time.Second * 8,
	})

	var (
		err      error
		release *github.RepositoryRelease
		retries  int = 3
	)

	for retries > 0 {
		release, _, err = client.Repositories.GetLatestRelease(context.Background(), "micro5k", "microg-unofficial-installer")
		if err != nil {
			logger.Log("Retrying to get info on the latest Micro5kMicroG release...")
			retries -= 1
		} else {
			break
		}
	}
	if err != nil || len(release.Assets) == 0 {
		logger.LogError("Get latest Micro5kMicroG github release:", err)
		return map[string]string{}, err
	}

	var installer github.ReleaseAsset

	for _, asset := range release.Assets {
		if strings.HasPrefix(*asset.Name, "microg-unofficial-installer-") && strings.HasSuffix(*asset.Name, ".zip") {
			installer = asset
		}
	}

	result := make(map[string]string)

	result["href"] = *installer.BrowserDownloadURL
	result["version"] = *release.TagName
	result["filename"] = helpers.ExtractFileNameFromHref(*installer.BrowserDownloadURL)

	return result, nil
}

func Micro5kMicroGLatestAvailableFull() (map[string]string, error) {
	url := "https://stuff.anarchy-droid.com/micro5k/full/"

	status_code, err := StatusCode(url)
	if err != nil {
		return map[string]string{}, err
	}
	if status_code == "404 Not Found" {
		return map[string]string{}, fmt.Errorf("not available")
	}

	versions_available := make([]string, 0)

	c := colly.NewCollector()

	c.OnError(func(_ *colly.Response, err error) {
		logger.LogError("Error getting latest available href for Micro5kMicroG-Full:", err)
	})

	c.OnHTML("a", func(e *colly.HTMLElement) {
		versions_available = append(versions_available, e.Attr("href"))
	})

	c.Visit(url)

	filter := func(s string) bool { return strings.HasPrefix(s, "microg-unofficial-installer") && !strings.HasSuffix(s, "sha256") && !strings.HasSuffix(s, "md5") }
	versions_available_filtered := helpers.FilterStringSlice(versions_available, filter)

	sort.Strings(versions_available_filtered)

	latest_available := versions_available_filtered[len(versions_available_filtered)-1]

	dl_url := url + latest_available

	result := make(map[string]string)

	result["href"] = dl_url
	result["version"] = helpers.GenericParseVersion(latest_available)
	result["filename"] = helpers.ExtractFileNameFromHref(latest_available)

	return result, nil
}

func Micro5kMicroGLatestAvailableGsync() (map[string]string, error) {
	url := "https://stuff.anarchy-droid.com/micro5k/gsync/"

	status_code, err := StatusCode(url)
	if err != nil {
		return map[string]string{}, err
	}
	if status_code == "404 Not Found" {
		return map[string]string{}, fmt.Errorf("not available")
	}

	versions_available := make([]string, 0)

	c := colly.NewCollector()

	c.OnError(func(_ *colly.Response, err error) {
		logger.LogError("Error getting latest available href for Micro5kMicroG-Gsync:", err)
	})

	c.OnHTML("a", func(e *colly.HTMLElement) {
		versions_available = append(versions_available, e.Attr("href"))
	})

	c.Visit(url)

	filter := func(s string) bool { return strings.HasPrefix(s, "google-sync-addon") && !strings.HasSuffix(s, "sha256") && !strings.HasSuffix(s, "md5") }
	versions_available_filtered := helpers.FilterStringSlice(versions_available, filter)

	sort.Strings(versions_available_filtered)

	latest_available := versions_available_filtered[len(versions_available_filtered)-1]

	dl_url := url + latest_available

	result := make(map[string]string)

	result["href"] = dl_url
	result["version"] = helpers.GenericParseVersion(latest_available)
	result["filename"] = helpers.ExtractFileNameFromHref(latest_available)

	return result, nil
}

// Returns variants of the latest available Micro5kMicroG zip
func Micro5kMicroGLatestAvailableHref(which string) (string, error) {
	if A1.Upstream.Micro5kMicroG[strings.ToLower(which)] != nil {
		if A1.Upstream.Micro5kMicroG[strings.ToLower(which)].Href != "" {
			return A1.Upstream.Micro5kMicroG[strings.ToLower(which)].Href, nil
		}
	}

	available := make(map[string]map[string]string)
	var err error

	switch strings.ToLower(which) {
	case "", "full":
		which = "full"
		available[which], err = Micro5kMicroGLatestAvailableFull()
		if err != nil {
			return "", err
		}
	case "oss":
		which = "oss"
		available[which], err = Micro5kMicroGLatestAvailableOSS()
		if err != nil {
			return "", err
		}
	case "gsync":
		which = "gsync"
		available[which], err = Micro5kMicroGLatestAvailableGsync()
		if err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("unclear which Micro5kMicroG package to get")
	}

	// Populate the A1 structs of availables
	A1.Mutex.Lock()
	defer A1.Mutex.Unlock()
	A1.Upstream.Micro5kMicroG[which] = &Item{}
	A1.Upstream.Micro5kMicroG[which].Href = available[which]["href"]
	A1.Upstream.Micro5kMicroG[which].Version = available[which]["version"]
	A1.Upstream.Micro5kMicroG[which].Checksum_url_suffix = ".sha256"
	A1.Upstream.Micro5kMicroG[which].Filename = available[which]["filename"]

	return available[which]["href"], nil
}