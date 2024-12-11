# run fish shell completion
set PROGNAME run

function __runfile_list_targets --description 'fetches all runnable tasks'
    for i in (commandline -opc)
        if contains -- $i help h
            return 1
        end
    end

    $PROGNAME --list 2>&1 | read -lz rawOutput

    # RETURN on non-zero exit code (in case of errors)
    if test $status -ne 0
      return
    end

    set -l output (echo $rawOutput | string split0)

    if test $output
      if test "$DEBUG" = "true"
        echo "$output" > /tmp/test.txt
      end
      echo $output
    end

    return 0
end

complete -c $PROGNAME -d "runs target with given name" -xa "(__runfile_list_targets)"
complete -c $PROGNAME -rF -l file -s f -d 'runs targets from this runfile' 
# complete -c $PROGNAME -n '__runfile_list_targets' -f -l help -s h -d 'show help'
# complete -c $PROGNAME -n '__runfile_list_targets' -f -l help -s h -d 'show help'
# complete -r -c $PROGNAME  -n '__runfile_list_targets' -a 'help h' -d 'Shows a list of commands or help for one command'

