package handler

import (
	"net/http"

	"nagae/internal/mail"
	"nagae/internal/session"
	composeview "nagae/views/compose"
)

func ComposeGet(w http.ResponseWriter, r *http.Request) {
	creds := session.GetCredentials(r)

	to := r.URL.Query().Get("to")
	subject := r.URL.Query().Get("subject")

	composeview.Compose(creds.Username, to, subject, "").Render(r.Context(), w)
}

func ComposePost(w http.ResponseWriter, r *http.Request) {
	creds := session.GetCredentials(r)

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	req := &mail.ComposeRequest{
		From:    creds.Username,
		To:      r.FormValue("to"),
		CC:      r.FormValue("cc"),
		Subject: r.FormValue("subject"),
		Body:    r.FormValue("body"),
	}

	rawMsg, err := mail.SendMessage(creds, req)
	if err != nil {
		composeview.Compose(creds.Username, req.To, req.Subject, err.Error()).Render(r.Context(), w)
		return
	}

	if imapClient, err := mail.NewIMAPClient(creds); err == nil {
		_ = imapClient.AppendToSent(rawMsg)
		imapClient.Close()
	}

	http.Redirect(w, r, "/inbox?sent=1", http.StatusSeeOther)
}
