package product

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func validDocument() *Document {
	return &Document{
		Mode:        Mode,
		ProjectName: "Sharing",
		Stories: []*Story{{
			ID:          "story-1",
			Title:       "Users can share",
			Description: "Users can share saved itineraries.",
			Priority:    1,
			Slices:      []*Slice{{ID: "slice-1", Behavior: "A user can share an itinerary."}},
		}},
	}
}

func TestSaveAndLoadOnlyPersistsProductFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "prd.json")
	doc := validDocument()
	if err := Save(path, doc); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Mode != Mode || loaded.Version != 1 {
		t.Fatalf("loaded document = %+v", loaded)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "red_hint") || strings.Contains(string(data), "context") {
		t.Fatalf("product document contains implementation fields: %s", data)
	}
}

func TestLoadRejectsUnknownFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "prd.json")
	data := `{"mode":"product","project_name":"x","stories":[],"context":"technical"}`
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("Load should reject implementation fields")
	}
}

func TestValidateRejectsCircularDependencies(t *testing.T) {
	doc := validDocument()
	doc.Stories = append(doc.Stories, &Story{
		ID: "story-2", Title: "Second", Description: "Second outcome", Priority: 2,
		DependsOn: []string{"story-1"},
		Slices:    []*Slice{{ID: "slice-2", Behavior: "Second behavior"}},
	})
	doc.Stories[0].DependsOn = []string{"story-2"}
	if err := doc.Validate(); err == nil || !strings.Contains(err.Error(), "circular") {
		t.Fatalf("Validate error = %v, want circular dependency error", err)
	}
}

func TestMigrateLegacyProduct(t *testing.T) {
	dir := t.TempDir()
	legacy := filepath.Join(dir, LegacyFilename)
	path := filepath.Join(dir, "prd.json")
	data := `{"project_name":"Legacy","stories":[{"id":"story-1","title":"Outcome","description":"Value","priority":1,"slices":[{"id":"slice-1","behavior":"Users see value"}]}]}`
	if err := os.WriteFile(legacy, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	migrated, err := MigrateLegacy(dir, path)
	if err != nil || !migrated {
		t.Fatalf("MigrateLegacy() = %v, %v", migrated, err)
	}
	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Fatal("legacy product should be removed")
	}
	if doc, err := Load(path); err != nil || doc.Mode != Mode {
		t.Fatalf("migrated document = %+v, error = %v", doc, err)
	}
}
