package service

import (
	"fmt"
	"strconv"

	"repo-nest/internal/db"
)

// ConfigData holds the application configuration sent to the frontend.
type ConfigData struct {
	Config    map[string]string `json:"config"`
	ScanRoots []string          `json:"scan_roots"`
}

// allowedConfigKeys is the allow-list of user-settable configuration keys.
var allowedConfigKeys = map[string]bool{
	"daily_code_standard": true,
	"scan_depth":          true,
	"git_author":          true,
	"auto_import":         true,
	// M1 Claude session auto-capture: "1" enables on-demand capture, anything
	// else (incl. unset) keeps it OFF. Reading session transcripts is
	// privacy-sensitive, so it is never implicitly enabled (ADR-0010).
	"claude_session_capture": true,
	// Target project for the agent-GLOBAL memory importers (openclaw / hermes).
	// Value is a project name or numeric id; unset -> those sources skip. See
	// memsrc.TargetProject and ADR-0011.
	"openclaw_project": true,
	"hermes_project":   true,
}

// stringConfigKeys are exempt from the numeric-value check: they carry
// free-text (author name, project name/id).
var stringConfigKeys = map[string]bool{
	"git_author":       true,
	"openclaw_project": true,
	"hermes_project":   true,
}

// GetConfig returns all configuration settings and scan roots.
func (s *Service) GetConfig() (*ConfigData, error) {
	configs, err := db.GetAllConfigs(s.db)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	roots, err := db.GetScanRoots(s.db)
	if err != nil {
		return nil, fmt.Errorf("failed to load scan roots: %w", err)
	}
	// A nil slice serialises to JSON null, which the frontend spreads/filters
	// directly and would throw. Always emit an empty array instead.
	if roots == nil {
		roots = []string{}
	}
	return &ConfigData{Config: configs, ScanRoots: roots}, nil
}

// UpdateConfig sets a single configuration key-value pair after validating
// the key against the allow-list and numeric values.
func (s *Service) UpdateConfig(key, value string) error {
	if !allowedConfigKeys[key] {
		return fmt.Errorf("unknown config key: %s", key)
	}
	if !stringConfigKeys[key] {
		if _, err := strconv.Atoi(value); err != nil {
			return fmt.Errorf("config value must be a number")
		}
	}
	if err := db.SetConfig(s.db, key, value); err != nil {
		return err
	}
	// git_author is applied immediately so "mine" stats, heatmap and recent
	// commits reflect the new author without a restart.
	if key == "git_author" {
		s.setGitUser(value)
	}
	return nil
}

// UpdateScanRoots replaces the entire scan root list atomically.
func (s *Service) UpdateScanRoots(scanRoots []string) error {
	if err := db.ReplaceScanRoots(s.db, scanRoots); err != nil {
		return fmt.Errorf("failed to update scan roots: %w", err)
	}
	return nil
}
