package session

import (
	"net/http"

	"nagae/internal/mail"

	"github.com/gorilla/sessions"
)

const sessionName = "webmail-session"

var store *sessions.CookieStore

func Init(secret string) {
	store = sessions.NewCookieStore([]byte(secret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

func SetCredentials(w http.ResponseWriter, r *http.Request, creds *mail.Credentials) error {
	s, err := store.Get(r, sessionName)
	if err != nil {
		return err
	}
	s.Values["imap_host"] = creds.IMAPHost
	s.Values["imap_port"] = creds.IMAPPort
	s.Values["smtp_host"] = creds.SMTPHost
	s.Values["smtp_port"] = creds.SMTPPort
	s.Values["username"] = creds.Username
	s.Values["password"] = creds.Password
	return store.Save(r, w, s)
}

func GetCredentials(r *http.Request) *mail.Credentials {
	s, err := store.Get(r, sessionName)
	if err != nil || s.IsNew {
		return nil
	}

	username, ok := s.Values["username"].(string)
	if !ok || username == "" {
		return nil
	}

	return &mail.Credentials{
		IMAPHost: s.Values["imap_host"].(string),
		IMAPPort: s.Values["imap_port"].(string),
		SMTPHost: s.Values["smtp_host"].(string),
		SMTPPort: s.Values["smtp_port"].(string),
		Username: username,
		Password: s.Values["password"].(string),
	}
}

func Clear(w http.ResponseWriter, r *http.Request) error {
	s, err := store.Get(r, sessionName)
	if err != nil {
		return err
	}
	s.Options.MaxAge = -1
	return store.Save(r, w, s)
}

func IsAuthenticated(r *http.Request) bool {
	return GetCredentials(r) != nil
}
