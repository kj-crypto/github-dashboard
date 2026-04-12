package tui

import (
	"fmt"
	"log"
	"time"

	contribution "github-dashboard/pkg"
	"github-dashboard/pkg/github"

	display "github-dashboard/pkg"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
)

type reposDataMsg struct {
	repositories  []github.Repository
	contributions string
}

func (d reposDataMsg) isEmpty() bool {
	return len(d.repositories) == 0 && d.contributions == ""
}

type terminalSize struct {
	width  int
	height int
}

type Alignment int

const (
	AlignmentHorizontal Alignment = iota
	AlignmentVertical
)

type BrowserModel struct {
	reposTable      table.Model
	readmeViewport  viewport.Model
	viewportFocused bool
	alignment       Alignment
	tableWidth      int
	minTableHeight  int
}

func (m BrowserModel) Init() tea.Cmd {
	return nil
}

type errorMsg struct {
	message string
}

type Model struct {
	browserModel *BrowserModel
	spinner      spinner.Model
	isLoading    bool
	username     string
	error        string
	data         reposDataMsg
	terminalSize terminalSize
}

const (
	MinWidth         = display.Width
	MinHeight        = display.Height + TopBottomPadding*2 + 3
	TopBottomPadding = 1
	LeftRightPadding = 2
)

var tableStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("63")).
	Padding(TopBottomPadding, LeftRightPadding)

func InitModel(username string) tea.Model {
	sp := spinner.New()
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	sp.Spinner = spinner.Points

	return Model{
		isLoading:    true,
		username:     username,
		spinner:      sp,
		browserModel: nil,
		error:        "",
		data:         reposDataMsg{},
		terminalSize: terminalSize{},
	}
}

