package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/kellyson71/supaco-cli/internal/api"
)

var ptWeekdays = map[string]int{
	"Segunda":       1,
	"Segunda-feira": 1,
	"Terça":         2,
	"Terça-feira":   2,
	"Quarta":        3,
	"Quarta-feira":  3,
	"Quinta":        4,
	"Quinta-feira":  4,
	"Sexta":         5,
	"Sexta-feira":   5,
	"Sábado":        6,
	"Sábado-feira":  6,
	"Domingo":       0,
}

var ptWeekdayNames = [7]string{"Domingo", "Segunda-feira", "Terça-feira", "Quarta-feira", "Quinta-feira", "Sexta-feira", "Sábado"}
var ptMonthNames = [13]string{"", "janeiro", "fevereiro", "março", "abril", "maio", "junho", "julho", "agosto", "setembro", "outubro", "novembro", "dezembro"}

func todayStr() string {
	now := time.Now()
	return fmt.Sprintf("%s, %d de %s de %d",
		ptWeekdayNames[now.Weekday()],
		now.Day(),
		ptMonthNames[now.Month()],
		now.Year(),
	)
}

func todayWeekdayIdx() int {
	return int(time.Now().Weekday())
}

// ── User Info (shown at start) ────────────────────────────────────────────────

func RenderUserInfo(academic *api.AcademicData, matricula string, width int) string {
	if academic == nil {
		return MutedStyle.Render(" Carregando dados...")
	}

	ira := academic.IRAFloat()
	var iraBadge string
	switch {
	case ira >= 8:
		iraBadge = BadgeGreenStyle.Render(fmt.Sprintf(" IRA %.2f ", ira))
	case ira >= 6:
		iraBadge = BadgeYellowStyle.Render(fmt.Sprintf(" IRA %.2f ", ira))
	default:
		iraBadge = BadgeRedStyle.Render(fmt.Sprintf(" IRA %.2f ", ira))
	}

	w := width - 6
	if w < 40 {
		w = 40
	}

	inner := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Top,
			BoldStyle.Render(academic.CursoNome()),
			"  ",
			iraBadge,
		),
		MutedStyle.Render("Matricula: ")+lipgloss.NewStyle().Foreground(cyan).Render(matricula)+
			MutedStyle.Render("  ·  ")+lipgloss.NewStyle().Foreground(white).Render(academic.EmailAcademico),
		MutedStyle.Render("Situacao: ")+SuccessStyle.Render(academic.Situacao)+
			MutedStyle.Render("  ·  Ingresso: ")+lipgloss.NewStyle().Foreground(white).Render(academic.Ingresso),
	)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(purple).
		Padding(0, 2).
		Width(w).
		Render(inner)
}

// ── Today's Classes ───────────────────────────────────────────────────────────

func RenderToday(diaries []api.Diary, width int) string {
	todayIdx := todayWeekdayIdx()
	todayName := ptWeekdayNames[todayIdx]

	title := PageTitleStyle.Render(" Aulas de Hoje")
	dateStr := MutedStyle.Padding(0, 0, 1, 0).Render("  " + todayStr())

	type classEntry struct {
		diary api.Diary
		block api.DayBlock
	}
	var entries []classEntry

	for _, d := range diaries {
		blocks := d.DayBlocks()
		for dia, b := range blocks {
			if ptWeekdays[dia] == todayIdx {
				entries = append(entries, classEntry{d, b})
				break
			}
		}
	}

	// Sort by start time
	for i := 1; i < len(entries); i++ {
		for j := i; j > 0 && entries[j].block.Start < entries[j-1].block.Start; j-- {
			entries[j], entries[j-1] = entries[j-1], entries[j]
		}
	}

	if len(entries) == 0 {
		todayDayName := todayName
		_ = todayDayName
		box := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(darkGray).
			Padding(1, 3).
			Margin(0, 2).
			Render(
				SuccessStyle.Render("Nenhuma aula hoje. Aproveite o dia!"),
			)
		return lipgloss.JoinVertical(lipgloss.Left, title, dateStr, "", box)
	}

	cardW := width - 6
	if cardW < 40 {
		cardW = 40
	}

	var cards []string
	for _, e := range entries {
		cards = append(cards, renderClassCard(e.diary, e.block, todayName, cardW))
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, dateStr, lipgloss.JoinVertical(lipgloss.Left, cards...))
}

