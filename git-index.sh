#!/bin/bash

# ZSH completion
_git_index_zsh_complete() {
    local -a completions
    completions=("${(@f)$(git-index complete)}")

    # Use native zsh array filtering to split flags (--*) and repos
    local -a flags=("${(@M)completions:#--*}")
    local -a repos=("${(@)completions:#--*}")

    # Add flags into a sorted 'options' group with a header
    if (( ${#flags[@]} )); then
        compadd -J 'options' -X 'options' -a flags
    fi

    # Add repositories into an unsorted 'repositories' group with a header
    if (( ${#repos[@]} )); then
        compadd -V 'repositories' -X 'repositories' -a repos
    fi
}

if [[ -n "$ZSH_VERSION" ]]; then
    # Unbind any old scm_breeze compctl/compdef just in case
    compdef -d c 2>/dev/null || true
    compdef _git_index_zsh_complete c
    
    # Bypass fzf-tab and Zsh's aggressive alphabetical sorting overrides
    zstyle ':completion:*:c:*' sort false
    zstyle ':completion:*:*:c:*' sort false
    zstyle ':completion:*:c:*' matcher-list ''
    zstyle ':fzf-tab:complete:c:*' fzf-flags '--no-sort' '--reverse'
fi

# Bash completion
_git_index_bash_complete() {
    local cur=${COMP_WORDS[COMP_CWORD]}
    local IFS=$'\n'
    local completions=($(git-index complete))
    COMPREPLY=( $(compgen -W "${completions[*]}" -- "$cur") )
}

if [[ -n "$BASH_VERSION" ]]; then
    # -o nosort prevents Bash 4.4+ from sorting our frecency ranking
    complete -o nosort -F _git_index_bash_complete c 2>/dev/null || complete -F _git_index_bash_complete c
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
        # Track frecency in background, suppressing job control messages
        (git-index track "$target_dir" >/dev/null 2>&1 &)
        cd -- "$target_dir"
    fi
}
