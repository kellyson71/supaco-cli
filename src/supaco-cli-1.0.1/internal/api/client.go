package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const baseURL = "https://suap.ifrn.edu.br"

type Client struct {
	http         *http.Client
	AccessToken  string
	RefreshToken string
	OnRefresh    func(access, refresh string)
}

func NewClient() *Client {
	return &Client{
		http: &http.Client{Timeout: 20 * time.Second},
	}
}

func (c *Client) Login(matricula, senha string) (*TokenPair, error) {
	body, _ := json.Marshal(map[string]string{
		"username": matricula,
		"password": senha,
	})
	resp, err := c.http.Post(baseURL+"/api/token/pair", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("sem conexao com o SUAP")
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("matricula ou senha incorretos")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("erro do servidor SUAP (%d)", resp.StatusCode)
	}

	var tokens TokenPair
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return nil, fmt.Errorf("resposta invalida do SUAP")
	}
	c.AccessToken = tokens.Access
	c.RefreshToken = tokens.Refresh
	return &tokens, nil
}

func (c *Client) refreshTokens() error {
	body, _ := json.Marshal(map[string]string{"refresh": c.RefreshToken})
	resp, err := c.http.Post(baseURL+"/api/token/refresh/", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("sessao expirada, faca login novamente")
	}
	var result struct {
		Access string `json:"access"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	c.AccessToken = result.Access
	if c.OnRefresh != nil {
		c.OnRefresh(c.AccessToken, c.RefreshToken)
	}
	return nil
}

func (c *Client) get(path string, out interface{}) error {
	req, err := http.NewRequest("GET", baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("sem conexao com o SUAP")
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		if err := c.refreshTokens(); err != nil {
			return err
		}
		return c.get(path, out)
	}
	if resp.StatusCode == 404 {
		return fmt.Errorf("endpoint nao encontrado: %s", path)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("erro %d em %s", resp.StatusCode, path)
	}

	data, _ := io.ReadAll(resp.Body)
	return json.Unmarshal(data, out)
}

func (c *Client) GetAcademicData() (*AcademicData, error) {
	var a AcademicData
	err := c.get("/api/ensino/meus-dados-aluno/", &a)
	return &a, err
}

func (c *Client) GetPeriods() ([]Period, error) {
	var paginated PaginatedPeriods
	err := c.get("/api/ensino/meus-periodos-letivos/", &paginated)
	if err != nil || len(paginated.Results) == 0 {
		// fallback
		var p2 PaginatedPeriods
		if err2 := c.get("/api/ensino/periodos/", &p2); err2 == nil && len(p2.Results) > 0 {
			return p2.Results, nil
		}
	}
	return paginated.Results, err
}

func (c *Client) GetDiaries(semestre string) ([]Diary, error) {
	var paginated PaginatedDiaries
	err := c.get("/api/ensino/diarios/"+semestre+"/", &paginated)
	return paginated.Results, err
}

func (c *Client) GetCompletionReqs() (*CompletionReqs, error) {
	var cr CompletionReqs
	err := c.get("/api/ensino/requisitos-conclusao/", &cr)
	return &cr, err
}

func (c *Client) GetUnreadMessages() (*MessagesResponse, error) {
	var msgs MessagesResponse
	err := c.get("/api/edu/mensagens/entrada/nao_lidas/?page=1", &msgs)
	return &msgs, err
}

// LatestSemester returns the most recent semestre string like "2024.1"
func LatestSemester(periods []Period) string {
	if len(periods) == 0 {
		return ""
	}
	latest := periods[0]
	for _, p := range periods[1:] {
		if p.AnoLetivo > latest.AnoLetivo || (p.AnoLetivo == latest.AnoLetivo && p.PeriodoLetivo > latest.PeriodoLetivo) {
			latest = p
		}
	}
	return fmt.Sprintf("%d.%d", latest.AnoLetivo, latest.PeriodoLetivo)
}
