package tui

import (
	"fmt"
	"time"

	"github-dashboard/pkg/github"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	table           table.Model
	viewport        viewport.Model
	repositories    []github.Repository
	contributions   string
	ready           bool
	viewportFocused bool
}

const tableHeight = 15
const contributionsWidth = 110

func NewModel(repositories []github.Repository, contributions string) Model {
	columns := []table.Column{
		{Title: "Name", Width: 20},
		{Title: "Description", Width: 30},
		{Title: "Language", Width: 12},
		{Title: "Updated", Width: 7},
		{Title: "Stars", Width: 5},
	}

	rows := []table.Row{}
	for _, repo := range repositories {
		rows = append(rows, table.Row{
			repo.Name,
			repo.Description,
			repo.Language,
			formatTimeAgo(repo.UpdatedAt),
			fmt.Sprintf("%d", repo.Stars),
		})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(tableHeight),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(true)
	t.SetStyles(s)

	vp := viewport.New(80, tableHeight+1)

	return Model{
		table:         t,
		viewport:      vp,
		repositories:  repositories,
		contributions: contributions,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport.Width = msg.Width/2 - 4
			m.viewport.Height = tableHeight + 1
			m.ready = true
			// Load initial README
			if len(m.repositories) > 0 {
				m.updateReadme()
			}
		}
	case tea.KeyMsg:
		if m.viewportFocused {
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "esc", "left", "h":
				// Return focus to table
				m.viewportFocused = false
				return m, nil
			default:
				// Let the viewport handle its own key events
				var vpCmd tea.Cmd
				m.viewport, vpCmd = m.viewport.Update(msg)
				return m, vpCmd
			}
		} else {
			// Existing table key handling
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "right", "l":
				m.viewportFocused = true
				return m, nil
			}
		}
	}

	// Handle table updates
	var tableCmd tea.Cmd
	m.table, tableCmd = m.table.Update(msg)
	if tableCmd != nil {
		cmds = append(cmds, tableCmd)
	}

	// Update viewport if focused
	if m.viewportFocused {
		var vpCmd tea.Cmd
		m.viewport, vpCmd = m.viewport.Update(msg)
		if vpCmd != nil {
			cmds = append(cmds, vpCmd)
		}
	} else {
		// Check if selection changed
		if len(m.repositories) > 0 && m.table.Cursor() < len(m.repositories) {
			m.updateReadme()
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) updateReadme() {
	if len(m.repositories) == 0 {
		return
	}

	selectedIdx := m.table.Cursor()
	if selectedIdx >= len(m.repositories) {
		return
	}

	// Get the README content
	readme := m.repositories[selectedIdx].Readme
	if readme == "" {
		readme = "# No README available\n\nThis repository doesn't have a README file."
	}

	// Render markdown
	width := m.viewport.Width - 4 // Account for padding
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(width),
	)

	content, _ := renderer.Render(readme)
	m.viewport.SetContent(content)
	m.viewport.GotoTop()
}

func (m Model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	// Create table view
	tableView := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		Render(m.table.View())

	view := ""
	if m.viewportFocused {
		view = lipgloss.NewStyle().
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2).
			Render(m.viewport.View())

	} else {
		view = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2).
			Render(m.viewport.View())
	}

	container := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.PlaceHorizontal(contributionsWidth, lipgloss.Center, m.contributions),
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			tableView,
			view,
		),
	)
	return container
}

func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)

	years := int(duration.Hours() / 24 / 365)
	if years > 0 {
		return fmt.Sprintf("%dy ago", years)
	}

	months := int(duration.Hours() / 24 / 30)
	if months > 0 {
		return fmt.Sprintf("%dmo ago", months)
	}

	days := int(duration.Hours() / 24)
	if days > 0 {
		return fmt.Sprintf("%dd ago", days)
	}

	hours := int(duration.Hours())
	if hours > 0 {
		return fmt.Sprintf("%dh ago", hours)
	}

	minutes := int(duration.Minutes())
	if minutes > 0 {
		return fmt.Sprintf("%dm ago", minutes)
	}

	return "now"
}
