function showsessions {
    ssh2 get --kind Session --template "{{ .Tag }}"
}

function ssh2_verify_go2s_ssh_tag {
    typeset ssh_tag="${1:-''}"
    test=`showsessions | grep "^${ssh_tag}$"`
    if [ "${test}" = "" ]
    then
        echo "Error: Session not found with tag<'$ssh_tag'>"
        return 1
    fi
    return 0
}

function go2s {
    typeset direct=""
    if [ "$1" = "-d" ] || [ "$1" = "--direct" ]
    then
        direct="--direct"
        shift
    fi

    typeset ssh_tag="$1"
    if [ "$ssh_tag" = "" ]
    then
        showsessions
        return 1
    fi
    ssh2_verify_go2s_ssh_tag "${ssh_tag}" || return 1

    ssh2 login ${direct} "${ssh_tag}"
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
