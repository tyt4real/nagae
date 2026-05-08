package handler

import (
	"encoding/json"
	"net/http"

	"nagae/internal/mail"
	"nagae/internal/session"
)

type pollResponse struct {
	Unseen uint32 `json:"unseen"`
}

func NotifyPoll(w http.ResponseWriter, r *http.Request) {
	creds := session.GetCredentials(r)

	client, err := mail.NewIMAPClient(creds)
	if err != nil {
		http.Error(w, "connection failed", http.StatusBadGateway)
		return
	}
	defer client.Close()

	unseen, err := client.UnseenCount("INBOX")
	if err != nil {
		http.Error(w, "status failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(pollResponse{Unseen: unseen})
}
