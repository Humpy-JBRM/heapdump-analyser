package heapdump

import (
	"humpy/src/util"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var hprofFileDuration = promauto.NewCounter(
	prometheus.CounterOpts{
		Namespace: "heapdump",
		Name:      "duration",
		Help:      "The total amount of time spent processing HPROF files",
	},
)

type HeapdumpWorker struct {
	hprofQueue  HeapdumpQueue
	controlChan chan interface{}
	wg          *sync.WaitGroup
}

func NewHeapdumpWorker(controlChan chan interface{}, wg *sync.WaitGroup) *HeapdumpWorker {
	return &HeapdumpWorker{
		hprofQueue:  GetHeapdumpQueue(),
		controlChan: controlChan,
		wg:          wg,
	}
}

func (w *HeapdumpWorker) Run() {
	log.Printf("HeapdumpWorker.Run(): starting worker")
	defer func() {
		log.Printf("HeapdumpWorker.Run(): terminating worker")
		if w.wg != nil {
			w.wg.Done()
		}
	}()
	for {
		// Terminate if we have been told to terminate
		select {
		case <-w.controlChan:
			return

		default:
		}

		heapdumpJob, err := w.hprofQueue.Next()
		if err != nil {
			log.Println(err)
			continue
		}

		// Do the analysis
		err = util.CreateHeapdumpReport(heapdumpJob.Id, heapdumpJob.HprofFile)
		if err != nil {
			log.Println(err)
			os.RemoveAll(filepath.Dir(heapdumpJob.HprofFile))
			continue
		}
	}
}
