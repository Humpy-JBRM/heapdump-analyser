package fswatcher

import (
	"bufio"
	"fmt"
	"heapdump/util"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
)

type HeapdumpWatcher struct {
	dir                 string
	modifiedPathChannel chan *string
	eventChannel        chan *string
	analyseHeapDump     bool
	exitOnOOMDump       bool
	wg                  sync.WaitGroup
}

func NewHeapdumpWatcher(dir string, analyseHeapDump bool, exitOnOOMDump bool) (*HeapdumpWatcher, error) {
	return &HeapdumpWatcher{
		dir:                 dir,
		modifiedPathChannel: make(chan *string, 100),
		eventChannel:        make(chan *string, 100),
		analyseHeapDump:     analyseHeapDump,
		exitOnOOMDump:       exitOnOOMDump,
	}, nil
}

func (w *HeapdumpWatcher) Start() (err error) {
	numWorkers := 1

	// Kick off a bunch of workers
	for i := 0; i < numWorkers; i++ {
		w.wg.Add(1)
		go func() {
			w.notifyWorker()
		}()
	}

	return w.runWatcher()
}

func (w *HeapdumpWatcher) runWatcher() error {
	log.Printf("INFO|RunWatcher()| Starting watcher on dir: %s|", w.dir)
	if w.analyseHeapDump {
		log.Printf("INFO|RunWatcher()| Heap dumps will be auto-analysed|")
	} else {
		log.Printf("INFO|RunWatcher()| Heap dumps will not be auto-analysed|")
	}
	defer log.Printf("INFO|RunWatcher()| Stopping watcher on dir: %s|", w.dir)

	// Run the correct command depending on the file system in use
	var cmd *exec.Cmd
	switch runtime.GOOS {

	// MacOS: fswatch
	case "darwin":
		cmd = exec.Command("fswatch", "-r", "-x", w.dir)

	// Linux: inotifywait
	//        The `fswait` library does not give us the CLOSE_WRITE event explicitly, so
	//        we cannot reliably determine that the heap dump is complete.
	//
	//		  This will not matter when we're writing directly to a s3:// mount
	case "linux":
		cmd = exec.Command("inotifywait", "-m", "-e", "close_write", "-r", "--format", "%w%f", w.dir)
	case "windows":
		panic("TODO: Implement windows watcher")
	default:
		panic("Unknown operating system")
	}

	cmd.Stderr = os.Stderr

	stdoutIn, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("cmd.StdoutPipe() failed with '%s'\n", err)
	}

	// Make sure we catch all incoming signals so we can exit gracefully
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func(cmd *exec.Cmd, sigs <-chan os.Signal) {
		sig := <-sigs
		cmd.Process.Kill()
		log.Printf("Received signal %s.  Exiting\n", sig)
	}(cmd, sigs)

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("cmd.Start() failed with '%s'\n", err)
	}

	scanner := bufio.NewScanner(stdoutIn)
	for scanner.Scan() {
		filePath := scanner.Text()
		log.Printf("DEBUG|RunWatcher()| %s", filePath)
		if runtime.GOOS == "darwin" {
			if strings.Index(filePath, " Renamed") > 0 {
				continue
			}
			if strings.Index(filePath, " Removed") > 0 {
				continue
			}
			if !strings.Contains(filePath, " IsFile") {
				continue
			}

			log.Printf("DEBUG|RunWatcher()| %s", filePath)
			filePath = strings.Split(filePath, " ")[0]
		}
		w.modifiedPathChannel <- &filePath
	}
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("cmd.Wait() failed with %s", err)
	}

	return nil
}

func (w *HeapdumpWatcher) notifyWorker() {
	log.Printf("INFO|NotifyWorker()|Starting notify worker ...")
	defer log.Printf("INFO|NotifyWorker()|Terminating notify worker ...")
	defer w.wg.Done()

	for modifiedPath := range w.modifiedPathChannel {
		if modifiedPath == nil {
			// no more entries
			return
		}

		log.Printf("INFO|NotifyWorker()|%s -> s3://...", *modifiedPath)
		var err error
		if w.analyseHeapDump {
			// We want to analyse the heapdump and create the report
			// so that it's available in s3
			err = util.CreateHeapdumpReport(*modifiedPath, os.Getenv("PARSE_HEAPDUMP_CMD"))
		} else {
			// We're not doing auto-analysis of the heap dump, just copy it
			// to S3
			err = util.CopyFileToS3(*modifiedPath, os.Getenv("AWS_S3_BUCKET"), true)
		}
		if err != nil {
			log.Println(err.Error())
		}
		os.Remove(*modifiedPath)
		if w.exitOnOOMDump && filepath.Base(*modifiedPath) == "java_pid_1.hprof" {
			os.Exit(0)
		}
	}
}
