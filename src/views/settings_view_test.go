package views

import (
	"testing"
	"time"
)

func TestSettingsClampDuration(t *testing.T) {
	tests := []struct {
		name                 string
		cur, delta, min, max time.Duration
		want                 time.Duration
	}{
		{"below minimum clamps up", time.Minute, -5 * time.Minute, 5 * time.Minute, 90 * time.Minute, 5 * time.Minute},
		{"above maximum clamps down", 85 * time.Minute, 10 * time.Minute, 5 * time.Minute, 90 * time.Minute, 90 * time.Minute},
		{"in range passes through", 25 * time.Minute, 5 * time.Minute, 5 * time.Minute, 90 * time.Minute, 30 * time.Minute},
		{"at minimum stays", 5 * time.Minute, -5 * time.Minute, 5 * time.Minute, 90 * time.Minute, 5 * time.Minute},
		{"at maximum stays", 90 * time.Minute, 5 * time.Minute, 5 * time.Minute, 90 * time.Minute, 90 * time.Minute},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := clampDuration(tc.cur, tc.delta, tc.min, tc.max)
			if got != tc.want {
				t.Fatalf("clampDuration(%v, %v, %v, %v) = %v, want %v", tc.cur, tc.delta, tc.min, tc.max, got, tc.want)
			}
		})
	}
}

func TestSettingsClampInt(t *testing.T) {
	tests := []struct {
		name              string
		v, min, max, want int
	}{
		{"below minimum clamps up", 0, 1, 12, 1},
		{"above maximum clamps down", 20, 1, 12, 12},
		{"in range passes through", 6, 1, 12, 6},
		{"at minimum stays", 1, 1, 12, 1},
		{"at maximum stays", 12, 1, 12, 12},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := clampInt(tc.v, tc.min, tc.max)
			if got != tc.want {
				t.Fatalf("clampInt(%d, %d, %d) = %d, want %d", tc.v, tc.min, tc.max, got, tc.want)
			}
		})
	}
}

func TestSettingsCursorBounds(t *testing.T) {
	last := setRowCount - 1
	cases := []struct {
		name, key   string
		start, want int
	}{
		{"down at last row stays", "j", last, last},
		{"arrow-down at last row stays", "down", last, last},
		{"up at first row stays", "k", 0, 0},
		{"arrow-up at first row stays", "up", 0, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := newBareModel()
			m.tab = TabSettings
			m.settingsCursor = tc.start
			m = driveKey(m, tc.key)
			if m.settingsCursor != tc.want {
				t.Fatalf("cursor after %q from %d = %d, want %d", tc.key, tc.start, m.settingsCursor, tc.want)
			}
			if m.settingsCursor < 0 || m.settingsCursor > last {
				t.Fatalf("cursor %d outside [0,%d] after %q", m.settingsCursor, last, tc.key)
			}
		})
	}

	// Sweep past both ends and confirm the cursor never leaves the row range.
	m := newBareModel()
	m.tab = TabSettings
	for i := 0; i < 50; i++ {
		m = driveKey(m, "j")
		if m.settingsCursor < 0 || m.settingsCursor > last {
			t.Fatalf("cursor %d outside [0,%d] after %d down presses", m.settingsCursor, last, i+1)
		}
	}
	if m.settingsCursor != last {
		t.Fatalf("cursor after repeated down = %d, want last row %d", m.settingsCursor, last)
	}
	for i := 0; i < 50; i++ {
		m = driveKey(m, "k")
		if m.settingsCursor < 0 || m.settingsCursor > last {
			t.Fatalf("cursor %d outside [0,%d] after %d up presses", m.settingsCursor, last, i+1)
		}
	}
	if m.settingsCursor != 0 {
		t.Fatalf("cursor after repeated up = %d, want first row 0", m.settingsCursor)
	}
}

func TestSettingsTogglePersistence(t *testing.T) {
	tests := []struct {
		name string
		row  int
		get  func(Model) bool
	}{
		{"auto-start breaks", setRowAutoBreak, func(m Model) bool { return m.settings.AutoStartBreaks }},
		{"auto-start focus", setRowAutoFocus, func(m Model) bool { return m.settings.AutoStartFocus }},
		{"mouse", setRowMouse, func(m Model) bool { return m.settings.Mouse }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := newBareModel()
			m.tab = TabSettings
			m.settingsCursor = tc.row
			before := tc.get(m)
			m = driveKey(m, "space")
			if got := tc.get(m); got != !before {
				t.Fatalf("space on %s row = %v, want toggled to %v", tc.name, got, !before)
			}
			// A second toggle must write the original value back.
			m = driveKey(m, "space")
			if got := tc.get(m); got != before {
				t.Fatalf("second space on %s row = %v, want restored %v", tc.name, got, before)
			}
		})
	}
}

func TestSettingsHookEditor(t *testing.T) {
	t.Run("enter commits typed value", func(t *testing.T) {
		m := newBareModel()
		m.tab = TabSettings
		m.settingsCursor = setRowHookFocus
		m = driveKey(m, "enter")
		if m.editingHook != setRowHookFocus {
			t.Fatalf("enter on hook row set editingHook = %d, want %d", m.editingHook, setRowHookFocus)
		}

		const want = "echo focus-start"
		for _, c := range want {
			m = driveKey(m, string(c))
		}
		if got := m.hookInput.Value(); got != want {
			t.Fatalf("hook input after typing = %q, want %q", got, want)
		}

		m = driveKey(m, "enter")
		if m.editingHook != -1 {
			t.Fatalf("editingHook after commit = %d, want -1 (closed)", m.editingHook)
		}
		if got := m.settings.HookFocusStart; got != want {
			t.Fatalf("HookFocusStart after commit = %q, want %q", got, want)
		}
	})

	t.Run("esc discards typed value", func(t *testing.T) {
		m := newBareModel()
		m.tab = TabSettings
		m.settingsCursor = setRowHookBreak
		original := m.settings.HookBreakStart
		m = driveKey(m, "enter")
		if m.editingHook != setRowHookBreak {
			t.Fatalf("enter on hook row set editingHook = %d, want %d", m.editingHook, setRowHookBreak)
		}
		for _, c := range "echo discard-me" {
			m = driveKey(m, string(c))
		}

		// The app-level esc case in handleKey returns before routing keys to
		// the settings tab, so deliver esc to the real settings key handler
		// (which forwards it to the editor) to exercise the discard path.
		next, _ := m.handleSettingsKey(keyPress("esc"), "esc")
		m = next.(Model)
		if m.editingHook != -1 {
			t.Fatalf("editingHook after esc = %d, want -1 (closed)", m.editingHook)
		}
		if got := m.settings.HookBreakStart; got != original {
			t.Fatalf("HookBreakStart after esc = %q, want original %q (discarded)", got, original)
		}
	})
}
