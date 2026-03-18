#!/usr/bin/env bash
# Tira screenshot da janela do terminal com o supaco aberto
# Uso: ./scripts/screenshot.sh [nome]
# Exemplo: ./scripts/screenshot.sh login
#
# Ferramentas usadas (Wayland/Hyprland/Sway): grim + slurp
# Instalar: sudo pacman -S grim slurp

set -e

ASSETS_DIR="$(dirname "$0")/../assets"
mkdir -p "$ASSETS_DIR"

NAME="${1:-screenshot}"
OUT="$ASSETS_DIR/$NAME.png"

echo "Posicione o terminal com o supaco aberto."
echo "Selecione a area do terminal com o mouse..."
echo ""

# grim + slurp: seleciona região interativamente
grim -g "$(slurp)" "$OUT"

echo "Screenshot salvo em: $OUT"

# Otimiza a imagem se optipng estiver disponível
if command -v optipng &>/dev/null; then
  optipng -quiet "$OUT"
fi

echo "Pronto! Adicione ao README:"
echo "  ![${NAME}](assets/${NAME}.png)"
