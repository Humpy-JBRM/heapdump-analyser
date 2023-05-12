package heapdump

import (
	"humpy/src/data"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/spf13/viper"
)

var hprofQueuePut = promauto.NewCounter(
	prometheus.CounterOpts{
		Namespace: "heapdump",
		Name:      "queue_put",
		Help:      "HPROF files put on the queue for processing",
	},
)
var hprofQueueFetch = promauto.NewCounter(
	prometheus.CounterOpts{
		Namespace: "heapdump",
		Name:      "files_processed",
		Help:      "HPROF files fetched from the queue",
	},
)
var hprofQueueWait = promauto.NewCounter(
	prometheus.CounterOpts{
		Namespace: "heapdump",
		Name:      "queue_wait",
		Help:      "Total number of milliseconds that jobs were in the queue",
	},
)

type HeapdumpQueue interface {
	Put(hprofFile *data.HeapdumpJob) error
	Next() (*data.HeapdumpJob, error)
}

type heapdumpQueueImpl struct {
	hprofFiles chan *data.HeapdumpJob
}

// Singleton instance
var queueInstance HeapdumpQueue = newHeapdumpQueue()

func GetHeapdumpQueue() HeapdumpQueue {
	return queueInstance
}

func newHeapdumpQueue() HeapdumpQueue {
	numWorkers := 1
	if nw := viper.GetInt(data.CONF_HEAPDUMP_WORKERS); nw > 0 {
		numWorkers = nw
	}
	return &heapdumpQueueImpl{
		hprofFiles: make(chan *data.HeapdumpJob, numWorkers),
	}
}

func (q *heapdumpQueueImpl) Put(hprofFile *data.HeapdumpJob) error {
	hprofFile.QueuedMillis = time.Now().UTC().UnixMilli()
	q.hprofFiles <- hprofFile
	hprofQueuePut.Inc()
	return nil
}

func (q *heapdumpQueueImpl) Next() (*data.HeapdumpJob, error) {
	job := <-q.hprofFiles
	hprofQueueFetch.Inc()
	hprofQueueWait.Add(float64(time.Now().UTC().UnixMilli() - job.QueuedMillis))
	return job, nil
}
