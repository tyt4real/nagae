package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"nagae/internal/mail"
	"nagae/internal/session"
)

func AttachmentGet(w http.ResponseWriter, r *http.Request) {
	creds := session.GetCredentials(r)

	mailbox := r.URL.Query().Get("mailbox")
	if mailbox == "" {
		mailbox = "INBOX"
	}

	uidStr := r.URL.Query().Get("uid")
	uid64, err := strconv.ParseUint(uidStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid uid", http.StatusBadRequest)
		return
	}

	filename := r.URL.Query().Get("filename")
	if filename == "" {
		http.Error(w, "filename required", http.StatusBadRequest)
		return
	}

	client, err := mail.NewIMAPClient(creds)
	if err != nil {
		http.Error(w, "could not connect to mail server", http.StatusBadGateway)
		return
	}
	defer client.Close()

	msg, err := client.FetchMessage(mailbox, uint32(uid64))
	if err != nil {
		http.Error(w, "could not fetch message", http.StatusInternalServerError)
		return
	}

	for _, att := range msg.Attachments {
		if att.Filename == filename {
			ct := att.ContentType
			if ct == "" {
				ct = "application/octet-stream"
			}
			w.Header().Set("Content-Type", ct)
			w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, att.Filename))
			w.Header().Set("Content-Length", strconv.Itoa(len(att.Data)))
			w.WriteHeader(http.StatusOK)
			w.Write(att.Data)
			return
		}
	}

	http.Error(w, "attachment not found", http.StatusNotFound)
}
