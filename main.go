package main

import "C"
import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/avishrantssh/GoLicenseClassifier/classifier"
)

var ROOT = ""

// Default Threshold for Filtering the results
var defaultThreshold = 0.8

// Default Licenses Root Directory
var default_path = "./classifier/default"
var licensePath string

// Regexp for Detecting Copyrights
var copyrightRE = regexp.MustCompile(`(?m)(?i:Copyright)\s+(?i:©\s+|\(c\)\s+)?(?:\d{2,4})(?:[-,]\s*\d{2,4})*,?\s*(?i:by)?\s*(.*?(?i:\s+Inc\.)?)[.,]?\s*(?i:All rights reserved\.?)?\s*$`)

// Removing in-text special code literals
var removeliteralRE = regexp.MustCompile(`\\n|\\f|\\r`)

// Create a classifier instance and load base licenses
func CreateClassifier() (*classifier.Classifier, error) {
	c := classifier.NewClassifier(defaultThreshold)
	return c, c.LoadLicenses(licensePath)
}

//export FindMatch
func FindMatch(root *C.char, fpaths *C.char) *C.char {
	ROOT = C.GoString(root)
	if licensePath == "" {
		licensePath = filepath.Join(ROOT, default_path)
	}
	patharr := GetPaths(C.GoString(fpaths))
	status := make([]string, len(patharr))

	// A simple channel implementation to lock function until execution is complete
	c, err := CreateClassifier()

	// fmt.Println("Finished Reading licenses")
	if err != nil {
		return C.CString("{ERROR:" + err.Error() + "}")
	}
	var wg sync.WaitGroup
	wg.Add(len(patharr))

	for index, path := range patharr {
		// Spawn a thread for each iteration in the loop.
		go func(index int, path string) {
			defer wg.Done()

			b, err := ioutil.ReadFile(path)
			// File Not Found
			if err != nil {
				status[index] = "{ERROR:" + err.Error() + "}"
				return
			}

			data := []byte(string(b))

			m := c.Match(data)
			var tmp string
			for i := 0; i < m.Len(); i++ {
				tmp += fmt.Sprintf("(%s,%f,%d,%d,%d,%d),", m[i].Name, m[i].Confidence, m[i].StartLine, m[i].EndLine, m[i].StartTokenIndex, m[i].EndTokenIndex)
			}

			cpInfo, holder := CopyrightInfo(string(b))
			status[index] = "{PATH:" + path + "},{EXT:" + filepath.Ext(path) + "},{LICENSE:[" + tmp + "]},{COP:[" + cpInfo + "]}{COP-HOLDER:[" + holder + "]}"

		}(index, path)
	}

	// Wait for `wg.Done()` to be exectued the number of times specified in the `wg.Add()` call.
	wg.Wait()
	return C.CString(strings.Join(status, "\n"))
}

// GetPaths function is used to convert new-line seperated filepaths to a string array.
func GetPaths(filepath string) []string {
	return strings.SplitN(filepath, "\n", -1)
}

//export LoadCustomLicenses
func LoadCustomLicenses(path *C.char) int {
	licensePath = C.GoString(path)
	return 1
}

//export SetThreshold
func SetThreshold(thresh int) int {
	if thresh < 0 || thresh > 100 {
		return 1
	}
	defaultThreshold = float64(thresh) / 100.0
	return 1
}

// CopyrightHolder finds a copyright notification, if it exists, and returns
// the copyright holder.
func CopyrightInfo(contents string) (string, string) {
	str := removeliteralRE.ReplaceAllString(contents, "\n")
	matches := copyrightRE.FindAllStringSubmatch(str, -1)
	var cpInfo, holder string
	for _, match := range matches {
		if len(match) == 2 {
			if len(cpInfo) == 0 {
				cpInfo = match[0]
				holder = match[1]
			} else {
				cpInfo += "," + match[0]
				holder += "," + match[1]
			}
		}
	}
	return cpInfo, holder
}

func main() {}
