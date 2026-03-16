# Completions for supaco
# https://github.com/kellyson71/supaco-cli

complete -c supaco -f

complete -c supaco -n '__fish_use_subcommand' -a hoje    -d 'Aulas de hoje'
complete -c supaco -n '__fish_use_subcommand' -a semana  -d 'Grade da semana'
complete -c supaco -n '__fish_use_subcommand' -a faltas  -d 'Frequência e limite de faltas'
complete -c supaco -n '__fish_use_subcommand' -a notas   -d 'Notas do semestre'
complete -c supaco -n '__fish_use_subcommand' -a status  -d 'Resumo rápido (IRA, hoje, alertas)'
complete -c supaco -n '__fish_use_subcommand' -a help    -d 'Mostrar ajuda'
