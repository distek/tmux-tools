# tmux-tools

Some tools/wrappers around tmux to make my life easier

## Commands

### `wm`

Pane movement. Not perfect but does it well enough for my use. Will probably improve upon it at some point.

`tmux-tools wm [top|bottom|left|right]`

TODO:

- [ ] It would be nice if neighbors swapped instead of moved I think? Or an option to swap? idk

---

### `nest`

Nested tmux session with auto-close. I use this in neovim config for my toggle terminal so I can have multiple term tabs.

`tmux-tools nest --tmux-config <path to nested.conf>`

In my nested config, I have special remaps setup so it doesn't interfere with the parent tmux mappings:

- Normal config:

```
source-file ~/.config/tmux/common.conf

unbind C-b
set -g prefix M-Space;
```

- Nested config:

```
source-file ~/.config/tmux/common.conf

unbind C-b
set -g prefix M-a;
```

---

### `sessions`

Save/restore tmux sessions including commands

Uses config file (default: ~/.config/tmux/tools/config.yaml):

```yaml
sessions:
  # Where to save sessions
  sessions_path: "~/.config/tmux/sessions"
  # Which commands to "restore" (Runs the command in the pane it was in when saved)
  restore_cmds:
    - "vim"
    - "nvim"
    - "less"
    - "tail"
    - "man"
```

Asks for name of session:

`tmux-tools sessions save`

Save session as provided name:

`tmux-tools sessions save --name <name>`

Load session by provided name:

`tmux-tools sessions load --name <name>`

Load session by `fzf` (requires `fzf` binary be in PATH):

`tmux-tools sessions load`

TODO:

- [ ] Should kill new session if restoring fails

---

### `clean`

Kills all non-`(attached)` sessions on `-S` socket

---

### `focus-pane`

Focus pane in a given direction:

`tmux-tools focus-pane {left | bottom | top | right}`

### notes

Pane-directory-local notes popup window

```
# How I use it (top right corner of tmux window):
bind-key -n M-n run-shell "~/.local/bin/tmux-tools notes -Y S -X $(tmux display-message -p '#{e|+|:#{window_width},#{e|-|:0,#{popup_width}}}')"
```

In the event that the parent directory has a `worktrees` dir, it will use the parent's path to keep track of the notes instead.

The notes are saved to `$HOME/.local/share/tmux-tools/notes/`

So, for example, if I'm in the tmux-tools project directory, it would save to `$HOME/.local/share/tmux-tools/notes/\#path\#to\#tmux-tools.md`

## TODO

- [ ] Standardize logging (style/formatting) across the tool
