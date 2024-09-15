# run fish shell completion
set PROGNAME run

function __fetch_runnable_tasks --description 'fetches all runnable tasks'
    for i in (commandline -opc)
        if contains -- $i help h
            return 1
        end
    end

    # Grab names and descriptions (if any) of the tasks
    set -l output (run --generate-shell-completion | string split0)
    echo "$output" > /tmp/test.txt
    if test $output
      echo $output
    end

    return 0
end

complete -c run -d "runs a task with given name" -xa "(__fish_run_no_subcommand)"
complete -c run -n '__fish_run_no_subcommand' -f -l help -s h -d 'show help'
complete -c run -n '__fish_run_no_subcommand' -f -l help -s h -d 'show help'
complete -r -c run -n '__fish_run_no_subcommand' -a 'help h' -d 'Shows a list of commands or help for one command'

