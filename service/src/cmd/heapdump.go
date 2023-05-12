package cmd

import (
	"humpy/src/api"
	"humpy/src/data"
	"humpy/src/heapdump"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var heapdumpCmd = &cobra.Command{
	Use:   "heapdump",
	Short: "Runs the heapdump analyser server",
	Run:   Runheapdump,
}

func Runheapdump(cmd *cobra.Command, args []string) {
	// Validate args and flags
	heapdumpAddress := viper.GetString(data.CONF_HEAPDUMP_LISTEN)
	if heapdumpAddress == "" {
		log.Fatalf("Runheapdump(): Must have a value for %s", data.CONF_HEAPDUMP_LISTEN)
	}
	heapdumpRoot := viper.GetString(data.CONF_HEAPDUMP_ROOT)
	if heapdumpRoot == "" {
		log.Fatalf("Runheapdump(): Must have a value for %s", data.CONF_HEAPDUMP_ROOT)
	}
	numWorkers := 1
	if nw := viper.GetInt(data.CONF_HEAPDUMP_WORKERS); nw > 0 {
		numWorkers = nw
	}

	_, err := os.Stat(filepath.Join(heapdumpRoot, "."))
	if err != nil {
		log.Fatalf("Runheapdump(): Directory '%s' must exist and must be a directory", data.CONF_HEAPDUMP_ROOT)
	}
	log.Println("Running Humpy heapdump on " + heapdumpAddress)
	log.Println("Heapdumps will be written ton " + heapdumpRoot)

	log.Printf("Starting %d queue workers", numWorkers)
	var wg sync.WaitGroup
	controlChan := make(chan interface{})
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go heapdump.NewHeapdumpWorker(controlChan, &wg).Run()
	}

	// Create the router
	ginRouter, err := api.NewRouter()
	if err != nil {
		log.Printf("Runheapdump(): Could not start server: %s", err.Error())
	}

	// Set the routes for this service
	ginRouter.POST("/api/heapdump", api.HeapdumpFile)
	ginRouter.StaticFS("/heapdump", http.Dir(heapdumpRoot))

	// Start the service
	err = ginRouter.Run(heapdumpAddress)
	if err != nil {
		log.Fatalf("Runheapdump(): Could not start server: %s", err.Error())
	}

	// Make sure all children get shut down
	close(controlChan)
	wg.Wait()
}
