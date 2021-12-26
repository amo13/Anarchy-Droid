package lookup

import(
	"fmt"
	"strings"
	"strconv"

	"github.com/amo13/anarchy-droid/get"
	"github.com/amo13/anarchy-droid/logger"
	"github.com/amo13/anarchy-droid/helpers"
	"github.com/amo13/anarchy-droid/device/adb"
	"github.com/amo13/anarchy-droid/device/fastboot"
)

var AliasYamlMap map[string]string
var CodenameToBrandYamlMap map[string]string
var ModelToCodenameYamlMap map[string]string
var SupportedYamlMap map[string]string
var RecoveryPartitionYamlMap map[string]string
var RecoveryKeyCombinationYamlMap map[string]string
var BootloaderKeyCombinationYamlMap map[string]string
var DeviceLookupCsvPath string = "bin/device-lookup.csv"
var DeviceLookupCsvLines [][]string
var CodenameSuffixes []string = []string{"_n", "_f", "_t", "_ds", "_nt", "_u2", "_ud2", "_uds", "_cdma", "_umts", "_udstv", "_umtsds"}

// Try to lookup in yaml first and CSV then
func ModelToCodename(model string) (string, error) {
	y, err := modelToCodenameYaml(model)
	if err != nil {
		return "", err
	}

	if y == "" {
		c, err := modelToCodenameCsv(model)
		if err != nil {
			return "", err
		}

		y = c
	}

	// if y == "" && adb.IsConnected() {
	// 	y = adb.Codename()
	// }

	// if strings.HasPrefix(y, "omni_") {
	// 	y = y[5:]
	// } else if strings.HasPrefix(y, "lineage_") {
	// 	y = y[8:]
	// }

	// // Remove trailing _ds, _cdma and similar from the codenames.
	// for i := range CodenameSuffixes {
	// 	if strings.HasSuffix(y, CodenameSuffixes[i]) {
	// 		y = y[:len(y)-len(CodenameSuffixes[i])]
	// 	}
	// }

	return y, nil
}

// Returns a slice of codename candidates for a given model
// The returned slice may contain 0, 1 or more elements
func ModelToCodenameCandidatesForApi(model string) ([]string, error) {
	y, err := modelToCodenameYaml(model)
	if err != nil {
		return []string{}, err
	}

	// Do not lookup in the CSV if the yml has an entry
	// In that case return this single codename in a slice
	if y != "" {
		return []string{y}, nil
	}

	c, err := modelToCodenameCandidatesCsv(model)
	if err != nil {
		return []string{}, err
	}

	matches := append([]string{y}, c...)

	return helpers.UniqueNonEmptyElementsOfSlice(matches), nil
}

// Returns codename candidates from device lookup CSV
// (Yaml table is unambiguous and takes precedence anyway)
func ModelToCodenameCandidates(model string) ([]string, error) {
	return modelToCodenameCandidatesCsv(model)
}

func modelToCodenameYaml(model string) (string, error) {
	m, err := modelToCodenameYamlMap()
	if err != nil {
		return "", err
	}

	return m[strings.ToLower(model)], nil
}

func modelToCodenameYamlMap() (map[string]string, error) {
	if ModelToCodenameYamlMap != nil && len(ModelToCodenameYamlMap) > 0 {
		return ModelToCodenameYamlMap, nil
	}

	content, err := helpers.ReadFromURL("https://raw.githubusercontent.com/amo13/Anarchy-Droid/master/lookup/codenames.yml")
	if err != nil {
		return nil, err
	}

	m, err := helpers.YamlToDowncaseFlatMap(content)
	if err != nil {
		return nil, err
	}

	ModelToCodenameYamlMap = m
	return ModelToCodenameYamlMap, nil
}

