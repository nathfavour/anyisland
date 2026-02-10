package tui

import "github.com/charmbracelet/lipgloss"

var (
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F591"}

	divider = lipgloss.NewStyle().
		Foreground(subtle).
		SetString(" • ").
		String()

	url = lipgloss.NewStyle().Foreground(special).Render

	docStyle = lipgloss.NewStyle().Padding(1, 2, 1, 2)

	tabStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, true, false, true).
			BorderForeground(subtle).
			Padding(0, 1)

	activeTabStyle = tabStyle.
			BorderForeground(highlight).
			Foreground(highlight)

	windowStyle = lipgloss.NewStyle().
			BorderForeground(highlight).
			Padding(1, 0).
			Align(lipgloss.Center).
			Border(lipgloss.NormalBorder(), false, true, true, true)

	titleStyle = lipgloss.NewStyle().
			MarginLeft(1).
			MarginRight(5).
			Padding(0, 1).
			Italic(true).
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#7D56F4")).
			Bold(true)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
			Background(lipgloss.AdaptiveColor{Light: "#D1D1D1", Dark: "#353533"})

	statusText = lipgloss.NewStyle().Inherit(statusStyle)

	statusNFO = statusText.
			Background(lipgloss.Color("#7D56F4")).
			Foreground(lipgloss.Color("#FFFDF5")).
			Padding(0, 1).
			MarginRight(1)

	statusKey = statusText.
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#FF5F56")).
			Padding(0, 1).
			MarginLeft(1)

	suggestStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#353533")).
			Padding(0, 1)

	suggestActiveStyle = suggestStyle.
				Background(lipgloss.Color("#7D56F4")).
				Bold(true)
)
