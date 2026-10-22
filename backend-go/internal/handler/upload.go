package handler

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/irvanmhndra/nexpos-api/internal/storage"
	"github.com/irvanmhndra/nexpos-api/pkg/apperror"
	"github.com/irvanmhndra/nexpos-api/pkg/httputil"
	"github.com/labstack/echo/v5"
)

const maxUploadBytes = 10 << 20 // 10 MB (client compresses first; server-side guard)

type UploadHandler struct {
	storage *storage.Client
}

func NewUploadHandler(s *storage.Client) *UploadHandler {
	return &UploadHandler{storage: s}
}

type uploadResponse struct {
	URL string `json:"url"`
	Key string `json:"key"`
}

// Upload accepts a multipart image field "file", validates it by real magic
// bytes (not the client Content-Type), stores it in R2 (server-proxied), and
// returns its public CDN URL. Keys are scoped per company for tenant isolation.
func (h *UploadHandler) Upload(c *echo.Context) error {
	if h.storage == nil {
		return httputil.Error(c, apperror.InternalError(errors.New("object storage is not configured")))
	}

	fh, err := c.FormFile("file")
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("file is required"))
	}
	if fh.Size > maxUploadBytes {
		return httputil.Error(c, apperror.BadRequest("file too large (max 10MB)"))
	}

	f, err := fh.Open()
	if err != nil {
		return httputil.Error(c, apperror.InternalError(err))
	}
	defer func() { _ = f.Close() }()

	head := make([]byte, 512)
	n, err := io.ReadFull(f, head)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return httputil.Error(c, apperror.InternalError(err))
	}
	ext, contentType, ok := sniffImage(head[:n])
	if !ok {
		return httputil.Error(c, apperror.BadRequest("only JPG, PNG, or WEBP images are allowed"))
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return httputil.Error(c, apperror.InternalError(err))
	}

	key := fmt.Sprintf("products/%d/%s%s", getCompanyID(c), uuid.NewString(), ext)
	url, err := h.storage.Upload(c.Request().Context(), key, contentType, f, fh.Size)
	if err != nil {
		return httputil.Error(c, apperror.InternalError(err))
	}

	return httputil.Success(c, http.StatusCreated, "Image uploaded", uploadResponse{URL: url, Key: key})
}

// sniffImage returns the extension (with leading dot) + content type from the
// file's magic bytes.
func sniffImage(head []byte) (ext, contentType string, ok bool) {
	switch {
	case len(head) >= 3 && head[0] == 0xFF && head[1] == 0xD8 && head[2] == 0xFF:
		return ".jpg", "image/jpeg", true
	case len(head) >= 8 && bytes.Equal(head[:8], []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}):
		return ".png", "image/png", true
	case len(head) >= 12 && string(head[0:4]) == "RIFF" && string(head[8:12]) == "WEBP":
		return ".webp", "image/webp", true
	}
	return "", "", false
}