func CodenamesToModels(codenames []string) ([]string, error) {
	models := []string{}

	for _, codename := range codenames {
		y, err := codenameToModelsYaml(codename)
		if err != nil {
			return []string{}, err
		}

		if len(y) == 0 {
			c, err := codenameToModelsCsv(codename)
			if err != nil {
				return []string{}, err
			}

			y = c
		}

		for _, model := range y {
			models = append(models, model)
		}
	}

	return helpers.UniqueNonEmptyElementsOfSlice(models), nil
}

func codenameToModelsYaml(codename string) ([]string, error) {
	m, err := modelToCodenameYamlMap()
	if err != nil {
		return []string{}, err
	}

	matches := []string{}

	for k, v := range m {
		if codename == v {
			matches = append(matches, k)
		}
	}

	return helpers.UniqueNonEmptyElementsOfSlice(matches), nil
}

// Try to lookup in yaml first and CSV then
// Resort to the adb prop if unsuccessful
func CodenameToBrand(codename string) (string, error) {
	y, err := codenameToBrandYaml(codename)
	if err != nil {
		return "", err
	}

	if y == "" {
		c, err := codenameToBrandCsv(codename)
		if err != nil {
			return "", err
		}

		y = c
	}

	if y == "" && adb.IsConnected() {
		y, err = adb.Brand()
		if err != nil {
			logger.LogError("CodenameToBrand: unable to query ADB for device brand", err)
			return "", err
		}
	}

	return y, nil
}

// Try to lookup in yaml first and CSV then
// Do not resort to the adb prop if unsuccessful
func CodenameToBrandForApi(codename string) (string, error) {
	y, err := codenameToBrandYaml(codename)
	if err != nil {
		return "", err
	}

	if y == "" {
		c, err := codenameToBrandCsv(codename)
		if err != nil {
			return "", err
		}

		y = c
	}

	return y, nil
}

func codenameToBrandYaml(codename string) (string, error) {
	m, err := codenameToBrandYamlMap()
	if err != nil {
		return "", err
	}

	return m[strings.ToLower(codename)], nil
}

func codenameToBrandYamlMap() (map[string]string, error) {
	if CodenameToBrandYamlMap != nil && len(CodenameToBrandYamlMap) > 0 {
		return CodenameToBrandYamlMap, nil
	}

	content, err := helpers.ReadFromURL("https://raw.githubusercontent.com/amo13/Anarchy-Droid/master/lookup/brands.yml")
	if err != nil {
		return nil, err
	}

	m, err := helpers.YamlToDowncaseFlatMap(content)
	if err != nil {
		return nil, err
	}

	CodenameToBrandYamlMap = m
	return CodenameToBrandYamlMap, nil
}

// Returns true if the support state is unknown
func IsSupported(codename string) (bool, error) {
	m, err := supportedYamlMap()
	if err != nil {
		return false, err
	}

	supported := m[strings.ToLower(codename)]
	if supported == "" {
		supported = "true"
	}

	supported_bool, err := strconv.ParseBool(supported)
	if err != nil {
		return false, err
	}

	return supported_bool, nil
}

func supportedYamlMap() (map[string]string, error) {
	if SupportedYamlMap != nil && len(SupportedYamlMap) > 0 {
		return SupportedYamlMap, nil
	}

	content, err := helpers.ReadFromURL("https://raw.githubusercontent.com/amo13/Anarchy-Droid/master/lookup/supported.yml")
	if err != nil {
		return nil, err
	}

	m, err := helpers.YamlToDowncaseFlatMap(content)
	if err != nil {
		return nil, err
	}

	SupportedYamlMap = m
	return SupportedYamlMap, nil
}

func RecoveryPartition(codename string) (string, error) {
	return recoveryPartitionYaml(codename)
}

func recoveryPartitionYaml(codename string) (string, error) {
	m, err := recoveryPartitionYamlMap()
	if err != nil {
		return "", err
	}

	return m[strings.ToLower(codename)], nil
}

