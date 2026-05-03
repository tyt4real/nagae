package main

import (
	"log"
	"net/http"

	"nagae/internal/config"
	"nagae/internal/handler"
	"nagae/internal/middleware"
	"nagae/internal/session"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	session.Init(cfg.SessionSecret)
	handler.SetMailServer(&cfg.Mail)

	mux := http.NewServeMux()

	// Files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Public
	mux.HandleFunc("GET /login", handler.LoginGet)
	mux.HandleFunc("POST /login", handler.LoginPost)
	mux.HandleFunc("GET /logout", handler.Logout)

	// Login required
	protected := http.NewServeMux()
	protected.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/inbox", http.StatusSeeOther)
	})
	protected.HandleFunc("GET /inbox", handler.InboxGet)
	protected.HandleFunc("GET /message", handler.MessageGet)
	protected.HandleFunc("POST /message/delete", handler.DeletePost)
	protected.HandleFunc("GET /attachment", handler.AttachmentGet)
	protected.HandleFunc("GET /compose", handler.ComposeGet)
	protected.HandleFunc("POST /compose", handler.ComposePost)

	mux.Handle("/", middleware.RequireAuth(protected))

	log.Printf("post/ listening on :%s (base: %s)", cfg.Port, cfg.BaseURL)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatal(err)
	}
}
