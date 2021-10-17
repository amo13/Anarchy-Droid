package get

import(
	"sync"
	"strings"

	"github.com/amo13/anarchy-droid/helpers"
	"github.com/amo13/anarchy-droid/logger"
)

var A1 = NewAvailable()

func NewAvailable() *Available {
	return &Available{
		Reloading: false,
		Mutex: &sync.Mutex{},
		Upstream: &Upstream{
			// Each rom distribution will be its own item
			Rom: make(map[string]*Item),
			Romlist: []string{},	// for selection on the gui
			Twrp: &Twrp{
				Img: &Item{},
				Zip: &Item{},
			},
			NanoDroid: make(map[string]*Item),
			// OpenGapps["arm"]["10.0"] --> [pico nano micro mini ...]
			OpenGapps: make(map[string]map[string][]string),
			Magisk: &Item{},
			CopyPartitions: &Item{},
		},
		Archive: &Archive{
			Rom: make(map[string]*Item),
			Romlist: []string{},	// for selection on the gui
			Twrp: &Twrp{
				Img: &Item{},
				Zip: &Item{},
			},
			Override_twrp: &Twrp{
				Img: &Item{},
				Zip: &Item{},
			},
			Logo: &Item{},
			Flashme_pre: &Item{},
			Flashme_post: &Item{},
		},
		User: &User{
			Rom: &Item{},
			Twrp: &Twrp{
				Img: &Item{},
				Zip: &Item{},
			},
			AdditionalZips: make([]string, 0),
		},
	}
}

type Available struct {
	Reloading bool
	Mutex *sync.Mutex
	Upstream *Upstream
	Archive *Archive
	User *User
}

type Upstream struct {
	Rom map[string]*Item
	Romlist []string
	Twrp *Twrp
	NanoDroid map[string]*Item
	OpenGapps map[string]map[string][]string
	Magisk *Item
	CopyPartitions *Item
}

type Archive struct {
	Rom map[string]*Item
	Romlist []string
	Twrp *Twrp
	Override_twrp *Twrp
	Logo *Item
	Flashme_pre *Item
	Flashme_post *Item
}

type User struct {
	Rom *Item
	Twrp *Twrp
	AdditionalZips []string
}

type Twrp struct {
	Img *Item
	Zip *Item
}

type Item struct {
	Name string
	Href string
	Filename string
	Version string
	Android_version string
	Checksum_url_suffix string
}

func (a *Available) CanFlash() bool {
	return a.User.Rom.Href != "" && a.User.Twrp.Img.Href != ""
}

type RetrievalError struct {
	What string
	Error error
}

