package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kellyson71/supaco-cli/internal/api"
)

// quickHeader renders a header line for quick commands
func quickHeader(title, semestre string) string {
	left := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F8FAFC")).
		Background(lipgloss.Color("#7C3AED")).
		Bold(true).
		Padding(0, 2).
		Render("SUPACO")
	mid := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#DDD6FE")).
		Background(lipgloss.Color("#7C3AED")).
		Padding(0, 1).
		Render(title)
	right := ""
	if semestre != "" {
		right = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#C4B5FD")).
			Background(lipgloss.Color("#7C3AED")).
			Padding(0, 1).
			Render(semestre)
	}
	bar := lipgloss.NewStyle().Background(lipgloss.Color("#7C3AED")).Render(
		left + "  " + mid + strings.Repeat(" ", 20) + right,
	)
	return bar
}

// QuickToday prints today's classes without opening TUI
func QuickToday(diaries []api.Diary, semestre string) string {
	var sb strings.Builder
	sb.WriteString(quickHeader("Aulas de Hoje", semestre))
	sb.WriteString("\n")
	sb.WriteString(MutedStyle.Render("  " + todayStr()))
	sb.WriteString("\n\n")

	todayIdx := todayWeekdayIdx()
	type entry struct {
		d api.Diary
		b api.DayBlock
	}
	var entries []entry
	for _, d := range diaries {
		for dia, b := range d.DayBlocks() {
			if ptWeekdays[dia] == todayIdx {
				entries = append(entries, entry{d, b})
			}
		}
	}
	// sort by start
	for i := 1; i < len(entries); i++ {
		for j := i; j > 0 && entries[j].b.Start < entries[j-1].b.Start; j-- {
			entries[j], entries[j-1] = entries[j-1], entries[j]
		}
	}

	if len(entries) == 0 {
		sb.WriteString("  " + SuccessStyle.Render("Nenhuma aula hoje!") + "\n")
		return sb.String()
	}

	for _, e := range entries {
		d := e.d
		b := e.b
		freq := d.Disciplina.Frequencia
		restante := d.PodeEFaltar()

		var freqBadge string
		switch {
		case freq >= 85:
			freqBadge = BadgeGreenStyle.Render(fmt.Sprintf(" pode faltar %dx ", restante))
		case freq >= 75:
			freqBadge = BadgeYellowStyle.Render(fmt.Sprintf(" atencao! %dx ", restante))
		case freq == 0:
			freqBadge = BadgeBlueStyle.Render(" iniciando ")
		default:
			freqBadge = BadgeRedStyle.Render(" reprovado por falta ")
		}

		timeStr := lipgloss.NewStyle().Foreground(lipgloss.Color("#06B6D4")).Bold(true).Width(16).Render(b.Start + " – " + b.End)
		nome := BoldStyle.Render(d.Disciplina.Descricao)
		prof := MutedStyle.Render(d.ProfNames())
		sala := ""
		if s := d.SalaShort(); s != "" {
			sala = MutedStyle.Render(" · " + s)
		}
		freqStr := lipgloss.NewStyle().Foreground(muted).Render(fmt.Sprintf("%.1f%%", freq))

		sb.WriteString("  " + timeStr + " " + nome + "\n")
		sb.WriteString("  " + strings.Repeat(" ", 16) + " " + prof + sala + "\n")
		sb.WriteString("  " + strings.Repeat(" ", 16) + " " + freqStr + "  " + freqBadge + "\n")
		sb.WriteString("\n")
	}

	return sb.String()
}

