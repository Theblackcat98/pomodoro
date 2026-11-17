package views

import (
	"testing"

	"pomodoro/src/models"
	"pomodoro/src/services"
)

// recordingStore captures SaveSettings calls; unneeded interface methods are
// satisfied by the embedded (nil) interface and are never reached here.
type recordingStore struct {
	services.PersistenceService
	data  services.AppData
	saves []models.Settings
}

func (s *recordingStore) GetAllData() (*services.AppData, error) {
	d := s.data
	return &d, nil
}

func (s *recordingStore) SaveSettings(settings *models.Settings) error {
	s.saves = append(s.saves, *settings)
	return nil
}

func (s *recordingStore) GetActiveSession() (*models.TimerSession, error) {
	return nil, nil
}

func newThemeModel(store services.PersistenceService) Model {
	m := NewModel(services.NewTimerService(nil), store)
	m.tab = TabSettings
	m.settingsCursor = setRowTheme
	return m
}

func TestThemeRowCyclesDarkLightDark(t *testing.T) {
	defer buildStyles(ThemeDefault)

	m := newThemeModel(nil)
	m.settings.Theme = ThemeDark
	buildStyles(ThemeDark)

	m = driveKey(m, "space")
	if got := m.settings.Theme; got != ThemeLight {
		t.Fatalf("theme after first cycle = %q, want %q", got, ThemeLight)
	}
	m = driveKey(m, "space")
	if got := m.settings.Theme; got != ThemeDark {
		t.Fatalf("theme after second cycle = %q, want %q", got, ThemeDark)
	}
	m = driveKey(m, "space")
	if got := m.settings.Theme; got != ThemeLight {
		t.Fatalf("theme after third cycle = %q, want %q", got, ThemeLight)
	}
}

func TestThemeChoiceIsWrittenToSettings(t *testing.T) {
	defer buildStyles(ThemeDefault)

	s := &recordingStore{}
	s.data.Settings = *models.DefaultSettings()
	s.data.Settings.Theme = ThemeDark
	s.data.Settings.SeenIntro = true
	m := newThemeModel(s)

	before := len(s.saves)
	m = driveKey(m, "space")
	if got := m.settings.Theme; got != ThemeLight {
		t.Fatalf("Settings.Theme after cycle = %q, want %q", got, ThemeLight)
	}
	if len(s.saves) <= before {
		t.Fatal("theme cycle did not persist settings")
	}
	if got := s.saves[len(s.saves)-1].Theme; got != ThemeLight {
		t.Fatalf("persisted Theme = %q, want %q", got, ThemeLight)
	}
}

func TestThemeSwitchChangesRenderedStyle(t *testing.T) {
	defer buildStyles(ThemeDefault)

	m := newThemeModel(nil)
	m.settings.Theme = ThemeDark
	buildStyles(ThemeDark)
	darkRender := appStyle.Render("theme")

	m = driveKey(m, "space")
	lightRender := appStyle.Render("theme")
	if darkRender == lightRender {
		t.Fatalf("rendered style unchanged after switching to light: %q", lightRender)
	}

	m = driveKey(m, "space")
	if got := appStyle.Render("theme"); got != darkRender {
		t.Fatalf("rendered style after switching back = %q, want original dark %q", got, darkRender)
	}
}

func TestDefaultThemeResolvesToDarkPalette(t *testing.T) {
	defer buildStyles(ThemeDefault)

	if got := models.DefaultSettings().Theme; got != ThemeDefault {
		t.Fatalf("DefaultSettings.Theme = %q, want %q", got, ThemeDefault)
	}

	buildStyles(ThemeDefault)
	got := appStyle.Render("theme")
	buildStyles(ThemeDark)
	dark := appStyle.Render("theme")
	if got != dark {
		t.Fatalf("default palette render = %q, want dark palette render %q", got, dark)
	}
	buildStyles(ThemeLight)
	if got := appStyle.Render("theme"); got == dark {
		t.Fatalf("light palette render equals dark palette render %q", got)
	}
	if !resolvePalette(ThemeDefault).dark {
		t.Fatal("resolvePalette(default) must be dark")
	}
}
