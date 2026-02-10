package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m MainModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("\n  Error: %v\n\n  Press q to quit.", m.err)
	}

	if m.loading {
		return fmt.Sprintf("\n  %s Loading tools...\n\n  Press q to quit.", m.spinner.View())
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
	doc.WriteString("\n")

	// Content
	var content string
	currentState := m.state
	if currentState == commandMode {
		currentState = m.prevState
	}

	switch currentState {
	case toolsView:
		content = m.list.View()
	case visualsView:
		content = "\n  Visual Gallery coming soon...\n  (Preview shots and recordings here)"
	case pulseView:
		content = "\n  Pulse Dashboard coming soon...\n  (Monitor connected tools and background tasks)"
	case helpView:
		content = "\n  Help & Instructions\n\n" +
			"  tab / shift+tab: Cycle tabs\n" +
			"  up / down: Navigate list\n" +
			"  / (in list): Filter tools\n" +
			"  / (anywhere): Enter slash command mode\n\n" +
			"  Commands:\n" +
			"  /quit / /q - Exit Anyisland\n" +
			"  /shot      - Take PNG screenshot\n" +
			"  /record    - Start/Stop recording\n" +
			"  /help      - Show this view\n" +
			"  q: Quit"
	}

	bodyHeight := m.height - 7
	if m.state == commandMode {
		bodyHeight--
	}

	body := windowStyle.Width(m.width - 2).Height(bodyHeight).Render(content)
	doc.WriteString(body)

	// Command Line / Status Bar
	if m.state == commandMode {
		doc.WriteString("\n")
		doc.WriteString(m.textInput.View())
	} else {
		status := m.renderStatusBar()
		doc.WriteString("\n")
		doc.WriteString(status)
	}

	return docStyle.Render(doc.String())
}

func (m MainModel) renderStatusBar() string {
	w := lipgloss.Width

	nfo := statusNFO.Render("ANYISLAND")
	key := statusKey.Render("TAB to switch")

	msg := " Managing your local tool ecosystem"
	if m.lastStatus != "" {
		msg = m.lastStatus
	}
	if m.recording != nil {
		msg = fmt.Sprintf("🔴 RECORDING (%d frames)", len(m.recording.Frames))
	}
	
	status := statusText.
		Width(m.width - w(nfo) - w(key) - 4).
		Render(msg)

	return lipgloss.JoinHorizontal(lipgloss.Top, nfo, status, key)
}