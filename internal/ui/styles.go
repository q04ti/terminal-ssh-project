package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Theme represents a color palette.
type Theme struct {
	Name      string
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Highlight lipgloss.Color
	Text      lipgloss.Color
	Muted     lipgloss.Color
	Border    lipgloss.Color
	Error     lipgloss.Color
	BgAlt     lipgloss.Color
}

var Themes = map[string]Theme{
	"cyan": {
		Name:      "Cyan Void",
		Primary:   lipgloss.Color("#00e5ff"),
		Secondary: lipgloss.Color("#9c27b0"),
		Highlight: lipgloss.Color("#69f0ae"),
		Text:      lipgloss.Color("#eceff1"),
		Muted:     lipgloss.Color("#546e7a"),
		Border:    lipgloss.Color("#00838f"),
		Error:     lipgloss.Color("#ff5252"),
		BgAlt:     lipgloss.Color("#102027"),
	},
	"purple": {
		Name:      "Neon Dusk",
		Primary:   lipgloss.Color("#d500f9"),
		Secondary: lipgloss.Color("#00e5ff"),
		Highlight: lipgloss.Color("#ff4081"),
		Text:      lipgloss.Color("#f3e5f5"),
		Muted:     lipgloss.Color("#7b1fa2"),
		Border:    lipgloss.Color("#aa00ff"),
		Error:     lipgloss.Color("#ff1744"),
		BgAlt:     lipgloss.Color("#210038"),
	},
	"green": {
		Name:      "Cyber Phosphor",
		Primary:   lipgloss.Color("#00e676"),
		Secondary: lipgloss.Color("#00b0ff"),
		Highlight: lipgloss.Color("#76ff03"),
		Text:      lipgloss.Color("#e8f5e9"),
		Muted:     lipgloss.Color("#2e7d32"),
		Border:    lipgloss.Color("#1b5e20"),
		Error:     lipgloss.Color("#ff5252"),
		BgAlt:     lipgloss.Color("#051e08"),
	},
	"amber": {
		Name:      "Amber CRT",
		Primary:   lipgloss.Color("#ffab00"),
		Secondary: lipgloss.Color("#ff6d00"),
		Highlight: lipgloss.Color("#ffd740"),
		Text:      lipgloss.Color("#fff8e1"),
		Muted:     lipgloss.Color("#bf360c"),
		Border:    lipgloss.Color("#ff8f00"),
		Error:     lipgloss.Color("#ff3d00"),
		BgAlt:     lipgloss.Color("#261400"),
	},
	"ice": {
		Name:      "Ice Monolith",
		Primary:   lipgloss.Color("#e0f7fa"),
		Secondary: lipgloss.Color("#80deea"),
		Highlight: lipgloss.Color("#ffffff"),
		Text:      lipgloss.Color("#cfd8dc"),
		Muted:     lipgloss.Color("#455a64"),
		Border:    lipgloss.Color("#37474f"),
		Error:     lipgloss.Color("#ff8a80"),
		BgAlt:     lipgloss.Color("#1a2226"),
	},
}

// GetTheme returns the theme by key or default cyan theme.
func GetTheme(themeKey string) Theme {
	if t, ok := Themes[strings.ToLower(themeKey)]; ok {
		return t
	}
	return Themes["cyan"]
}

// Styles container computed for a given theme.
type Styles struct {
	Theme Theme

	Title       lipgloss.Style
	Subtitle    lipgloss.Style
	HeaderBox   lipgloss.Style
	StatusBox   lipgloss.Style
	ContentBox  lipgloss.Style
	MenuItem    lipgloss.Style
	MenuKey     lipgloss.Style
	SelectedKey lipgloss.Style
	Highlight   lipgloss.Style
	Muted       lipgloss.Style
	Error       lipgloss.Style
	BorderBox   lipgloss.Style
	Prompt      lipgloss.Style
	Tag         lipgloss.Style
}

// NewStyles builds Lip Gloss styles for the given theme.
func NewStyles(theme Theme) Styles {
	return Styles{
		Theme: theme,

		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Primary),

		Subtitle: lipgloss.NewStyle().
			Foreground(theme.Muted).
			Italic(true),

		HeaderBox: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(theme.Border).
			Padding(0, 1),

		StatusBox: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(theme.Border).
			Foreground(theme.Muted).
			Padding(0, 1),

		ContentBox: lipgloss.NewStyle().
			Padding(1, 2),

		MenuItem: lipgloss.NewStyle().
			Foreground(theme.Text),

		MenuKey: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Primary),

		SelectedKey: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Highlight).
			Background(theme.BgAlt),

		Highlight: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Highlight),

		Muted: lipgloss.NewStyle().
			Foreground(theme.Muted),

		Error: lipgloss.NewStyle().
			Foreground(theme.Error).
			Bold(true),

		BorderBox: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Border).
			Padding(0, 1),

		Prompt: lipgloss.NewStyle().
			Foreground(theme.Secondary).
			Bold(true),

		Tag: lipgloss.NewStyle().
			Foreground(theme.Highlight).
			Background(theme.BgAlt).
			Padding(0, 1),
	}
}

const ASCIILogoBig = `
  █████╗ ███████╗████████╗███████╗██████╗ ██████╗  █████╗ ██████╗ ██╗  ██╗
 ██╔══██╗██╔════╝╚══██╔══╝██╔════╝██╔══██╗██╔══██╗██╔══██╗██╔══██╗██║ ██╔╝
 ███████║█████╗     ██║   █████╗  ██████╔╝██║  ██║███████║██████╔╝█████╔╝ 
 ██╔══██║██╔══╝     ██║   ██╔══╝  ██╔══██╗██║  ██║██╔══██║██╔══██╗██╔═██╗ 
 ██║  ██║██║        ██║   ███████╗██║  ██║██████╔╝██║  ██║██║  ██║██║  ██╗
 ╚═╝  ╚═╝╚═╝        ╚═╝   ╚══════╝╚═╝  ╚═╝╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝`

const ASCIILogoCompact = `
  █▀█ █▀▀ ▀█▀ █▀▀ █▀▄ █▀▄ █▀█ █▀▄ █▄▀
  █▀█ █▀▀  █  █▀▀ █▀▄ █ █ █▀█ █▀▄ █ █
  ▀ ▀ ▀    ▀  ▀▀▀ ▀ ▀ ▀▀  ▀ ▀ ▀ ▀ ▀ ▀`
