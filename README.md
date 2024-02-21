# tmux-tools

Some tools/wrappers around tmux to make my life easier

## Commands

### `wm`

Pane movement. Not perfect but does it well enough for my use. Will probably improve upon it at some point.

`tmux-tools wm [up|down|left|right]`

#### TODO

- [ ] It would be nice if neighbors swapped instead of moved I think? Or an option to swap? idk

### `nest`

Nested tmux session with auto-close. I use this in neovim config for my toggle terminal so I can have multiple term tabs.

`tmux-tools nest --tmux-config <path to nested.conf>`

In my nested config, I have special remaps setup so it doesn't interfere with the parent tmux mappings

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

- asks for name of session
  `tmux-tools sessions save`

- save session as provided name
  `tmux-tools sessions save --name <name>`

- load session by provided name
  `tmux-tools sessions load --name <name>`

- load session by `fzf` (requires `fzf` binary be in PATH)
  `tmux-tools sessions load`

#### TODO

- [ ] Should kill new session if restoring fails

## TODO

- [ ] Standardize logging (style/formatting) across the tool
