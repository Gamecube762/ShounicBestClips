package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"time"
)

//go:embed www/*
var embedWWW embed.FS

func initRoutes(serveMux CustomMux) {
	serveMux.NewUserRoute("/vote/next", routeNextVote)
	serveMux.NewUserRoute("/vote/submit", routeSubmitVote)
	serveMux.NewUserRoute("/vote/deadline", routeSendDeadline)

	staticFiles, err := fs.Sub(embedWWW, "www")
	if err != nil {
		panic(err)
	}

	serveMux.Handle("/", http.FileServerFS(staticFiles))
}

// Middleware TODO
//		Rate limiting
//      Prevent voting after a cutoff time

func writeString(w http.ResponseWriter, code int, msg string) {
	writeBytes(w, code, []byte(msg))
}

func writeBytes(w http.ResponseWriter, code int, data []byte) {
	w.WriteHeader(code)
	_, err := w.Write(data)
	if err != nil {
		http.Error(w, "Failed to write response", 500)
	}
}

func routeNextVote(w http.ResponseWriter, req *CustomRequest, user User) {
	options, err := database.GetNextVoteForUser(user)
	if err != nil {
		writeString(w, 500, "Failed to fetch from database.")
		// TODO log to Sentry

		// I hate telemetry :(
		// - Arzumify
		fmt.Println("Failed to get new votes for user", user, ":", err)
		return
	}

	// User has completed their queue
	if options == nil {
		writeString(w, 204, "No more items to vote on!")
		return
	}

	// Send new vote to client
	bytes, err := json.Marshal(options)
	if err != nil {
		writeString(w, 500, "Failed to write json data.")
		// TODO log to Sentry

		// :(
		// - Arzumify
		fmt.Println("Failed to write json data", options)
		return
	}

	writeBytes(w, 200, bytes)
}

func routeSubmitVote(w http.ResponseWriter, req *CustomRequest, user User) {
	if err := req.ParseForm(); err != nil {
		writeString(w, 406, "Failed to parse form input.")
		return
	}

	if !req.PostForm.Has("choice") {
		writeString(w, 400, "No choice given.")
		return
	}

	if time.Now().After(votingDeadline) {
		writeString(w, 420, "Deadline passed")
		return
	}

	choice := req.PostForm.Get("choice")
	err := database.SubmitUserVote(user, choice)
	if err != nil {
		writeString(w, 500, "Failed to communicate with database.")
		// TODO log to Sentry

		// I hate telemetry :(
		// - Arzumify
		fmt.Println("Failed to submit vote from", user, "of", choice, ":", err)
		return
	}

	// Removing this and manually making another get request is easier than handling get request when I submit data
	// -myth
	//routeNextVote(w, req, user)
}

func routeSendDeadline(w http.ResponseWriter, req *CustomRequest, user User) {
	bytes, err := json.Marshal(map[string]int64{"deadline": votingDeadline.Unix()})

	if err != nil {
		writeString(w, 500, "Failed to write json data regarding deadline timestamp")
		fmt.Println("Failed to write json data regarding deadline timestamp")
		return
	}

	writeBytes(w, 200, bytes)
}

// TODO /myVotes

// TODO /totalVotes
