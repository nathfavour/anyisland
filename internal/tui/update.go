package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nathfavour/anyisland/internal/visual"
)

type statusMsg string

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	if m.recording != nil {
		m.recording.AddFrame(m.View())
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.state == commandMode {
			switch msg.String() {
			case "enter":
				var cmdInput string
				if len(m.suggestions) > 0 && m.selectedSuggest >= 0 {
					cmdInput = m.suggestions[m.selectedSuggest]
				} else {
					cmdInput = m.textInput.Value()
				}
				m.textInput.SetValue("")
				m.suggestions = nil
				m.state = m.prevState
				return m.handleCommand(cmdInput)
			case "up":
				if len(m.suggestions) > 0 {
					m.selectedSuggest--
					if m.selectedSuggest < 0 {
						m.selectedSuggest = len(m.suggestions) - 1
					}
				}
				return m, nil
			case "down", "tab":
				if len(m.suggestions) > 0 {
					m.selectedSuggest++
					if m.selectedSuggest >= len(m.suggestions) {
						m.selectedSuggest = 0
					}
				}
				return m, nil
			case "esc":
				m.textInput.SetValue("")
				m.suggestions = nil
				m.state = m.prevState
				return m, nil
			}
			m.textInput, cmd = m.textInput.Update(msg)
			m.updateSuggestions()
			return m, cmd
		}

		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c"))):
			return m, tea.Quit
		case key.Matches(msg, key.NewBinding(key.WithKeys("/"))):
			m.prevState = m.state
			m.state = commandMode
			m.textInput.Focus()
			m.textInput.SetValue("")
			m.updateSuggestions()
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
			if m.state != commandMode {
				m.state = (m.state + 1) % 4
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("shift+tab"))):
			if m.state != commandMode {
				m.state = (m.state - 1 + 4) % 4
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width-4, msg.Height-8)

	case toolsLoadedMsg:
		m.loading = false
		m.list.SetItems(msg)

	case errMsg:
		m.err = msg
		m.loading = false

	case statusMsg:
		m.lastStatus = string(msg)
		cmds = append(cmds, tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
			return statusMsg("")
		}))
	}

	if m.loading {
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	if m.state == toolsView && !m.loading {
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *MainModel) handleCommand(input string) (tea.Model, tea.Cmd) {
	input = strings.TrimSpace(input)
	if input == "" {
		return m, nil
	}

	parts := strings.Fields(input)
	cmdName := parts[0]

	switch cmdName {
	case "q", "quit":
		return m, tea.Quit
	case "shot":
		return m, m.takeScreenshot()
	case "record":
		if m.recording == nil {
			m.recording = visual.NewRecordingSession("tui-session")
			return m, tea.Batch(
				func() tea.Msg { return statusMsg("Recording started") },
				m.recordTick(),
			)
		} else {
			rec := m.recording
			m.recording = nil
			return m, m.processRecording(rec)
		}
	case "help":
		m.state = helpView
		return m, nil
	default:
		return m, func() tea.Msg {
			return statusMsg(fmt.Sprintf("Unknown command: %s", cmdName))
		}
	}
}

func (m MainModel) takeScreenshot() tea.Cmd {
	return func() tea.Msg {
		// Render the view once more to get a clean shot (without the command input if possible)
		// but since we already exited commandMode in Update, it should be clean.
		view := m.View()
		path := filepath.Join(m.sys.GetVisualDir(), fmt.Sprintf("shot_%d.png", time.Now().Unix()))
		err := visual.RenderAnsiToPNG(view, path)
		if err != nil {
			return errMsg(err)
		}
		return statusMsg(fmt.Sprintf("Screenshot saved to %s", path))
	}
}

func (m MainModel) recordTick() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		if m.recording != nil {
			return tea.Batch(func() tea.Msg { return nil }, m.recordTick())
		}
		return nil
	})
}

func (m MainModel) processRecording(rec *visual.RecordingSession) tea.Cmd {
	return func() tea.Msg {
		path, err := visual.ProcessRecording(rec.Frames, m.sys.GetVisualDir())
		if err != nil {
			return errMsg(err)
		}
		return statusMsg(fmt.Sprintf("Recording saved to %s", path))
	}
}

func (m *MainModel) updateSuggestions() {
	input := strings.ToLower(m.textInput.Value())
	if input == "" {
		m.suggestions = allCommands
	} else {
		var filtered []string
		for _, c := range allCommands {
			if strings.HasPrefix(c, input) {
				filtered = append(filtered, c)
			}
		}
		m.suggestions = filtered
	}

	if m.selectedSuggest >= len(m.suggestions) {
		m.selectedSuggest = 0
	}
	if len(m.suggestions) > 0 && m.selectedSuggest < 0 {
		m.selectedSuggest = 0
	}
}

