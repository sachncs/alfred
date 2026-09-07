package runtime

import (
	"io"
	"net/http"
)

const maxAttachmentBytes = 50 * 1024 * 1024 // 50 MiB

func (r *LocalRuntime) handleUploadAttachment(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseMultipartForm(maxAttachmentBytes); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	file, header, err := req.FormFile("file")
	if err != nil {
		http.Error(w, "file is required: "+err.Error(), 400)
		return
	}
	defer func() { _ = file.Close() }()

	body, err := io.ReadAll(io.LimitReader(file, maxAttachmentBytes+1))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if int64(len(body)) > maxAttachmentBytes {
		http.Error(w, "attachment too large", http.StatusRequestEntityTooLarge)
		return
	}

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	threadID := req.FormValue("threadId")
	name := header.Filename

	att, err := r.AttachmentStore().Save(threadID, mimeType, name, body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	RouteJSON(w, att, 201)
}
