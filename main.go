package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kellyson71/supaco-cli/internal/api"
	"github.com/kellyson71/supaco-cli/internal/config"
	"github.com/kellyson71/supaco-cli/internal/ui"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		runTUI()
		return
	}

	switch strings.ToLower(args[0]) {
	case "hoje", "today":
		runQuick("hoje")
	case "semana", "week":
		runQuick("semana")
	case "faltas", "freq", "frequencia":
		runQuick("faltas")
	case "notas", "grades":
		runQuick("notas")
	case "status":
		runQuick("status")
	case "help", "--help", "-h":
		printHelp()
	default:
		fmt.Fprintf(os.Stderr, "Comando desconhecido: %s\n\nUse 'supaco help' para ver os comandos disponíveis.\n", args[0])
		os.Exit(1)
	}
}

func runTUI() {
	p := tea.NewProgram(ui.NewApp(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
		os.Exit(1)
	}
}

func runQuick(cmd string) {
	cfg, _ := config.Load()
	if cfg.AccessToken == "" {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Bold(true)
		fmt.Println(errorStyle.Render("✗ Não autenticado. Execute 'supaco' para fazer login."))
		os.Exit(1)
	}

	client := api.NewClient()
	client.AccessToken = cfg.AccessToken
	client.RefreshToken = cfg.RefreshToken
	client.OnRefresh = func(access, refresh string) {
		cfg.AccessToken = access
		cfg.RefreshToken = refresh
		cfg.Save()
	}

	// Fetch data needed for the command
	academic, _ := client.GetAcademicData()
	periods, _ := client.GetPeriods()
	semestre := api.LatestSemester(periods)
	latest := api.LatestPeriod(periods)
	var diaries []api.Diary
	if semestre != "" {
		diaries, _ = client.GetDiaries(semestre)
		if latest != nil {
			b, err := client.GetBoletim(latest.AnoLetivo, latest.PeriodoLetivo)
			if err == nil {
				diaries = api.MergeBoletim(diaries, b)
			}
		}
	}

	switch cmd {
	case "hoje":
		fmt.Println(ui.QuickToday(diaries, semestre))
	case "semana":
		fmt.Println(ui.QuickWeek(diaries, semestre))
	case "faltas":
		fmt.Println(ui.QuickFaltas(diaries, semestre))
	case "notas":
		fmt.Println(ui.QuickNotas(diaries, academic, semestre))
	case "status":
		msgs, _ := client.GetUnreadMessages()
		fmt.Println(ui.QuickStatus(diaries, academic, msgs, semestre, cfg.Matricula))
	}
}

func printHelp() {
	purple := lipgloss.Color("#7C3AED")
	cyan := lipgloss.Color("#06B6D4")
	white := lipgloss.Color("#F8FAFC")
	muted := lipgloss.Color("#94A3B8")

	logo := lipgloss.NewStyle().Foreground(purple).Bold(true).Render(
		"  ____  _   _ ____   _    ____ ___\n" +
			" / ___|| | | |  _ \\/ \\  / ___/ _ \\\n" +
			" \\___ \\| | | | |_) / _ \\| |  | | | |\n" +
			"  ___) | |_| |  __/ ___ \\ |__| |_| |\n" +
			" |____/ \\___/|_| /_/   \\_\\____\\___/",
	)
	sub := lipgloss.NewStyle().Foreground(lipgloss.Color("#8B5CF6")).Render("         CLI do SUAP · IFRN")

	cmdStyle := lipgloss.NewStyle().Foreground(cyan).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(white)
	mutedStyle := lipgloss.NewStyle().Foreground(muted)

	fmt.Println(logo)
	fmt.Println(sub)
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Foreground(white).Bold(true).Render("Uso:"))
	fmt.Printf("  %s %s\n\n", cmdStyle.Render("supaco"), mutedStyle.Render("[comando]"))
	fmt.Println(lipgloss.NewStyle().Foreground(white).Bold(true).Render("Comandos:"))
	cmds := [][]string{
		{"          ", "Abre o painel interativo (TUI)"},
		{"hoje      ", "Aulas de hoje"},
		{"semana    ", "Grade da semana"},
		{"faltas    ", "Frequência e limite de faltas"},
		{"notas     ", "Notas do semestre"},
		{"status    ", "Resumo rápido (IRA, hoje, alertas)"},
	}
	for _, c := range cmds {
		fmt.Printf("  %s  %s\n", cmdStyle.Render(c[0]), descStyle.Render(c[1]))
	}
	fmt.Println()
	fmt.Println(mutedStyle.Render("  Sem comando, abre o painel completo. Credenciais salvas em ~/.config/supaco/config.json"))
}
