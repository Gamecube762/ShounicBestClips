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

	serveMux.NewRoute("/vote/deadline", routeSendDeadline)
	serveMux.NewRoute("/vote/totals", routeTotals)

	fs, err := fs.Sub(embedWWW, "www")
	if err != nil {
		panic(err)
	}

	serveMux.Handle("/", http.FileServerFS(fs))
}

// Middleware TODO
//		Rate limiting
//      Prevent voting after a cutoff time

const (
	DBFailure         = `{"message": "Failed to fetch from Database."}`
	VoteClosed        = `{"message": "Voting is closed."}`
	SerializationFail = `{"message": "Failed to serialize JSON payload"}`
)

func routeNextVote(w http.ResponseWriter, req *CustomRequest, user User) {
	if time.Now().UTC().Unix() > votingDeadlineUnix {
		w.WriteHeader(420) // http.StatusTeapot for the funny, http.StatusGone would b more appropriate?
		w.Write([]byte(VoteClosed))
		return
	}

	options, err := database.GetNextVoteForUser(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(DBFailure))
		// TODO log to Sentry
		fmt.Printf("Failed to get new votes for user %v \"%s\"\n", user, err)
		return
	}

	// User has completed their queue
	if options == nil {
		w.WriteHeader(http.StatusNoContent) // NO_CONTENT
		return
	}

	// Send new vote to client
	bytes, err := json.Marshal(options)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(SerializationFail))
		// TODO log to Sentry
		fmt.Printf("Failed to write json data %v\n", options)
		return
	}

	w.Write(bytes)
}

func routeSubmitVote(w http.ResponseWriter, req *CustomRequest, user User) {
	if time.Now().UTC().Unix() > votingDeadlineUnix {
		w.WriteHeader(420)
		w.Write([]byte(VoteClosed))
		return
	}

	if err := req.ParseForm(); err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte(`{"message": "Failed to parse form input."}`))
		return
	}

	choice := req.PostForm.Get("choice")
	if choice == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "No choice given."}`))
		return
	}

	err := database.SubmitUserVote(user, choice)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(DBFailure))
		// TODO log to Sentry
		fmt.Printf("Failed to submit vote from %v of \"%s\": %v\n", user, choice, err)
		return
	}

	// Removing this and manually making another get request is easier than handling get request when I submit data
	// -myth
	//routeNextVote(w, req, user)
}

func routeSendDeadline(w http.ResponseWriter, req *CustomRequest) {
	bytes, err := json.Marshal(map[string]int64{"deadline": votingDeadlineUnix})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(SerializationFail))
		fmt.Printf("Failed to write json data regarding deadline timestamp")
		return
	}

	w.Write(bytes)
}

// TODO /myVotes

func routeTotals(w http.ResponseWriter, req *CustomRequest) {
	count, err := database.TallyVotes()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "Failed to count votes"}`))
		fmt.Println(err.Error())
		return
	}

	bytes, err := json.Marshal(count)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(SerializationFail))
		fmt.Println(err.Error())
		return
	}

	w.Write(bytes)
}
