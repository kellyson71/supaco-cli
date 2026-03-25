#!/usr/bin/env bash
set -e

# ─────────────────────────────────────────────
#  supaco-cli — instalador rapido
#  https://github.com/kellyson71/supaco-cli
# ─────────────────────────────────────────────

REPO="kellyson71/supaco-cli"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
BINARY="supaco"

# cores
RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
BOLD='\033[1m'
DIM='\033[2m'
NC='\033[0m'

info()    { echo -e "  ${CYAN}>${NC} $*"; }
success() { echo -e "  ${GREEN}✓${NC} $*"; }
warn()    { echo -e "  ${YELLOW}!${NC} $*"; }
error()   { echo -e "  ${RED}✗${NC} $*" >&2; exit 1; }
step()    { echo -e "\n${BOLD}$*${NC}"; }

echo -e ""
echo -e "${BOLD}  ____  _   _ ____   _    ____ ___${NC}"
echo -e "${BOLD} / ___|| | | |  _ \\ / \\  / ___/ _ \\${NC}"
echo -e "${BOLD} \\___ \\| | | | |_) / _ \\| |  | | | |${NC}"
echo -e "${BOLD}  ___) | |_| |  __/ ___ \\ |__| |_| |${NC}"
echo -e "${BOLD} |____/ \\___/|_| /_/   \\_\\____\\___/${NC}"
echo -e "${DIM}  CLI do SUAP · IFRN${NC}"
echo -e ""

# ── Detecta OS e arch ───────────────────────────────────
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) warn "Arquitetura $ARCH nao testada — tentando compilar do zero." ;;
esac

# ── Tenta baixar binario pre-compilado ──────────────────
step "Buscando versao mais recente..."
LATEST=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": "\(.*\)".*/\1/')

if [ -z "$LATEST" ]; then
  warn "Nao foi possivel obter a versao mais recente. Tentando compilar do zero."
  LATEST="source"
fi

ASSET_NAME="${BINARY}_${OS}_${ARCH}"
ASSET_URL="https://github.com/${REPO}/releases/download/${LATEST}/${ASSET_NAME}"

TMPDIR_WORK="$(mktemp -d)"
trap 'rm -rf "$TMPDIR_WORK"' EXIT

download_binary() {
  info "Tentando baixar ${ASSET_NAME} (${LATEST})..."
  if curl -fsSL "$ASSET_URL" -o "$TMPDIR_WORK/$BINARY" 2>/dev/null; then
    chmod +x "$TMPDIR_WORK/$BINARY"
    success "Binario baixado."
    return 0
  fi
  return 1
}

build_from_source() {
  step "Compilando do codigo-fonte..."

  if ! command -v go &>/dev/null; then
    error "Go nao encontrado. Instale em https://go.dev/dl/ e tente novamente."
  fi

  GO_VERSION="$(go version | awk '{print $3}' | sed 's/go//')"
  info "Go ${GO_VERSION} encontrado."

  info "Clonando repositorio..."
  git clone --depth=1 "https://github.com/${REPO}.git" "$TMPDIR_WORK/src" -q

  info "Compilando..."
  cd "$TMPDIR_WORK/src"
  CGO_ENABLED=0 go build -trimpath -mod=vendor -ldflags="-s -w" -o "$TMPDIR_WORK/$BINARY" . 2>&1
  success "Compilado com sucesso."
}

# Tenta binario primeiro, fallback para source
if ! download_binary; then
  warn "Binario pre-compilado nao disponivel para ${OS}/${ARCH}."
  build_from_source
fi

# ── Instala ─────────────────────────────────────────────
step "Instalando..."
mkdir -p "$INSTALL_DIR"
mv "$TMPDIR_WORK/$BINARY" "$INSTALL_DIR/$BINARY"
success "Instalado em ${INSTALL_DIR}/${BINARY}"

# ── Instala completions fish ─────────────────────────────
FISH_COMPLETIONS_DIR="$HOME/.config/fish/completions"
SRC_COMPLETIONS=""

if command -v fish &>/dev/null && [ -d "$FISH_COMPLETIONS_DIR" ]; then
  # tenta pegar do clone se existir, senão baixa
  if [ -f "$TMPDIR_WORK/src/completions/${BINARY}.fish" ]; then
    SRC_COMPLETIONS="$TMPDIR_WORK/src/completions/${BINARY}.fish"
  else
    curl -fsSL "https://raw.githubusercontent.com/${REPO}/main/completions/${BINARY}.fish" \
      -o "$TMPDIR_WORK/${BINARY}.fish" 2>/dev/null && SRC_COMPLETIONS="$TMPDIR_WORK/${BINARY}.fish"
  fi
  if [ -n "$SRC_COMPLETIONS" ]; then
    cp "$SRC_COMPLETIONS" "$FISH_COMPLETIONS_DIR/${BINARY}.fish"
    success "Fish completions instalados."
  fi
fi

# ── Verifica PATH ────────────────────────────────────────
echo ""
if ! command -v "$BINARY" &>/dev/null; then
  warn "${INSTALL_DIR} nao esta no seu PATH."
  echo ""
  echo -e "  Adicione ao seu shell:"
  echo -e ""
  echo -e "  ${DIM}# bash / zsh${NC}"
  echo -e "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.bashrc"
  echo -e ""
  echo -e "  ${DIM}# fish${NC}"
  echo -e "  fish_add_path \$HOME/.local/bin"
  echo ""
else
  success "supaco esta disponivel no PATH."
fi

echo -e "  ${GREEN}${BOLD}Pronto! Execute:  supaco${NC}"
echo ""
