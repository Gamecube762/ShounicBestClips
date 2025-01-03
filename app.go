package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

var database *Database

var envDBFile = getEnvOrDefault("CLIPS_DB", "votes.db?_mutex=full")
var envBindAddr = getEnvOrDefault("CLIPS_BIND", ":8081")
var envBehindProxy = os.Getenv("CLIPS_BEHIND_PROXY")

// I made it less stupid for you
// - Arzumify
var votingDeadline = time.Date(2025, time.January, 15, 24, 30, 50, 0, time.UTC)

func main() {
	var err error

	fmt.Printf("Loading database %s\n", envDBFile)
	database, err = LoadDatabase(envDBFile)
	if err != nil {
		// There is no point catching the error if you are just going to panic, smh
		// - Arzumify
		panic(err)
	}

	serveMux := CustomMux{http.NewServeMux()}
	initRoutes(serveMux)

	fmt.Printf("Starting http server on %s\n", envBindAddr)
	if err = http.ListenAndServe(envBindAddr, serveMux); err != nil {
		panic(err)
	}
}

func getEnvOrDefault(key string, defValue string) (value string) {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = defValue
	}
	return
}
