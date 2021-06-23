function showsessions {
    ssh2 get session --format ".tag"
}

function ssh2_quick_login_command {
    typeset ssh_tag="$1"
    command=$(ssh2 quick-login-command $ssh_tag)
    echo $command
    return 0
}

function ssh2_verify_go2s_ssh_tag {
    typeset ssh_tag="$1"
    if [ ! ssh2_quick_login_command $ssh_tag ]
    then
        echo "Error: Session not found with tag<'$ssh_tag'>"
        return 1
    fi
    return 0
}

function go2s {
    typeset -a in_args
    typeset -a out_args

    in_args=( "$@" )

    if [ -n "$ZSH_VERSION" ]
    then
        i=1
        tst="-le"
    else
        i=0
        tst="-lt"
    fi


    typeset ssh_tag="$1"

    if [ "$ssh_tag" = "" ]
    then
        showsessions
        return 1
    fi
    ssh2_verify_go2s_ssh_tag "ssh_tag" || return 1
    command=$(ssh2_quick_login_command $ssh_tag)
    eval $command

}

function ssh2_setup_tab_completion {
    if [ -n "${BASH:-}" ] ; then

    elif [ -n "$ZSH_VERSION" ] ; then
        _show_sessions () {
            reply=( $(showsessions) )
        }

        compctl -K _show_sessions go2s
    fi
}

ssh2_setup_tab_completion
