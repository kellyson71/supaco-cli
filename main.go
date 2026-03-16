package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"supaco-cli/internal/ui"
)

func main() {
	p := tea.NewProgram(
		ui.NewApp(),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao iniciar o Supaco: %v\n", err)
		os.Exit(1)
	}
}
