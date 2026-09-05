package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	releaseapp "github.com/wuge-xu/shipguard/internal/application/release"
	"github.com/wuge-xu/shipguard/internal/release"
)

type ReleaseHandler struct {
	service *releaseapp.Service
}

func NewReleaseHandler(
	service *releaseapp.Service,
) *ReleaseHandler {
	return &ReleaseHandler{
		service: service,
	}
}

type createReleaseRequest struct {
	Service     string `json:"service"`
	Environment string `json:"environment"`
	SourceSHA   string `json:"source_sha"`
	ImageDigest string `json:"image_digest"`
}

type releaseResponse struct {
	ID          string    `json:"id"`
	Service     string    `json:"service"`
	Environment string    `json:"environment"`
	SourceSHA   string    `json:"source_sha"`
	ImageDigest string    `json:"image_digest"`
	GitOpsSHA   string    `json:"gitops_sha"`
	Status      string    `json:"status"`
	Version     int64     `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *ReleaseHandler) CreateRelease(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req createReleaseRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(
			w,
			http.StatusBadRequest,
			errorResponse{
				Error: "invalid request body",
			},
		)
		return
	}

	item, err := h.service.CreateRelease(
		r.Context(),
		releaseapp.CreateReleaseInput{
			Service:     req.Service,
			Environment: req.Environment,
			SourceSHA:   req.SourceSHA,
			ImageDigest: req.ImageDigest,
		},
	)
	if err != nil {
		if errors.Is(
			err,
			release.ErrInvalidRelease,
		) {
			writeJSON(
				w,
				http.StatusBadRequest,
				errorResponse{
					Error: "invalid release",
				},
			)
			return
		}

		writeJSON(
			w,
			http.StatusInternalServerError,
			errorResponse{
				Error: "internal server error",
			},
		)
		return
	}

	response := releaseResponse{
		ID:          item.ID,
		Service:     item.Service,
		Environment: item.Environment,
		SourceSHA:   item.SourceSHA,
		ImageDigest: item.ImageDigest,
		GitOpsSHA:   item.GitOpsSHA,
		Status:      string(item.Status),
		Version:     item.Version,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}

	writeJSON(
		w,
		http.StatusCreated,
		response,
	)
}

func writeJSON(
	w http.ResponseWriter,
	status int,
	value any,
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(value)
}