func recoveryPartitionYamlMap() (map[string]string, error) {
	if RecoveryPartitionYamlMap != nil && len(RecoveryPartitionYamlMap) > 0 {
		return RecoveryPartitionYamlMap, nil
	}

	content, err := helpers.ReadFromURL("https://raw.githubusercontent.com/amo13/Anarchy-Droid/master/lookup/recovery_partition_names.yml")
	if err != nil {
		return nil, err
	}

	m, err := helpers.YamlToFlatMap(content)
	if err != nil {
		return nil, err
	}

	RecoveryPartitionYamlMap = m
	return RecoveryPartitionYamlMap, nil
}

func RecoveryKeyCombination(codename_or_brand string) (string, error) {
	return recoveryKeyCombinationYaml(codename_or_brand)
}

func recoveryKeyCombinationYaml(codename_or_brand string) (string, error) {
	m, err := recoveryKeyCombinationYamlMap()
	if err != nil {
		return "", err
	}

	return m[strings.ToLower(codename_or_brand)], nil
}

func recoveryKeyCombinationYamlMap() (map[string]string, error) {
	if RecoveryKeyCombinationYamlMap != nil && len(RecoveryKeyCombinationYamlMap) > 0 {
		return RecoveryKeyCombinationYamlMap, nil
	}

	content, err := helpers.ReadFromURL("https://raw.githubusercontent.com/amo13/Anarchy-Droid/master/lookup/recovery_key_combinations.yml")
	if err != nil {
		return nil, err
	}

	m, err := helpers.YamlToFlatMap(content)
	if err != nil {
		return nil, err
	}

	RecoveryKeyCombinationYamlMap = m
	return RecoveryKeyCombinationYamlMap, nil
}

func BootloaderKeyCombination(codename_or_brand string) (string, error) {
	return bootloaderKeyCombinationYaml(codename_or_brand)
}

func bootloaderKeyCombinationYaml(codename_or_brand string) (string, error) {
	m, err := bootloaderKeyCombinationYamlMap()
	if err != nil {
		return "", err
	}

	return m[strings.ToLower(codename_or_brand)], nil
}

func bootloaderKeyCombinationYamlMap() (map[string]string, error) {
	if BootloaderKeyCombinationYamlMap != nil && len(BootloaderKeyCombinationYamlMap) > 0 {
		return BootloaderKeyCombinationYamlMap, nil
	}

	content, err := helpers.ReadFromURL("https://raw.githubusercontent.com/amo13/Anarchy-Droid/master/lookup/bootloader_key_combinations.yml")
	if err != nil {
		return nil, err
	}

	m, err := helpers.YamlToFlatMap(content)
	if err != nil {
		return nil, err
	}

	BootloaderKeyCombinationYamlMap = m
	return BootloaderKeyCombinationYamlMap, nil
}

func Alias(codename string) (string, error) {
	return aliasYaml(codename)
}

func aliasYaml(codename string) (string, error) {
	m, err := aliasYamlMap()
	if err != nil {
		return "", err
	}

	return m[strings.ToLower(codename)], nil
}

func aliasYamlMap() (map[string]string, error) {
	if AliasYamlMap != nil && len(AliasYamlMap) > 0 {
		return AliasYamlMap, nil
	}

	content, err := helpers.ReadFromURL("https://raw.githubusercontent.com/amo13/Anarchy-Droid/master/lookup/aliases.yml")
	if err != nil {
		return nil, err
	}

	m, err := helpers.YamlToDowncaseFlatMap(content)
	if err != nil {
		return nil, err
	}

	AliasYamlMap = m
	return AliasYamlMap, nil
}

// Check if the given model name already is the codename
func IsCodename(model string) (bool, error) {
	// Check in codename yaml
	cm, err := modelToCodenameYamlMap()
	if err != nil {
		return false, err
	}

	for _, v := range cm {
		if strings.ToLower(v) == strings.ToLower(model) {
			return true, nil
		}
	}

	// Check in Device Lookup CSV
	table, err := lookupCsvToTable()
	if err != nil {
		return false, err
	}

	for line := range table {
		if strings.ToLower(table[line][2]) == strings.ToLower(model) {
			return true, nil
		}
	}

	return false, nil
}

