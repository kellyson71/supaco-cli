package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Simple ASCII art — sem Unicode block chars que quebram em terminais
const logoASCII = `
  ____  _   _ ____   _    ____ ___
 / ___|| | | |  _ \ / \  / ___/ _ \
 \___ \| | | | |_) / _ \| |  | | | |
  ___) | |_| |  __/ ___ \ |__| |_| |
 |____/ \___/|_| /_/   \_\____\___/`

type loginModel struct {
	inputs      []textinput.Model
	focused     int
	err         string
	loading     bool
	lembrarSenha bool
}

func newLoginModel() loginModel {
	mat := textinput.New()
	mat.Placeholder = "ex: 20241001000001"
	mat.Focus()
	mat.CharLimit = 20
	mat.Width = 30
	mat.Prompt = ""

	pwd := textinput.New()
	pwd.Placeholder = "senha"
	pwd.EchoMode = textinput.EchoPassword
	pwd.EchoCharacter = '*'
	pwd.CharLimit = 64
	pwd.Width = 30
	pwd.Prompt = ""

	return loginModel{
		inputs:  []textinput.Model{mat, pwd},
		focused: 0,
	}
}

func (m loginModel) Update(msg tea.Msg) (loginModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// On the toggle step, handle toggle keys before passing to inputs
		if m.focused == 2 {
			switch msg.String() {
			case " ", "s", "S":
				m.lembrarSenha = !m.lembrarSenha
				return m, nil
			case "n", "N":
				m.lembrarSenha = false
				return m, nil
			}
		}
		switch msg.String() {
		case "tab", "down":
			m.err = ""
			steps := len(m.inputs) + 1 // +1 for toggle step
			m.focused = (m.focused + 1) % steps
			for i := range m.inputs {
				if i == m.focused {
					m.inputs[i].Focus()
				} else {
					m.inputs[i].Blur()
				}
			}
			if m.focused == 2 {
				m.inputs[1].Blur()
			}
		case "shift+tab", "up":
			m.err = ""
			steps := len(m.inputs) + 1
			m.focused = (m.focused - 1 + steps) % steps
			for i := range m.inputs {
				if i == m.focused {
					m.inputs[i].Focus()
				} else {
					m.inputs[i].Blur()
				}
			}
			if m.focused == 2 {
				m.inputs[1].Blur()
			}
		}
	}

	var cmds []tea.Cmd
	for i := range m.inputs {
		var cmd tea.Cmd
		m.inputs[i], cmd = m.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m loginModel) View(width, height int) string {
	logoLine := lipgloss.NewStyle().
		Foreground(purple).
		Bold(true).
		Render(logoASCII)

	tagLine := lipgloss.NewStyle().
		Foreground(violet).
		Render("         CLI do SUAP · IFRN")

	// Form
	matLabel := InputLabelStyle.Render(" Matricula")
	var matInput string
	if m.focused == 0 {
		matInput = InputFocusedStyle.Width(32).Render(m.inputs[0].View())
	} else {
		matInput = InputBlurredStyle.Width(32).Render(m.inputs[0].View())
	}

	pwdLabel := InputLabelStyle.Render(" Senha")
	var pwdInput string
	if m.focused == 1 {
		pwdInput = InputFocusedStyle.Width(32).Render(m.inputs[1].View())
	} else {
		pwdInput = InputBlurredStyle.Width(32).Render(m.inputs[1].View())
	}

	var btn string
	if m.loading {
		btn = InfoStyle.Render("  Autenticando...")
	} else {
		btn = lipgloss.NewStyle().
			Foreground(white).
			Background(purple).
			Bold(true).
			Padding(0, 4).
			Render("Entrar →")
	}

	// Lembrar senha toggle
	var toggleLine string
	if m.focused == 2 {
		check := "[ ]"
		if m.lembrarSenha {
			check = "[x]"
		}
		toggleLine = lipgloss.NewStyle().
			Foreground(violet).Bold(true).
			Render(" > ") +
			InputLabelStyle.Render("Lembrar senha? ") +
			lipgloss.NewStyle().Foreground(cyan).Bold(true).Render(check) +
			MutedStyle.Render("  espaco para alternar")
	} else {
		check := "[ ]"
		if m.lembrarSenha {
			check = "[x]"
		}
		toggleLine = MutedStyle.Render("   Lembrar senha? "+check)
	}

	errMsg := ""
	if m.err != "" {
		errMsg = DangerStyle.Render(" x " + m.err)
	}

	help := MutedStyle.Render("tab · enter · ctrl+c sair")

	formInner := lipgloss.JoinVertical(lipgloss.Left,
		matLabel,
		matInput,
		"",
		pwdLabel,
		pwdInput,
		"",
		toggleLine,
		"",
		" "+btn,
		"",
		errMsg,
		"",
		help,
	)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(purple).
		Padding(1, 2).
		Width(40).
		Render(formInner)

	// Centraliza verticalmente: logo + box
	content := lipgloss.JoinVertical(lipgloss.Left,
		logoLine,
		tagLine,
		"",
		strings.Repeat(" ", 2)+box,
	)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}

func (m loginModel) Matricula() string {
	return strings.TrimSpace(m.inputs[0].Value())
}

func (m loginModel) Password() string {
	return m.inputs[1].Value()
}

func (m loginModel) LembrarSenha() bool {
	return m.lembrarSenha
}

func (m loginModel) IsReady() bool {
	return m.Matricula() != "" && m.Password() != ""
}
