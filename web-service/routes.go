package main

import (
	"net/http"
)

func (app *App) routeHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(200)
	w.Write([]byte("AuditVault: OK :-)"))
}
