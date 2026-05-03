package handler

import (
	"net/http"

	"nagae/internal/mail"
	"nagae/internal/session"
	"nagae/views/auth"
)

func LoginGet(w http.ResponseWriter, r *http.Request) {
	if session.IsAuthenticated(r) {
		http.Redirect(w, r, "/inbox", http.StatusSeeOther)
		return
	}
	auth.Login("").Render(r.Context(), w)
}

func LoginPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	creds := &mail.Credentials{
		IMAPHost: mailServer.IMAPHost,
		IMAPPort: mailServer.IMAPPort,
		SMTPHost: mailServer.SMTPHost,
		SMTPPort: mailServer.SMTPPort,
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
	}

	client, err := mail.NewIMAPClient(creds)
	if err != nil {
		auth.Login("Invalid credentials or server unreachable.").Render(r.Context(), w)
		return
	}
	client.Close()

	if err := session.SetCredentials(w, r, creds); err != nil {
		http.Error(w, "session error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/inbox", http.StatusSeeOther)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	_ = session.Clear(w, r)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
