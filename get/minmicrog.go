package get

import (
	"github.com/amo13/anarchy-droid/helpers"
	"github.com/amo13/anarchy-droid/logger"

	"github.com/google/go-github/github"

	"net/http"
	"strings"
	"context"
	"time"
	"fmt"
)

// Download latest OpenGapps zip into flash folder and return the file name
func MinMicroG(variant string) (string, error) {
	variants, err := MinMicroGLatestAvailableVariants()
	if err != nil {
		return "", err
	}
	if variants[variant] != nil {
		return "", fmt.Errorf("Variant %s not available. Available variants are: %s", variant, variants)
	}

	dl_url := variants[variant].Href

	// Parse the file name from the href
	file_name := variants[variant].Filename

	err = DownloadFile("flash/" + file_name, dl_url, "")
	if err != nil {
		return "", err
	}

	return file_name, nil
}

// Returns file name of the latest available OpenGapps zip
func MinMicroGLatestAvailableVariants() (map[string]*Item, error) {
	if len(A1.Upstream.MinMicroG) > 0 {
		return A1.Upstream.MinMicroG, nil
	}

	client := github.NewClient(&http.Client{
		Timeout: time.Second * 8,
	})

	var (
		err      error
		release *github.RepositoryRelease
		retries  int = 3
	)

	for retries > 0 {
		release, _, err = client.Repositories.GetLatestRelease(context.Background(), "FriendlyNeighborhoodShane", "MinMicroG_releases")
		if err != nil {
			logger.Log("Retrying to get info on the latest MinMicroG release...")
			retries -= 1
		} else {
			break
		}
	}
	if err != nil {
		logger.LogError("Get latest MinMicroG github release:", err)
		return map[string]*Item{}, err
	}

	variants := make(map[string]*Item)

	for _, asset := range release.Assets {
		if strings.Contains(*asset.Name, "MinMicroG-") {
			variant := strings.Split(*asset.Name, "-")[1]
			variants[variant] = &Item{}
			variants[variant].Name = variant
			variants[variant].Href = *asset.BrowserDownloadURL
			variants[variant].Filename = helpers.ExtractFileNameFromHref(variants[variant].Href)
			variants[variant].Version = helpers.GenericParseVersion(variants[variant].Filename)
			variants[variant].Checksum_url_suffix = ""
		}
	}

	if variants["Playstore"] == nil {
		variant := "Playstore"
		variants[variant] = &Item{}
		variants[variant].Name = variant
		variants[variant].Href = "https://stuff.anarchy-droid.com/MinAddon-Playstore-Playstore-UPDATELY-20221022011031.zip"
		variants[variant].Filename = helpers.ExtractFileNameFromHref(variants[variant].Href)
		variants[variant].Version = "Random nightly"
		variants[variant].Checksum_url_suffix = ""
	}

	// Populate the A1 structs of availables
	A1.Mutex.Lock()
	defer A1.Mutex.Unlock()
	A1.Upstream.MinMicroG = variants

	return variants, nil
}

// Returns download link of the latest available NanoDroid zip
func MinMicroGLatestAvailableHref(variant string) (string, error) {
	variants, err := MinMicroGLatestAvailableVariants()
	if err != nil {
		return "", err
	}

	return variants[variant].Href, nil
}