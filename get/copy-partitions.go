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
func CopyPartitionsZip() (string, error) {
	dl_url, err := CopyPartitionsZipLatestAvailableHref()
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

// Returns download link of the latest available Magisk zip
func CopyPartitionsZipLatestAvailableHref() (string, error) {
	if A1.Upstream.CopyPartitions.Href != "" {
		return A1.Upstream.CopyPartitions.Href, nil
	}

	url := "https://stuff.free-droid.com"

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
		logger.LogError("Error getting latest available href for copy-partitions.zip:", err)
	})

	c.OnHTML("a", func(e *colly.HTMLElement) {
		versions_available = append(versions_available, e.Attr("href"))
	})

	c.Visit(url)

	filter := func(s string) bool { return strings.HasPrefix(s, "copy-partitions") && !strings.HasSuffix(s, "sha256") && !strings.HasSuffix(s, "md5") }
	versions_available_filtered := helpers.FilterStringSlice(versions_available, filter)

	sort.Strings(versions_available_filtered)

	latest_available := versions_available_filtered[len(versions_available_filtered)-1]

	dl_url := "https://stuff.free-droid.com/" + latest_available

	A1.Mutex.Lock()
	defer A1.Mutex.Unlock()
	A1.Upstream.CopyPartitions.Name = "CopyPartitions"
	A1.Upstream.CopyPartitions.Href = dl_url
	A1.Upstream.CopyPartitions.Filename = latest_available

	return dl_url, nil
}