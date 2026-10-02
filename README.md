# git-index (c)

A blazingly fast git repository jumper written in Go. A modern, customizable replacement for the `git_index` tool from `scm_breeze`.

## Features
- **Fast Traversal**: Scans your repositories quickly using Go concurrency.
- **Scattered Repositories**: Supports multiple recursive search roots and one-off repository paths anywhere on your disk.
- **Smart Matching**: Jump to a repo by exact name, prefix, substring, or path.
- **Interactive Disambiguation**: If a search matches multiple repositories, `git-index` instantly pops open an interactive `fzf` menu so you can pick the exact one.
- **CLI Config Management**: Add new search roots or one-off directories dynamically from the terminal.

## Requirements
- **[fzf](https://github.com/junegunn/fzf)**: Required for the interactive disambiguation menu. (`brew install fzf`)

## Installation

1. Compile the binary using `just`:
   ```bash
   just build
   ```
2. Move it to your PATH (e.g. `~/bin/` or `~/.local/bin/`):
   ```bash
   cp git-index ~/bin/
   ```
3. Source the shell hooks in your `~/.zshrc` or `~/.bashrc`:
   ```bash
   # Add this to your shell config
   source /path/to/this/repo/git-index.sh
   ```

*(Optional but highly recommended)*: To make `fzf` pop up when you hit `<tab>` (instead of just when you hit `Enter`), install the [fzf-tab](https://github.com/Aloxaf/fzf-tab) plugin for Zsh.

## Usage

Use the `c` command to jump:

```bash
# Jump to a repo (e.g. jump to ~/code/billing-api)
c billing

# If "billing" matches multiple repos, you'll be dropped into an fzf menu to pick!

# Add the current directory as a one-off repo
c --add

# Add a specific directory as a one-off repo
c --add /etc/nixos

# Add a new recursive search root with a max depth of 4
c --add-root ~/work/clients 4

# List all repositories currently in the index
c --list

# Rebuild the index manually (happens automatically when adding paths)
c --rebuild
```

### Configuration
Your configuration is stored at `$XDG_CONFIG_HOME/git-index/config.toml` (which defaults to `~/.config/git-index/config.toml`). 
The cache is stored at `$XDG_CACHE_HOME/git-index/index.txt` (which defaults to `~/.cache/git-index/index.txt`).

**Example `config.toml`**:
```toml
# Glob patterns to ignore during traversal
ignore = ['**/archive/**', '**/node_modules/**', '**/.terraform/**']

# A recursive search root
[[search_dirs]]
path = '/Users/username/code'
max_depth = 4

# A one-off repository (max_depth 0)
[[search_dirs]]
path = '/etc/nixos'
max_depth = 0
```
