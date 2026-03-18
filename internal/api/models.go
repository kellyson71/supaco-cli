package api

import (
	"strconv"
	"strings"
)

// Auth
type TokenPair struct {
	Access  string `json:"access"`
	Refresh string `json:"refresh"`
}

// Academic data — from /api/ensino/meus-dados-aluno/
type AcademicData struct {
	Ingresso          string `json:"ingresso"`
	EmailAcademico    string `json:"email_academico"`
	PeriodoReferencia int    `json:"periodo_referencia"`
	IRA               string `json:"ira"` // "81,05" — comma decimal
	Curso             string `json:"curso"`
	Matriz            string `json:"matriz"`
	QtdPeriodos       int    `json:"qtd_periodos"`
	Situacao          string `json:"situacao"`
}

func (a AcademicData) IRAFloat() float64 {
	s := strings.ReplaceAll(a.IRA, ",", ".")
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// Short course name (removes code prefix and campus suffix)
func (a AcademicData) CursoNome() string {
	// "09404 - Tecnologia em ADS (2012) - Campus X" → "Tecnologia em ADS (2012)"
	parts := strings.SplitN(a.Curso, " - ", 3)
	if len(parts) >= 2 {
		return strings.TrimSpace(parts[1])
	}
	return a.Curso
}

// Periods — paginated
type Period struct {
	AnoLetivo     int `json:"ano_letivo"`
	PeriodoLetivo int `json:"periodo_letivo"`
}

type PaginatedPeriods struct {
	Results []Period `json:"results"`
	Count   int      `json:"count"`
}

// Diary structures — from /api/ensino/diarios/{semestre}/
type DiaryHorario struct {
	Dia     string `json:"dia"`     // "Terça"
	Horario string `json:"horario"` // "13:00 - 13:45"
}

func (h DiaryHorario) Start() string {
	parts := strings.SplitN(h.Horario, " - ", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[0])
	}
	return h.Horario
}

func (h DiaryHorario) End() string {
	parts := strings.SplitN(h.Horario, " - ", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[1])
	}
	return h.Horario
}

type DiaryLocal struct {
	ID   int    `json:"id"`
	Sala string `json:"sala"`
}

type Professor struct {
	Nome  string `json:"nome"`
	Email string `json:"email"`
}

type DiaryDisciplina struct {
	ID         int     `json:"id"`
	Descricao  string  `json:"descricao"`
	Sigla      string  `json:"sigla"`
	CHTotal    int     `json:"ch_total_aula"`
	CHCumprida int     `json:"ch_cumprida_aula"`
	QtdFaltas  int     `json:"qtd_faltas"`
	Frequencia float64 `json:"frequencia"`
	Notas      []Note  `json:"notas"`
	Medias     []Media `json:"medias"`
}

type Note struct {
	Etapa int     `json:"etapa"`
	Nota  float64 `json:"nota"`
}

type Media struct {
	Descricao string  `json:"descricao"`
	Media     float64 `json:"media"`
}

type Diary struct {
	ID          int             `json:"id"`
	Disciplina  DiaryDisciplina `json:"disciplina"`
	Professores []Professor     `json:"professores"`
	Horarios    []DiaryHorario  `json:"horarios"`
	Local       *DiaryLocal     `json:"local"`
	// Boletim é preenchido após merge com /api/ensino/meu-boletim/
	Boletim     *BoletimEntry   `json:"-"`
}

type PaginatedDiaries struct {
	Results []Diary `json:"results"`
	Count   int     `json:"count"`
}

// DayBlock groups all horarios of a diary for a specific day
type DayBlock struct {
	Dia   string
	Start string
	End   string
}

func (d Diary) DayBlocks() map[string]DayBlock {
	blocks := map[string]DayBlock{}
	for _, h := range d.Horarios {
		if b, ok := blocks[h.Dia]; !ok {
			blocks[h.Dia] = DayBlock{Dia: h.Dia, Start: h.Start(), End: h.End()}
		} else {
			// keep earliest start, latest end
			if h.Start() < b.Start {
				b.Start = h.Start()
			}
			if h.End() > b.End {
				b.End = h.End()
			}
			blocks[h.Dia] = b
		}
	}
	return blocks
}

func (d Diary) ProfNames() string {
	names := make([]string, len(d.Professores))
	for i, p := range d.Professores {
		// "Breno Mauricio de Freitas Viana" → "Breno Viana"
		parts := strings.Fields(p.Nome)
		if len(parts) >= 2 {
			names[i] = parts[0] + " " + parts[len(parts)-1]
		} else {
			names[i] = p.Nome
		}
	}
	return strings.Join(names, ", ")
}

func (d Diary) SalaShort() string {
	if d.Local == nil {
		return ""
	}
	// "Chave 117 - Sala de Aula 16 (Piso 02) - Bloco 10 - Salas de Aula (PF)" → "Sala 16"
	sala := d.Local.Sala
	if idx := strings.Index(sala, "Sala de Aula "); idx >= 0 {
		rest := sala[idx+len("Sala de Aula "):]
		num := strings.Fields(rest)[0]
		return "Sala " + num
	}
	// fallback: first 20 chars
	if len(sala) > 20 {
		return sala[:20] + "…"
	}
	return sala
}