// QuickWeek prints the week schedule
func QuickWeek(diaries []api.Diary, semestre string) string {
	var sb strings.Builder
	sb.WriteString(quickHeader("Grade da Semana", semestre))
	sb.WriteString("\n\n")

	todayIdx := todayWeekdayIdx()
	days := []struct {
		name string
		idx  int
	}{
		{"Segunda-feira", 1}, {"Terca-feira", 2}, {"Quarta-feira", 3},
		{"Quinta-feira", 4}, {"Sexta-feira", 5},
	}

	for _, day := range days {
		var classes []struct {
			start, end, nome, prof, sala string
		}
		for _, d := range diaries {
			for dia, b := range d.DayBlocks() {
				if ptWeekdays[dia] == day.idx {
					classes = append(classes, struct {
						start, end, nome, prof, sala string
					}{b.Start, b.End, d.Disciplina.Descricao, d.ProfNames(), d.SalaShort()})
				}
			}
		}
		// sort
		for i := 1; i < len(classes); i++ {
			for j := i; j > 0 && classes[j].start < classes[j-1].start; j-- {
				classes[j], classes[j-1] = classes[j-1], classes[j]
			}
		}

		var dayLabel string
		if day.idx == todayIdx {
			dayLabel = lipgloss.NewStyle().Foreground(lipgloss.Color("#F8FAFC")).Background(lipgloss.Color("#7C3AED")).Bold(true).Padding(0, 1).Render(" " + day.name + " ← hoje ")
		} else {
			dayLabel = SubtitleStyle.Render(day.name)
		}

		sb.WriteString("  " + dayLabel + "\n")
		if len(classes) == 0 {
			sb.WriteString("    " + MutedStyle.Render("– sem aulas –") + "\n")
		} else {
			for _, c := range classes {
				timeStr := lipgloss.NewStyle().Foreground(lipgloss.Color("#06B6D4")).Bold(true).Width(14).Render(c.start + " – " + c.end)
				extra := ""
				if c.sala != "" {
					extra = MutedStyle.Render("  " + c.sala)
				}
				sb.WriteString("    " + timeStr + "  " + BoldStyle.Render(c.nome) + extra + "\n")
				if c.prof != "" {
					sb.WriteString("    " + strings.Repeat(" ", 14) + "  " + MutedStyle.Render(c.prof) + "\n")
				}
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// QuickFaltas prints attendance status per subject
func QuickFaltas(diaries []api.Diary, semestre string) string {
	var sb strings.Builder
	sb.WriteString(quickHeader("Frequencia e Faltas", semestre))
	sb.WriteString("\n\n")

	if len(diaries) == 0 {
		sb.WriteString("  " + MutedStyle.Render("Nenhuma disciplina encontrada.") + "\n")
		return sb.String()
	}

	// Column widths
	nameW := 36

	for _, d := range diaries {
		freq := d.Disciplina.Frequencia
		faltas := d.Disciplina.QtdFaltas
		maxF := d.MaxFaltas()
		restante := d.PodeEFaltar()

		var badge string
		var nameStyle lipgloss.Style
		switch {
		case faltas == 0 && freq == 0:
			badge = BadgeBlueStyle.Render(" iniciando ")
			nameStyle = BoldStyle
		case freq >= 85:
			badge = BadgeGreenStyle.Render(fmt.Sprintf(" PODE FALTAR %dx ", restante))
			nameStyle = BoldStyle
		case freq >= 75:
			badge = BadgeYellowStyle.Render(fmt.Sprintf(" ATENCAO! %dx restante ", restante))
			nameStyle = WarningStyle
		default:
			badge = BadgeRedStyle.Render(" REPROVADO POR FALTA ")
			nameStyle = DangerStyle
		}

		nome := d.Disciplina.Descricao
		if len(nome) > nameW {
			nome = nome[:nameW-1] + "…"
		}
		nameCol := nameStyle.Width(nameW).Render(nome)
		bar := ProgressBar(freq, 14)
		freqStr := lipgloss.NewStyle().Foreground(muted).Width(7).Render(fmt.Sprintf("%.1f%%", freq))
		faltasStr := MutedStyle.Render(fmt.Sprintf("%d/%d", faltas, maxF))

		sb.WriteString("  " + nameCol + "  " + bar + "  " + freqStr + " " + faltasStr + "  " + badge + "\n")
	}

	sb.WriteString("\n")
	return sb.String()
}

// QuickNotas prints grades per subject
func QuickNotas(diaries []api.Diary, academic *api.AcademicData, semestre string) string {
	var sb strings.Builder
	if academic != nil {
		sb.WriteString(quickHeader(fmt.Sprintf("Notas  ·  IRA: %.2f", academic.IRAFloat()), semestre))
	} else {
		sb.WriteString(quickHeader("Notas do Semestre", semestre))
	}
	sb.WriteString("\n\n")

	if len(diaries) == 0 {
		sb.WriteString("  " + MutedStyle.Render("Nenhuma nota disponível.") + "\n")
		return sb.String()
	}

	nameW := 36
	for _, d := range diaries {
		nome := d.Disciplina.Descricao
		if len(nome) > nameW {
			nome = nome[:nameW-1] + "…"
		}
		nameCol := BoldStyle.Width(nameW).Render(nome)

		var notasParts []string
		for _, n := range d.Disciplina.Notas {
			var s lipgloss.Style
			switch {
			case n.Nota >= 7:
				s = SuccessStyle
			case n.Nota >= 5:
				s = WarningStyle
			default:
				s = DangerStyle
			}
			notasParts = append(notasParts, MutedStyle.Render(fmt.Sprintf("N%d:", n.Etapa))+s.Render(fmt.Sprintf("%.1f", n.Nota)))
		}

		notasStr := strings.Join(notasParts, "  ")
		if notasStr == "" {
			notasStr = MutedStyle.Render("sem notas")
		}

		var mediaStr string
		for _, m := range d.Disciplina.Medias {
			var s lipgloss.Style
			switch {
			case m.Media >= 7:
				s = SuccessStyle
			case m.Media >= 5:
				s = WarningStyle
			default:
				s = DangerStyle
			}
			mediaStr += "  " + MutedStyle.Render(m.Descricao+":") + s.Render(fmt.Sprintf("%.1f", m.Media))
		}

		sb.WriteString("  " + nameCol + "  " + notasStr + mediaStr + "\n")
	}

	sb.WriteString("\n")
	return sb.String()
}

// QuickStatus prints a quick overview
func QuickStatus(diaries []api.Diary, academic *api.AcademicData, msgs *api.MessagesResponse, semestre, matricula string) string {
	var sb strings.Builder
	sb.WriteString(quickHeader("Status Rapido", semestre))
	sb.WriteString("\n\n")

	// User info
	if academic != nil {
		ira := academic.IRAFloat()
		iraColor := lipgloss.Color("#10B981")
		if ira < 6 {
			iraColor = lipgloss.Color("#EF4444")
		} else if ira < 8 {
			iraColor = lipgloss.Color("#F59E0B")
		}

		sb.WriteString("  " + MutedStyle.Render("Matricula:  ") + lipgloss.NewStyle().Foreground(lipgloss.Color("#06B6D4")).Bold(true).Render(matricula) + "\n")
		sb.WriteString("  " + MutedStyle.Render("Curso:      ") + BoldStyle.Render(academic.CursoNome()) + "\n")
		sb.WriteString("  " + MutedStyle.Render("Situacao:   ") + SuccessStyle.Render(academic.Situacao) + "\n")
		sb.WriteString("  " + MutedStyle.Render("IRA:        ") + lipgloss.NewStyle().Foreground(iraColor).Bold(true).Render(fmt.Sprintf("%.2f", ira)) + "\n")
		sb.WriteString("\n")
	}

	// Today
	todayIdx := todayWeekdayIdx()
	todayCount := 0
	for _, d := range diaries {
		for dia := range d.DayBlocks() {
			if ptWeekdays[dia] == todayIdx {
				todayCount++
			}
		}
	}
	sb.WriteString("  " + SubtitleStyle.Render("Hoje") + MutedStyle.Render(" — "+todayStr()) + "\n")
	if todayCount == 0 {
		sb.WriteString("  " + SuccessStyle.Render("  Sem aulas hoje!") + "\n")
	} else {
		sb.WriteString("  " + lipgloss.NewStyle().Foreground(lipgloss.Color("#06B6D4")).Render(fmt.Sprintf("  %d aula(s)", todayCount)) + "\n")
		for _, d := range diaries {
			for dia, b := range d.DayBlocks() {
				if ptWeekdays[dia] == todayIdx {
					sb.WriteString("    " +
						lipgloss.NewStyle().Foreground(lipgloss.Color("#06B6D4")).Width(14).Render(b.Start+"–"+b.End) +
						"  " + BoldStyle.Render(d.Disciplina.Descricao) + "\n")
				}
			}
		}
	}
	sb.WriteString("\n")

	// Frequency alerts
	sb.WriteString("  " + SubtitleStyle.Render("Frequencia") + "\n")
	atRisk := 0
	critical := 0
	ok := 0
	for _, d := range diaries {
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
	_ = ok
	if critical > 0 {
		sb.WriteString("  " + BadgeRedStyle.Render(fmt.Sprintf(" %d reprovada(s) por falta ", critical)) + "\n")
	}
	if atRisk > 0 {
		sb.WriteString("  " + BadgeYellowStyle.Render(fmt.Sprintf(" %d em risco ", atRisk)) + "\n")
	}
	if critical == 0 && atRisk == 0 && len(diaries) > 0 {
		sb.WriteString("  " + BadgeGreenStyle.Render(" Frequencia ok! ") + "\n")
	}
	sb.WriteString("\n")

	// Notifications
	if msgs != nil && msgs.Count > 0 {
		sb.WriteString("  " + BadgeYellowStyle.Render(fmt.Sprintf(" %d notificacao(oes) nao lida(s) ", msgs.Count)) + "\n\n")
	}

	return sb.String()
}
