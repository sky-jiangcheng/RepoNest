// Package platform centralises OS-specific behaviour: default scan roots,
// user data locations (database, plugins, log file) and git user detection.
//
// User data paths are stable across versions: an installation that already
// has a legacy data directory (gitboard, then gitbuddy) is renamed to dirName
// ("reponest") on first launch after the rename, so the database, plugins and
// logs carry over instead of being orphaned.
package platform

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// dirName is the user data directory and the log file basename.
const dirName = "reponest"

// legacyDirName is the data directory used before the GitBuddy -> RepoNest
// rename (which itself succeeded the GitBoard era). The spelling is
// intentional: it names data written by older builds and is only read by the
// migration below.
const legacyDirName = "gitbuddy"

// DefaultScanRoots returns platform-specific default scan root directories.
// Windows: all drive letters except C: (the system drive).
// macOS / Linux: the user's home directory.
func DefaultScanRoots() []string {
	home, _ := os.UserHomeDir()

	switch runtime.GOOS {
	case "windows":
		roots := []string{}
		for _, drive := range getWindowsDrives() {
			upper := strings.ToUpper(drive)
			if upper != "C:" && upper != "C:\\" {
				roots = append(roots, drive)
			}
		}
		if len(roots) == 0 {
			roots = append(roots, home)
		}
		return roots
	default: // darwin, linux and others
		return []string{home}
	}
}

// getWindowsDrives enumerates available drive letters on Windows.
// On non-Windows platforms, returns an empty slice.
func getWindowsDrives() []string {
	if runtime.GOOS != "windows" {
		return nil
	}
	drives := []string{}
	for c := 'A'; c <= 'Z'; c++ {
		path := string(c) + ":\\"
		if _, err := os.Stat(path); err == nil {
			drives = append(drives, path)
		}
	}
	return drives
}

// GetGitUserName returns the git user.name from global or local config.
func GetGitUserName() string {
	cmd := exec.Command("git", "config", "user.name")
	out, err := cmd.Output()
	if err != nil {
		// fallback to OS username
		if u, e := os.UserHomeDir(); e == nil {
			return filepath.Base(u)
		}
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

var migrateLegacyOnce sync.Once

// migrateLegacyData renames the pre-rename "gitboard" data directory to
// dirName, moving the database and plugins across. It is a no-op on a fresh
// install or once the rename has happened. Failures are ignored on purpose:
// a move that cannot complete (permissions, cross-device) must not stop the
// app from starting, and the old directory stays in place to be moved by hand.
func migrateLegacyData() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return
	}
	old := filepath.Join(configDir, legacyDirName)
	if _, err := os.Stat(old); err != nil {
		return
	}
	current := filepath.Join(configDir, dirName)
	if _, err := os.Stat(current); err == nil {
		return
	}
	_ = os.Rename(old, current)
}

// fallbackDir returns a private (0700) directory for degraded environments
// where the platform config dir is unavailable. os.MkdirTemp creates the
// directory with an unpredictable name, so another local process cannot
// pre-create it to plant files (a fixed /tmp/reponest* path could be —
// CWE-379). It is created once per process and reused, so the DB and plugin
// paths stay stable within a run.
var fallbackDirOnce struct {
	sync.Once
	path string
}

func fallbackDir() string {
	fallbackDirOnce.Do(func() {
		dir, err := os.MkdirTemp("", dirName+"-")
		if err != nil {
			// MkdirTemp failing means the temp root itself is unusable; keep
			// going with a best-effort name rather than failing the app.
			dir = filepath.Join(os.TempDir(), dirName+"-degraded")
		}
		fallbackDirOnce.path = dir
	})
	return fallbackDirOnce.path
}

// GetDbPath returns the path to the SQLite database file.
func GetDbPath() string {
	migrateLegacyOnce.Do(migrateLegacyData)

	if configDir, err := os.UserConfigDir(); err == nil {
		dir := filepath.Join(configDir, dirName)
		if err := os.MkdirAll(dir, 0750); err == nil {
			return filepath.Join(dir, "dashboard.db")
		}
	}
	return filepath.Join(fallbackDir(), "dashboard.db")
}

// GetPluginsDir returns the directory that holds plugin directories
// (one subdirectory per plugin, each containing plugin.go).
func GetPluginsDir() string {
	if configDir, err := os.UserConfigDir(); err == nil {
		dir := filepath.Join(configDir, dirName, "plugins")
		if err := os.MkdirAll(dir, 0750); err == nil {
			return dir
		}
	}
	return filepath.Join(fallbackDir(), "plugins")
}

// migrateLegacyLog renames the log file left behind by earlier versions
// (gitboard.log -> reponest.log). It runs after the log directory is known,
// which covers every platform: on Windows the log lives inside the config
// directory that migrateLegacyData already renamed, so only the file name is
// left to fix there.
func migrateLegacyLog(dir string) {
	old := filepath.Join(dir, legacyDirName+".log")
	if _, err := os.Stat(old); err != nil {
		return
	}
	current := filepath.Join(dir, dirName+".log")
	if _, err := os.Stat(current); err == nil {
		return
	}
	_ = os.Rename(old, current)
}

// GetLogPath returns the platform-appropriate log file path.
//
//	darwin:  ~/Library/Logs/reponest.log
//	windows: %APPDATA%\reponest\logs\reponest.log
//	linux:   $XDG_STATE_HOME/reponest/reponest.log (default ~/.local/state/...)
func GetLogPath() string {
	var dir string
	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return filepath.Join(os.TempDir(), dirName+".log")
		}
		dir = filepath.Join(home, "Library", "Logs")
	case "windows":
		appData, err := os.UserConfigDir()
		if err != nil {
			return filepath.Join(os.TempDir(), dirName+".log")
		}
		dir = filepath.Join(appData, dirName, "logs")
	default:
		state := os.Getenv("XDG_STATE_HOME")
		if state == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return filepath.Join(os.TempDir(), dirName+".log")
			}
			state = filepath.Join(home, ".local", "state")
		}
		dir = filepath.Join(state, dirName)
	}
	if err := os.MkdirAll(dir, 0750); err != nil {
		return filepath.Join(os.TempDir(), dirName+".log")
	}
	migrateLegacyLog(dir)
	return filepath.Join(dir, dirName+".log")
}