func (a *Available) Populate(codename string) error {
	a.Reloading = true

	var wg sync.WaitGroup
	errs := make(chan RetrievalError)
	wg.Add(17)

	go func() {
		defer wg.Done()

		_, err := TwrpImgLatestAvailableHref(codename)
		if err != nil {
			errs <- RetrievalError{"TWRP-Img", err}
		}

		logger.Log("Finished looking for TWRP-Img")
	}()

	go func() {
		defer wg.Done()

		_, err := TwrpZipLatestAvailableHref(codename)
		if err != nil {
			errs <- RetrievalError{"TWRP-Zip", err}
		}

		logger.Log("Finished looking for TWRP-Zip")
	}()	

	go func() {
		defer wg.Done()

		_, err := NanoDroidLatestAvailableHref("full")
		if err != nil {
			errs <- RetrievalError{"NanoDroid-Full", err}
		}

		logger.Log("Finished looking for NanoDroid-Full")
	}()	

	go func() {
		defer wg.Done()

		_, err := NanoDroidLatestAvailableHref("fdroid")
		if err != nil {
			errs <- RetrievalError{"NanoDroid-F-Droid", err}
		}

		logger.Log("Finished looking for NanoDroid-F-Droid")
	}()	

	go func() {
		defer wg.Done()

		_, err := NanoDroidLatestAvailableHref("microG")
		if err != nil {
			errs <- RetrievalError{"NanoDroid-microG", err}
		}

		logger.Log("Finished looking for NanoDroid-microG")
	}()

	go func() {
		defer wg.Done()

		_, err := NanoDroidLatestAvailableHref("google")
		if err != nil {
			errs <- RetrievalError{"NanoDroid-Google", err}
		}

		logger.Log("Finished looking for NanoDroid-Google")
	}()

	go func() {
		defer wg.Done()

		_, err := NanoDroidLatestAvailableHref("patcher")
		if err != nil {
			errs <- RetrievalError{"NanoDroid-Patcher", err}
		}

		logger.Log("Finished looking for NanoDroid-Patcher")
	}()

	go func() {
		defer wg.Done()

		_, err := CopyPartitionsZipLatestAvailableHref()
		if err != nil {
			errs <- RetrievalError{"CopyPartitions", err}
		}

		logger.Log("Finished looking for CopyPartitions")
	}()

	go func() {
		defer wg.Done()

		err := OpenGappsPopulateAvailablesStruct()
		if err != nil {
			errs <- RetrievalError{"OpenGapps", err}
		}

		logger.Log("Finished looking for OpenGapps")
	}()

	go func() {
		defer wg.Done()

		_, err := ArchiveLatestAvailableHrefMap(codename)
		if err != nil {
			errs <- RetrievalError{"Archive", err}
		}

		logger.Log("Finished looking for Archive")
	}()

	go func() {
		defer wg.Done()

		_, err := LineageosLatestAvailableHref(codename)
		if err != nil {
			errs <- RetrievalError{"LineageOS", err}
		}

		logger.Log("Finished looking for LineageOS")
	}()

	go func() {
		defer wg.Done()

		_, err := LineageosMicrogLatestAvailableHref(codename)
		if err != nil {
			errs <- RetrievalError{"LineageOSMicroG", err}
		}

		logger.Log("Finished looking for LineageOSMicroG")
	}()

	go func() {
		defer wg.Done()

		_, err := CarbonromLatestAvailableHref(codename)
		if err != nil {
			errs <- RetrievalError{"Carbonrom", err}
		}

		logger.Log("Finished looking for Carbonrom")
	}()

	go func() {
		defer wg.Done()

		_, err := ResurrectionRemixLatestAvailableHref(codename)
		if err != nil {
			errs <- RetrievalError{"ResurrectionRemix", err}
		}

		logger.Log("Finished looking for ResurrectionRemix")
	}()

	go func() {
		defer wg.Done()

		_, err := CrDroidLatestAvailableHref(codename)
		if err != nil {
			errs <- RetrievalError{"CrDroid", err}
		}
		logger.Log("Finished looking for CrDroid")
	}()

	go func() {
		defer wg.Done()

		_, err := EOSLatestAvailableHref(codename)
		if err != nil {
			errs <- RetrievalError{"e-OS", err}
		}

		logger.Log("Finished looking for e-OS")
	}()

	go func() {
		defer wg.Done()

		_, err := AospExtendedLatestAvailableHref(codename)
		if err != nil {
			errs <- RetrievalError{"AospExtended", err}
		}

		logger.Log("Finished looking for AospExtended")
	}()

	go func() {
		wg.Wait()
		close(errs)
	}()

	for e := range errs {
		// Only send error to sentry if it is something else than "not found" or "not available"
		if !helpers.IsStringInSlice(strings.ToLower(e.Error.Error()), []string{"not available", "not found"}) {
			logger.LogError("Error retrieving latest available item from " + e.What + ":", e.Error)
		}
	}

	// Push names of available roms to Romlists
	for romname, _ := range a.Upstream.Rom {
		if romname == "LineageOSMicroG" || romname == "LineageOS" {
			if helpers.IsStringInSlice("LineageOS", a.Upstream.Romlist) {
				continue
			}
			romname = "LineageOS"
		}
		a.Mutex.Lock()
		a.Upstream.Romlist = append(a.Upstream.Romlist, romname)
		a.Mutex.Unlock()
	}
	for romname, _ := range a.Archive.Rom {
		if romname == "LineageOSMicroG" || romname == "LineageOS" {
			if helpers.IsStringInSlice("LineageOS", a.Archive.Romlist) {
				continue
			}
			romname = "LineageOS"
		}
		a.Mutex.Lock()
		a.Archive.Romlist = append(a.Archive.Romlist, romname)
		a.Mutex.Unlock()
	}

	a.Reloading = false

	logger.Log("Finished reloading roms.")

	return nil
}

