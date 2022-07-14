package prometheus

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

func StartServer(serverAddr string) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	go func() {
		if err := http.ListenAndServe(serverAddr, mux); err != nil {
			fmt.Printf("Error starting prometheus server: %s\n", err.Error())
		}
	}()
}
