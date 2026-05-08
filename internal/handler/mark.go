package handler

import (
	"net/http"
	"strconv"

	"nagae/internal/mail"
	"nagae/internal/session"
)

func MarkPost(w http.ResponseWriter, r *http.Request) {
	creds := session.GetCredentials(r)

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	mailbox := r.FormValue("mailbox")
	if mailbox == "" {
		mailbox = "INBOX"
	}

	uid64, err := strconv.ParseUint(r.FormValue("uid"), 10, 32)
	if err != nil {
		http.Error(w, "invalid uid", http.StatusBadRequest)
		return
	}

	seen := r.FormValue("seen") == "1"

	client, err := mail.NewIMAPClient(creds)
	if err != nil {
		http.Error(w, "could not connect to mail server", http.StatusBadGateway)
		return
	}
	defer client.Close()

	if err := client.MarkMessage(mailbox, uint32(uid64), seen); err != nil {
		http.Error(w, "could not update flag: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/inbox?mailbox="+mailbox, http.StatusSeeOther)
}
