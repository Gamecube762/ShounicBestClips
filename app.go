package main

import (
	"fmt"
	"net/http"
	"os"
)

var database *Database

var envDBFile = getEnvOrDefault("CLIPS_DB", "votes.db?_mutex=full")
var envBindAddr = getEnvOrDefault("CLIPS_BIND", ":8081")
var envBehindProxy = os.Getenv("CLIPS_BEHIND_PROXY")

// TODO: Make this less stupid -myth
// NOTE: THIS WILL BE TIMEZONE SENSITIVE!!!!!!!!!!
var votingDeadlineUnix int64 = 1736496000

func main() {
	var err error

	fmt.Printf("Loading database %s\n", envDBFile)
	database, err = LoadDatabase(envDBFile)
	if err != nil {
		panic(err)
	}
	defer database.Close()

	serveMux := CustomMux{http.NewServeMux()}
	initRoutes(serveMux)

	fmt.Printf("Starting http server on http://localhost%s\n", envBindAddr)
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