func dlDeviceLookupCsv() error {
	err := get.DownloadFile(DeviceLookupCsvPath, "http://storage.googleapis.com/play_public/supported_devices.csv", "")
	if err != nil {
		return err
	}

	return nil
}

func lookupCsvToTable() ([][]string, error) {
	if DeviceLookupCsvLines != nil {
		return DeviceLookupCsvLines, nil
	}

	err := dlDeviceLookupCsv()
	if err != nil {
		return make([][]string, 0), err
	}

	table, err := helpers.CsvUTF16ToSlice(DeviceLookupCsvPath)
	if err != nil {
		return make([][]string, 0), err
	}

	return table, nil
}

func queryDeviceLookupCsvTable(item string, match_in_column int, lookup_from_column int) ([]string, error) {
	table, err := lookupCsvToTable()
	if err != nil {
		return []string{}, err
	}

	matches := make([]string, 0)

	for line := range table {
		if strings.ToLower(table[line][match_in_column]) == strings.ToLower(item) {
			matches = append(matches, strings.ToLower(table[line][lookup_from_column]))
		}
	}

	return matches, nil
}

// Returns error if ambiguous
// examples: "Moto G Play", "moto z4"
func modelToCodenameCsv(model string) (string, error) {
	// Check if the given model name already is the codename
	ic, err := IsCodename(model)
	if err != nil {
		return "", err
	}
	if ic {
		return model, nil
	}

	// Lookup the codenames matching the given model
	matches, err := modelToCodenameCandidatesCsv(model)
	if err != nil {
		return "", err
	}

	// If multiple codenames result and they all start with one of them
	// then take the shortest one of them. Ex. "gts28wifi" and "gts28wifichn"
	// If only one codename matched, return this one as string
	result, err := helpers.PrefixOfAll(matches)
	if err != nil {
		// Triggers if ambiguous
		return "", err
	}

	// Look for matches in the ADB props and/or fastboot vars of the device
	// if there are still multiple codename candidates
	if result == "" {
		matchedmatches := make([]string, 0)
		adb_state := adb.State()
		if helpers.IsStringInSlice(adb_state, []string{"android", "recovery"}) {
			// look for a match in the ADB props
			props, err := adb.GetPropMap()
			if err != nil {
				return "", err
			}

			for _, match := range matches {
				for _, prop := range props {
					if strings.Contains(prop, match) {
						matchedmatches = append(matchedmatches, match)
					}
				}
			}

			matchedmatches = helpers.UniqueNonEmptyElementsOfSlice(matchedmatches)

			// If multiple codenames result and they all start with one of them
			// then take the shortest one of them. Ex. "gts28wifi" and "gts28wifichn"
			// If only one codename matched, return this one as string
			result, err = helpers.PrefixOfAll(matchedmatches)
			if err != nil {
				// Triggers if ambiguous
				return "", err
			}
		} else if fastboot.State() == "connected" {
			// look for a match in the fastboot vars
			vars, err := fastboot.GetVarMap()
			if err != nil {
				return "", err
			}

			for _, match := range matches {
				for _, v := range vars {
					if strings.Contains(v, match) {
						matchedmatches = append(matchedmatches, match)
					}
				}
			}

			matchedmatches = helpers.UniqueNonEmptyElementsOfSlice(matchedmatches)

			// If multiple codenames result and they all start with one of them
			// then take the shortest one of them. Ex. "gts28wifi" and "gts28wifichn"
			// If only one codename matched, return this one as string
			result, err = helpers.PrefixOfAll(matchedmatches)
			if err != nil {
				// Triggers if ambiguous
				return "", err
			}
		} else {
			logger.Log("Unable to query adb or fastboot for props or vars to check for matches with one of the codename candidates")
			result = ""
		}
	}

	if result == "" {
		if len(matches) > 1 {
			return "", fmt.Errorf("ambiguous")
		} else {
			return "", fmt.Errorf("Unable to lookup codename for model %s with CSV", model)
		}
	} else {
		return result, nil
	}
}

