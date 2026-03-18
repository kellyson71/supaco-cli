# Completions for supaco
# https://github.com/kellyson71/supaco-cli

complete -c supaco -f

complete -c supaco -n '__fish_use_subcommand' -a hoje      -d 'Aulas de hoje'
complete -c supaco -n '__fish_use_subcommand' -a semana    -d 'Grade da semana'
complete -c supaco -n '__fish_use_subcommand' -a faltas    -d 'Frequencia e limite de faltas'
complete -c supaco -n '__fish_use_subcommand' -a notas     -d 'Notas e medias do semestre'
complete -c supaco -n '__fish_use_subcommand' -a status    -d 'Resumo rapido (IRA, hoje, alertas)'
complete -c supaco -n '__fish_use_subcommand' -a perfil    -d 'Perfil academico e progresso do curso'
complete -c supaco -n '__fish_use_subcommand' -a msgs      -d 'Mensagens nao lidas'
complete -c supaco -n '__fish_use_subcommand' -a help      -d 'Mostrar ajuda'
