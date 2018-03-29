package mhtml

import (
	"bufio"
	"bytes"
	"io"
	"mime"
	"mime/multipart"
	"net/textproto"
)

func GetHtml(mhtml []byte) []byte {
	br := bufio.NewReader(bytes.NewReader(mhtml))
	tr := textproto.NewReader(br)

	mimeHeader, err := tr.ReadMIMEHeader()
	if err != nil {
		return nil
	}

	contentType := mimeHeader.Get("Content-Type")
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil
	}

	boundary := params["boundary"]
	if len(boundary) == 0 {
		return nil
	}

	mr := multipart.NewReader(br, boundary)

	for {
		part, err := mr.NextPart()
		if err != nil {
			break
		}

		contentType := part.Header["Content-Type"]
		if len(contentType) == 0 {
			continue
		}

		if contentType[0] != "text/html" {
			continue
		}

		b := make([]byte, len(mhtml))
		n, err := part.Read(b)
		if err != nil && err != io.EOF {
			return nil
		}

		return b[:n]
	}

	return nil
}