func (a *Available) PopulateForApi(codename string) error {
	var wg sync.WaitGroup
	errs := make(chan RetrievalError)
	wg.Add(8)

	go func() {
		defer wg.Done()

		h, err := TwrpImgLatestAvailableHref(codename)
		if err != nil {
			errs <- RetrievalError{"TWRP-Img", err}
		}

		if h != "" {
			a.Mutex.Lock()
			defer a.Mutex.Unlock()
			a.Upstream.Twrp.Img.Href = h
		}
	}()

	go func() {
		defer wg.Done()

		_, err := a.PopulateArchive(codename)
		if err != nil {
			errs <- RetrievalError{"Archive", err}
		}
	}()

	go func() {
		defer wg.Done()

		h, err := LineageosLatestAvailableHref(codename)
		if err != nil {
			errs <- RetrievalError{"LineageOS", err}
		}

		if h != "" {
			a.Mutex.Lock()
			defer a.Mutex.Unlock()
			a.Upstream.Romlist = append(a.Upstream.Romlist, "LineageOS")
			a.Upstream.Rom["LineageOS"] = &Item{}
			a.Upstream.Rom["LineageOS"].Href = h
		}
	}()

	go func() {
		defer wg.Done()

		h, err := LineageosMicrogLatestAvailableHref(codename)
		if err != nil {
			errs <- RetrievalError{"LineageOSMicroG", err}
		}

		if h != "" {
			a.Mutex.Lock()
			defer a.Mutex.Unlock()
			a.Upstream.Romlist = append(a.Upstream.Romlist, "LineageOSMicroG")
			a.Upstream.Rom["LineageOSMicroG"] = &Item{}
			a.Upstream.Rom["LineageOSMicroG"].Href = h
		}
	}()

	go func() {
		defer wg.Done()

		h, err := CarbonromLatestAvailableHref(codename)
		if err != nil {
			errs <- RetrievalError{"Carbonrom", err}
		}

		if h != "" {
			a.Mutex.Lock()
			defer a.Mutex.Unlock()
			a.Upstream.Romlist = append(a.Upstream.Romlist, "Carbonrom")
			a.Upstream.Rom["Carbonrom"] = &Item{}
			a.Upstream.Rom["Carbonrom"].Href = h
		}
	}()

	go func() {
		defer wg.Done()

		h, err := ResurrectionRemixLatestAvailableHref(codename)
		if err != nil {
			errs <- RetrievalError{"ResurrectionRemix", err}
		}

		if h != "" {
			a.Mutex.Lock()
			defer a.Mutex.Unlock()
			a.Upstream.Romlist = append(a.Upstream.Romlist, "ResurrectionRemix")
			a.Upstream.Rom["ResurrectionRemix"] = &Item{}
			a.Upstream.Rom["ResurrectionRemix"].Href = h
		}
	}()

	go func() {
		defer wg.Done()

		h, err := CrDroidLatestAvailableHref(codename)
		if err != nil {
			errs <- RetrievalError{"CrDroid", err}
		}

		if h != "" {
			a.Mutex.Lock()
			defer a.Mutex.Unlock()
			a.Upstream.Romlist = append(a.Upstream.Romlist, "CrDroid")
			a.Upstream.Rom["CrDroid"] = &Item{}
			a.Upstream.Rom["CrDroid"].Href = h
		}
	}()

	go func() {
		defer wg.Done()

		h, err := EOSLatestAvailableHref(codename)
		if err != nil {
			errs <- RetrievalError{"e-OS", err}
		}

		if h != "" {
			a.Mutex.Lock()
			defer a.Mutex.Unlock()
			a.Upstream.Romlist = append(a.Upstream.Romlist, "e-OS")
			a.Upstream.Rom["e-OS"] = &Item{}
			a.Upstream.Rom["e-OS"].Href = h
		}
	}()

	go func() {
		wg.Wait()
		close(errs)
	}()

	for e := range errs {
		// Only send error to sentry if it is something else than "not found" or "not available"
		if !helpers.IsStringInSlice(strings.ToLower(e.Error.Error()), []string{"not available", "not found"}) {
			logger.LogError("API-Server: Error retrieving latest available item from " + e.What + ":", e.Error)
		}
	}

	// Push names of available roms to Archive romlist
	for romname, _ := range a.Archive.Rom {
		if romname == "LineageOSMicroG" || romname == "LineageOS" {
			if helpers.IsStringInSlice("LineageOS", a.Archive.Romlist) {
				continue
			}
			romname = "LineageOS"
		}
		a.Mutex.Lock()
		a.Archive.Romlist = append(a.Archive.Romlist, romname)
		a.Mutex.Unlock()
	}

	return nil
}

