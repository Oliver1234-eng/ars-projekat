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

	router.HandleFunc("/configs/", count(server.IdempotencyCheck(server.createConfigHandler), "createConfigHandler")).Methods("POST")
	router.HandleFunc("/configs/{uuid}/", count(server.IdempotencyCheck(server.createConfigVersionHandler), "createConfigVersionHandler")).Methods("POST")
	router.HandleFunc("/groups/", count(server.IdempotencyCheck(server.createGroupHandler), "createGroupHandler")).Methods("POST")
	router.HandleFunc("/groups/{uuid}/", count(server.IdempotencyCheck(server.createGroupVersionHandler), "createGroupVersionHandler")).Methods("POST")
	router.HandleFunc("/configs/{uuid}/{ver}/", count(server.getConfigHandler, "getConfigHandler")).Methods("GET")
	router.HandleFunc("/groups/{uuid}/{ver}/", count(server.getGroupHandler, "getGroupHandler")).Methods("GET")
	router.HandleFunc("/configs/{uuid}/{ver}/", count(server.delConfigHandler, "delConfigHandler")).Methods("DELETE")
	router.HandleFunc("/groups/{uuid}/{ver}/", count(server.delGroupHandler, "delGroupHandler")).Methods("DELETE")
	router.HandleFunc("/groups/{uuid}/{ver}/configs/", count(server.IdempotencyCheck(server.addConfigToGroupHandler), "addConfigToGroupHandler")).Methods("POST")
	router.Path("/metrics").Handler(metricsHandler())

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