func modelToCodenameCandidatesCsv(model string) ([]string, error) {
	// Lookup the codenames matching the given model
	matches, err := queryDeviceLookupCsvTable(model, 3, 2)
	if err != nil {
		return []string{}, err
	}

	if len(matches) == 0 {
		matches, err = queryDeviceLookupCsvTable(model, 1, 2)
		if err != nil {
			return []string{}, err
		}
	}

	// Remove trailing _ds, _cdma and similar from the codenames.
	for i := range CodenameSuffixes {
		for j := range matches {
			if strings.HasSuffix(matches[j], CodenameSuffixes[i]) {
				matches[j] = matches[j][:len(matches[j])-len(CodenameSuffixes[i])]
			}
		}
	}

	return helpers.UniqueNonEmptyElementsOfSlice(matches), nil
}

func CodenameToNamesCsv(codename string) ([]string, error) {
	matches, err := queryDeviceLookupCsvTable(codename, 2, 1)
	if err != nil {
		return []string{}, err
	}

	// Retry with codename suffixes if nothing was found
	if len(matches) == 0 {
		for _, suffix := range CodenameSuffixes {
			matches, err = queryDeviceLookupCsvTable(codename + suffix, 2, 1)
		}
	}

	return helpers.UniqueNonEmptyElementsOfSlice(matches), nil
}

// Tries to return a unique marketing name with a given codename
// If multiple name candidates result and they all start with one of them
// then returns the shortest one of them. Ex. "galaxy s5" and "galaxy s5 dual sim"
// Returns error if ambiguous
func CodenameToNameCsv(codename string) (string, error) {
	candidates, err := CodenameToNamesCsv(codename)
	if err != nil {
		return "", err
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("not found")
	} else if len(candidates) == 1 {
		return candidates[0], nil
	} else {
		r, err := helpers.PrefixOfAll(candidates)
		if err != nil {
			// Triggers if ambiguous
			return "", err
		}
		return r, nil
	}
}

func codenameToModelsCsv(codename string) ([]string, error) {
	// Lookup the models matching the given codename
	matches, err := queryDeviceLookupCsvTable(codename, 2, 3)
	if err != nil {
		return []string{}, err
	}

	// Retry with codename suffixes if nothing was found
	if len(matches) == 0 {
		for _, suffix := range CodenameSuffixes {
			matches, err = queryDeviceLookupCsvTable(codename + suffix, 2, 3)
			if err != nil {
				return []string{}, err
			}
		}
	}

	return helpers.UniqueNonEmptyElementsOfSlice(matches), nil
}

func codenameToBrandCsv(codename string) (string, error) {
	// Lookup the brands matching the given codename
	matches, err := queryDeviceLookupCsvTable(codename, 2, 0)
	if err != nil {
		return "", err
	}

	// Retry with codename suffixes if nothing was found
	if len(matches) == 0 {
		for _, suffix := range CodenameSuffixes {
			matches_suffix, err := queryDeviceLookupCsvTable(codename + suffix, 2, 0)
			if err != nil {
				return "", err
			}

			if len(matches_suffix) > 0 {
				matches = append(matches, matches_suffix...)
			}
		}
	}

	matches = helpers.UniqueNonEmptyElementsOfSlice(matches)

	if len(matches) == 0 {
		return "", fmt.Errorf("Brand of " + codename + " not found")
	} else if len(matches) == 1 {
		return matches[0], nil
	} else {
		return "", fmt.Errorf("Brand of " + codename + " is ambiguous")
	}
}