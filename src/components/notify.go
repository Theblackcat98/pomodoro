package components

import (
	"fmt"
	"os"
	"os/exec"

	clog "github.com/charmbracelet/log"
	"github.com/gen2brain/beeep"
	"pomodoro/src/models"
)

// hookLogger receives hook/notification failures; nil disables logging.
var hookLogger *clog.Logger

// SetLogger installs the diagnostics logger for hooks and notifications.
func SetLogger(l *clog.Logger) { hookLogger = l }

// External notifications, layered per terminal support (see the improvement
// plan's research): BEL is the universal floor, OSC 9/777 turn into real
// desktop notifications on terminals that support them (kitty, WezTerm,
// Ghostty, foot, iTerm2 — ignored harmlessly elsewhere), and beeep covers
// the rest via the OS notification daemon.
//
// Everything is written to stderr: bubbletea owns stdout for its renderer,
// and a raw escape interleaved with a frame would corrupt the display. The
// terminal emulator parses BEL/OSC from the shared pty either way.

// NotifyExternal delivers the alert matching the configured mode.
// The visual mode is intentionally a no-op here: the in-app flash banner
// already covers it, and sound is played by the Notifier.
func NotifyExternal(mode models.NotificationType, title, body string) {
	switch mode {
	case models.BellNotification:
		BellAndOSC(title, body)
	case models.DesktopNotification:
		// Best-effort; without a notification daemon this fails silently.
		if err := beeep.Notify(title, body, ""); err != nil && hookLogger != nil {
			hookLogger.Error("desktop notification", "error", err)
		}
	}
}

// BellAndOSC rings the terminal bell and emits the OSC 9 and OSC 777
// notification sequences.
func BellAndOSC(title, body string) {
	// BEL — every terminal honors it (often mapped to "urgent" or a
	// desktop notification when unfocused).
	fmt.Fprint(os.Stderr, "\a")
	// OSC 9 (iTerm2-origin) and OSC 777 (urxvt-origin): terminals that
	// understand them surface a desktop notification; the rest ignore
	// unknown OSC codes.
	fmt.Fprintf(os.Stderr, "\x1b]9;%s: %s\x07", title, body)
	fmt.Fprintf(os.Stderr, "\x1b]777;notify;%s;%s\x07", title, body)
}

// HookEnv runs a user-configured hook command with POMODORO_STATE in the
// environment. Fire-and-forget: hooks must never block the UI.
func HookEnv(command, state string) {
	if command == "" {
		return
	}
	go func() {
		cmd := exec.Command("sh", "-c", command)
		cmd.Env = append(os.Environ(), "POMODORO_STATE="+state)
		if err := cmd.Run(); err != nil && hookLogger != nil {
			hookLogger.Error("hook failed", "state", state, "error", err)
		}
	}()
}
