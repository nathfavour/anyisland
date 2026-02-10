package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m MainModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("
  Error: %v

  Press q to quit.", m.err)
	}

	if m.loading {
		return fmt.Sprintf("
  %s Loading tools...

  Press q to quit.", m.spinner.View())
	}

	doc := strings.Builder{}

	// Tabs
	var renderedTabs []string
	for i, t := range m.tabs {
		var style lipgloss.Style
		if int(m.state) == i {
			style = activeTabStyle
		} else {
			style = tabStyle
		}
		renderedTabs = append(renderedTabs, style.Render(t))
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	doc.WriteString("
")

	// Content
	var content string
	switch m.state {
	case toolsView:
		content = m.list.View()
	case visualsView:
		content = "
  Visual Gallery coming soon...
  (Preview shots and recordings here)"
	case pulseView:
		content = "
  Pulse Dashboard coming soon...
  (Monitor connected tools and background tasks)"
	case helpView:
		content = "
  Help & Instructions

" +
			"  tab / shift+tab: Cycle tabs
" +
			"  up / down: Navigate list
" +
			"  /: Filter tools
" +
			"  q: Quit"
	}

	body := windowStyle.Width(m.width - 2).Height(m.height - 6).Render(content)
	doc.WriteString(body)

	// Status bar
	status := m.renderStatusBar()
	doc.WriteString("
")
	doc.WriteString(status)

	return docStyle.Render(doc.String())
}

func (m MainModel) renderStatusBar() string {
	w := lipgloss.Width

	nfo := statusNFO.Render("ANYISLAND")
	key := statusKey.Render("TAB to switch")
	
	status := statusText.
		Width(m.width - w(nfo) - w(key) - 4).
		Render(" Managing your local tool ecosystem")

	return lipgloss.JoinHorizontal(lipgloss.Top, nfo, status, key)
}
