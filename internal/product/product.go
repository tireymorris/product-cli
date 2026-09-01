package product

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const Mode = "product"
const LegacyFilename = "product.json"

const (
	MaxStories       = 1000
	MaxStoryDescSize = 100 * 1024
)

type Document struct {
	Mode        string   `json:"mode"`
	Version     int64    `json:"version"`
	ProjectName string   `json:"project_name"`
	BranchName  string   `json:"branch_name,omitempty"`
	Stories     []*Story `json:"stories"`
}

type Story struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Slices      []*Slice `json:"slices"`
	Priority    int      `json:"priority"`
	DependsOn   []string `json:"depends_on,omitempty"`
	Passes      bool     `json:"passes"`
}

type Slice struct {
	ID       string `json:"id"`
	Behavior string `json:"behavior"`
	Passes   bool   `json:"passes"`
}

func (d *Document) Validate() error {
	if d == nil {
		return errors.New("product document cannot be nil")
	}
	if d.Mode != Mode {
		return fmt.Errorf("mode must be %q", Mode)
	}
	if len(d.Stories) > MaxStories {
		return fmt.Errorf("story count %d exceeds maximum %d", len(d.Stories), MaxStories)
	}
	seen := make(map[string]bool, len(d.Stories))
	for i, story := range d.Stories {
		if err := story.validate(seen); err != nil {
			return fmt.Errorf("story %d: %w", i, err)
		}
		seen[story.ID] = true
	}
	return validateDependencies(d)
}

func (s *Story) validate(seen map[string]bool) error {
	if s == nil {
		return errors.New("story cannot be nil")
	}
	if s.ID == "" {
		return errors.New("story ID cannot be empty")
	}
	if seen[s.ID] {
		return fmt.Errorf("duplicate story ID %q", s.ID)
	}
	if strings.TrimSpace(s.Title) == "" {
		return errors.New("story title cannot be empty")
	}
	if len(s.Description) > MaxStoryDescSize {
		return fmt.Errorf("story description exceeds %d bytes", MaxStoryDescSize)
	}
	if s.Priority < 0 {
		return errors.New("story priority cannot be negative")
	}
	if len(s.Slices) == 0 {
		return fmt.Errorf("story %q must have at least one slice", s.ID)
	}
	sliceIDs := make(map[string]bool, len(s.Slices))
	for i, slice := range s.Slices {
		if slice == nil {
			return fmt.Errorf("slice %d cannot be nil", i)
		}
		if slice.ID == "" || slice.Behavior == "" {
			return fmt.Errorf("slice %d needs an id and behavior", i)
		}
		if sliceIDs[slice.ID] {
			return fmt.Errorf("duplicate slice ID %q", slice.ID)
		}
		sliceIDs[slice.ID] = true
	}
	for _, dependency := range s.DependsOn {
		if dependency == "" || dependency == s.ID {
			return fmt.Errorf("invalid dependency %q", dependency)
		}
	}
	return nil
}

func validateDependencies(d *Document) error {
	stories := make(map[string]*Story, len(d.Stories))
	for _, story := range d.Stories {
		stories[story.ID] = story
	}
	visited := map[string]bool{}
	var visit func(string, []string) error
	visit = func(id string, path []string) error {
		for _, parent := range path {
			if parent == id {
				return fmt.Errorf("circular dependency detected: %v", append(path, id))
			}
		}
		if visited[id] {
			return nil
		}
		story, ok := stories[id]
		if !ok {
			return fmt.Errorf("story %q depends on non-existent story", id)
		}
		for _, dependency := range story.DependsOn {
			if err := visit(dependency, append(path, id)); err != nil {
				return err
			}
		}
		visited[id] = true
		return nil
	}
	for _, story := range d.Stories {
		if err := visit(story.ID, nil); err != nil {
			return err
		}
	}
	return nil
}

func Load(path string) (*Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read product document %q: %w", path, err)
	}
	var d Document
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&d); err != nil {
		return nil, fmt.Errorf("parse product document %q: %w", path, err)
	}
	if err := d.Validate(); err != nil {
		return nil, fmt.Errorf("validate product document %q: %w", path, err)
	}
	return &d, nil
}

func Save(path string, d *Document) error {
	if err := d.Validate(); err != nil {
		return err
	}
	d.Version++
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return fmt.Errorf("encode product document: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create product directory: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".product-cli-*")
	if err != nil {
		return fmt.Errorf("create temporary product document: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0600); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temporary product document: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temporary product document: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("replace product document: %w", err)
	}
	return nil
}

func MigrateLegacy(workDir, path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, err
	}
	legacy := filepath.Join(workDir, LegacyFilename)
	d, err := LoadLegacy(legacy)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if err := Save(path, d); err != nil {
		return false, err
	}
	if err := os.Remove(legacy); err != nil {
		return false, err
	}
	return true, nil
}

func LoadLegacy(path string) (*Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var d Document
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&d); err != nil {
		return nil, fmt.Errorf("parse legacy product document: %w", err)
	}
	if d.Mode == "" {
		d.Mode = Mode
	}
	if err := d.Validate(); err != nil {
		return nil, fmt.Errorf("validate legacy product document: %w", err)
	}
	return &d, nil
}

func (d *Document) Progress() (completed, total int) {
	if d == nil {
		return 0, 0
	}
	for _, story := range d.Stories {
		total++
		if story.Passes {
			completed++
		}
	}
	return completed, total
}