func renderClassCard(d api.Diary, b api.DayBlock, dayName string, cardW int) string {
	timeStr := lipgloss.NewStyle().Foreground(cyan).Bold(true).Render(b.Start + " – " + b.End)
	name := BoldStyle.Render(d.Nome())
	prof := MutedStyle.Render(d.ProfNames())
	sala := ""
	if s := d.SalaShort(); s != "" {
		sala = MutedStyle.Render("  ·  " + s)
	}

	faltas := d.NumeroFaltas()
	maxF := d.MaxFaltas()
	chTotal := d.CargaHoraria()
	chDada := d.CargaHorariaCumprida()
	restante := d.PodeEFaltar()
	usedPct := d.AbsenceUsagePct()

	barW := 18
	bar := AbsenceBar(usedPct, barW)

	var badge string
	var borderC lipgloss.Color
	switch {
	case maxF == 0:
		badge = BadgeBlueStyle.Render(" Sem dados ")
		borderC = violet
	case restante <= 0:
		badge = BadgeRedStyle.Render(" REPROVADO por falta! ")
		borderC = red
	case restante <= int(float64(maxF)*0.3):
		badge = BadgeYellowStyle.Render(fmt.Sprintf(" Atencao! %dx restante ", restante))
		borderC = yellow
	default:
		badge = BadgeGreenStyle.Render(fmt.Sprintf(" Pode faltar %dx ", restante))
		borderC = green
	}

	row1 := lipgloss.JoinHorizontal(lipgloss.Top,
		timeStr,
		"   ",
		name,
	)
	row2 := "     " + prof + sala
	row3 := "     " + bar + "  " +
		MutedStyle.Render(fmt.Sprintf("%d/%d faltas", faltas, maxF)) +
		"   " + badge
	row4 := "     " + MutedStyle.Render(fmt.Sprintf("CH: %dh dadas / %dh total  ·  25%% = max %d faltas",
		chDada, chTotal, maxF))

	inner := lipgloss.JoinVertical(lipgloss.Left, row1, row2, "", row3, row4)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderC).
		Padding(1, 2).
		Width(cardW).
		Margin(0, 2, 1, 2).
		Render(inner)
}

// ── Week Schedule ─────────────────────────────────────────────────────────────

func RenderWeek(diaries []api.Diary, width int) string {
	title := PageTitleStyle.Render(" Grade da Semana")
	todayIdx := todayWeekdayIdx()

	type dayEntry struct {
		dia   string
		idx   int
		short string
	}
	days := []dayEntry{
		{"Segunda", 1, "Seg"},
		{"Terça", 2, "Ter"},
		{"Quarta", 3, "Qua"},
		{"Quinta", 4, "Qui"},
		{"Sexta", 5, "Sex"},
	}

	type classRow struct {
		start string
		end   string
		nome  string
		prof  string
		sala  string
	}

	var sections []string

	for _, day := range days {
		var rows []classRow
		for _, d := range diaries {
			blocks := d.DayBlocks()
			for dia, b := range blocks {
				if ptWeekdays[dia] == day.idx {
					rows = append(rows, classRow{
						start: b.Start,
						end:   b.End,
						nome:  d.Nome(),
						prof:  d.ProfNames(),
						sala:  d.SalaShort(),
					})
				}
			}
		}

		// Sort rows by start time
		for i := 1; i < len(rows); i++ {
			for j := i; j > 0 && rows[j].start < rows[j-1].start; j-- {
				rows[j], rows[j-1] = rows[j-1], rows[j]
			}
		}

		var dayLabel string
		if day.idx == todayIdx {
			dayLabel = lipgloss.NewStyle().
				Foreground(white).Background(purple).Bold(true).
				Padding(0, 1).
				Render(" " + ptWeekdayNames[day.idx] + " ← hoje ")
		} else {
			dayLabel = SubtitleStyle.Render("  " + ptWeekdayNames[day.idx])
		}

		if len(rows) == 0 {
			sections = append(sections, lipgloss.JoinVertical(lipgloss.Left,
				dayLabel,
				MutedStyle.Padding(0, 4).Render("– sem aulas –"),
				"",
			))
			continue
		}

		var rowLines []string
		for _, r := range rows {
			timeC := lipgloss.NewStyle().Foreground(cyan).Bold(true).Width(15).Render(r.start + " – " + r.end)
			nameC := BoldStyle.Render(r.nome)
			extra := ""
			if r.sala != "" {
				extra = MutedStyle.Render("  " + r.sala)
			}
			rowLines = append(rowLines, "    "+timeC+" "+nameC+extra)
			if r.prof != "" {
				rowLines = append(rowLines, "    "+strings.Repeat(" ", 16)+MutedStyle.Render(r.prof))
			}
		}
		rowLines = append(rowLines, "")

		sections = append(sections, lipgloss.JoinVertical(lipgloss.Left,
			append([]string{dayLabel}, rowLines...)...,
		))
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, "", lipgloss.JoinVertical(lipgloss.Left, sections...))
}

