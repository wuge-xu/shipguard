package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	releaseapp "github.com/wuge-xu/shipguard/internal/application/release"
	"github.com/wuge-xu/shipguard/internal/release"
)

type fakeReleaseRepository struct {
	created     release.Release
	createCalls int
}

func (f *fakeReleaseRepository) Create(
	_ context.Context,
	item release.Release,
) error {
	f.createCalls++
	f.created = item

	return nil
}

func (f *fakeReleaseRepository) GetByID(
	_ context.Context,
	_ string,
) (release.Release, error) {
	return release.Release{}, nil
}

func (f *fakeReleaseRepository) UpdateTransition(
	_ context.Context,
	_ release.Release,
	_ release.Release,
) error {
	return nil
}

func TestCreateReleaseReturnsCreatedJSON(t *testing.T) {
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
		"service": "demo-service",
		"environment": "production",
		"source_sha": "commit-http-001",
		"image_digest": "sha256:http001"
	}
	`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/releases",
		body,
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	rec := httptest.NewRecorder()

	handler.CreateRelease(
		rec,
		req,
	)

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"status = %d, want %d",
			rec.Code,
			http.StatusCreated,
		)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf(
			"Content-Type = %q, want application/json",
			got,
		)
	}

	var response map[string]any

	if err := json.Unmarshal(
		rec.Body.Bytes(),
		&response,
	); err != nil {
		t.Fatalf(
			"decode response JSON: %v",
			err,
		)
	}

	requiredKeys := []string{
		"id",
		"service",
		"environment",
		"source_sha",
		"image_digest",
		"gitops_sha",
		"status",
		"version",
		"created_at",
		"updated_at",
	}

	for _, key := range requiredKeys {
		if _, ok := response[key]; !ok {
			t.Fatalf(
				"response missing key %q: %s",
				key,
				rec.Body.String(),
			)
		}
	}

	for _, forbiddenKey := range []string{
		"ID",
		"Service",
		"Environment",
		"SourceSHA",
		"ImageDigest",
		"Status",
		"Version",
	} {
		if _, ok := response[forbiddenKey]; ok {
			t.Fatalf(
				"response contains Go field name %q",
				forbiddenKey,
			)
		}
	}

	if response["service"] != "demo-service" {
		t.Fatalf(
			"service = %v, want demo-service",
			response["service"],
		)
	}

	if response["environment"] != "production" {
		t.Fatalf(
			"environment = %v, want production",
			response["environment"],
		)
	}

	if response["source_sha"] != "commit-http-001" {
		t.Fatalf(
			"source_sha = %v, want commit-http-001",
			response["source_sha"],
		)
	}

	if response["status"] != string(release.StatusPendingApproval) {
		t.Fatalf(
			"status = %v, want %q",
			response["status"],
			release.StatusPendingApproval,
		)
	}

	if response["version"] != float64(1) {
		t.Fatalf(
			"version = %v, want 1",
			response["version"],
		)
	}

	id, ok := response["id"].(string)
	if !ok || !strings.HasPrefix(id, "rel-") {
		t.Fatalf(
			"id = %v, want rel- prefix",
			response["id"],
		)
	}

	if repository.createCalls != 1 {
		t.Fatalf(
			"Create calls = %d, want 1",
			repository.createCalls,
		)
	}

	if repository.created.ID != id {
		t.Fatalf(
			"persisted ID = %q, response ID = %q",
			repository.created.ID,
			id,
		)
	}
}

func TestCreateReleaseRejectsMalformedJSON(t *testing.T) {
	repository := &fakeReleaseRepository{}

	service := releaseapp.NewService(
		repository,
		nil,
	)

	handler := NewReleaseHandler(
		service,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/releases",
		bytes.NewBufferString(`{"service":`),
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

	if repository.createCalls != 0 {
		t.Fatalf(
			"Create calls = %d, want 0",
			repository.createCalls,
		)
	}
}
