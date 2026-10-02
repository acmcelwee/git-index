#!/bin/bash

# ZSH completion
_git_index_zsh_complete() {
    local -a completions
    completions=("${(@f)$(git-index complete)}")
    _describe 'repositories' completions
}

if [[ -n "$ZSH_VERSION" ]]; then
    # Unbind any old scm_breeze compctl/compdef just in case
    compdef -d c 2>/dev/null || true
    compdef _git_index_zsh_complete c
fi

# Bash completion
_git_index_bash_complete() {
    local cur=${COMP_WORDS[COMP_CWORD]}
    local completions=$(git-index complete)
    COMPREPLY=( $(compgen -W "$completions" -- $cur) )
}

if [[ -n "$BASH_VERSION" ]]; then
    complete -F _git_index_bash_complete c
fi

# Remove any existing alias for c to ensure our function is called
unalias c 2>/dev/null || true

# The main wrapper function
c() {
    if [ "$1" = "--rebuild" ]; then
        git-index index
        return
    elif [ "$1" = "--add" ]; then
        git-index add "${2:-.}"
        return
    elif [ "$1" = "--add-root" ]; then
        git-index config add-root "${2:-.}" --depth="${3:-3}"
        return
    elif [ "$1" = "--list" ] || [ "$1" = "-l" ]; then
        git-index list
        return
    fi
    
    local target_dirs=$(git-index match "$1")
    if [ -z "$target_dirs" ]; then
        echo "No matching repository found for '$1'"
        return 1
    fi
    
    local count=$(echo "$target_dirs" | wc -l | tr -d ' ')
    local target_dir=""
    
    if [ "$count" -gt 1 ]; then
        if command -v fzf >/dev/null; then
            target_dir=$(echo "$target_dirs" | fzf --height 40% --reverse)
            # If user pressed Escape in fzf, target_dir will be empty
            if [ -z "$target_dir" ]; then
                return 0
            fi
        else
            echo "Multiple matches found. Install 'fzf' for interactive selection:"
            echo "$target_dirs"
            return 1
        fi
    else
        target_dir="$target_dirs"
    fi

    if [ -n "$target_dir" ]; then
        cd "$target_dir"
    fi
}
