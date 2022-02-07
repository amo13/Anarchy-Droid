package helpers

import (
	"os"
    "io"
	"fmt"
	"bufio"
    "bytes"
    "regexp"
    "os/exec"
    "runtime"
    "strings"
    "net/http"
    "io/ioutil"
    "archive/zip"
    "encoding/csv"
    "path/filepath"
    "gopkg.in/yaml.v3"
    "golang.org/x/text/transform"
    "golang.org/x/text/encoding/unicode"

    "github.com/amo13/anarchy-droid/logger"
)

func Cmd(command string, args ...string) (stdout string, stderr string) {
    if strings.HasPrefix(command, "sudo ") || strings.Contains(command, "| sudo ") {
        return Cmd("/bin/sh", "-c", command + " " + strings.Join(args, " "))
    }

    c := exec.Command(command, args...)

    cOut, err := c.StdoutPipe()
    if err != nil {
        logger.LogError("Error binding STDOUT for command", err)
    }
    cErr, err := c.StderrPipe()
    if err != nil {
        logger.LogError("Error binding STDERR for command", err)
    }

    err = c.Start()
    // Do not send a bug report containing a sudo password
    if err != nil && !strings.Contains(err.Error(), "sudo ") {
        logger.LogError("Could not execute command " + command + ":", err)
    }

    outBytes, err := io.ReadAll(cOut)
    // Do not send a bug report containing a sudo password
    if err != nil && !strings.Contains(err.Error(), "sudo ") {
        logger.LogError("Unable to read STDOUT", err)
    }
    errBytes, err := io.ReadAll(cErr)
    // Do not send a bug report containing a sudo password
    if err != nil && !strings.Contains(err.Error(), "sudo ") {
        logger.LogError("Unable to read STDERR", err)
    }

    c.Wait()

    return string(outBytes), string(errBytes)
}

func ReadFromURL(url string) ([]byte, error) {
    client := &http.Client{}

    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        logger.LogError("Unable to create http request for " + url, err)
        return nil, err
    }

    req.Header.Set("User-Agent", "Mozilla/5.0 (platform; rv:geckoversion) Gecko/geckotrail Firefox/firefoxversion Anarchy-Droid/current")

    resp, err := client.Do(req)
    if err != nil {
        logger.LogError("Unable to get http response for " + url, err)
        return nil, err
    }
    defer resp.Body.Close()
    
    content, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    return content, nil
}

func StringToLinesSlice(s string) []string {
    windows := strings.Replace(s, "\n", "\r\n", -1)
    return strings.Split(strings.Trim(strings.Replace(windows, "\r\n", "\n", -1), "\n"), "\n")
}

// Filter a slice of strings. Use an anonymous function as selector, e.g.:
// func(s string) bool { return strings.HasPrefix(s, "foo_") && len(s) <= 7 }
func FilterStringSlice(ss []string, test func(string) bool) (ret []string) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}

func FirstWordInFile(file_path string) (word string, err error) {
    file, err := os.Open(file_path)
    if err != nil {
        return "", err
    }
    scanner := bufio.NewScanner(file)

    scanner.Split(bufio.ScanWords)

    // Scan for next token. 
    success := scanner.Scan() 
    if success == false {
        // False on error or EOF. Check error
        err = scanner.Err()
        if err != nil {
            return "", err
        }
        return "", fmt.Errorf("Unable to scan the first word in %s", file_path)
    }

    // Get data from scan with Bytes() or Text()
    return scanner.Text(), nil
}

func Unzip(src, dest string) error {
    dest = filepath.Clean(dest) + string(os.PathSeparator)

    r, err := zip.OpenReader(src)
    if err != nil {
        return err
    }
    defer func() {
        if err := r.Close(); err != nil {
            panic(err)
        }
    }()

    os.MkdirAll(dest, 0755)

    // Closure to address file descriptors issue with all the deferred .Close() methods
    extractAndWriteFile := func(f *zip.File) error {
        path := filepath.Join(dest, f.Name)
        // Check for ZipSlip: https://snyk.io/research/zip-slip-vulnerability
        if !strings.HasPrefix(path, dest) {
            return fmt.Errorf("%s: illegal file path", path)
        }

        rc, err := f.Open()
        if err != nil {
            return err
        }
        defer func() {
            if err := rc.Close(); err != nil {
                panic(err)
            }
        }()

        if f.FileInfo().IsDir() {
            os.MkdirAll(path, 0755)
        } else {
            os.MkdirAll(filepath.Dir(path), 0755)
            f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
            if err != nil {
                return err
            }
            defer func() {
                if err := f.Close(); err != nil {
                    panic(err)
                }
            }()

            _, err = io.Copy(f, rc)
            if err != nil {
                return err
            }
        }
        return nil
    }

    for _, f := range r.File {
        err := extractAndWriteFile(f)
        if err != nil {
            return err
        }
    }

    return nil
}

