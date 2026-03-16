package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"supaco-cli/internal/api"
	"supaco-cli/internal/config"
)

// ── Screens ───────────────────────────────────────────────────────────────────

type screen int

const (
	screenLogin screen = iota
	screenLoading
	screenWelcome // shows user info briefly
	screenMenu
	screenToday
	screenWeek
	screenAbsences
	screenGrades
	screenProfile
	screenNotifications
)

// ── Messages ──────────────────────────────────────────────────────────────────

type loginSuccessMsg struct{ access, refresh string }
type loginErrMsg struct{ err string }

type dataLoadedMsg struct {
	academic   *api.AcademicData
	diaries    []api.Diary
	completion *api.CompletionReqs
	messages   *api.MessagesResponse
	semestre   string
}

type dataErrMsg struct {
	err            string
	sessionExpired bool
}

type advanceScreenMsg struct{}

// ── Menu Items ────────────────────────────────────────────────────────────────

type menuItem struct {
	icon   string
	label  string
	target screen
	badge  string
}

// ── App Model ─────────────────────────────────────────────────────────────────

type App struct {
	screen screen
	width  int
	height int

	// Login
	login loginModel

	// Loading
	spinner spinner.Model
	loadMsg string

	// Data
	client     *api.Client
	cfg        *config.Config
	academic   *api.AcademicData
	diaries    []api.Diary
	completion *api.CompletionReqs
	messages   *api.MessagesResponse
	semestre   string
	errMsg     string

	// Menu
	menuItems  []menuItem
	menuCursor int

	// Content viewport
	viewport  viewport.Model
	viewReady bool
}

func NewApp() *App {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(purple)

	app := &App{spinner: s}

	cfg, _ := config.Load()
	app.cfg = cfg
	app.login = newLoginModel()

	app.client = api.NewClient()
	app.client.OnRefresh = func(access, refresh string) {
		app.cfg.AccessToken = access
		app.cfg.RefreshToken = refresh
		app.cfg.Save()
	}

	if cfg.AccessToken != "" && cfg.RefreshToken != "" {
		app.client.AccessToken = cfg.AccessToken
		app.client.RefreshToken = cfg.RefreshToken
		app.screen = screenLoading
		app.loadMsg = "Carregando seus dados..."
	} else {
		app.screen = screenLogin
	}

	return app
}

func (a *App) Init() tea.Cmd {
	cmds := []tea.Cmd{textinput.Blink, a.spinner.Tick}
	if a.screen == screenLoading {
		cmds = append(cmds, a.loadData())
	}
	return tea.Batch(cmds...)
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		if a.viewReady {
			a.viewport.Width = a.width - 4
			a.viewport.Height = a.height - 5
		}
		a.refreshViewport()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit
		case "q":
			switch a.screen {
			case screenMenu, screenWelcome:
				return a, tea.Quit
			case screenLogin, screenLoading:
				// do nothing
			default:
				a.screen = screenMenu
				a.refreshViewport()
				return a, nil
			}
		case "esc":
			switch a.screen {
			case screenLogin, screenLoading:
				// do nothing
			case screenWelcome:
				a.screen = screenMenu
				a.refreshViewport()
				return a, nil
			default:
				a.screen = screenMenu
				a.refreshViewport()
				return a, nil
			}
		case "enter", " ":
			if a.screen == screenWelcome {
				a.screen = screenMenu
				a.refreshViewport()
				return a, nil
			}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(msg)
		return a, cmd

	case loginSuccessMsg:
		a.cfg.AccessToken = msg.access
		a.cfg.RefreshToken = msg.refresh
		a.cfg.Save()
		a.client.AccessToken = msg.access
		a.client.RefreshToken = msg.refresh
		a.screen = screenLoading
		a.loadMsg = "Buscando seus dados no SUAP..."
		return a, a.loadData()

	case loginErrMsg:
		a.login.loading = false
		a.login.err = msg.err
		a.screen = screenLogin
		return a, nil

	case dataLoadedMsg:
		a.academic = msg.academic
		a.diaries = msg.diaries
		a.completion = msg.completion
		a.messages = msg.messages
		a.semestre = msg.semestre
		a.buildMenu()
		a.screen = screenWelcome // show user info first
		return a, nil

	case dataErrMsg:
		if msg.sessionExpired {
			a.cfg.AccessToken = ""
			a.cfg.RefreshToken = ""
			a.cfg.Save()
			a.login = newLoginModel()
			a.login.err = "Sessao expirada. Faca login novamente."
			a.screen = screenLogin
		} else {
			// data load failed but maybe still authed — show menu empty
			a.buildMenu()
			a.errMsg = msg.err
			a.screen = screenMenu
		}
		return a, nil

	case advanceScreenMsg:
		if a.screen == screenWelcome {
			a.screen = screenMenu
		}
		return a, nil
	}

	switch a.screen {
	case screenLogin:
		return a.updateLogin(msg)
	case screenMenu:
		return a.updateMenu(msg)
	case screenToday, screenWeek, screenAbsences, screenGrades, screenProfile, screenNotifications:
		return a.updateContent(msg)
	}

	return a, nil
}

