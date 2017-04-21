package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path"
	"sync"
	"time"

	"github.com/ubuntu/tutorial-deployment/consts"
	"github.com/ubuntu/tutorial-deployment/paths"
)

func startHTTPServer(port int, wg *sync.WaitGroup, stop chan bool) {
	s := &http.Server{Addr: fmt.Sprintf(":%d", port)}
	log.Printf("Serving on http://localhost:%d\n", port)

	p := paths.New()

	http.Handle(consts.APIURL, http.StripPrefix(consts.APIURL, http.FileServer(http.Dir(p.API))))
	http.Handle(consts.CodelabSrcURL, http.StripPrefix(consts.CodelabSrcURL, http.FileServer(http.Dir(p.Export))))
	// always serve root file for tutorials if page refreshed
	http.HandleFunc(consts.ServeRootURL, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path.Join(p.Website, "index.html"))
	})
	http.Handle("/", http.FileServer(http.Dir(p.Website)))

	wg.Add(2)
	// Serve
	go func() {
		defer wg.Done()
		defer s.Close()
		if err := s.ListenAndServe(); err != nil {
			log.Printf("Server quit: %s", err)
		}
	}()
	// Stop server
	go func() {
		defer wg.Done()
		<-stop
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := s.Shutdown(ctx); err != nil {
			panic(err)
		}
		cancel()
	}()

}