// ── Absences / Frequency ──────────────────────────────────────────────────────

func RenderAbsences(diaries []api.Diary, width int) string {
	title := PageTitleStyle.Render(" Frequencia e Faltas")
	subtitle := MutedStyle.Padding(0, 0, 1, 0).Render("  Disciplinas do semestre  ·  limite: 25% de faltas")

	cardW := width - 6
	if cardW < 40 {
		cardW = 40
	}

	if len(diaries) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, title, subtitle,
			MutedStyle.Padding(2, 3).Render("Nenhuma disciplina encontrada."))
	}

	var cards []string
	for _, d := range diaries {
		faltas := d.NumeroFaltas()
		maxF := d.MaxFaltas()
		chTotal := d.CargaHoraria()
		chDada := d.CargaHorariaCumprida()
		restante := d.PodeEFaltar()
		usedPct := d.AbsenceUsagePct()

		var badge string
		var borderC lipgloss.Color
		switch {
		case maxF == 0:
			badge = BadgeBlueStyle.Render(" Semestre iniciando ")
			borderC = violet
		case restante <= 0:
			badge = BadgeRedStyle.Render(" REPROVADO POR FALTA ")
			borderC = red
		case restante <= int(float64(maxF)*0.3):
			badge = BadgeYellowStyle.Render(fmt.Sprintf(" ATENCAO! %dx restante ", restante))
			borderC = yellow
		default:
			badge = BadgeGreenStyle.Render(fmt.Sprintf(" PODE FALTAR %dx ", restante))
			borderC = green
		}

		bar := AbsenceBar(usedPct, 22)
		nameW := cardW - 28
		if nameW < 20 {
			nameW = 20
		}

		faltasDetail := fmt.Sprintf("%d faltas  ·  max %d  (25%% de %dh)", faltas, maxF, chTotal)
		chDetail := fmt.Sprintf("CH dada: %dh / %dh total", chDada, chTotal)

		inner := lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.JoinHorizontal(lipgloss.Top,
				lipgloss.NewStyle().Width(nameW).Render(BoldStyle.Render(d.Nome())),
				badge,
			),
			MutedStyle.Render(faltasDetail),
			MutedStyle.Render(chDetail),
			"",
			bar+" "+MutedStyle.Render(fmt.Sprintf("%.0f%% do limite usado", usedPct)),
		)

		cards = append(cards, lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderC).
			Padding(1, 2).
			Width(cardW).
			Margin(0, 2, 1, 2).
			Render(inner),
		)
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, subtitle, lipgloss.JoinVertical(lipgloss.Left, cards...))
}

// ── Grades ────────────────────────────────────────────────────────────────────

