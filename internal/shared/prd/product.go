package prd

import (
	"errors"
	"fmt"
)

type productDocument struct {
	Mode        string          `json:"mode"`
	Version     int64           `json:"version"`
	ProjectName string          `json:"project_name"`
	BranchName  string          `json:"branch_name,omitempty"`
	Stories     []*productStory `json:"stories"`
}

type productStory struct {
	ID          string          `json:"id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Slices      []*productSlice `json:"slices,omitempty"`
	Priority    int             `json:"priority"`
	DependsOn   []string        `json:"depends_on,omitempty"`
	Passes      bool            `json:"passes"`
}

type productSlice struct {
	ID       string `json:"id"`
	Behavior string `json:"behavior"`
	Passes   bool   `json:"passes"`
}

func (p *productDocument) validate() error {
	if len(p.Stories) > MaxStories {
		return fmt.Errorf("story count %d exceeds maximum %d", len(p.Stories), MaxStories)
	}

	seenIDs := make(map[string]bool)
	for i, story := range p.Stories {
		storyID := "<nil>"
		if story != nil {
			storyID = story.ID
		}
		if err := story.validate(seenIDs); err != nil {
			return fmt.Errorf("story %d (%q): %w", i, storyID, err)
		}
		seenIDs[storyID] = true
	}
	if err := p.validateDependencies(); err != nil {
		return fmt.Errorf("invalid dependencies: %w", err)
	}
	return nil
}

func (p *productDocument) validateDependencies() error {
	visited := make(map[string]bool)
	var dfs func(id string, path []string) error
	dfs = func(id string, path []string) error {
		if id == "" {
			return nil
		}
		for _, pathID := range path {
			if pathID == id {
				cycle := append(path, id)
				return fmt.Errorf("circular dependency detected: %v", cycle)
			}
		}
		if visited[id] {
			return nil
		}
		visited[id] = true
		story := p.getStory(id)
		if story == nil {
			return fmt.Errorf("story %q depends on non-existent story", id)
		}
		for _, depID := range story.DependsOn {
			if err := dfs(depID, append(path, id)); err != nil {
				return err
			}
		}
		return nil
	}

	for _, story := range p.Stories {
		if story == nil {
			return errors.New("story cannot be nil")
		}
		if err := dfs(story.ID, nil); err != nil {
			return err
		}
	}
	return nil
}

func (p *productDocument) getStory(id string) *productStory {
	for _, story := range p.Stories {
		if story.ID == id {
			return story
		}
	}
	return nil
}

func (s *productStory) validate(seenIDs map[string]bool) error {
	if s == nil {
		return errors.New("story cannot be nil")
	}
	if s.ID == "" {
		return errors.New("story ID cannot be empty")
	}
	if seenIDs[s.ID] {
		return fmt.Errorf("duplicate story ID %q", s.ID)
	}
	if s.Title == "" {
		return errors.New("story title cannot be empty")
	}
	if len(s.Description) > MaxStoryDescSize {
		return fmt.Errorf("story description size %d exceeds maximum %d bytes", len(s.Description), MaxStoryDescSize)
	}
	if s.Priority < 0 {
		return fmt.Errorf("story priority %d cannot be negative", s.Priority)
	}
	if len(s.Slices) == 0 {
		return fmt.Errorf("story %q must have at least one slice", s.ID)
	}
	sliceIDs := make(map[string]bool)
	for i, sl := range s.Slices {
		if err := sl.validate(s.ID, i, sliceIDs); err != nil {
			return err
		}
		sliceIDs[sl.ID] = true
	}
	for _, dep := range s.DependsOn {
		if dep == "" {
			return fmt.Errorf("story %q has empty dependency ID", s.ID)
		}
		if dep == s.ID {
			return fmt.Errorf("story %q cannot depend on itself", s.ID)
		}
	}
	return nil
}

func (s *productSlice) validate(storyID string, index int, seenIDs map[string]bool) error {
	if s == nil {
		return fmt.Errorf("story %q slice %d cannot be nil", storyID, index)
	}
	if s.ID == "" {
		return fmt.Errorf("story %q slice %d id cannot be empty", storyID, index)
	}
	if seenIDs[s.ID] {
		return fmt.Errorf("story %q has duplicate slice ID %q", storyID, s.ID)
	}
	if s.Behavior == "" {
		return fmt.Errorf("story %q slice %q behavior cannot be empty", storyID, s.ID)
	}
	return nil
}

func toProductDocument(p *PRD) *productDocument {
	out := &productDocument{
		Mode:        ModeProduct,
		Version:     p.Version,
		ProjectName: p.ProjectName,
		BranchName:  p.BranchName,
		Stories:     make([]*productStory, 0, len(p.Stories)),
	}
	for _, story := range p.Stories {
		out.Stories = append(out.Stories, toProductStory(story))
	}
	return out
}

func toProductStory(story *Story) *productStory {
	if story == nil {
		return nil
	}
	out := &productStory{
		ID:          story.ID,
		Title:       story.Title,
		Description: story.Description,
		Priority:    story.Priority,
		DependsOn:   append([]string(nil), story.DependsOn...),
		Passes:      story.Passes,
		Slices:      make([]*productSlice, 0, len(story.Slices)),
	}
	for _, slice := range story.Slices {
		out.Slices = append(out.Slices, toProductSlice(slice))
	}
	return out
}

func toProductSlice(slice *Slice) *productSlice {
	if slice == nil {
		return nil
	}
	return &productSlice{
		ID:       slice.ID,
		Behavior: slice.Behavior,
		Passes:   slice.Passes,
	}
}

func (p *productDocument) toPRD() *PRD {
	out := &PRD{
		Mode:        ModeProduct,
		Version:     p.Version,
		ProjectName: p.ProjectName,
		BranchName:  p.BranchName,
		Stories:     make([]*Story, 0, len(p.Stories)),
	}
	for _, story := range p.Stories {
		out.Stories = append(out.Stories, story.toPRD())
	}
	return out
}

func (s *productStory) toPRD() *Story {
	if s == nil {
		return nil
	}
	out := &Story{
		ID:          s.ID,
		Title:       s.Title,
		Description: s.Description,
		Priority:    s.Priority,
		DependsOn:   append([]string(nil), s.DependsOn...),
		Passes:      s.Passes,
		Slices:      make([]*Slice, 0, len(s.Slices)),
	}
	for _, slice := range s.Slices {
		out.Slices = append(out.Slices, slice.toPRD())
	}
	return out
}

func (s *productSlice) toPRD() *Slice {
	if s == nil {
		return nil
	}
	return &Slice{
		ID:       s.ID,
		Behavior: s.Behavior,
		Passes:   s.Passes,
	}
}
