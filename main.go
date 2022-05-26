package main

import (
	poststore "ars-projekat/configstore"
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

	store, err := poststore.New()
	if err != nil {
		log.Fatal(err)
	}

	server := Service{
		store: store,
	}

	router.HandleFunc("/configs/", server.IdempotencyCheck(server.createConfigHandler)).Methods("POST")
	router.HandleFunc("/configs/{uuid}/", server.IdempotencyCheck(server.createConfigVersionHandler)).Methods("POST")
	router.HandleFunc("/groups/", server.IdempotencyCheck(server.createGroupHandler)).Methods("POST")
	router.HandleFunc("/groups/{uuid}/", server.IdempotencyCheck(server.createGroupVersionHandler)).Methods("POST")
	router.HandleFunc("/configs/{uuid}/{ver}/", server.getConfigHandler).Methods("GET")
	router.HandleFunc("/groups/{uuid}/{ver}/", server.getGroupHandler).Methods("GET")
	router.HandleFunc("/configs/{uuid}/{ver}/", server.delConfigHandler).Methods("DELETE")
	router.HandleFunc("/groups/{uuid}/{ver}/", server.delGroupHandler).Methods("DELETE")
	router.HandleFunc("/groups/{uuid}/{ver}/configs/", server.IdempotencyCheck(server.addConfigToGroupHandler)).Methods("POST")

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
