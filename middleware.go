package main

import (
	"fmt"
	"net"
	"net/http"
	"time"
)

type CustomMux struct {
	*http.ServeMux
	db Database
}
type RouteFunc func(http.ResponseWriter, *CustomRequest, Database)
type UserRouteFunc func(http.ResponseWriter, *CustomRequest, User, Database)

// Basic HTTP route with logging
func (mux *CustomMux) NewRoute(pattern string, handler RouteFunc) {
	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		cw := &CustomResponseWriter{w, 200}
		cr := &CustomRequest{r, ""}

		if cr.GetRealIP() == "" {
			fmt.Printf(
				"Empty IP? cf-=%s x-real=%s remoteAddr=%s",
				r.Header.Get("CF-Connecting-IP"),
				r.Header.Get("X-Real-Ip"),
				r.RemoteAddr,
			)
			// TODO log to sentry
			cw.WriteHeader(511)
			cw.Write([]byte("Empty IP? Try again, if this is persistent, contact @Gamecube762"))
			return
		}

		start := time.Now()
		handler(cw, cr, mux.db)
		end := time.Since(start).Milliseconds()

		fmt.Printf(
			"[%s] %dms %d %s %s\n",
			cr.GetRealIP(),
			end,
			cw.statusCode,
			r.Method,
			r.RequestURI,
		)
	})
}

// HTTP route with User and logging.
func (mux *CustomMux) NewUserRoute(pattern string, handler UserRouteFunc) {
	mux.NewRoute(pattern, func(w http.ResponseWriter, r *CustomRequest, db Database) {
		user, err := mux.loadUser(w, r)
		if err != nil {
			return
		}

		handler(w, r, user, db)
	})
}

// Load the user from database.
// If there is an error, a 500 error will automatically be written to the ResponseWriter.
func (mux *CustomMux) loadUser(w http.ResponseWriter, r *CustomRequest) (user User, err error) {
	user, err = mux.db.GetUser(r.GetRealIP())
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("Failed to get user!"))
		// TODO log to Sentry
		fmt.Printf("Failed to get User \"%s\"\n", err)
	}

	return
}

// Custom Writer so we can pull the statusCode for logging
type CustomResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *CustomResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

type CustomRequest struct {
	*http.Request
	realIp string
}

func (r *CustomRequest) GetRealIP() string {
	if r.realIp == "" {
		switch envBehindProxy {
		case "cloudflare":
			r.realIp = r.Header.Get("CF-Connecting-IP")
		case "nginx":
			r.realIp = r.Header.Get("X-Real-Ip")
		default:
			// IP forwarding headers do not include the port.
			// We'll strip the port from r.RemoteAddr for consistency.
			var err error
			r.realIp, _, err = net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				fmt.Printf("Failed to split port from \"%s\" %s", r.RemoteAddr, err)
				r.realIp = r.RemoteAddr
			}
		}
	}
	return r.realIp
}
