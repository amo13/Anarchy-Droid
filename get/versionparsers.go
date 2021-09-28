package get

import (
	"fmt"
	"strings"

	"anarchy-droid/helpers"
)

func GuessRomNameAndAndroidVersion(filename string) (name string, av string, err error) {
	filename = strings.ToLower(filename)
	version := helpers.GenericParseVersion(filename)
	if strings.Contains(filename, "lineage") && strings.Contains(filename, "microg") {
		name = "LineageOSMicroG"
		av, err = LineageosParseAndroidVersion(version)
	} else if strings.Contains(filename, "lineage") && !strings.Contains(filename, "microg") {
		name = "LineageOS"
		av, err = LineageosParseAndroidVersion(version)
	} else if strings.HasPrefix(filename, "cm1") || strings.HasPrefix(filename, "cm-") || strings.HasPrefix(filename, "cyanogenmod") {
		name = "CyanogenMod"
		av, err = LineageosParseAndroidVersion(version)
	} else if strings.HasPrefix(filename, "resurrectionremix") || strings.HasPrefix(filename, "rr-") || strings.HasPrefix(filename, "rros-") {
		name = "ResurrectionRemix"
		av, err = ResurrectionRemixParseAndroidVersion(version)
	} else if strings.HasPrefix(filename, "aospextended") || strings.HasPrefix(filename, "aex") {
		name = "AospExtended"
		av, err = AospExtendedParseAndroidVersion(version)
	} else if strings.Contains(filename, "omnirom") {
		name = "Omnirom"
		av, err = OmniromParseAndroidVersion(version)
	} else if strings.Contains(filename, "carbon") {
		name = "Carbonrom"
		av, err = CarbonromParseAndroidVersion(version)
	} else if strings.Contains(filename, "crdroid") {
		name = "crDroid"
		av, err = CrDroidParseAndroidVersion(filename)	// takes filename, not romversion!
	} else if strings.Contains(filename, "calyx") {
		name = "CalyxOS"
		av, err = CalyxOsParseAndroidVersion(version)
	} else if strings.Contains(filename, "graphene") {
		name = "GrapheneOS"
		av, err = GrapheneOsParseAndroidVersion(version)
	} else if strings.Contains(filename, "pixelexperience") {
		name = "PixelExperience"
		av, err = PixelExperienceParseAndroidVersion(version)
	} else if strings.HasPrefix(filename, "mk") {
		name = "MoKee"
		av, err = MokeeParseAndroidVersion(version)
	} else if strings.HasPrefix(filename, "evolution") {
		name = "EvolutionX"
		av, err = EvolutionXParseAndroidVersion(version)
	} else {
		return "", "", fmt.Errorf("Unable to guess rom name or android version")
	}

	return name, av, err
}

func OmniromParseAndroidVersion(romversion string) (string, error) {
	return romversion, nil
}

func PixelExperienceParseAndroidVersion(romversion string) (string, error) {
	return romversion, nil
}

func EvolutionXParseAndroidVersion(romversion string) (string, error) {
	return "", fmt.Errorf("Unable to guess android version")
}

func CalyxOsParseAndroidVersion(romversion string) (string, error) {
	return "", fmt.Errorf("Unable to guess android version")
}

func GrapheneOsParseAndroidVersion(romversion string) (string, error) {
	return "", fmt.Errorf("Unable to guess android version")
}

func MokeeParseAndroidVersion(romversion string) (string, error) {
	if romversion != "" {
		s1 := strings.Split(romversion, ".")[0]	// "71.2"
		s2 := s1[:len(s1)-2]	// "71"
		s3 := s2[:len(s2)-1] + "." + s2[len(s2)-1:]	// "7.1"

		return s3, nil
	} else {
		return "", fmt.Errorf("Unable to tell android version of MoKee: %s", romversion)
	}
}