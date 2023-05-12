package api

import (
	"fmt"
	"humpy/src/data"
	"humpy/src/heapdump"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var hprofRequests = promauto.NewCounter(
	prometheus.CounterOpts{
		Namespace: "heapdump",
		Name:      "hprof_requests",
		Help:      "The number of HPROF files sent to the heapdump analyser",
	},
)
var hprofFileResponses = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "heapdump",
		Name:      "hprof_responses",
		Help:      "HPROF files responses, keyed by http code",
	},
	[]string{"code"},
)

// POST /api/heapdump
//
// This is a multipart-form upload with a single field called "file" (type "file")
func HeapdumpFile(c *gin.Context) {
	hprofRequests.Inc()

	// Get the file being uploaded
	file, err := c.FormFile("file")
	if err != nil {
		log.Printf("HeapdumpFile(): %s", err.Error())
		hprofFileResponses.WithLabelValues(fmt.Sprint(http.StatusBadRequest))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	mpf, err := file.Open()
	if err != nil {
		log.Printf("HeapdumpFile(): %s", err.Error())
		hprofFileResponses.WithLabelValues(fmt.Sprint(http.StatusBadRequest))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// Save it to a temporary directory
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		e := fmt.Errorf("HeapdumpFile(): %s", err.Error())
		hprofFileResponses.WithLabelValues(fmt.Sprint(http.StatusBadRequest))
		c.AbortWithError(http.StatusBadRequest, e)
		return
	}
	tempFile, err := ioutil.TempFile(tempDir, "*-"+file.Filename)
	if err != nil {
		os.RemoveAll(tempDir)
		e := fmt.Errorf("HeapdumpFile(): %s", err.Error())
		hprofFileResponses.WithLabelValues(fmt.Sprint(http.StatusBadRequest))
		c.AbortWithError(http.StatusBadRequest, e)
		return
	}

	io.Copy(tempFile, mpf)
	mpf.Close()
	tempFile.Close()

	// Add this file to the heapdump analysis queue
	heapdumpJob := data.NewHeapdumpJob(tempFile.Name())
	err = heapdump.GetHeapdumpQueue().Put(heapdumpJob)
	if err != nil {
		e := fmt.Errorf("HeapdumpFile(): %s", err.Error())
		hprofFileResponses.WithLabelValues(fmt.Sprint(http.StatusInternalServerError))
		c.AbortWithError(http.StatusInternalServerError, e)
		return
	}

	hprofFileResponses.WithLabelValues(fmt.Sprint(http.StatusAccepted))
	c.JSON(http.StatusAccepted, heapdumpJob)
}