func (d Diary) CargaHoraria() int {
	if d.Boletim != nil {
		return d.Boletim.CargaHoraria
	}
	return d.Disciplina.CHTotal
}

func (d Diary) CargaHorariaCumprida() int {
	if d.Boletim != nil {
		return d.Boletim.CargaHorariaCumprida
	}
	return d.Disciplina.CHCumprida
}

// AbsenceUsagePct returns how much of the absence budget has been used (0-100+)
func (d Diary) AbsenceUsagePct() float64 {
	maxF := d.MaxFaltas()
	if maxF == 0 {
		return 0
	}
	return float64(d.NumeroFaltas()) / float64(maxF) * 100
}

func (d Diary) NumeroFaltas() int {
	if d.Boletim != nil {
		return d.Boletim.NumeroFaltas
	}
	return d.Disciplina.QtdFaltas
}

func (d Diary) Frequencia() float64 {
	if d.Boletim != nil {
		return d.Boletim.PercentualCargaHorariaFrequentada
	}
	return d.Disciplina.Frequencia
}

func (d Diary) MaxFaltas() int {
	ch := d.CargaHoraria()
	if ch == 0 {
		return 0
	}
	return int(float64(ch) * 0.25)
}

func (d Diary) PodeEFaltar() int {
	return d.MaxFaltas() - d.NumeroFaltas()
}

func (d Diary) Nome() string {
	if d.Boletim != nil && d.Boletim.Disciplina != "" {
		// "TEC.0017 - Arquitetura de Computadores" → "Arquitetura de Computadores"
		parts := strings.SplitN(d.Boletim.Disciplina, " - ", 2)
		if len(parts) == 2 {
			return strings.TrimSpace(parts[1])
		}
		return d.Boletim.Disciplina
	}
	return d.Disciplina.Descricao
}

// Boletim — from /api/ensino/meu-boletim/{ano}/{periodo}/
// Esta é a fonte correta para faltas, frequência e notas reais.

type BoletimNota struct {
	Nota   *float64 `json:"nota"`
	Faltas int      `json:"faltas"`
}

type BoletimEntry struct {
	CodigoDiario                    string       `json:"codigo_diario"`
	Disciplina                      string       `json:"disciplina"`
	CargaHoraria                    int          `json:"carga_horaria"`
	CargaHorariaCumprida            int          `json:"carga_horaria_cumprida"`
	NumeroFaltas                    int          `json:"numero_faltas"`
	PercentualCargaHorariaFrequentada float64    `json:"percentual_carga_horaria_frequentada"`
	Situacao                        string       `json:"situacao"`
	QuantidadeAvaliacoes            int          `json:"quantidade_avaliacoes"`
	NotaEtapa1                      BoletimNota  `json:"nota_etapa_1"`
	NotaEtapa2                      BoletimNota  `json:"nota_etapa_2"`
	NotaEtapa3                      BoletimNota  `json:"nota_etapa_3"`
	NotaEtapa4                      BoletimNota  `json:"nota_etapa_4"`
	MediaDisciplina                 *float64     `json:"media_disciplina"`
	NotaAvaliacaoFinal              BoletimNota  `json:"nota_avaliacao_final"`
	MediaFinalDisciplina            *float64     `json:"media_final_disciplina"`
}

type PaginatedBoletim struct {
	Results []BoletimEntry `json:"results"`
	Count   int            `json:"count"`
}

// Notas retorna as etapas com nota lançada
func (b BoletimEntry) Notas() []struct {
	Etapa int
	Nota  float64
	Faltas int
} {
	etapas := []struct {
		N *float64
		F int
		E int
	}{
		{b.NotaEtapa1.Nota, b.NotaEtapa1.Faltas, 1},
		{b.NotaEtapa2.Nota, b.NotaEtapa2.Faltas, 2},
		{b.NotaEtapa3.Nota, b.NotaEtapa3.Faltas, 3},
		{b.NotaEtapa4.Nota, b.NotaEtapa4.Faltas, 4},
	}
	var out []struct {
		Etapa  int
		Nota   float64
		Faltas int
	}
	for _, e := range etapas {
		if e.N != nil {
			out = append(out, struct {
				Etapa  int
				Nota   float64
				Faltas int
			}{e.E, *e.N, e.F})
		}
	}
	return out
}

// Completion requirements — from /api/ensino/requisitos-conclusao/
type CHBlock struct {
	CHEsperada int `json:"ch_esperada"`
	CHCumprida int `json:"ch_cumprida"`
	CHPendente int `json:"ch_pendente"`
}

type CompletionReqs struct {
	PercentualCumprida string  `json:"percentual_cumprida"` // "20.97"
	Obrigatorios       CHBlock `json:"regulares_obrigatorios"`
	Optativos          CHBlock `json:"regulares_optativos"`
}

func (c CompletionReqs) Percentual() float64 {
	f, _ := strconv.ParseFloat(c.PercentualCumprida, 64)
	return f
}

// Messages
type Message struct {
	ID        int    `json:"id"`
	Assunto   string `json:"assunto"`
	Data      string `json:"data"`
	Remetente string `json:"remetente"`
}

type MessagesResponse struct {
	Results []Message `json:"results"`
	Count   int       `json:"count"`
}
