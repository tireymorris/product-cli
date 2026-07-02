package prd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"ralph/internal/shared/config"
	"ralph/internal/shared/constants"
)

// Load reads and parses the PRD under a shared lock.
func Load(cfg *config.Config) (*PRD, error) {
	prdPath := cfg.PRDPath()

	fileLock, err := acquireSharedLock(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock for reading %q: %w", prdPath, err)
	}
	defer fileLock.Unlock()

	data, err := os.ReadFile(prdPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read PRD file %q: %w", prdPath, err)
	}

	if err := rejectLegacyAcceptanceCriteriaInJSON(data); err != nil {
		return nil, fmt.Errorf("PRD validation failed for %q: %w", prdPath, err)
	}

	if isProductDocumentPath(prdPath) {
		var product productDocument
		dec := json.NewDecoder(bytes.NewReader(data))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&product); err != nil {
			return nil, fmt.Errorf("failed to parse product file %q: %w", prdPath, err)
		}
		if err := product.validate(); err != nil {
			return nil, fmt.Errorf("product validation failed for %q: %w", prdPath, err)
		}
		return product.toPRD(), nil
	}

	var p PRD
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("failed to parse PRD file %q: %w", prdPath, err)
	}

	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("PRD validation failed for %q: %w", prdPath, err)
	}

	return &p, nil
}

// Save writes the PRD atomically under an exclusive lock and increments Version.
func Save(cfg *config.Config, p *PRD) error {
	prdPath := cfg.PRDPath()

	isProduct := isProductDocumentPath(prdPath)
	if isProduct {
		if err := toProductDocument(p).validate(); err != nil {
			return fmt.Errorf("product validation failed before saving %q: %w", prdPath, err)
		}
	} else {
		if err := p.Validate(); err != nil {
			return fmt.Errorf("PRD validation failed before saving %q: %w", prdPath, err)
		}
	}

	fileLock, err := acquireExclusiveLock(cfg)
	if err != nil {
		return fmt.Errorf("failed to acquire lock for writing %q: %w", prdPath, err)
	}
	defer fileLock.Unlock()

	p.Version++

	var (
		data []byte
	)
	if isProduct {
		data, err = json.MarshalIndent(toProductDocument(p), "", "  ")
	} else {
		data, err = json.MarshalIndent(p, "", "  ")
	}
	if err != nil {
		kind := "PRD"
		if isProduct {
			kind = "product document"
		}
		return fmt.Errorf("failed to marshal %s %q (version %d): %w", kind, prdPath, p.Version, err)
	}

	tmpDir := filepath.Join(filepath.Dir(prdPath), ".ralph")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return fmt.Errorf("failed to create state dir %q: %w", tmpDir, err)
	}

	tmpPath := filepath.Join(tmpDir, fmt.Sprintf("prd.tmp.%d.%d", time.Now().Unix(), rand.Intn(constants.TempFileRandomRange)))

	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write temporary PRD file %q: %w", tmpPath, err)
	}

	if err := os.Rename(tmpPath, prdPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to atomically replace PRD file %q from temporary %q: %w", prdPath, tmpPath, err)
	}

	return nil
}

func Exists(cfg *config.Config) (bool, error) {
	_, err := os.Stat(cfg.PRDPath())
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("failed to stat PRD file %q: %w", cfg.PRDPath(), err)
}

func isProductDocumentPath(path string) bool {
	return filepath.Base(path) == "product.json"
}

// IsProductDocument reports whether path points at a product.json artifact.
func IsProductDocument(path string) bool {
	return isProductDocumentPath(path)
}

// ResolveExistingDocument returns cfg when its PRD file exists, otherwise a
// product.json config when that file exists in the same work directory.
func ResolveExistingDocument(cfg *config.Config) (*config.Config, error) {
	if exists, err := Exists(cfg); err != nil {
		return nil, err
	} else if exists {
		return cfg, nil
	}
	if cfg.PRDFile == "product.json" {
		return nil, nil
	}
	productCfg := *cfg
	productCfg.PRDFile = "product.json"
	exists, err := Exists(&productCfg)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}
	return &productCfg, nil
}

// EncodeDocument writes p using the JSON shape for documentPath.
func EncodeDocument(w io.Writer, documentPath string, p *PRD) error {
	enc := json.NewEncoder(w)
	if isProductDocumentPath(documentPath) {
		return enc.Encode(toProductDocument(p))
	}
	return enc.Encode(p)
}