// ── Login Update ──────────────────────────────────────────────────────────────

func (a *App) updateLogin(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if a.login.focused == 1 {
				if !a.login.IsReady() {
					a.login.err = "Preencha matricula e senha"
					return a, nil
				}
				a.login.loading = true
				a.screen = screenLoading
				a.loadMsg = "Autenticando..."
				return a, a.doLogin(a.login.Matricula(), a.login.Password())
			}
			// advance to password
			a.login.focused = 1
			a.login.inputs[0].Blur()
			a.login.inputs[1].Focus()
			return a, textinput.Blink
		}
	}

	a.login, cmd = a.login.Update(msg)
	return a, cmd
}

// ── Menu Update ───────────────────────────────────────────────────────────────

func (a *App) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if a.menuCursor > 0 {
				a.menuCursor--
			}
		case "down", "j":
			if a.menuCursor < len(a.menuItems)-1 {
				a.menuCursor++
			}
		case "enter", " ":
			if len(a.menuItems) == 0 {
				return a, nil
			}
			item := a.menuItems[a.menuCursor]
			switch item.label {
			case "Sair":
				return a, tea.Quit
			case "Atualizar dados":
				a.screen = screenLoading
				a.loadMsg = "Atualizando dados..."
				return a, a.loadData()
			default:
				a.screen = item.target
				a.refreshViewport()
			}
		case "r":
			a.screen = screenLoading
			a.loadMsg = "Atualizando dados..."
			return a, a.loadData()
		}
	}
	return a, nil
}

// ── Content Update ────────────────────────────────────────────────────────────

func (a *App) updateContent(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	a.viewport, cmd = a.viewport.Update(msg)
	return a, cmd
}

// ── View ──────────────────────────────────────────────────────────────────────

func (a *App) View() string {
	if a.width == 0 {
		return ""
	}
	switch a.screen {
	case screenLogin:
		return a.login.View(a.width, a.height)
	case screenLoading:
		return a.viewLoading()
	case screenWelcome:
		return a.viewWelcome()
	case screenMenu:
		return a.viewMenu()
	default:
		return a.viewContent()
	}
}

// ── Loading ───────────────────────────────────────────────────────────────────

func (a *App) viewLoading() string {
	logo := lipgloss.NewStyle().Foreground(purple).Bold(true).Render(
		"  ____  _   _ ____   _    ____ ___\n" +
			" / ___|| | | |  _ \\ / \\  / ___/ _ \\\n" +
			" \\___ \\| | | | |_) / _ \\| |  | | | |\n" +
			"  ___) | |_| |  __/ ___ \\ |__| |_| |\n" +
			" |____/ \\___/|_| /_/   \\_\\____\\___/",
	)
	sub := lipgloss.NewStyle().Foreground(violet).Render("         CLI do SUAP · IFRN")
	loading := a.spinner.View() + "  " + lipgloss.NewStyle().Foreground(muted).Render(a.loadMsg)

	content := lipgloss.JoinVertical(lipgloss.Center, logo, sub, "", "", loading)
	return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, content)
}

