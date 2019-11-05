package main

import (
	"fmt"
	"html"
	"net/http"
	"time"

	"github.com/joshrwolf/rad/pkg/handlers"
	"github.com/joshrwolf/rad/pkg/mutators"

	log "github.com/sirupsen/logrus"
)

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello %q", html.EscapeString(r.URL.Path))
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handleRoot)
	mux.Handle("/mutate", handlers.AdmitFuncHandler(mutators.PrependRegistry))

	s := &http.Server{
		Addr:           ":8443",
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Printf("Starting webhook service on :8443")
	log.Fatal(s.ListenAndServeTLS("./ssl/tls.crt", "./ssl/tls.key"))
}