func RenderGrades(diaries []api.Diary, academic *api.AcademicData, width int) string {
	title := PageTitleStyle.Render(" Notas do Semestre")

	iraStr := ""
	if academic != nil {
		iraStr = fmt.Sprintf("  ·  IRA: %.2f", academic.IRAFloat())
	}
	subtitle := MutedStyle.Padding(0, 0, 1, 0).Render("  Notas por disciplina" + iraStr)

	cardW := width - 6
	if cardW < 40 {
		cardW = 40
	}

	if len(diaries) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, title, subtitle,
			MutedStyle.Padding(2, 3).Render("Nenhuma nota disponivel."))
	}

	var cards []string
	for _, d := range diaries {
		var notasParts []string
		var mediaLine string
		var situacaoBadge string
		borderC := darkGray

		if b := d.Boletim; b != nil {
			for _, n := range b.Notas() {
				var s lipgloss.Style
				switch {
				case n.Nota >= 7:
					s = lipgloss.NewStyle().Foreground(green).Bold(true)
				case n.Nota >= 5:
					s = lipgloss.NewStyle().Foreground(yellow).Bold(true)
				default:
					s = lipgloss.NewStyle().Foreground(red).Bold(true)
				}
				label := MutedStyle.Render(fmt.Sprintf("N%d:", n.Etapa))
				nota := s.Render(fmt.Sprintf("%.1f", n.Nota))
				if n.Faltas > 0 {
					nota += MutedStyle.Render(fmt.Sprintf("(%df)", n.Faltas))
				}
				notasParts = append(notasParts, label+" "+nota)
			}
			if b.MediaDisciplina != nil {
				m := *b.MediaDisciplina
				var ms lipgloss.Style
				switch {
				case m >= 7:
					ms = SuccessStyle
				case m >= 5:
					ms = WarningStyle
				default:
					ms = DangerStyle
				}
				mediaLine = MutedStyle.Render("Media: ") + ms.Render(fmt.Sprintf("%.2f", m))
				if b.MediaFinalDisciplina != nil && *b.MediaFinalDisciplina != m {
					mf := *b.MediaFinalDisciplina
					var mfs lipgloss.Style
					switch {
					case mf >= 7:
						mfs = SuccessStyle
					case mf >= 5:
						mfs = WarningStyle
					default:
						mfs = DangerStyle
					}
					mediaLine += "   " + MutedStyle.Render("Final: ") + mfs.Render(fmt.Sprintf("%.2f", mf))
				}
			}
			switch b.Situacao {
			case "Aprovado":
				situacaoBadge = BadgeGreenStyle.Render(" Aprovado ")
				borderC = green
			case "Reprovado":
				situacaoBadge = BadgeRedStyle.Render(" Reprovado ")
				borderC = red
			default:
				situacaoBadge = BadgeBlueStyle.Render(" " + b.Situacao + " ")
			}
		} else {
			for _, n := range d.Disciplina.Notas {
				var s lipgloss.Style
				switch {
				case n.Nota >= 7:
					s = lipgloss.NewStyle().Foreground(green).Bold(true)
				case n.Nota >= 5:
					s = lipgloss.NewStyle().Foreground(yellow).Bold(true)
				default:
					s = lipgloss.NewStyle().Foreground(red).Bold(true)
				}
				notasParts = append(notasParts,
					MutedStyle.Render(fmt.Sprintf("N%d:", n.Etapa))+" "+s.Render(fmt.Sprintf("%.1f", n.Nota)),
				)
			}
		}

		nameW := cardW - 20
		if nameW < 20 {
			nameW = 20
		}
		header := lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(nameW).Render(BoldStyle.Render(d.Nome())),
			situacaoBadge,
		)

		gradesLine := strings.Join(notasParts, "   ")
		if gradesLine == "" {
			gradesLine = MutedStyle.Render("Sem notas lancadas ainda")
		}

		lines := []string{header, "", gradesLine}
		if mediaLine != "" {
			lines = append(lines, "", mediaLine)
		}

		cards = append(cards, lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderC).
			Padding(1, 2).
			Width(cardW).
			Margin(0, 2, 1, 2).
			Render(lipgloss.JoinVertical(lipgloss.Left, lines...)),
		)
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, subtitle, lipgloss.JoinVertical(lipgloss.Left, cards...))
}

// ── Profile ───────────────────────────────────────────────────────────────────

