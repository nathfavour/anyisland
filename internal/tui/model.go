package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nathfavour/anyisland/internal/pal"
	"github.com/nathfavour/anyisland/internal/registry"
)

type sessionState int

const (
	toolsView sessionState = iota
	visualsView
	pulseView
	helpView
)

type toolItem struct {
	name, desc, version string
}

func (i toolItem) Title() string       { return i.name }
func (i toolItem) Description() string { return i.desc + " (" + i.version + ")" }
func (i toolItem) FilterValue() string { return i.name }

type MainModel struct {
	state    sessionState
	sys      pal.System
	reg      *registry.Registry
	list     list.Model
	spinner  spinner.Model
	width    int
	height   int
	loading  bool
	err      error
	tabs     []string
}

func NewMainModel(sys pal.System, reg *registry.Registry) MainModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(highlight)

	tabs := []string{"Tools", "Visuals", "Pulse", "Help"}

	items := []list.Item{} // Will be populated by Init
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Installed Tools"
	l.Styles.Title = titleStyle

	return MainModel{
		state:   toolsView,
		sys:     sys,
		reg:     reg,
		spinner: s,
		list:    l,
		tabs:    tabs,
		loading: true,
	}
}

func (m MainModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.loadTools,
	)
}

type toolsLoadedMsg []list.Item
type errMsg error

func (m MainModel) loadTools() tea.Msg {
	tools, err := m.reg.ListTools()
	if err != nil {
		return errMsg(err)
	}

	items := make([]list.Item, len(tools))
	for i, t := range tools {
		items[i] = toolItem{
			name:    t.Name,
			desc:    t.Source, // Use Source as placeholder for desc
			version: t.Version,
		}
	}
	return toolsLoadedMsg(items)
}
