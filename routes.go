package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
)

//go:embed templates/*.html
var embedTemplates embed.FS
var templates *template.Template

func initRoutes(serveMux CustomMux) {
	serveMux.NewRoute("/", routeRoot)
	serveMux.NewRoute("/styling.css", stylingCSS)
	serveMux.NewUserRoute("/vote/next", routeNextVote)
	serveMux.NewUserRoute("/vote/submit", routeSubmitVote)

	templates = template.Must(template.ParseFS(embedTemplates, "templates/*.html"))
}

// Middleware TODO
//		Rate limiting

// Base route, return HTML template
func routeRoot(w http.ResponseWriter, req *http.Request) {
	if err := templates.ExecuteTemplate(w, "index.html", nil); err != nil {
		w.WriteHeader(500)
		w.Write([]byte("Failed to execute template."))
		fmt.Printf("Failed to execute template.")
	}

}

// Serve that solja boy
// becuase index.html has `styling.css` linked it makes a request for it, so  by just adding a route it automatically gets called and used
// Probs suboptimal. idc tho lol.
func stylingCSS(w http.ResponseWriter, req *http.Request) {
	http.ServeFile(w, req, "templates/styling.css")
}

func routeNextVote(w http.ResponseWriter, req *http.Request, user User) {
	options, err := database.GetNextVoteForUser(user)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("Failed to fetch from database."))
		// TODO log to Sentry
		fmt.Printf("Failed to get new votes for user %v \"%s\"\n", user, err)
		return
	}

	// User has completed their queue
	if options == nil {
		w.WriteHeader(204) // NO_CONTENT
		w.Write([]byte("No more items to vote on!"))
		return
	}

	// Send new vote to client
	bytes, err := json.Marshal(options)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("Failed to write JSON data?"))
		// TODO log to Sentry
		fmt.Printf("Failed to write json data %v\n", options)
		return
	}

	w.Write(bytes)
}

func routeSubmitVote(w http.ResponseWriter, req *http.Request, user User) {
	if err := req.ParseForm(); err != nil {
		w.WriteHeader(406)
		w.Write([]byte("Failed to parse form input."))
		return
	}

	choice := req.PostForm.Get("choice")
	if choice == "" {
		w.WriteHeader(400)
		w.Write([]byte("No choice given"))
		return
	}

	err := database.SubmitUserVote(user, choice)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("Failed to communicate with database."))
		// TODO log to Sentry
		fmt.Printf("Failed to submit vote from %v of \"%s\": %v\n", user, choice, err)
		return
	}

	routeNextVote(w, req, user)
}

// TODO /myVotes

// TODO /totalVotes
