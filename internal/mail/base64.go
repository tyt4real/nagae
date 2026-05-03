package mail

import (
	"encoding/base64"
	"io"
	"strings"
)

func newBase64Reader(r io.Reader) io.Reader {
	raw, _ := io.ReadAll(r)
	clean := strings.ReplaceAll(string(raw), "\r\n", "")
	clean = strings.ReplaceAll(clean, "\n", "")
	clean = strings.ReplaceAll(clean, "\r", "")
	return base64.NewDecoder(base64.StdEncoding, strings.NewReader(clean))
}
