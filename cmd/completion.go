package main

const zshCompletion = `#compdef petowner

_petowner()
{
    local command port pet word
    local -a values
    local -i command_index i

    for ((i = 2; i < CURRENT; i++)); do
        word=$words[i]
        if [[ $word == -p && $((i + 1)) -lt $CURRENT ]]; then
            port=$words[i+1]
            ((i++))
        elif [[ $word != -* && -z $command ]]; then
            command=$word
            command_index=$i
        fi
    done

    if [[ $words[CURRENT-1] == -p ]]; then
        _message 'MCP server port'
        return
    fi

    if [[ -z $command ]]; then
        compadd -- -p -v animation pet pets show hide vet completion
        return
    fi

    case $command in
        completion)
            compadd -- zsh
            ;;
        animation)
            if [[ -n $port ]]; then
                values=("${(@f)$("${words[1]}" -p "$port" pets 2>/dev/null |
                    awk '/^  / {print $1}' | sort -u)}")
                compadd -- $values
            fi
            ;;
        pet)
            if [[ -n $port && $CURRENT -eq $((command_index + 1)) ]]; then
                values=("${(@f)$("${words[1]}" -p "$port" pets 2>/dev/null |
                    awk '/^[^ ]/ {print $1}')}")
                compadd -- $values
            elif [[ -n $port && $CURRENT -eq $((command_index + 2)) ]]; then
                pet=$words[command_index+1]
                values=("${(@f)$("${words[1]}" -p "$port" pets 2>/dev/null |
                    awk -v pet="$pet" '/^[^ ]/ {active=($1==pet); next} active && /^  / {print $1}')}")
                compadd -- $values
            fi
            ;;
    esac
}

compdef _petowner petowner
`
