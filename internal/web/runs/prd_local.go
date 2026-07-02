package runs

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runpaths"
	"ralph/internal/shared/runstate"
)

// LocalPRDRunID is the stable API id for an in-progress TUI/CLI run backed only by prd.json.
const LocalPRDRunID = runstate.LocalRunID

// OngoingLocalPRD reports a synthetic run when prd.json exists and the run is still
// active (implementation in progress, or a product planning document). Product
// documents remain visible even when every story is marked passes.
func OngoingLocalPRD(cfg *config.Config, registry *Registry) (*Run, bool) {
	if _, ok := registry.ActiveForWorkDir(cfg.WorkDir); ok {
		return nil, false
	}

	if _, err := prd.MigrateLegacyProductIfNeeded(cfg); err != nil {
		return nil, false
	}

	selectedCfg, ok := localPRDConfig(cfg)
	if !ok {
		return nil, false
	}

	p, err := prd.Load(selectedCfg)
	if err != nil {
		return nil, false
	}
	if !p.IsProduct() && p.AllCompleted() {
		return nil, false
	}

	prdPath := selectedCfg.PRDPath()
	info, err := os.Stat(prdPath)
	if err != nil {
		return nil, false
	}
	mod := info.ModTime()
	if mod.IsZero() {
		mod = time.Now()
	}

	meta := loadLocalPRDMeta(cfg.WorkDir)
	status, phase := localPRDStatus(p, meta.Checkpoint)
	run := &Run{
		ID:              LocalPRDRunID,
		WorkDir:         cfg.WorkDir,
		Prompt:          localPRDPrompt(p),
		Status:          status,
		Phase:           phase,
		CreatedAt:       mod,
		UpdatedAt:       mod,
		PRDPath:         cfg.PRDFile,
		ReviewLoopState: meta.ReviewLoopState,
	}
	if !meta.UpdatedAt.IsZero() {
		run.UpdatedAt = meta.UpdatedAt
	}
	return run, true
}

func localPRDConfig(cfg *config.Config) (*config.Config, bool) {
	return LocalDocumentConfig(cfg)
}

// LocalDocumentConfig reports whether prd.json exists in the work directory.
func LocalDocumentConfig(cfg *config.Config) (*config.Config, bool) {
	if _, err := prd.MigrateLegacyProductIfNeeded(cfg); err != nil {
		return nil, false
	}
	exists, err := prd.Exists(cfg)
	if err != nil || !exists {
		return nil, false
	}
	return cfg, true
}

type localPRDMeta struct {
	runstate.ReviewLoopState
	UpdatedAt time.Time `json:"updated_at"`
}

func loadLocalPRDMeta(workDir string) localPRDMeta {
	path := runpaths.MetaPath(workDir, LocalPRDRunID)
	data, err := os.ReadFile(path)
	if err != nil {
		return localPRDMeta{}
	}
	var m localPRDMeta
	if err := json.Unmarshal(data, &m); err != nil {
		return localPRDMeta{}
	}
	return m
}

func localPRDStatus(p *prd.PRD, checkpoint string) (status, phase string) {
	return runstate.LocalPRDStatusPhase(p, checkpoint, "")
}

func localPRDPrompt(p *prd.PRD) string {
	if name := strings.TrimSpace(p.ProjectName); name != "" {
		return name
	}
	ctx := strings.TrimSpace(p.Context)
	if ctx == "" {
		return "Local PRD run"
	}
	if i := strings.IndexByte(ctx, '\n'); i >= 0 {
		ctx = strings.TrimSpace(ctx[:i])
	}
	const maxLen = 120
	if len(ctx) > maxLen {
		return ctx[:maxLen] + "…"
	}
	return ctx
}
