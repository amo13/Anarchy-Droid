package get

import (
	"github.com/gocolly/colly"
	"anarchy-droid/helpers"
	"anarchy-droid/logger"
	"strings"
	"sort"
	"fmt"
)

func FromArchive(codename string, what string) (string, error) {
	available_hrefs, err := ArchiveLatestAvailableHrefMap(codename)
	if err != nil {
		return "", err
	}

	dl_url := available_hrefs[what]

	// Parse the file name from the href
	file_name := helpers.ExtractFileNameFromHref(dl_url)

	sha256_status_code, _ := StatusCode(dl_url + ".sha256")

	if sha256_status_code == "200 OK" {
		err = DownloadFile("flash/" + file_name, dl_url, ".sha256")
		if err != nil {
			return "", err
		}
	} else {
		err = DownloadFile("flash/" + file_name, dl_url, "")
		if err != nil {
			return "", err
		}
	}

	return file_name, nil
}

func ArchiveLatestAvailableHrefMap(codename string) (map[string]string, error) {
	url := "https://archive.free-droid.com/" + codename

	files, err := ArchiveLatestAvailableFileNamesMap(codename)
	if err != nil {
		return map[string]string{}, err
	}

	for folder, filename := range files {
		files[folder] = url + "/" + folder + "/" + filename

		// Populate the A1 structs of availables
		archivePopulateAvailablesStruct(folder, filename, files[folder])
	}

	return files, nil
}

// Populate the A1 structs of availables
func archivePopulateAvailablesStruct(folder string, filename string, href string) error {
	A1.Mutex.Lock()
	defer A1.Mutex.Unlock()
	
	switch strings.ToLower(folder) {
	case "twrp":
		if strings.HasSuffix(filename, ".img") {
			A1.Archive.Twrp.Img.Name = "TWRP"
			A1.Archive.Twrp.Img.Href = href
			A1.Archive.Twrp.Img.Checksum_url_suffix = ".sha256"
			A1.Archive.Twrp.Img.Filename = filename
			v, err := TwrpImgParseVersion(filename)
			if err != nil {
				return fmt.Errorf("unable to parse TWRP version in %s", filename)
			}
			A1.Archive.Twrp.Img.Version = v
		} else if strings.HasSuffix(filename, ".zip") {
			A1.Archive.Twrp.Zip.Name = "TWRP"
			A1.Archive.Twrp.Zip.Href = href
			A1.Archive.Twrp.Zip.Checksum_url_suffix = ".sha256"
			A1.Archive.Twrp.Zip.Filename = filename
			v, err := TwrpZipParseVersion(filename)
			if err != nil {
				return fmt.Errorf("unable to parse TWRP version in %s", filename)
			}
			A1.Archive.Twrp.Zip.Version = v
		}
	case "override_twrp":
		if strings.HasSuffix(filename, ".img") {
			A1.Archive.Override_twrp.Img.Name = "TWRP"
			A1.Archive.Override_twrp.Img.Href = href
			A1.Archive.Override_twrp.Img.Checksum_url_suffix = ".sha256"
			A1.Archive.Override_twrp.Img.Filename = filename
			v, err := TwrpZipParseVersion(filename)
			if err != nil {
				return fmt.Errorf("unable to parse TWRP version in %s", filename)
			}
			A1.Archive.Override_twrp.Img.Version = v
		} else if strings.HasSuffix(filename, ".zip") {
			A1.Archive.Override_twrp.Zip.Name = "TWRP"
			A1.Archive.Override_twrp.Zip.Href = href
			A1.Archive.Override_twrp.Zip.Checksum_url_suffix = ".sha256"
			A1.Archive.Override_twrp.Zip.Filename = filename
			v, err := TwrpZipParseVersion(filename)
			if err != nil {
				return fmt.Errorf("unable to parse TWRP version in %s", filename)
			}
			A1.Archive.Override_twrp.Zip.Version = v
		}
	case "logo", "startup-logo":
		A1.Archive.Logo.Href = href
		A1.Archive.Logo.Filename = filename
		A1.Archive.Logo.Checksum_url_suffix = ".sha256"
	case "flashme_pre":
		A1.Archive.Flashme_pre.Href = href
		A1.Archive.Flashme_pre.Filename = filename
		A1.Archive.Flashme_pre.Checksum_url_suffix = ".sha256"
	case "flashme_post":
		A1.Archive.Flashme_post.Href = href
		A1.Archive.Flashme_post.Filename = filename
		A1.Archive.Flashme_post.Checksum_url_suffix = ".sha256"
	default: // Assume rom
		if A1.Archive.Rom[folder] == nil {
			// Initialize item
			A1.Archive.Rom[folder] = &Item{}
		}
		if strings.Contains(strings.ToLower(filename), "lineage") && strings.Contains(strings.ToLower(filename), "microg") {
			A1.Archive.Rom[folder].Name = "LineageOSMicroG"
		} else {
			A1.Archive.Rom[folder].Name = folder
		}
		A1.Archive.Rom[folder].Href = href
		A1.Archive.Rom[folder].Checksum_url_suffix = ".sha256"
		A1.Archive.Rom[folder].Filename = filename
		A1.Archive.Rom[folder].Version = helpers.GenericParseVersion(filename)

		romname, av, err := GuessRomNameAndAndroidVersion(filename)
		if err != nil {
			return err
		}
		if romname != "" {
			A1.Archive.Rom[folder].Name = romname
		}
		if av != "" {
			A1.Archive.Rom[folder].Android_version = av
		}
	}

	return nil
}

func ArchiveLatestAvailableFileNamesMap(codename string) (map[string]string, error) {
	url := "https://archive.free-droid.com/" + codename

	folders, err := archiveAvailableFolders(codename)
	if err != nil {
		return map[string]string{}, err
	}

	availables := map[string]string{}

	for _, folder := range folders {
		folder_url := url + "/" + folder

		versions_available := make([]string, 0)

		c := colly.NewCollector()

		c.OnError(func(_ *colly.Response, err error) {
			logger.LogError("Error opening " + url + " :", err)
		})

		c.OnHTML("a", func(e *colly.HTMLElement) {
			versions_available = append(versions_available, e.Attr("href"))
		})

		c.Visit(folder_url)

		filter := func(s string) bool { return !strings.HasSuffix(s, "md5") && !strings.HasSuffix(s, "sha256") && !strings.HasSuffix(s, "sum") }
		versions_available_filtered := helpers.FilterStringSlice(versions_available, filter)

		sort.Strings(versions_available_filtered)

		latest_available := versions_available_filtered[len(versions_available_filtered)-1]

		availables[folder] = latest_available
	}

	return availables, nil
}

func archiveAvailableFolders(codename string) ([]string, error) {
	url := "https://archive.free-droid.com/" + codename

	status_code, err := StatusCode(url)
	if err != nil {
		return []string{}, err
	}
	if status_code != "200 OK" {
		return []string{}, fmt.Errorf("not available")
	}

	folders := make([]string, 0)

	c := colly.NewCollector()

	c.OnError(func(_ *colly.Response, err error) {
		logger.LogError("Error opening " + url + " :", err)
	})

	c.OnHTML("a", func(e *colly.HTMLElement) {
		if e.Text != "../" {	
			// strip the trailing slash before appending
			folders = append(folders, e.Text[:len(e.Text)-1])
		}
	})

	c.Visit(url)

	return folders, nil
}