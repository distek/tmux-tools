package lib

// GlobalArgs are the collected global args specified by the user such as -S, -L, -D, etc.
var (
	GlobalArgs map[string]string

	// Pane caching so we don't have to run through all of the panes each time GetPanes is called
	PaneCache []Pane

	// Set to true to force new PaneCache
	UsePaneCache bool
)
