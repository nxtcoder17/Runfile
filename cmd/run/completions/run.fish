# run fish shell completion
set PROGNAME run

# list_targets fetches all targets, with all flags provided on the cli
function list_targets
  eval $PROGNAME --list
  # eval (commandline -b) --list
end

complete -c $PROGNAME -d "runs named task" -xa '(list_targets)'
complete -c $PROGNAME -l help -s h -d 'show help'
complete -c $PROGNAME -l list -s l -d 'list all tasks'
complete -c $PROGNAME -rF -l file -s f -d 'runs targets from this runfile' 