// ── Welcome (user info confirmation) ─────────────────────────────────────────

func (a *App) viewWelcome() string {
	logoLine := lipgloss.NewStyle().Foreground(purple).Bold(true).Render(
		"  ____  _   _ ____   _    ____ ___\n" +
			" / ___|| | | |  _ \\ / \\  / ___/ _ \\\n" +
			" \\___ \\| | | | |_) / _ \\| |  | | | |\n" +
			"  ___) | |_| |  __/ ___ \\ |__| |_| |\n" +
			" |____/ \\___/|_| /_/   \\_\\____\\___/",
	)
	sub := lipgloss.NewStyle().Foreground(violet).Render("         CLI do SUAP · IFRN")

	sem := ""
	if a.semestre != "" {
		sem = "  ·  Semestre " + a.semestre
	}
	semLine := lipgloss.NewStyle().Foreground(muted).Render("         Dados carregados" + sem)

	userCard := RenderUserInfo(a.academic, a.cfg.Matricula, a.width-10)

	hint := MutedStyle.Render("  Pressione ENTER ou qualquer tecla para continuar...")

	content := lipgloss.JoinVertical(lipgloss.Left,
		logoLine,
		sub,
		semLine,
		"",
		userCard,
		"",
		hint,
	)

	return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, content)
}

// ── Menu ──────────────────────────────────────────────────────────────────────

func (a *App) viewMenu() string {
	// ── Header bar ──
	semStr := ""
	if a.semestre != "" {
		semStr = " · " + a.semestre
	}

	courseStr := ""
	if a.academic != nil {
		short := a.academic.CursoNome()
		if len(short) > 30 {
			short = short[:28] + "…"
		}
		courseStr = short
	}

	matStr := a.cfg.Matricula

	headerLeft := lipgloss.NewStyle().
		Foreground(white).Background(purple).Bold(true).
		Padding(0, 1).
		Render("SUPACO" + semStr)

	headerRight := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#DDD6FE")).Background(purple).
		Padding(0, 1).
		Render(matStr)

	headerMid := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#C4B5FD")).Background(purple).
		Render(courseStr)

	lw := lipgloss.Width(headerLeft)
	rw := lipgloss.Width(headerRight)
	mw := lipgloss.Width(headerMid)
	gap := a.width - lw - rw - mw - 2
	if gap < 1 {
		gap = 1
	}

	header := lipgloss.NewStyle().Background(purple).Width(a.width).Render(
		headerLeft + strings.Repeat(" ", gap/2) + headerMid + strings.Repeat(" ", gap-gap/2) + headerRight,
	)

	// ── Sidebar ──
	sideW := 32

	sideTitle := lipgloss.NewStyle().
		Foreground(white).Bold(true).
		Render(" SUPACO")
	sideSub := MutedStyle.Render(" CLI do SUAP · IFRN")
	sideSep := SepStyle.Render(" " + strings.Repeat("─", sideW-2))

	var menuLines []string
	for i, item := range a.menuItems {
		label := item.icon + " " + item.label
		badge := ""
		if item.badge != "" {
			badge = " " + BadgeYellowStyle.Render(" "+item.badge+" ")
		}
		if i == a.menuCursor {
			menuLines = append(menuLines,
				MenuItemSelectedStyle.Width(sideW).Render(label+badge),
			)
		} else {
			menuLines = append(menuLines,
				MenuItemStyle.Width(sideW).Render(label+badge),
			)
		}
	}

	sideHelp := "\n" + lipgloss.NewStyle().Foreground(gray).Padding(0, 1).Render(
		"↑↓ navegar  enter selecionar\n r atualizar    q sair",
	)

	sidebar := lipgloss.NewStyle().Width(sideW).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			"",
			sideTitle,
			sideSub,
			"",
			sideSep,
			"",
			lipgloss.JoinVertical(lipgloss.Left, menuLines...),
			sideHelp,
		),
	)

	// ── Status panel ──
	mainW := a.width - sideW - 3
	statusPanel := lipgloss.NewStyle().Width(mainW).Padding(0, 2).Render(
		lipgloss.JoinVertical(lipgloss.Left, a.buildStatusPanel(mainW)...),
	)

	divider := SepStyle.Render(
		strings.Repeat("│\n", max(0, a.height-2)),
	)
	_ = divider

	body := lipgloss.JoinHorizontal(lipgloss.Top,
		sidebar,
		SepStyle.Render(" │"),
		statusPanel,
	)

	return lipgloss.JoinVertical(lipgloss.Left, header, body)
}

