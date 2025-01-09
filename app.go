package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var envDBFile = getEnvOrDefault("CLIPS_DB", "votes.db?_txlock=immediate")
var envBindAddr = getEnvOrDefault("CLIPS_BIND", ":8081")
var envBehindProxy = os.Getenv("CLIPS_BEHIND_PROXY")

// TODO: Make this less stupid -myth
// NOTE: THIS WILL BE TIMEZONE SENSITIVE!!!!!!!!!!
var votingDeadlineUnix int64 = 1736496000

func main() {
	fmt.Printf("Loading database %s\n", envDBFile)
	// we are passing this by ref on the http server
	database, err := LoadDatabase(envDBFile)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Starting http server on %s\n", envBindAddr)
	serveMux := CustomMux{
		ServeMux: http.NewServeMux(),
		db:       *database,
	}
	initRoutes(serveMux)
	server := &http.Server{Addr: envBindAddr, Handler: serveMux}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			// err can happen if same port used etc...
			database.Close()
			panic(err)
		}
	}()

	// this one will lockup the process NOT the ListenAndServe, so you can gracefully exit
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGSEGV)
	sig := <-sigc
	switch sig {
	case os.Interrupt:
	case syscall.SIGQUIT:
	case syscall.SIGTERM:
	case syscall.SIGSEGV: // gnome sys manager sends this on 'end process'
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		// triggers shutdown, also exits sig loop (if no err)
		if err := server.Shutdown(ctx); err != nil {
			panic(err)
		}
	}

	database.Close()
	fmt.Printf("Program exiting!")
	os.Exit(0)
}

func getEnvOrDefault(key string, defValue string) (value string) {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = defValue
	}
	return
}
