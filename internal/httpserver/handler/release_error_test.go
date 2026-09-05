package handler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	releaseapp "github.com/wuge-xu/shipguard/internal/application/release"
	"github.com/wuge-xu/shipguard/internal/release"
)

var errRepositoryCreate = errors.New("database write failed")

type failingReleaseRepository struct{}

func (f *failingReleaseRepository) Create(
	_ context.Context,
	_ release.Release,
) error {
	return errRepositoryCreate
}

func (f *failingReleaseRepository) GetByID(
	_ context.Context,
	_ string,
) (release.Release, error) {
	return release.Release{}, nil
}

func (f *failingReleaseRepository) UpdateTransition(
	_ context.Context,
	_ release.Release,
	_ release.Release,
) error {
	return nil
}

func TestCreateReleaseRejectsInvalidDomainInput(t *testing.T) {
	repository := &fakeReleaseRepository{}

	service := releaseapp.NewService(
		repository,
		nil,
	)

	handler := NewReleaseHandler(
		service,
	)

	body := bytes.NewBufferString(`
	{
		"service": "",
		"environment": "production",
		"source_sha": "commit-invalid-001",
		"image_digest": "sha256:invalid001"
	}
	`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/releases",
		body,
	)

	rec := httptest.NewRecorder()

	handler.CreateRelease(
		rec,
		req,
	)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"status = %d, want %d",
			rec.Code,
			http.StatusBadRequest,
		)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf(
			"Content-Type = %q, want application/json",
			got,
		)
	}

	if !strings.Contains(
		rec.Body.String(),
		`"error":"invalid release"`,
	) {
		t.Fatalf(
			"body = %q, want invalid release error",
			rec.Body.String(),
		)
	}

	if repository.createCalls != 0 {
		t.Fatalf(
			"Create calls = %d, want 0",
			repository.createCalls,
		)
	}
}

func TestCreateReleaseHidesRepositoryError(t *testing.T) {
	repository := &failingReleaseRepository{}

	service := releaseapp.NewService(
		repository,
		nil,
	)

	handler := NewReleaseHandler(
		service,
	)

	body := bytes.NewBufferString(`
	{
		"service": "demo-service",
		"environment": "production",
		"source_sha": "commit-failure-001",
		"image_digest": "sha256:failure001"
	}
	`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/releases",
		body,
	)

	rec := httptest.NewRecorder()

	handler.CreateRelease(
		rec,
		req,
	)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"status = %d, want %d",
			rec.Code,
			http.StatusInternalServerError,
		)
	}

	if !strings.Contains(
		rec.Body.String(),
		`"error":"internal server error"`,
	) {
		t.Fatalf(
			"body = %q, want generic internal error",
			rec.Body.String(),
		)
	}

	if strings.Contains(
		rec.Body.String(),
		errRepositoryCreate.Error(),
	) {
		t.Fatalf(
			"response leaked repository error: %q",
			rec.Body.String(),
		)
	}
}
