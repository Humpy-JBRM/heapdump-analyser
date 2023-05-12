package util

import (
	"fmt"
	"humpy/src/data"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// createHeapdumpReport will automatically analyse the heap dump and create
// the various reports in HTML format.
//
// This has two main benefits:
//
//   - the heap dump is pre-indexed, so any further analysis in MAT will
//     be massively speeded up
//
//   - the HTML files can be served and viewed, making it possible to
//     potentially identify the cause of the problem without needing to
//     download anything
func CreateHeapdumpReport(jobId string, pathToHprof string) error {
	// Parse the heap dump
	//
	// This will create files called `heapdump.*.index` and `heapdump_<report>.zip`
	cmdline := []string{
		viper.GetString(data.CONF_HEAPDUMP_PARSE_CMD),
		pathToHprof,
		"org.eclipse.mat.api:suspects",
		"org.eclipse.mat.api:overview",
		"org.eclipse.mat.api:top_components",
	}
	log.Printf("INFO|createHeapdumpReport()|Execute '%s'", strings.Join(cmdline, " "))
	defer os.RemoveAll(filepath.Dir(pathToHprof))
	cmd := exec.Command(cmdline[0], cmdline[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("ERROR|createHeapdumpReport()|%s: %s", cmdline[0], err.Error())
	}

	// Now we want to:
	//
	//	- unzip the reports
	//
	//	- clean up the *.prof and index files
	var reportZips = map[string]string{
		"Leak Suspects":   "Leak_Suspects.zip",
		"System Overview": "System_Overview.zip",
		"Top Components":  "Top_Components.zip",
	}

	// Remove everything that isn't a zip file
	hprofDir := filepath.Dir(pathToHprof)
	entries, err := os.ReadDir(hprofDir)
	if err != nil {
		return fmt.Errorf("ERROR|createHeapdumpReport()|%s: %s", cmdline[0], err.Error())
	}
	for _, entry := range entries {
		for reportName, zipSuffix := range reportZips {
			if strings.Contains(entry.Name(), zipSuffix) {
				log.Printf("CreateHeapdumpReport(%s): creating report '%s'", pathToHprof, reportName)
				targetDir := filepath.Clean(
					filepath.Join(
						viper.GetString(data.CONF_HEAPDUMP_ROOT),
						jobId,
						strings.ReplaceAll(reportName, " ", "_"),
					),
				)
				os.MkdirAll(targetDir, 0755)
				err = Unzip(filepath.Join(filepath.Dir(pathToHprof), entry.Name()), targetDir)
				if err != nil {
					os.RemoveAll(targetDir)
					return fmt.Errorf("ERROR|createHeapdumpReport()|%s: %s", cmdline[0], err.Error())
				}
			}
		}
	}
	return nil
}