func (a *Available) Print() {
	logger.Log("Upstream:")
	logger.Log("  Roms:")
	for rom, _ := range a.Upstream.Rom {
		logger.Log("   ", rom, ":")
		logger.Log("      name:", a.Upstream.Rom[rom].Name)
		logger.Log("      href:", a.Upstream.Rom[rom].Href)
		logger.Log("      file:", a.Upstream.Rom[rom].Filename)
		logger.Log("      ver :", a.Upstream.Rom[rom].Version)
		logger.Log("      Aver:", a.Upstream.Rom[rom].Android_version)
	}
	logger.Log("  Romlist:", strings.Join(a.Upstream.Romlist[:], " "))
	logger.Log("  TWRP Img:")
	logger.Log("    href:", a.Upstream.Twrp.Img.Href)
	logger.Log("    file:", a.Upstream.Twrp.Img.Filename)
	logger.Log("    ver :", a.Upstream.Twrp.Img.Version)
	logger.Log("  TWRP Zip:")
	logger.Log("    href:", a.Upstream.Twrp.Zip.Href)
	logger.Log("    file:", a.Upstream.Twrp.Zip.Filename)
	logger.Log("    ver :", a.Upstream.Twrp.Zip.Version)
	logger.Log("  NanoDroid:")
	for module, _ := range a.Upstream.NanoDroid {
		logger.Log("   ", module, ":")
		logger.Log("      href:", a.Upstream.NanoDroid[module].Href)
		logger.Log("      file:", a.Upstream.NanoDroid[module].Filename)
		logger.Log("      ver :", a.Upstream.NanoDroid[module].Version)
	}
	
	// Do not print OpenGapps to prevent flood

	logger.Log("Archive:")
	logger.Log("  Roms:")
	for rom, _ := range a.Archive.Rom {
		logger.Log("   ", rom, ":")
		logger.Log("      name:", a.Archive.Rom[rom].Name)
		logger.Log("      href:", a.Archive.Rom[rom].Href)
		logger.Log("      file:", a.Archive.Rom[rom].Filename)
		logger.Log("      ver :", a.Archive.Rom[rom].Version)
		logger.Log("      Aver:", a.Archive.Rom[rom].Android_version)
	}
	logger.Log("  Romlist:", strings.Join(a.Archive.Romlist[:], " "))
	logger.Log("  TWRP Img:")
	logger.Log("    href:", a.Archive.Twrp.Img.Href)
	logger.Log("    file:", a.Archive.Twrp.Img.Filename)
	logger.Log("    ver :", a.Archive.Twrp.Img.Version)
	logger.Log("  TWRP Zip:")
	logger.Log("    href:", a.Archive.Twrp.Zip.Href)
	logger.Log("    file:", a.Archive.Twrp.Zip.Filename)
	logger.Log("    ver :", a.Archive.Twrp.Zip.Version)
	logger.Log("  TWRP Img Override:")
	logger.Log("    href:", a.Archive.Override_twrp.Img.Href)
	logger.Log("    file:", a.Archive.Override_twrp.Img.Filename)
	logger.Log("    ver :", a.Archive.Override_twrp.Img.Version)
	logger.Log("  TWRP Zip Override:")
	logger.Log("    href:", a.Archive.Override_twrp.Zip.Href)
	logger.Log("    file:", a.Archive.Override_twrp.Zip.Filename)
	logger.Log("    ver :", a.Archive.Override_twrp.Zip.Version)
	logger.Log("  Logo:")
	logger.Log("    href:", a.Archive.Logo.Href)
	logger.Log("    file:", a.Archive.Logo.Filename)
	logger.Log("    ver :", a.Archive.Logo.Version)
	logger.Log("  Flashme_pre:")
	logger.Log("    href:", a.Archive.Flashme_pre.Href)
	logger.Log("    file:", a.Archive.Flashme_pre.Filename)
	logger.Log("    ver :", a.Archive.Flashme_pre.Version)
	logger.Log("  Flashme_post:")
	logger.Log("    href:", a.Archive.Flashme_post.Href)
	logger.Log("    file:", a.Archive.Flashme_post.Filename)
	logger.Log("    ver :", a.Archive.Flashme_post.Version)

	logger.Log("User:")
	logger.Log("  Rom:")
	logger.Log("    name:", a.User.Rom.Name)
	logger.Log("      href:", a.User.Rom.Href)
	logger.Log("      file:", a.User.Rom.Filename)
	logger.Log("      ver :", a.User.Rom.Version)
	logger.Log("      Aver:", a.User.Rom.Android_version)
	logger.Log("  TWRP Img:")
	logger.Log("    href:", a.User.Twrp.Img.Href)
	logger.Log("    file:", a.User.Twrp.Img.Filename)
	logger.Log("    ver :", a.User.Twrp.Img.Version)
	logger.Log("  TWRP Zip:")
	logger.Log("    href:", a.User.Twrp.Zip.Href)
	logger.Log("    file:", a.User.Twrp.Zip.Filename)
	logger.Log("    ver :", a.User.Twrp.Zip.Version)
}