# # args: (1)
# # run fish shell completion
# set PROGNAME run
#
# # list_targets fetches all targets, with all flags provided on the cli
# function list_targets
#   # eval $PROGNAME --list
#   eval (commandline -b) --list
#   echo "shell:completion"
# end
#
# complete -c $PROGNAME -d "runs named task" -xa '(list_targets)'
# complete -c $PROGNAME -d "runs named task" -xa '(list_targets)'
# complete -c $PROGNAME -l help -s h -d 'show help'
# complete -c $PROGNAME -l list -s l -d 'list all tasks'
# complete -c $PROGNAME -rF -l file -s f -d 'runs targets from this runfile' 

# completion fish shell completion

function __fish_completion_no_subcommand --description 'Test if there has been any subcommand yet'
    for i in (commandline -opc)
        if contains -- $i
            return 1
        end
    end
    return 0
end

complete -c run -n '__fish_completion_no_subcommand' -f -l help -s h -d 'show help'
complete -c run -n '__fish_completion_no_subcommand' -f -l list -s l -d 'list all tasks'
complete -c run -n '__fish_completion_no_subcommand' -rF -l file -s f -d 'runs targets from this runfile'

