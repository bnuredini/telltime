package main

import (
	"net/http"

	"github.com/bnuredini/telltime/internal/httphandler"
	"github.com/bnuredini/telltime/ui"
)

func routes(uni *universe) http.Handler {
	httpHandler := httphandler.New(uni.DB, uni.TemplateManager)

	mux := http.NewServeMux()
	mux.HandleFunc("/", httpHandler.HomeGet)
	mux.HandleFunc("/activity", httpHandler.ActivityGet)
	mux.Handle("/static/", http.FileServer(http.FS(ui.Files)))

	return mux
}
