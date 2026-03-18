# VHS tape — requires: ttyd + ffmpeg
# Install ttyd: https://github.com/tsl0922/ttyd
# Run: vhs demo.tape

Output assets/demo.gif
Output assets/demo.mp4

Set Theme "Catppuccin Mocha"
Set FontSize 14
Set Width 1200
Set Height 700
Set Padding 20
Set Framerate 30
Set PlaybackSpeed 1.0

# Abre o supaco
Type "supaco"
Enter
Sleep 2s

# Aguarda login aparecer e digita credenciais
Sleep 1s
Type "20251094040030"
Tab
Type "SuaSenhaAqui"
Enter

# Aguarda carregar
Sleep 4s

# Na tela de boas-vindas
Enter
Sleep 1s

# Navega no menu: Aulas de Hoje
Enter
Sleep 2s

# Volta ao menu
Escape
Sleep 500ms

# Navega para Grade da Semana
Down
Enter
Sleep 2s

Escape
Sleep 500ms

# Navega para Frequência
Down
Enter
Sleep 2s

Escape
Sleep 500ms

# Navega para Notas
Down
Enter
Sleep 2s

Escape
Sleep 500ms

# Sai
Type "q"
Sleep 500ms
