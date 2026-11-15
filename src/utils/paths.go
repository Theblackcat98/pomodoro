package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// StateDir returns (and creates) the pomodoro state directory for logs and
// other runtime artifacts: $XDG_STATE_HOME/pomodoro, falling back to
// ~/.local/state/pomodoro.
func StateDir() (string, error) {
	base := os.Getenv("XDG_STATE_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home dir: %w", err)
		}
		base = filepath.Join(home, ".local", "state")
	}
	dir := filepath.Join(base, "pomodoro")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create state dir: %w", err)
	}
	return dir, nil
}
