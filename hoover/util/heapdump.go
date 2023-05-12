package util

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
func CreateHeapdumpReport(pathToHprof string, parseHeapdumpCmd string) error {
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("ERROR|createHeapdumpReport(%s)|Could not get hostname: %s", pathToHprof, err.Error())
	}

	// Create a temporary directory to store our results
	tempdir, err := os.MkdirTemp("", "")
	if err != nil {
		return fmt.Errorf("ERROR|createHeapdumpReport()|Could not create temp dir: %s", err.Error())
	}

	// Move the hprof file into this directory
	_, err = copy(pathToHprof, filepath.Join(tempdir, "heapdump.hprof"))
	if err != nil {
		return fmt.Errorf("ERROR|createHeapdumpReport()|Could not copy hprof file: %s", err.Error())
	}
	os.Remove(pathToHprof)

	// Parse the heap dump
	//
	// This will create files called `heapdump.*.index` and `heapdump_<report>.zip`
	cmdline := []string{
		parseHeapdumpCmd,
		filepath.Join(tempdir, "heapdump.hprof"),
		"org.eclipse.mat.api:suspects",
		"org.eclipse.mat.api:overview",
		"org.eclipse.mat.api:top_components",
	}
	log.Printf("INFO|createHeapdumpReport()|Execute '%s'", strings.Join(cmdline, " "))
	cmd := exec.Command(cmdline[0], cmdline[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("ERROR|createHeapdumpReport()|%s: %s", cmdline[0], err.Error())
	}

	// Copy all of these files to the s3 bucket
	hprofBase := strings.Split(filepath.Base(pathToHprof), ".")[0]
	err = CopyDirToS3(filepath.Join(hostname, hprofBase), tempdir, os.Getenv("AWS_S3_BUCKET"), true)
	if err != nil {
		return err
	}
	return nil
}

func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}
