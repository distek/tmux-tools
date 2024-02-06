# tmux-tools

Some tools/wrappers around tmux to make my life easier

## Commands

### wm

Pane movement. Not perfect but does it well enough for my use. Will probably improve upon it at some point.

`tmux-tools wm [up|down|left|right]`

### nest

Nested tmux session with auto-close. I use this in neovim config for my toggle terminal so I can have multiple term tabs.

`tmux-tools nest --tmux-config <path to nested.conf>`

In my nested config, I have special remaps setup so it doesn't interfere with the parent tmux mappings
