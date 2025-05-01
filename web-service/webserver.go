package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func (app *App) configureRoutes() {
	http.HandleFunc("/health-check/", app.webServerHeaders(app.routeHealthCheck))
	http.HandleFunc("/dictation/", app.webServerHeaders(app.routeDictation))
	http.HandleFunc("/dashboard/", app.webServerHeaders(app.routeDashboard))
	http.HandleFunc("/server-side-events/", app.webServerHeaders(app.routeServerSideEvents))
	http.HandleFunc("/user/", app.webServerHeaders(app.routeUser))

	http.HandleFunc("/orphan/", app.routeOrphans)
	http.HandleFunc("/store/", app.routeStore)
	http.HandleFunc("/stream/", app.routeStream)
	http.HandleFunc("/waveform/", app.routeWaveForm)
}

// The web server always emits these default HTTP response headers
// The response headers have been defined by the OWAPS secure headers project
// https://owasp.org/www-project-secure-headers/
func (app *App) defaultResponseHeaders(w http.ResponseWriter) {
	if w.Header().Get("Content-Type") == "text/html" {
		// w.Header().Set("Content-Security-Policy", "default-src 'self';")
	}

	// w.Header().Set("AAccess-Control-Allow-Origin", "turso.io")
	// w.Header().Set("Access-Control-Allow-Headers", "Authorization")
	w.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
	w.Header().Set("Cross-Origin-Opener-Policy", "same-origin-allow-popups")
	w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")
	w.Header().Set("Referrer-Policy", "same-origin")
	w.Header().Set("Server", "AuditVault")
	w.Header().Set("Strict-Transport-Security", "max-age=63072000")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "0")
}

// Calls the default security headers and then conditionally sets
// HTTP Response Headers based on the content being served.
func (app *App) webServerHeaders(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if strings.HasPrefix(r.URL.Path, "/data") {
			w.Header().Set("Content-Type", "application/json")
			app.defaultResponseHeaders(w)
			fn(w, r)
			return
		}

		if strings.Contains(r.URL.Path, "static-assets") {
			// change the cache control settings for static assets
			w.Header().Set("Cache-Control", "public, max-age=63072000, immutable")

			if strings.Contains(r.URL.Path, ".css") {
				w.Header().Set("Content-Type", "text/css")
			}
			if strings.Contains(r.URL.Path, ".js") {
				w.Header().Set("Content-Type", "text/javascript")
			}
			if strings.Contains(r.URL.Path, ".svg") {
				w.Header().Set("Content-Type", "image/svg+xml")
			}
			if strings.Contains(r.URL.Path, ".png") {
				w.Header().Set("Content-Type", "image/png")
			}
		} else {
			w.Header().Set("Cache-Control", "private, max-age=0, no-cache")
			w.Header().Set("Content-Type", "text/html")
		}

		app.defaultResponseHeaders(w)
		fn(w, r)
	}
}

// Start a HTTP web server on the defined port number.
func (app *App) startWebServer() {
	s := &http.Server{
		Addr:         app.portNumber,
		IdleTimeout:  time.Minute,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// start the web service and log a fatal error if
	// the service could not be started for any reason
	log.Println("INFO: starting web service")
	log.Println("INFO:" + app.executableFolder)

	if err := s.ListenAndServe(); err != nil {
		log.Println("ERR:" + err.Error())
		os.Exit(1)
	}
}

func (app *App) webServerPassthrough(fn http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn.ServeHTTP(w, r)
	}
}
