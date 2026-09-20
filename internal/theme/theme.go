package theme

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

type ThemeInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Author      string `json:"author"`
	Description string `json:"description"`
	Screenshot  string `json:"screenshot,omitempty"`
	IsActive    bool   `json:"is_active"`
	HasBuild    bool   `json:"has_build"`
}

type SettingsStore interface {
	Get(ctx context.Context, key string) (json.RawMessage, error)
	Set(ctx context.Context, key string, value interface{}) error
}

type ThemeManager struct {
	themesDir    string
	settingsRepo SettingsStore
}

func NewThemeManager(themesDir string, settingsRepo SettingsStore) *ThemeManager {
	return &ThemeManager{
		themesDir:    themesDir,
		settingsRepo: settingsRepo,
	}
}

// GetActiveThemeID haalt het actieve thema op uit de database met fallback naar "aesthetic"
func (m *ThemeManager) GetActiveThemeID(ctx context.Context) string {
	raw, err := m.settingsRepo.Get(ctx, "active_theme")
	if err == nil && len(raw) > 0 {
		var themeID string
		if err := json.Unmarshal(raw, &themeID); err == nil && themeID != "" {
			return themeID
		}
	}
	return "aesthetic"
}

// SetActiveThemeID activeert een ander thema in de database
func (m *ThemeManager) SetActiveThemeID(ctx context.Context, themeID string) error {
	return m.settingsRepo.Set(ctx, "active_theme", themeID)
}

// ListThemes ontdekt alle thema's in de ./themes/ map
func (m *ThemeManager) ListThemes(ctx context.Context) ([]ThemeInfo, error) {
	activeTheme := m.GetActiveThemeID(ctx)
	var result []ThemeInfo

	entries, err := os.ReadDir(m.themesDir)
	if err != nil {
		if os.IsNotExist(err) {
			_ = os.MkdirAll(m.themesDir, 0755)
			return []ThemeInfo{}, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		themeJSONPath := filepath.Join(m.themesDir, entry.Name(), "theme.json")
		data, err := os.ReadFile(themeJSONPath)
		if err != nil {
			continue
		}

		var info ThemeInfo
		if err := json.Unmarshal(data, &info); err != nil {
			log.Printf("Fout bij parsen van %s: %v", themeJSONPath, err)
			continue
		}

		if info.ID == "" {
			info.ID = entry.Name()
		}

		distIndex := filepath.Join(m.themesDir, entry.Name(), "dist", "index.html")
		_, statErr := os.Stat(distIndex)
		info.HasBuild = statErr == nil
		info.IsActive = (info.ID == activeTheme)

		result = append(result, info)
	}

	return result, nil
}

// ResolveDistDir geeft de map terug waaruit assets en index.html moeten worden geserveerd
func (m *ThemeManager) ResolveDistDir(ctx context.Context) string {
	activeID := m.GetActiveThemeID(ctx)

	// 1. Probeer de dist van het actieve thema
	candidate := filepath.Join(m.themesDir, activeID, "dist")
	if _, err := os.Stat(filepath.Join(candidate, "index.html")); err == nil {
		return candidate
	}

	// 2. Fallback naar ./frontend/dist
	if _, err := os.Stat("./frontend/dist/index.html"); err == nil {
		return "./frontend/dist"
	}

	return candidate
}