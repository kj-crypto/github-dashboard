package tui

import (
	"fmt"
	"log"
	"time"

	contribution "github-dashboard/pkg"
	"github-dashboard/pkg/github"
	"github-dashboard/pkg/utils"

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

type tabbleSizes struct {
	headerWidth  int
	headerHeight int
	cellHeight   int
}

type Readme struct {
	vp       viewport.Model
	markdown string
	renderer *glamour.TermRenderer
}

type BrowserModel struct {
	reposTable    table.Model
	readme        Readme
	readmeFocused bool
	showReadme    bool
	tableSizes    tabbleSizes
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

// Styes
var tableStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder())

var contributionsStyle = lipgloss.NewStyle()

var reposTableStyle = table.DefaultStyles()
var reposTableHeaderStyle = reposTableStyle.Header.
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240")).
	BorderBottom(true).
	Bold(true)

var headerHeight = lipgloss.Height(reposTableHeaderStyle.Render("Temp"))
var MinWidth = 2*53 + 5 - 1 + contributionsStyle.GetHorizontalFrameSize()
var MinHeight = 7 + 1 + tableStyle.GetVerticalFrameSize() + headerHeight + 1

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
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
	)
	reposTableStyle.Header = reposTableHeaderStyle
	reposTableStyle.Selected = reposTableStyle.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(true)
	t.SetStyles(reposTableStyle)

	headerWidth := utils.GetHeaderWidth(&t, &reposTableHeaderStyle)
	log.Printf("Header width x height: %d x %d", headerWidth, headerHeight)
	log.Printf("Total table width: %d", headerWidth+tableStyle.GetHorizontalFrameSize())

	vp := viewport.New()
	vp.Style = lipgloss.NewStyle().Padding(0)

	m := &BrowserModel{
		reposTable: t,
		readme: Readme{
			vp:       vp,
			markdown: "",
			renderer: nil,
		},
		readmeFocused: false,
		showReadme:    true,
		tableSizes: tabbleSizes{
			headerWidth:  headerWidth,
			headerHeight: headerHeight,
			cellHeight:   1 + reposTableStyle.Cell.GetVerticalFrameSize(),
		},
	}
	m.updateReadme(data.repositories)
	m.resize(size)
	return m
}

func (m *BrowserModel) resize(term terminalSize) {
	log.Printf("[RESIZE] Terminal size: %dx%d", term.width, term.height)
	ts := m.tableSizes
	reposTable := &m.reposTable
	readme := &m.readme
	readmeVp := &readme.vp

	freeHorizontalSpace := term.width - ts.headerWidth - tableStyle.GetHorizontalFrameSize()
	readmeWidth := freeHorizontalSpace - tableStyle.GetHorizontalFrameSize()
	log.Printf("[RESIZE] Free horizontal space: %d, readme width: %d", freeHorizontalSpace, readmeWidth)
	if readmeWidth < ts.headerWidth/2 {
		m.showReadme = false
	} else {
		w := min(readmeWidth, ts.headerWidth)
		log.Printf("[RESIZE] Setting width to: %d", w)
		readmeVp.SetWidth(w)
		renderer, _ := glamour.NewTermRenderer(
			glamour.WithStandardStyle("dark"),
			glamour.WithWordWrap(w),
		)
		readme.renderer = renderer

		content, _ := renderer.Render(readme.markdown)
		readmeVp.SetContent(content)
		readmeVp.GotoTop()
		m.showReadme = true

	}
	freeVerticalSpace := term.height - 8 - contributionsStyle.GetVerticalFrameSize() - tableStyle.GetVerticalFrameSize()
	maxHeight := len(reposTable.Rows())*ts.cellHeight + ts.headerHeight
	log.Printf("[RESIZE] Free vertical space: %d, max height: %d", freeVerticalSpace, maxHeight)
	h := min(freeVerticalSpace, maxHeight)
	log.Printf("[RESIZE] Setting height to: %d", h)
	reposTable.SetHeight(h)
	reposTable.SetWidth(ts.headerWidth)
	readmeVp.SetHeight(h)
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
		fetchData(m.username, utils.GetToken()),
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
			m.browserModel.resize(m.terminalSize)
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
		m.readmeFocused = !m.readmeFocused
		return nil
	default:
		if m.readmeFocused {
			var cmd tea.Cmd
			m.readme.vp, cmd = m.readme.vp.Update(msg)
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

	text := repos[selectedIdx].Readme
	if text == "" {
		text = "# No README available\n\nThis repository doesn't have a README file."
	}
	m.readme.markdown = text
	if m.readme.renderer != nil {
		content, _ := m.readme.renderer.Render(text)
		m.readme.vp.SetContent(content)
		m.readme.vp.GotoTop()
	}
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
	v := m.browserModel.view(m.data.contributions, m.terminalSize.width)
	v.AltScreen = true
	return v
}

func (m BrowserModel) view(contributions string, width int) tea.View {
	style := tableStyle
	if m.readmeFocused {
		style = style.BorderStyle(lipgloss.ThickBorder())
	}

	tableView := tableStyle.Render(m.reposTable.View())
	view := style.Render(m.readme.vp.View())

	var details string
	if m.showReadme {
		details = lipgloss.JoinHorizontal(
			lipgloss.Top,
			tableView,
			view,
		)
	} else {
		details = tableView
	}
	container := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.PlaceHorizontal(width, lipgloss.Center, contributionsStyle.Render(contributions)),
		lipgloss.PlaceHorizontal(width, lipgloss.Center, details),
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