func (a *App) buildStatusPanel(w int) []string {
	var lines []string
	lines = append(lines, "")
	lines = append(lines, SubtitleStyle.Render("Hoje"))
	lines = append(lines, MutedStyle.Render(todayStr()))
	lines = append(lines, "")

	// Today's classes count
	todayIdx := todayWeekdayIdx()
	var todayClasses []string
	for _, d := range a.diaries {
		for dia, b := range d.DayBlocks() {
			if ptWeekdays[dia] == todayIdx {
				line := lipgloss.NewStyle().Foreground(cyan).Render(b.Start+" – "+b.End) +
					"  " + BoldStyle.Render(d.Disciplina.Descricao)
				todayClasses = append(todayClasses, "  "+line)
			}
		}
	}

	if len(todayClasses) == 0 {
		lines = append(lines, "  "+SuccessStyle.Render("Sem aulas hoje!"))
	} else {
		lines = append(lines, MutedStyle.Render(fmt.Sprintf("  %d aula(s):", len(todayClasses))))
		lines = append(lines, todayClasses...)
	}

	// Attendance alerts
	lines = append(lines, "")
	lines = append(lines, SubtitleStyle.Render("Frequencia"))
	lines = append(lines, "")

	atRisk := 0
	critical := 0
	ok := 0
	for _, d := range a.diaries {
		f := d.Disciplina.Frequencia
		switch {
		case f < 75 && (f > 0 || d.Disciplina.QtdFaltas > 0):
			critical++
		case f < 85 && f > 0:
			atRisk++
		default:
			ok++
		}
	}

	if critical > 0 {
		lines = append(lines, "  "+BadgeRedStyle.Render(fmt.Sprintf(" %d reprovada(s) por falta ", critical)))
	}
	if atRisk > 0 {
		lines = append(lines, "  "+BadgeYellowStyle.Render(fmt.Sprintf(" %d disciplina(s) em risco ", atRisk)))
	}
	if critical == 0 && atRisk == 0 && len(a.diaries) > 0 {
		lines = append(lines, "  "+BadgeGreenStyle.Render(" Frequencia ok em tudo! "))
	}

	// Notifications
	if a.messages != nil && a.messages.Count > 0 {
		lines = append(lines, "")
		lines = append(lines, "  "+BadgeYellowStyle.Render(fmt.Sprintf(" %d notificacao(oes) ", a.messages.Count)))
	}

	// IRA
	if a.academic != nil {
		lines = append(lines, "")
		lines = append(lines, SubtitleStyle.Render("IRA"))
		ira := a.academic.IRAFloat()
		iraColor := green
		if ira < 6 {
			iraColor = red
		} else if ira < 8 {
			iraColor = yellow
		}
		lines = append(lines, "  "+lipgloss.NewStyle().Foreground(iraColor).Bold(true).
			Render(fmt.Sprintf("%.2f", ira)))
	}

	// Error message if any
	if a.errMsg != "" {
		lines = append(lines, "")
		lines = append(lines, DangerStyle.Render("  Aviso: "+a.errMsg))
	}

	return lines
}

// ── Content view ──────────────────────────────────────────────────────────────

