package handler

import (
	"net/http"

	"nagae/internal/mail"
	"nagae/internal/session"
	inboxview "nagae/views/inbox"
)

func SearchGet(w http.ResponseWriter, r *http.Request) {
	creds := session.GetCredentials(r)

	query := r.URL.Query().Get("q")
	mailbox := r.URL.Query().Get("mailbox")
	if mailbox == "" {
		mailbox = "INBOX"
	}

	client, err := mail.NewIMAPClient(creds)
	if err != nil {
		http.Error(w, "could not connect to mail server", http.StatusBadGateway)
		return
	}
	defer client.Close()

	mailboxes, _ := client.ListMailboxes()

	var results []mail.MessageHeader
	if query != "" {
		results, err = client.SearchMessages(mailbox, query)
		if err != nil {
			http.Error(w, "search failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Cache-Control", "no-store")
	inboxview.Search(creds.Username, mailbox, mailboxes, query, results).Render(r.Context(), w)
}
