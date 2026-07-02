package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ralph/internal/shared/config"
	"ralph/internal/web/handlers"
	"ralph/internal/web/runs"
)

func TestGetRunPRDReturnsStories(t *testing.T) {
	workDir := t.TempDir()
	prdJSON := `{
  "version": 1,
  "project_name": "test",
  "stories": [
    {"id": "s1", "title": "a", "description": "d", "slices": [{"id": "slice-1", "behavior": "c", "red_hint": "write failing test for: c", "passes": false}], "priority": 1, "passes": false}
  ]
}`
	if err := os.WriteFile(filepath.Join(workDir, "prd.json"), []byte(prdJSON), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	reg := runs.NewRegistry()
	now := time.Now()
	if err := reg.Register(&runs.Run{
		ID:        "run-prd",
		WorkDir:   workDir,
		Prompt:    "goal",
		Status:    "waiting_review",
		Phase:     "review",
		CreatedAt: now,
		UpdatedAt: now,
		PRDPath:   "prd.json",
	}); err != nil {
		t.Fatal(err)
	}

	api := handlers.NewAPI(cfg, reg)
	req := httptest.NewRequest(http.MethodGet, "/api/runs/run-prd/prd", nil)
	req.SetPathValue("id", "run-prd")
	rec := httptest.NewRecorder()
	api.GetRunPRD(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var body struct {
		Stories []struct {
			ID     string `json:"id"`
			Slices []struct {
				ID       string `json:"id"`
				Behavior string `json:"behavior"`
				RedHint  string `json:"red_hint"`
				Passes   bool   `json:"passes"`
			} `json:"slices"`
			Priority int `json:"priority"`
		} `json:"stories"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if len(body.Stories) != 1 {
		t.Fatalf("len(stories) = %d, want 1", len(body.Stories))
	}
	if body.Stories[0].ID != "s1" {
		t.Fatalf("story id = %q, want s1", body.Stories[0].ID)
	}
	if len(body.Stories[0].Slices) != 1 || body.Stories[0].Slices[0].Behavior != "c" {
		t.Fatalf("slices = %#v, want one slice for c", body.Stories[0].Slices)
	}
	if body.Stories[0].Priority != 1 {
		t.Fatalf("priority = %d, want 1", body.Stories[0].Priority)
	}
}

func TestGetRunPRDReturnsProductStories(t *testing.T) {
	workDir := t.TempDir()
	productJSON := `{
  "version": 1,
  "project_name": "product",
  "stories": [
    {"id": "s1", "title": "a", "description": "d", "slices": [{"id": "slice-1", "behavior": "c", "passes": false}], "priority": 1, "passes": false}
  ]
}`
	if err := os.WriteFile(filepath.Join(workDir, "product.json"), []byte(productJSON), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	reg := runs.NewRegistry()
	now := time.Now()
	if err := reg.Register(&runs.Run{
		ID:        "run-product",
		WorkDir:   workDir,
		Prompt:    "goal",
		Status:    "waiting_review",
		Phase:     "review",
		CreatedAt: now,
		UpdatedAt: now,
		PRDPath:   "product.json",
	}); err != nil {
		t.Fatal(err)
	}

	api := handlers.NewAPI(cfg, reg)
	req := httptest.NewRequest(http.MethodGet, "/api/runs/run-product/prd", nil)
	req.SetPathValue("id", "run-product")
	rec := httptest.NewRecorder()
	api.GetRunPRD(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var body struct {
		ProjectName string `json:"project_name"`
		Stories     []struct {
			ID     string `json:"id"`
			Slices []struct {
				ID       string `json:"id"`
				Behavior string `json:"behavior"`
				Passes   bool   `json:"passes"`
			} `json:"slices"`
		} `json:"stories"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if body.ProjectName != "product" {
		t.Fatalf("project name = %q, want product", body.ProjectName)
	}
	if len(body.Stories) != 1 || len(body.Stories[0].Slices) != 1 {
		t.Fatalf("stories = %#v, want one product story with one slice", body.Stories)
	}
	if body.Stories[0].Slices[0].Behavior != "c" {
		t.Fatalf("behavior = %q, want c", body.Stories[0].Slices[0].Behavior)
	}
	if strings.Contains(rec.Body.String(), "red_hint") {
		t.Fatalf("product response should not include red_hint: %s", rec.Body.String())
	}
}