func (a *App) viewContent() string {
	if !a.viewReady {
		a.viewport = viewport.New(a.width-4, a.height-5)
		a.viewport.Style = lipgloss.NewStyle().Padding(0, 1)
		a.viewReady = true
		a.refreshViewport()
	}

	header := a.buildContentHeader()
	help := HelpStyle.Render(" ↑↓ rolar  ·  esc/q voltar  ·  ctrl+c sair")

	return lipgloss.JoinVertical(lipgloss.Left, header, a.viewport.View(), help)
}

func (a *App) buildContentHeader() string {
	names := map[screen]string{
		screenToday:         " Aulas de Hoje",
		screenWeek:          " Grade da Semana",
		screenAbsences:      " Frequencia e Faltas",
		screenGrades:        " Notas",
		screenProfile:       " Meu Perfil",
		screenNotifications: " Notificacoes",
	}
	screenName := names[a.screen]
	title := lipgloss.NewStyle().Foreground(white).Bold(true).Background(purple).Padding(0, 2).Render("SUPACO" + screenName)
	back := lipgloss.NewStyle().Foreground(lipgloss.Color("#DDD6FE")).Background(purple).Render("esc voltar ")

	gap := a.width - lipgloss.Width(title) - lipgloss.Width(back)
	if gap < 0 {
		gap = 0
	}
	return lipgloss.NewStyle().Background(purple).Width(a.width).Render(
		title + strings.Repeat(" ", gap) + back,
	)
}

func (a *App) refreshViewport() {
	if !a.viewReady {
		return
	}
	a.viewport.Width = a.width - 4
	a.viewport.Height = a.height - 5

	var content string
	switch a.screen {
	case screenToday:
		content = RenderToday(a.diaries, a.width)
	case screenWeek:
		content = RenderWeek(a.diaries, a.width)
	case screenAbsences:
		content = RenderAbsences(a.diaries, a.width)
	case screenGrades:
		content = RenderGrades(a.diaries, a.academic, a.width)
	case screenProfile:
		content = RenderProfile(a.academic, a.completion, a.cfg.Matricula, a.width)
	case screenNotifications:
		content = RenderNotifications(a.messages, a.width)
	}

	a.viewport.SetContent(content)
	a.viewport.GotoTop()
}

func (a *App) buildMenu() {
	badge := ""
	if a.messages != nil && a.messages.Count > 0 {
		badge = fmt.Sprintf("%d", a.messages.Count)
	}
	a.menuItems = []menuItem{
		{"*", "Aulas de Hoje", screenToday, ""},
		{"~", "Grade da Semana", screenWeek, ""},
		{"%", "Frequencia e Faltas", screenAbsences, ""},
		{"#", "Notas do Semestre", screenGrades, ""},
		{"@", "Meu Perfil", screenProfile, ""},
		{"!", "Notificacoes", screenNotifications, badge},
		{"+", "Atualizar dados", 0, ""},
		{"x", "Sair", 0, ""},
	}
}

// ── Commands ──────────────────────────────────────────────────────────────────

func (a *App) doLogin(matricula, senha string) tea.Cmd {
	// Save matricula for display
	a.cfg.Matricula = matricula
	return func() tea.Msg {
		tokens, err := a.client.Login(matricula, senha)
		if err != nil {
			return loginErrMsg{err: err.Error()}
		}
		return loginSuccessMsg{access: tokens.Access, refresh: tokens.Refresh}
	}
}

func (a *App) loadData() tea.Cmd {
	return func() tea.Msg {
		academic, err := a.client.GetAcademicData()
		if err != nil {
			if err.Error() == "sessao expirada, faca login novamente" {
				return dataErrMsg{err: err.Error(), sessionExpired: true}
			}
			return dataErrMsg{err: "Falha ao carregar dados: " + err.Error()}
		}

		periods, _ := a.client.GetPeriods()
		semestre := api.LatestSemester(periods)

		var diaries []api.Diary
		if semestre != "" {
			diaries, _ = a.client.GetDiaries(semestre)
		}

		completion, _ := a.client.GetCompletionReqs()
		messages, _ := a.client.GetUnreadMessages()

		return dataLoadedMsg{
			academic:   academic,
			diaries:    diaries,
			completion: completion,
			messages:   messages,
			semestre:   semestre,
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
