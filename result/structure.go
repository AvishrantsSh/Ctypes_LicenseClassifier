package result

import (
	"time"
)

// JSON structure for storing scan results
type JSON_struct struct {
	header Header
	files  []FileInfo
}

// Header Info struct
type Header struct {
	Tool_name       string
	Start_timestamp time.Time
	End_timestamp   time.Time
	Duration        float64
	Files_count     int
}

type FileInfo struct {
	Path       string
	Licenses   []License
	Copyrights []CpInfo
	Errors     string
}

type License struct {
	Expression string
	Confidence float64
	Startline  int
	Endline    int
	Starttoken int
	Endttoken  int
}

type CpInfo struct {
	Expression []string
	Holders    []string
	// Timestamp  []string
}
