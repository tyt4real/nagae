package handler

import (
	"net/http"
	"strconv"

	"nagae/internal/mail"
	"nagae/internal/session"
)

func DeletePost(w http.ResponseWriter, r *http.Request) {
	creds := session.GetCredentials(r)

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	mailbox := r.FormValue("mailbox")
	if mailbox == "" {
		mailbox = "INBOX"
	}

	uidStr := r.FormValue("uid")
	uid64, err := strconv.ParseUint(uidStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid uid", http.StatusBadRequest)
		return
	}

	client, err := mail.NewIMAPClient(creds)
	if err != nil {
		http.Error(w, "could not connect to mail server", http.StatusBadGateway)
		return
	}
	defer client.Close()

	if err := client.DeleteMessage(mailbox, uint32(uid64)); err != nil {
		http.Error(w, "could not delete message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/inbox?mailbox="+mailbox+"&deleted=1", http.StatusSeeOther)
}
