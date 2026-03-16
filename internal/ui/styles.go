package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	purple    = lipgloss.Color("#7C3AED")
	violet    = lipgloss.Color("#8B5CF6")
	blue      = lipgloss.Color("#3B82F6")
	cyan      = lipgloss.Color("#06B6D4")
	green     = lipgloss.Color("#10B981")
	yellow    = lipgloss.Color("#F59E0B")
	red       = lipgloss.Color("#EF4444")
	white     = lipgloss.Color("#F8FAFC")
	gray      = lipgloss.Color("#64748B")
	darkGray  = lipgloss.Color("#334155")
	bgDark    = lipgloss.Color("#0F172A")
	bgCard    = lipgloss.Color("#1E293B")
	muted     = lipgloss.Color("#94A3B8")

	// Header
	HeaderStyle = lipgloss.NewStyle().
		Foreground(white).
		Background(purple).
		Bold(true).
		Padding(0, 2)

	HeaderTitleStyle = lipgloss.NewStyle().
		Foreground(white).
		Bold(true).
		Padding(0, 1)

	HeaderInfoStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#DDD6FE")).
		Padding(0, 1)

	// Logo
	LogoStyle = lipgloss.NewStyle().
		Foreground(purple).
		Bold(true)

	LogoSubStyle = lipgloss.NewStyle().
		Foreground(violet)

	// Card
	CardStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(darkGray).
		Padding(1, 2).
		Margin(0, 0, 1, 0)

	CardHighlightStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(purple).
		Padding(1, 2).
		Margin(0, 0, 1, 0)

	// Text styles
	TitleStyle = lipgloss.NewStyle().
		Foreground(violet).
		Bold(true)

	SubtitleStyle = lipgloss.NewStyle().
		Foreground(cyan).
		Bold(true)

	MutedStyle = lipgloss.NewStyle().
		Foreground(muted)

	BoldStyle = lipgloss.NewStyle().
		Foreground(white).
		Bold(true)

	// Status
	SuccessStyle = lipgloss.NewStyle().
		Foreground(green).
		Bold(true)

	WarningStyle = lipgloss.NewStyle().
		Foreground(yellow).
		Bold(true)

	DangerStyle = lipgloss.NewStyle().
		Foreground(red).
		Bold(true)

	InfoStyle = lipgloss.NewStyle().
		Foreground(cyan).
		Bold(true)

	// Menu
	MenuItemStyle = lipgloss.NewStyle().
		Foreground(white).
		Padding(0, 2)

	MenuItemSelectedStyle = lipgloss.NewStyle().
		Foreground(white).
		Background(purple).
		Bold(true).
		Padding(0, 2)

	MenuItemIconStyle = lipgloss.NewStyle().
		Foreground(violet)

	// Input
	InputLabelStyle = lipgloss.NewStyle().
		Foreground(cyan).
		Bold(true)

	InputFocusedStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(purple).
		Padding(0, 1)

	InputBlurredStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(darkGray).
		Padding(0, 1)

	// Help
	HelpStyle = lipgloss.NewStyle().
		Foreground(gray).
		Padding(1, 2)

	// Separator
	SepStyle = lipgloss.NewStyle().
		Foreground(darkGray)

	// Status bar
	StatusBarStyle = lipgloss.NewStyle().
		Foreground(muted).
		Padding(0, 2)

	// Page title
	PageTitleStyle = lipgloss.NewStyle().
		Foreground(white).
		Background(violet).
		Bold(true).
		Padding(0, 2).
		Margin(0, 0, 1, 0)

	// Badge
	BadgeGreenStyle = lipgloss.NewStyle().
		Foreground(bgDark).
		Background(green).
		Bold(true).
		Padding(0, 1)

	BadgeRedStyle = lipgloss.NewStyle().
		Foreground(bgDark).
		Background(red).
		Bold(true).
		Padding(0, 1)

	BadgeYellowStyle = lipgloss.NewStyle().
		Foreground(bgDark).
		Background(yellow).
		Bold(true).
		Padding(0, 1)

	BadgeBlueStyle = lipgloss.NewStyle().
		Foreground(bgDark).
		Background(blue).
		Bold(true).
		Padding(0, 1)
)

// ProgressBar renders a simple colored progress bar
func ProgressBar(percent float64, width int) string {
	if width <= 0 {
		width = 20
	}
	filled := int(float64(width) * percent / 100.0)
	if filled > width {
		filled = width
	}
	empty := width - filled

	var color lipgloss.Color
	switch {
	case percent >= 75:
		color = green
	case percent >= 50:
		color = yellow
	default:
		color = red
	}

	bar := lipgloss.NewStyle().Foreground(color).Render(repeat("█", filled)) +
		lipgloss.NewStyle().Foreground(darkGray).Render(repeat("░", empty))
	return bar
}

func repeat(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
