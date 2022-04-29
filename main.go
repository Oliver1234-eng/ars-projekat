package main

import (
	"context"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	router := mux.NewRouter()
	router.StrictSlash(true)

	server := Service{
		configs: map[string]*Config{},
		groups:  map[string]*Group{},
	}

	router.HandleFunc("/configs", server.createConfigHandler).Methods("POST")
	//router.HandleFunc("/groups/", server.someHandler).Methods("POST")
	router.HandleFunc("/configs/{uuid}/", server.getConfigHandler).Methods("GET")
	//router.HandleFunc("/groups/{uuid}/", server.someHandler).Methods("GET")
	router.HandleFunc("/configs/{uuid}/", server.delConfigHandler).Methods("DELETE")
	//router.HandleFunc("/groups/{uuid}/", server.someHandler).Methods("DELETE")
	//router.HandleFunc("/groups/{uuid}/configs", server.someHandler).Methods("POST")

	// start server
	srv := &http.Server{Addr: "0.0.0.0:8000", Handler: router}
	go func() {
		log.Println("server starting")
		if err := srv.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Fatal(err)
			}
		}
	}()

	<-quit

	log.Println("service shutting down ...")

	// gracefully stop server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
	log.Println("server stopped")
}
