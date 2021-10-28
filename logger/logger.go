package logger

import (
	"os"
	"fmt"
	"log"
	"time"
	"bytes"
	"strings"
	"runtime"
	"net/http"
    "io/ioutil"
    "math/rand"

	"gopkg.in/yaml.v3"
	"github.com/getsentry/sentry-go"
)

// Set to "0" after first report so we can see if the program had to be restarted
var newVisit string = "1"
var Sessionmap map[string]string
var sessionIdChars = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var AppName string
var AppVersion string
var BuildDate string
var Consent bool
var Device_model string
var Device_codename string

func Report(params map[string]string) {
	if params["tracking_consent"] == "false" || !Consent || AppVersion == "DEVELOPMENT" {
		Log("Skipped reporting:")
		Log(mapToString(params))
		return
	}

	report := "idsite=3&rec=1&send_image=0"
	report = report + "&_id=" + session()["id"]
	report = report + "&new_visit=" + newVisit
	newVisit = "0"
	report = report + "&ua=" + userAgent() + " Firefox/" + AppVersion
	report = report + "&_cvar={\"1\":[\"Version\",\"" + AppVersion + "\"],\"2\":[\"Build\",\"" + BuildDate + "\"]}"
	report = report + "&e_c=" + params["category"]
	if params["action"] != "" { report = report + "&e_a=" + params["action"] }
	if params["name"] != "" { report = report + "&e_n=" + params["name"] }
	if params["value"] != "" { report = report + "&e_v=" + params["value"] }
	if params["progress"] != "" { report = report + "&url=https://app/progress/" + params["progress"] }
	if params["tracking"] != "" { report = report + "&url=https://app/tracking/" + params["tracking"] }
	
	// If more parameters are given, add them to the report with their respective values
	// For this, remove the params taken into account and range over the rest
	delete(params, "tracking_consent")
	delete(params, "_id")
	delete(params, "ua")
	delete(params, "_cvar")
	delete(params, "category")
	delete(params, "action")
	delete(params, "name")
	delete(params, "value")
	delete(params, "progress")
	delete(params, "tracking")

	for k, v := range params {
		report = report + "&" + k + "=" + v
	}

	// Submit the report
	resp, err := http.Get("https://stats.free-droid.com/piwik.php?" + report)
	if err != nil {
	   LogError("Error submitting the report:", err)
	}
	if resp.Status != "204 No Content" {
		Log("Reporting failed with unexpected http status code", resp.Status)
	}
	resp.Body.Close()
}

func userAgent() string {
	switch runtime.GOOS {
	case "windows":
        return "(Windows)"
    case "darwin":
        return "(Macintosh)"
    case "linux":
        return "(Linux)"
    default:
        return "unknown"
	}
}

func randSeq(n int) string {
    b := make([]rune, n)
    for i := range b {
        b[i] = sessionIdChars[rand.Intn(len(sessionIdChars))]
    }
    return string(b)
}

func session() map[string]string {
	if Sessionmap != nil {
		return Sessionmap
	} else {
		Sessionmap = make(map[string]string)
	}

	// Create a session file if it does not exist
	_, err := os.Stat("log/session")
	if os.IsNotExist(err) {
		rand.Seed(time.Now().UnixNano())
		sessionId := randSeq(16)

		yamldata, err := yaml.Marshal(&map[string]string{"id":sessionId})
		if err != nil {
			LogError("Unable to marshal the session id to yaml:", err)
		}
		err = ioutil.WriteFile("log/session", yamldata, 0644)
		if err != nil {
			LogError("Error writing log/session file:", err)
		}

		Sessionmap, err = yamlToFlatMap(yamldata)
		if err != nil {
			LogError("Error converting yaml to flat map:", err)
		}
	} else {
		sessionfile, err := ioutil.ReadFile("log/session")
		if err != nil {
			LogError("Error reading session file:", err)
		}
		Sessionmap, err = yamlToFlatMap(sessionfile)
		if err != nil {
			LogError("Error converting yaml to flat map:", err)
		}
	}

	if Sessionmap["id"] == "" {
		Sessionmap["id"] = "XXXXXXXXXXXXXXXX"
	}

	return Sessionmap
}

func Log(s ...string) {
	fmt.Println(strings.Join(s[:]," "))
	appendToLogFile(strings.Join(s[:]," "))
}

func LogError(message string, err error) {
	if err.Error() != "cancelled" {
		sentry.WithScope(func(scope *sentry.Scope) {
			scope.SetUser(sentry.User{ID: session()["id"]})
			if Device_codename != "" {
				scope.SetTag("codename", Device_codename)
				scope.SetTag("model", Device_model)
			}
			sentry.CaptureException(fmt.Errorf(message + " " + err.Error()))
		})
		Log("ERROR: " + message + " " + err.Error())
	}
}

func appendToLogFile(s string) {
	f, err := os.OpenFile("log/" + AppName + ".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
	    fmt.Println("Cannot open log file for writing:", err)
	}
	defer f.Close()

	log2file := log.New(f, AppName + " ", log.LstdFlags)
	log2file.Println(s)
}

// Redefine from helpers to make this package free of internal
// dependencies so we can include it in every other package
func mapToString(m map[string]string) string {
    b := new(bytes.Buffer)
    for key, value := range m {
        fmt.Fprintf(b, "\"%s\"=\"%s\"\n", key, value)
    }
    return b.String()
}

// Redefine from helpers to make this package free of internal
// dependencies so we can include it in every other package
func yamlToFlatMap(yamldata []byte) (map[string]string, error) {
    themap := make(map[string]string)
    err := yaml.Unmarshal(yamldata, &themap)
    if err != nil {
        return nil, err
    }

    return themap, nil
}