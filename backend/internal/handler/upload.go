package handler

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/irvanmhndra/nexpos-api/internal/storage"
	"github.com/irvanmhndra/nexpos-api/pkg/apperror"
	"github.com/irvanmhndra/nexpos-api/pkg/httputil"
	"github.com/labstack/echo/v5"
)

type UploadHandler struct {
	storage *storage.Client
}

func NewUploadHandler(s *storage.Client) *UploadHandler {
	return &UploadHandler{storage: s}
}

type presignRequest struct {
	ContentType string `json:"content_type"`
	Kind        string `json:"kind"`
}

type presignResponse struct {
	UploadURL string `json:"upload_url"`
	PublicURL string `json:"public_url"`
	Key       string `json:"key"`
}

// allowedImageTypes maps accepted MIME types to the object-key extension.
var allowedImageTypes = map[string]string{
	"image/webp": "webp",
	"image/jpeg": "jpg",
	"image/png":  "png",
}

// Presign mints a short-lived URL the client uploads the image bytes to (R2),
// plus the public CDN URL to store on the product. Keys are scoped per company.
func (h *UploadHandler) Presign(c *echo.Context) error {
	if h.storage == nil {
		return httputil.Error(c, apperror.InternalError(errors.New("object storage is not configured")))
	}

	var req presignRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, httputil.BindError(err))
	}

	ext, ok := allowedImageTypes[req.ContentType]
	if !ok {
		return httputil.Error(c, apperror.BadRequest("unsupported content_type (allowed: image/webp, image/jpeg, image/png)"))
	}
	if req.Kind != "product" {
		return httputil.Error(c, apperror.BadRequest("unsupported kind"))
	}

	companyID := getCompanyID(c)
	key := fmt.Sprintf("products/%d/%s.%s", companyID, uuid.NewString(), ext)

	ctx := c.Request().Context()
	uploadURL, err := h.storage.PresignPut(ctx, key, req.ContentType, 5*time.Minute)
	if err != nil {
		return httputil.Error(c, apperror.InternalError(err))
	}

	return httputil.Success(c, http.StatusOK, "Presigned URL created", presignResponse{
		UploadURL: uploadURL,
		PublicURL: h.storage.PublicURL(key),
		Key:       key,
	})
}
