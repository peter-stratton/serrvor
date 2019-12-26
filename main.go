package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

func errorRoll(threshold int) bool {
	rand.Seed(time.Now().UnixNano())
	x := rand.Intn(100)
	if x <= threshold {
		return true
	}
	return false
}

func endpointFactory(responseBody string, errorStatusCode, errorFrequency int) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if errorRoll(errorFrequency) {
			w.WriteHeader(errorStatusCode)
			w.Write([]byte(responseBody))
		} else {
			w.WriteHeader(200)
			w.Write([]byte("happy path"))
		}
	}
}

func main() {
	endpointPath := "/foo"
	responseBody := "bad path"
	errorCode := 404
	errorFrequency := 50

	router := mux.NewRouter()
	router.HandleFunc(endpointPath, endpointFactory(responseBody, errorCode, errorFrequency)).Methods("GET")

	srv := &http.Server{
		Addr:    ":8181",
		Handler: router,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	log.Print("Server Started")

	<-done
	log.Print("Server Stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// extra handling here
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
	log.Print("Server Exited Properly")
}