func RenderProfile(academic *api.AcademicData, completion *api.CompletionReqs, matricula string, width int) string {
	title := PageTitleStyle.Render(" Meu Perfil")

	if academic == nil {
		return lipgloss.JoinVertical(lipgloss.Left, title, MutedStyle.Padding(2, 3).Render("Carregando..."))
	}

	cardW := width - 6
	if cardW < 40 {
		cardW = 40
	}

	ira := academic.IRAFloat()
	var iraBadge string
	switch {
	case ira >= 8:
		iraBadge = BadgeGreenStyle.Render(fmt.Sprintf(" %.2f ", ira))
	case ira >= 6:
		iraBadge = BadgeYellowStyle.Render(fmt.Sprintf(" %.2f ", ira))
	default:
		iraBadge = BadgeRedStyle.Render(fmt.Sprintf(" %.2f ", ira))
	}

	profileLines := []string{
		SubtitleStyle.Render(" Dados Pessoais"),
		"",
		MutedStyle.Render("Matricula:  ") + lipgloss.NewStyle().Foreground(cyan).Bold(true).Render(matricula),
		MutedStyle.Render("E-mail:     ") + lipgloss.NewStyle().Foreground(white).Render(academic.EmailAcademico),
		"",
		SubtitleStyle.Render(" Dados Academicos"),
		"",
		MutedStyle.Render("Curso:      ") + BoldStyle.Render(academic.CursoNome()),
		MutedStyle.Render("Situacao:   ") + SuccessStyle.Render(academic.Situacao),
		MutedStyle.Render("Ingresso:   ") + lipgloss.NewStyle().Foreground(white).Render(academic.Ingresso),
		MutedStyle.Render("Periodo:    ") + lipgloss.NewStyle().Foreground(white).Render(fmt.Sprintf("%d", academic.PeriodoReferencia)),
		lipgloss.JoinHorizontal(lipgloss.Top,
			MutedStyle.Render("IRA:        "),
			iraBadge,
		),
	}

	var completionLines []string
	if completion != nil {
		pct := completion.Percentual()
		bar := ProgressBar(pct, 26)
		completionLines = []string{
			"",
			SubtitleStyle.Render(" Conclusao do Curso"),
			"",
			bar + "  " + lipgloss.NewStyle().Foreground(cyan).Bold(true).Render(fmt.Sprintf("%.1f%%", pct)),
			MutedStyle.Render(fmt.Sprintf("Obrigatorias: %dh / %dh  ·  Pendente: %dh",
				completion.Obrigatorios.CHCumprida,
				completion.Obrigatorios.CHEsperada,
				completion.Obrigatorios.CHPendente)),
		}
	}

	allLines := append(profileLines, completionLines...)

	card := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(purple).
		Padding(1, 2).
		Width(cardW).
		Margin(0, 2).
		Render(lipgloss.JoinVertical(lipgloss.Left, allLines...))

	return lipgloss.JoinVertical(lipgloss.Left, title, "", card)
}

// ── Notifications ─────────────────────────────────────────────────────────────

func RenderNotifications(msgs *api.MessagesResponse, width int) string {
	title := PageTitleStyle.Render(" Notificacoes")

	cardW := width - 6
	if cardW < 40 {
		cardW = 40
	}

	if msgs == nil || len(msgs.Results) == 0 {
		empty := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(darkGray).
			Padding(2, 4).
			Margin(0, 2).
			Render(SuccessStyle.Render("Nenhuma mensagem nao lida."))
		return lipgloss.JoinVertical(lipgloss.Left, title, "", empty)
	}

	subtitle := MutedStyle.Padding(0, 0, 1, 0).Render(fmt.Sprintf("  %d mensagem(ns) nao lida(s)", msgs.Count))

	var cards []string
	for _, m := range msgs.Results {
		inner := lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Foreground(yellow).Render("  ")+BoldStyle.Render(m.Assunto),
			MutedStyle.Render("De: ")+lipgloss.NewStyle().Foreground(cyan).Render(m.Remetente)+
				MutedStyle.Render("  ·  "+m.Data),
		)
		cards = append(cards, lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(yellow).
			Padding(1, 2).
			Width(cardW).
			Margin(0, 2, 1, 2).
			Render(inner),
		)
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, subtitle, lipgloss.JoinVertical(lipgloss.Left, cards...))
}
