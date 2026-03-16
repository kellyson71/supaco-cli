# SUPACO CLI

**Terminal moderno para o SUAP do IFRN.**
Veja suas aulas, frequência, notas e notificações direto no terminal — sem abrir o navegador.

```
  ____  _   _ ____   _    ____ ___
 / ___|| | | |  _ \ / \  / ___/ _ \
 \___ \| | | | |_) / _ \| |  | | | |
  ___) | |_| |  __/ ___ \ |__| |_| |
 |____/ \___/|_| /_/   \_\____\___/
        CLI do SUAP · IFRN
```

> ⚠️ **Projeto não oficial.** Desenvolvido pela comunidade, sem vínculo com o IFRN.

---

## Screenshots

<!-- SCREENSHOTS: rode o app e tire prints com seu sistema (Print Screen / Spectacle no KDE, Flameshot, etc.)
     Depois salve em assets/ e substitua os links abaixo. -->

> 📸 *Screenshots em breve. Contribuições bem-vindas!*
>
> Para gerar um GIF automaticamente, instale o [ttyd](https://github.com/tsl0922/ttyd) e rode:
> ```bash
> vhs demo.tape
> ```
> O arquivo `demo.tape` já está incluso no repositório.

---

## Funcionalidades

| Tela | O que mostra |
|------|-------------|
| 📅 **Aulas de Hoje** | Horário, sala e professor das aulas do dia. Indicador de frequência em tempo real. |
| 📆 **Grade da Semana** | Visão semanal completa com todos os horários. |
| 📊 **Frequência e Faltas** | Quantas faltas você tem e quantas ainda pode dar por disciplina. |
| 📝 **Notas do Semestre** | Notas por etapa e médias parciais/finais. |
| 👤 **Meu Perfil** | Dados acadêmicos, IRA e progresso de conclusão do curso. |
| 🔔 **Notificações** | Mensagens não lidas enviadas pelo SUAP. |

---

## Instalação

### Pré-requisitos

- [Go 1.21+](https://go.dev/dl/)

### Via `go install`

```bash
go install github.com/kellyson71/supaco-cli@latest
```

### Compilando do fonte

```bash
git clone https://github.com/kellyson71/supaco-cli.git
cd supaco-cli
go build -o supaco .

# Instalar globalmente (Linux/macOS)
sudo mv supaco /usr/local/bin/
```

### Arch Linux (manual)

```bash
git clone https://github.com/kellyson71/supaco-cli.git
cd supaco-cli
go build -o supaco .
cp supaco ~/.local/bin/supaco   # ou /usr/local/bin/ com sudo
```

---

## Uso

```bash
supaco
```

Na primeira execução, uma tela de login aparece. Digite sua **matrícula** e **senha** do SUAP.
Os tokens são salvos em `~/.config/supaco/config.json` (modo `600`) — você não precisará fazer login novamente até o token expirar.

### Navegação

| Tecla | Ação |
|-------|------|
| `↑` / `k` | Subir no menu |
| `↓` / `j` | Descer no menu |
| `Enter` | Selecionar / Confirmar |
| `Esc` / `q` | Voltar / Fechar |
| `r` | Atualizar dados do SUAP |
| `Ctrl+C` | Sair |

---

## Como funciona

O SUPACO usa a API JWT do SUAP:

1. **Login** → `POST /api/token/pair` com matrícula e senha
2. **Token refresh** → automático ao receber `401 Unauthorized`
3. **Dados** carregados em sequência:
   - `/api/ensino/meus-dados-aluno/` — perfil e IRA
   - `/api/ensino/meus-periodos-letivos/` — semestres cursados
   - `/api/ensino/diarios/{semestre}/` — aulas, faltas e notas
   - `/api/ensino/requisitos-conclusao/` — progresso do curso
   - `/api/edu/mensagens/entrada/nao_lidas/` — notificações

---

## Gerando screenshots / GIF

Para gerar o GIF de demonstração incluso no README:

```bash
# 1. Instale o ttyd (necessário pelo VHS)
# Arch: yay -S ttyd
# Ubuntu: snap install ttyd --classic
# macOS: brew install tsl0922/tools/ttyd

# 2. Instale o VHS
go install github.com/charmbracelet/vhs@latest

# 3. Rode o tape
vhs demo.tape
# Gera assets/demo.gif e assets/demo.mp4
```

Para screenshots manuais no Linux:

```bash
# Script incluso (usa grim + slurp no Wayland)
./scripts/screenshot.sh login      # salva assets/login.png
./scripts/screenshot.sh menu       # salva assets/menu.png
./scripts/screenshot.sh hoje       # salva assets/hoje.png

# Ou manualmente:
# Wayland (Hyprland/Sway/KDE Wayland)
grim -g "$(slurp)" assets/screenshot.png

# Flameshot
flameshot gui -p assets/

# GNOME
gnome-screenshot -a -f assets/screenshot.png

# KDE Spectacle
spectacle -r -b -o assets/screenshot.png
```

---

## Estrutura do projeto

```
supaco-cli/
├── main.go                    # Entry point
├── demo.tape                  # VHS tape para gerar GIF
├── assets/                    # Screenshots e GIFs
└── internal/
    ├── api/
    │   ├── client.go          # HTTP client + auth
    │   └── models.go          # Structs da API do SUAP
    ├── config/
    │   └── config.go          # Persistência de tokens
    └── ui/
        ├── app.go             # App principal (Bubble Tea)
        ├── login.go           # Tela de login
        ├── views.go           # Telas de conteúdo
        └── styles.go          # Estilos (Lip Gloss)
```

---

## Stack

- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** — framework TUI (Elm architecture)
- **[Lip Gloss](https://github.com/charmbracelet/lipgloss)** — estilização de terminal
- **[Bubbles](https://github.com/charmbracelet/bubbles)** — componentes (textinput, viewport, spinner)

---

## Contribuindo

PRs são bem-vindos! Algumas ideias:

- [ ] Histórico de notas (semestres anteriores)
- [ ] Notificações de alertas de frequência via sistema
- [ ] Modo offline com cache persistente
- [ ] Suporte a outros campi / instâncias do SUAP

```bash
git clone https://github.com/kellyson71/supaco-cli.git
cd supaco-cli
go run .
```

---

## Aviso Legal

Este projeto **não é oficial** e não tem qualquer vínculo com o IFRN ou com os desenvolvedores do SUAP.
Usa apenas endpoints públicos da API do SUAP que são acessíveis com credenciais válidas de aluno.
Não armazena senhas — apenas os tokens JWT gerados após login.

---

<p align="center">
  Feito com ♥ por alunos do IFRN
</p>
