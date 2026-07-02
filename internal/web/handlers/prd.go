package handlers

import (
	"net/http"
	"os"

	"ralph/internal/shared/prd"
	"ralph/internal/web/runs"
)

func (a *API) GetRunPRD(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	runCfg := *a.cfg
	switch id {
	case runs.LocalPRDRunID:
		selectedCfg, ok := runs.LocalDocumentConfig(a.cfg)
		if !ok {
			writeJSONError(w, http.StatusNotFound, "run not found")
			return
		}
		runCfg = *selectedCfg
	default:
		run, ok := a.registry.Get(id)
		if !ok {
			writeJSONError(w, http.StatusNotFound, "run not found")
			return
		}
		runCfg.WorkDir = run.WorkDir
		if run.PRDPath != "" {
			runCfg.PRDFile = run.PRDPath
		}
	}
	prdPath := runCfg.PRDPath()
	if _, err := os.Stat(prdPath); err != nil {
		writeJSONError(w, http.StatusNotFound, "prd not found")
		return
	}

	p, err := prd.Load(&runCfg)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "load prd")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = prd.EncodeDocument(w, p)
}