func IsStringInSlice(a string, list []string) bool {
    for _, b := range list {
        if b == a {
            return true
        }
    }
    return false
}

func RemoveStringFromSlice(s []string, r string) []string {
    for i, v := range s {
        if v == r {
            return append(s[:i], s[i+1:]...)
        }
    }
    return s
}

func MapToString(m map[string]string) string {
    b := new(bytes.Buffer)
    for key, value := range m {
        fmt.Fprintf(b, "\"%s\"=\"%s\"\n", key, value)
    }
    return b.String()
}

func KeysOfMap(m map[string]string) []string {
    keys := make([]string, 0, len(m))
    for key := range m {
        keys = append(keys, key)
    }
    return keys
}

func YamlToFlatMap(yamldata []byte) (map[string]string, error) {
    themap := make(map[string]string)
    err := yaml.Unmarshal(yamldata, &themap)
    if err != nil {
        return nil, err
    }

    return themap, nil
}

// Only downcase the keys, not the values
func YamlToDowncaseFlatMap(yamldata []byte) (map[string]string, error) {
    orig, err := YamlToFlatMap(yamldata)
    if err != nil {
        return nil, err
    }

    downcasecopy := make(map[string]string)
    for k, v := range orig {
        downcasecopy[strings.ToLower(k)] = v
    }

    return downcasecopy, nil
}

func CsvUTF16ToSlice(csv_path string) ([][]string, error)  {
    content, err := ioutil.ReadFile("bin/device-lookup.csv")
    if err != nil {
        return make([][]string, 0), err
    }

    var lines [][]string

    dec := unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder()
    utf16r := bytes.NewReader(content)
    utf8r := transform.NewReader(utf16r, dec)

    r := csv.NewReader(utf8r)

    // read the header
    if _, err := r.Read(); err != nil {
        return make([][]string, 0), err
    }
    for {
        rec, err := r.Read()
        if err != nil {
            if err == io.EOF {
                break
            }
            continue
        }

        lines = append(lines, rec)
    }

    return lines, nil
}

func UniqueNonEmptyElementsOfSlice(s []string) []string {
  unique := make(map[string]bool, len(s))
    us := make([]string, len(unique))
    for _, elem := range s {
        if len(elem) != 0 {
            if !unique[elem] {
                us = append(us, elem)
                unique[elem] = true
            }
        }
    }

    return us
}

func PrefixOfAll(sl []string) (string, error) {
    if len(sl) > 1 {
        checks := make(map[string][]bool)
        for _, match := range sl {
            checks[match] = make([]bool, len(sl))
            for j, othermatch := range sl {
                if strings.HasPrefix(othermatch, match) {
                    checks[match][j] = true
                } else {
                    checks[match][j] = false
                }
            }
        }
        for k, _ := range checks {
            for i, v := range checks[k] {
                if v == true {
                    if i != len(checks[k])-1 {
                        continue
                    }
                } else {
                    break
                }
                return k, nil
            }
        }
        return "", fmt.Errorf("ambiguous")
    } else if len(sl) == 1 {
        return sl[0], nil
    }

    return "", nil
}

func Intersection(a, b []string) (c []string) {
    m := make(map[string]bool)

    for _, item := range a {
        m[item] = true
    }

    for _, item := range b {
        if _, ok := m[item]; ok {
            c = append(c, item)
        }
    }
    return
}

func ExtractFileNameFromHref(href string) string {
    parts := strings.Split(href, "/")
    return parts[len(parts) - 1]
}

func GenericParseVersion(s string) string {
    r := regexp.MustCompile(`(?:(\d+\.[.\d]*\d+))`)
    return r.FindString(s)
}


////////////////////////////
// For debugging purposes //
////////////////////////////

func GetCallerFunctionName() string {
    // Skip GetCallerFunctionName and the function to get the caller of
    return getFrame(2).Function
}

func getFrame(skipFrames int) runtime.Frame {
    // We need the frame at index skipFrames+2, since we never want runtime.Callers and getFrame
    targetFrameIndex := skipFrames + 2

    // Set size to targetFrameIndex+2 to ensure we have room for one more caller than we need
    programCounters := make([]uintptr, targetFrameIndex+2)
    n := runtime.Callers(0, programCounters)

    frame := runtime.Frame{Function: "unknown"}
    if n > 0 {
        frames := runtime.CallersFrames(programCounters[:n])
        for more, frameIndex := true, 0; more && frameIndex <= targetFrameIndex; frameIndex++ {
            var frameCandidate runtime.Frame
            frameCandidate, more = frames.Next()
            if frameIndex == targetFrameIndex {
                frame = frameCandidate
            }
        }
    }

    return frame
}