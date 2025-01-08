package main

import (
	"fmt"
	"net"
	"net/http"
	"time"
)

type CustomMux struct{ *http.ServeMux }
type RouteFunc func(http.ResponseWriter, *CustomRequest)
type UserRouteFunc func(http.ResponseWriter, *CustomRequest, User)

// Basic HTTP route with logging
func (mux *CustomMux) NewRoute(pattern string, handler RouteFunc) {
	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		cw := &CustomResponseWriter{w, http.StatusOK}
		cr := &CustomRequest{r, ""}

		// assume endpoint is json
		cw.Header().Add("Content-Type", "application/json; charset=utf-8")

		// user must have a useragent
		if r.UserAgent() == "" {
			cw.WriteHeader(http.StatusBadRequest)
			cw.Write([]byte(`{"message": "User-Agent is invalid."}`))
			return
		}

		if cr.GetRealIP() == "" {
			fmt.Printf(
				"Empty IP? cf-=%s x-real=%s remoteAddr=%s",
				r.Header.Get("CF-Connecting-IP"),
				r.Header.Get("X-Real-Ip"),
				r.RemoteAddr,
			)
			// TODO log to sentry
			cw.WriteHeader(http.StatusNetworkAuthenticationRequired)
			cw.Write([]byte(`{"message": "Empty IP? Try again, if this is persistent, contact @Gamecube762"}`))
			return
		}

		start := time.Now()
		handler(cw, cr)
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
	mux.NewRoute(pattern, func(w http.ResponseWriter, r *CustomRequest) {
		user, err := mux.loadUser(w, r)
		if err != nil {
			return
		}

		handler(w, r, user)
	})
}

// Load the user from database.
// If there is an error, a 500 error will automatically be written to the ResponseWriter.
func (mux *CustomMux) loadUser(w http.ResponseWriter, r *CustomRequest) (user User, err error) {
	user, err = database.GetUser(r.GetRealIP())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "Failed to get user!"}`))
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