func initBrowserModel(data reposDataMsg, size terminalSize) *BrowserModel {
	columns := []table.Column{
		{Title: "Name", Width: 20},
		{Title: "Description", Width: 30},
		{Title: "Language", Width: 12},
		{Title: "Updated", Width: 7},
		{Title: "Stars", Width: 5},
	}

	rows := []table.Row{}
	for _, repo := range data.repositories {
		rows = append(rows, table.Row{
			repo.Name,
			repo.Description,
			repo.Language,
			formatTimeAgo(repo.UpdatedAt),
			fmt.Sprintf("%d", repo.Stars),
		})
	}

	tableWidth := 0
	for _, col := range columns {
		tableWidth += col.Width
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithWidth(tableWidth),
		// TODO: remove it here
		table.WithHeight(20),
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

	cellXoffset := s.Cell.GetHorizontalFrameSize()
	log.Printf("Table width: %d", tableWidth)
	log.Printf("Header horizontal frame size: %d", s.Header.GetHorizontalFrameSize())
	log.Printf("Style frame size: %d", tableStyle.GetHorizontalFrameSize())
	log.Printf("Cell X offset: %d, with all columns: %d", cellXoffset, cellXoffset*len(t.Columns()))
	tableWidth += s.Header.GetHorizontalFrameSize() + tableStyle.GetHorizontalFrameSize() + 2*LeftRightPadding
	log.Printf("Total table width: %d", tableWidth)

	vp := viewport.New(viewport.WithWidth(80), viewport.WithHeight(21))
	// vp := viewport.New(viewport.WithWidth(size.width), viewport.WithHeight(size.height))

	m := &BrowserModel{
		reposTable:      t,
		readmeViewport:  vp,
		viewportFocused: false,
		alignment:       AlignmentHorizontal,
		tableWidth:      tableWidth,
	}
	m.updateReadme(data.repositories)
	m.resize(size)
	return m
}

func (m *BrowserModel) resize(term terminalSize) *BrowserModel {
	log.Printf("Actual table width: %d", m.reposTable.Width())
	// maxHorizontalSpace := 2 * (m.reposTable.Width() + 2*LeftRightPadding)
	// minHorizontalSpace := int(0.75 * float64(maxHorizontalSpace))

	// if term.width >= maxHorizontalSpace {
	// 	m.readmeViewport.SetWidth(m.reposTable.Width())
	// } else if term.width < minHorizontalSpace {
	// 	m.alignment = AlignmentVertical
	// 	m.readmeViewport.SetWidth(m.reposTable.Width())
	// } else {
	// 	width := term.width - m.reposTable.Width() - 4*LeftRightPadding
	// 	m.readmeViewport.SetWidth(width)
	// }

	// if m.alignment == AlignmentVertical {
	// 	m.readmeViewport.SetHeight(term.height - 3)
	// }

	return m
}

func fetchData(username string, token string) tea.Cmd {
	return func() tea.Msg {
		results := make(chan struct {
			contributions string
			repos         []github.Repository
			err           error
		}, 2)

		go func() {
			contributions, err := contribution.GetContributionsFromApi(token, username)
			if err != nil {
				results <- struct {
					contributions string
					repos         []github.Repository
					err           error
				}{"", nil, err}
				return
			}
			calendar := contribution.FormatCalendar(contributions, 0, true)
			results <- struct {
				contributions string
				repos         []github.Repository
				err           error
			}{calendar, nil, nil}
		}()

		go func() {
			repos, err := github.GetRepositories(token, username)
			if err != nil {
				results <- struct {
					contributions string
					repos         []github.Repository
					err           error
				}{"", nil, err}
				return
			}
			results <- struct {
				contributions string
				repos         []github.Repository
				err           error
			}{"", repos, nil}
		}()
		var contributions string
		var repos []github.Repository

		for i := 0; i < 2; i++ {
			result := <-results
			if result.err != nil {
				return errorMsg{message: result.err.Error()}
			}
			if result.contributions != "" {
				contributions = result.contributions
			}
			if result.repos != nil {
				repos = result.repos
			}
		}

		return reposDataMsg{
			repositories:  repos,
			contributions: contributions,
		}
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		fetchData(m.username, contribution.GetToken()),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		log.Printf("[UI] Window resize: %dx%d (min: %dx%d)", msg.Width, msg.Height, MinWidth, MinHeight)
		m.terminalSize.height = msg.Height
		m.terminalSize.width = msg.Width
		if msg.Width < MinWidth || msg.Height < MinHeight {
			log.Printf("[UI] Window too small - setting error")
			m.error = "Terminal too small"
			m.isLoading = false
			return m, nil
		} else {
			if m.browserModel == nil && !m.data.isEmpty() {
				m.browserModel = initBrowserModel(m.data, m.terminalSize)
			}
			m.error = ""
		}
		if !m.isLoading && m.error == "" {
			m.browserModel = m.browserModel.resize(m.terminalSize)
		}
		return m, nil
	case tea.KeyMsg:
		log.Printf("[UI] Key pressed: %s", msg.String())
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		default:
			if !m.isLoading && m.error == "" {
				log.Printf("[UI] Forwarding key to browser model")
				cmd := m.browserModel.update(msg, m.data.repositories)
				return m, cmd
			}
			return m, nil
		}
	case reposDataMsg:
		log.Printf("[UI] Received repos data message")
		m.data = msg
		if m.error == "" {
			m.browserModel = initBrowserModel(msg, m.terminalSize)
			m.isLoading = false
		}
		return m, nil
	case spinner.TickMsg:
		if m.isLoading {
			log.Printf("[UI] Spinner tick")
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil
	case errorMsg:
		log.Printf("[UI] Error message: %s", msg.message)
		m.error = msg.message
		m.isLoading = false
		return m, nil
	}
	return m, nil
}

func (m *BrowserModel) update(msg tea.KeyMsg, repos []github.Repository) tea.Cmd {
	switch msg.String() {
	case "esc", "left", "h", "right", "l":
		m.viewportFocused = !m.viewportFocused
		return nil
	default:
		if m.viewportFocused {
			var cmd tea.Cmd
			m.readmeViewport, cmd = m.readmeViewport.Update(msg)
			return cmd

		} else {
			var cmd tea.Cmd
			m.reposTable, cmd = m.reposTable.Update(msg)
			m.updateReadme(repos)
			return cmd
		}
	}
}

func (m *BrowserModel) updateReadme(repos []github.Repository) {
	if len(repos) == 0 {
		return
	}

	selectedIdx := m.reposTable.Cursor()
	if selectedIdx >= len(repos) {
		return
	}

	// Get the README content
	readme := repos[selectedIdx].Readme
	if readme == "" {
		readme = "# No README available\n\nThis repository doesn't have a README file."
	}

	// Render markdown
	width := m.readmeViewport.Width() - 4 // TODO: Account for padding
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(width),
	)

	content, _ := renderer.Render(readme)
	m.readmeViewport.SetContent(content)
	m.readmeViewport.GotoTop()
}

func (m Model) View() tea.View {
	if m.error != "" {
		textStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render
		v := tea.NewView(fmt.Sprintf("\n  Error: %s\n\n  Press 'q' to quit\n", textStyle(m.error)))
		v.AltScreen = true
		return v
	}
	if m.isLoading {
		textStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render
		v := tea.NewView(fmt.Sprintf("\n %s  %s\n", m.spinner.View(), textStyle("Repositories loading ...")))
		v.AltScreen = true
		return v
	}
	v := m.browserModel.view(m.data.contributions)
	v.AltScreen = true
	return v
}

func (m BrowserModel) view(contributions string) tea.View {
	style := tableStyle
	if m.viewportFocused {
		style = style.BorderStyle(lipgloss.ThickBorder())
	}

	tableView := tableStyle.Render(m.reposTable.View())
	view := style.Render(m.readmeViewport.View())

	var details string
	if m.alignment == AlignmentHorizontal {
		details = lipgloss.JoinHorizontal(
			lipgloss.Top,
			tableView,
			view,
		)
	} else {
		details = lipgloss.JoinVertical(
			lipgloss.Left,
			tableView,
			view,
		)
	}
	style = lipgloss.NewStyle().
		BorderStyle(lipgloss.BlockBorder()).
		BorderForeground(lipgloss.Color("240"))
	container := lipgloss.JoinVertical(
		lipgloss.Left,
		// lipgloss.PlaceHorizontal(contribution.Width, lipgloss.Center, contributions),
		style.Render(contributions),
		details,
	)
	return tea.NewView(container)
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
