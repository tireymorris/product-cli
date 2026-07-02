package workdir

import (
	"os"
	"path/filepath"
	"strings"
)

var projectManifests = []string{
	"go.mod",
	"package.json",
	"Cargo.toml",
	"pyproject.toml",
	"Gemfile",
	"pom.xml",
	"build.gradle",
	"build.gradle.kts",
	"composer.json",
	"mix.exs",
	"Package.swift",
	"CMakeLists.txt",
	"Makefile",
}

var sourceExtensions = map[string]bool{
	".go": true, ".py": true, ".js": true, ".ts": true, ".jsx": true, ".tsx": true,
	".rb": true, ".java": true, ".rs": true, ".c": true, ".cpp": true, ".cs": true,
	".php": true, ".swift": true, ".kt": true, ".ex": true, ".hs": true, ".scala": true,
	".sh": true, ".ml": true, ".r": true, ".pl": true, ".lua": true, ".dart": true,
	".vue": true, ".svelte": true, ".html": true, ".css": true, ".scss": true,
}

var skipDirNames = map[string]bool{
	"node_modules": true,
	"vendor":       true,
	"__pycache__":  true,
}

func ContainsSource(workDir string) bool {
	if workDir == "" {
		return false
	}
	for _, name := range projectManifests {
		if fileExists(filepath.Join(workDir, name)) {
			return true
		}
	}
	return containsSourceFile(workDir)
}

func containsSourceFile(workDir string) bool {
	found := false
	_ = filepath.WalkDir(workDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || skipDirNames[name] {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if sourceExtensions[ext] {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
