package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Repository struct {
	FullName        string `json:"full_name"`
	RepoDescription string `json:"description"`
	Stars           int    `json:"stargazers_count"`
	URL             string `json:"html_url"`
	HasManifest     bool   `json:"-"`
}

func (r Repository) Title() string {
	if r.HasManifest {
		return "🏝️ " + r.FullName
	}
	return r.FullName
}
func (r Repository) Description() string {
	return fmt.Sprintf("⭐ %d | %s", r.Stars, r.RepoDescription)
}
func (r Repository) FilterValue() string { return r.FullName }

type Resolver struct {
	client *http.Client
}

func NewResolver() *Resolver {
	return &Resolver{
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (r *Resolver) Resolve(ctx context.Context, query string) (string, error) {
	fmt.Printf("🔍 Searching for '%s'...\n", query)

	// Tier 1: Beacon Search (Code Search for anyisland.json)
	beaconRepos, _ := r.searchBeacon(ctx, query)

	// Tier 2: Broad Search (Repo Search by popularity)
	broadRepos, _ := r.searchBroad(ctx, query)

	// Merge and deduplicate
	repoMap := make(map[string]Repository)
	for _, repo := range beaconRepos {
		repo.HasManifest = true
		repoMap[repo.FullName] = repo
	}
	for _, repo := range broadRepos {
		if _, exists := repoMap[repo.FullName]; !exists {
			repoMap[repo.FullName] = repo
		}
	}

	if len(repoMap) == 0 {
		return "", fmt.Errorf("no suitable repositories found for '%s'", query)
	}

	var candidates []Repository
	for _, repo := range repoMap {
		candidates = append(candidates, repo)
	}

	// Sort: Native (Manifest) first, then stars
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].HasManifest != candidates[j].HasManifest {
			return candidates[i].HasManifest
		}
		return candidates[i].Stars > candidates[j].Stars
	})

	if len(candidates) == 1 {
		fmt.Printf("Found exact match: %s\n", candidates[0].URL)
		return candidates[0].URL, nil
	}

	// Interactive Picker
	selected, err := r.pickRepository(candidates)
	if err != nil {
		return "", err
	}
	return selected.URL, nil
}

func (r *Resolver) searchBeacon(ctx context.Context, query string) ([]Repository, error) {
	q := fmt.Sprintf("%s filename:anyisland.json", query)
	u := fmt.Sprintf("https://api.github.com/search/code?q=%s", url.QueryEscape(q))

	req, _ := http.NewRequestWithContext(ctx, "GET", u, nil)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "anyisland-cli")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api error: %d", resp.StatusCode)
	}

	var result struct {
		Items []struct {
			Repository Repository `json:"repository"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var repos []Repository
	for _, item := range result.Items {
		repos = append(repos, item.Repository)
	}
	return repos, nil
}

func (r *Resolver) searchBroad(ctx context.Context, query string) ([]Repository, error) {
	u := fmt.Sprintf("https://api.github.com/search/repositories?q=%s&sort=stars&order=desc", url.QueryEscape(query))

	req, _ := http.NewRequestWithContext(ctx, "GET", u, nil)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "anyisland-cli")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api error: %d", resp.StatusCode)
	}

	var result struct {
		Items []Repository `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Items, nil
}

// TUI Picker Model
type pickerModel struct {
	list     list.Model
	choice   *Repository
	quitting bool
}

func (m pickerModel) Init() tea.Cmd { return nil }
func (m pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			i, ok := m.list.SelectedItem().(Repository)
			if ok {
				m.choice = &i
			}
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m pickerModel) View() string {
	if m.quitting {
		return ""
	}
	return docStyle.Render(m.list.View())
}

var docStyle = lipgloss.NewStyle().Margin(1, 2)

func (r *Resolver) pickRepository(candidates []Repository) (*Repository, error) {
	items := make([]list.Item, len(candidates))
	for i, c := range candidates {
		items[i] = c
	}

	l := list.New(items, list.NewDefaultDelegate(), 40, 20)
	l.Title = "Select a repository to install"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ADD8")).Bold(true)

	m := pickerModel{list: l}
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	res := finalModel.(pickerModel)
	if res.choice == nil {
		return nil, fmt.Errorf("no repository selected")
	}

	return res.choice, nil
}
