package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/distek/tmux-tools/lib"
	"github.com/spf13/cobra"
)

var wmCmd = &cobra.Command{
	Use:   "wm",
	Short: "Window manager",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Usage()
	},
}

func mergeMaps(dst, src map[string]string) map[string]string {
	for k, v := range src {
		dst[k] = v
	}

	return dst
}

func splitFull(pane lib.Pane, dir string) {
	targetPane, err := lib.GetFurthestPaneInDir(dir)
	if err != nil {
		log.Println(err)
		targetPane, err = lib.GetFurthestPaneInDir("top-" + dir)
		if err != nil {
			log.Fatal(err)
		}
	}

	defaultArgs := map[string]string{
		"-f": "",
		"-t": targetPane.ID,
	}
	extraArgs := make(map[string]string, 2)

	switch dir {
	case "up":
		extraArgs["-b"] = ""
		extraArgs["-v"] = ""
	case "down":
		extraArgs["-v"] = ""
	case "left":
		extraArgs["-b"] = ""
		extraArgs["-h"] = ""
	case "right":
		extraArgs["-h"] = ""
	}

	merged := mergeMaps(defaultArgs, extraArgs)

	o, e, err := lib.Tmux(lib.GlobalArgs, "split-window", merged, "cat")
	if err != nil {
		log.Fatal(fmt.Errorf("cmd: moveWindowInDir: lib.Tmux: up: command failed: err=%s, stdout=%s, err=%s", err, o, e))
	}

	newPane, err := lib.GetCurrentPane()
	if err != nil {
		log.Fatal(err)
	}

	if newPane.ID == pane.ID {
		return
	}

	err = lib.SwapPanes(pane, newPane)
	if err != nil {
		log.Fatal(err)
	}

	err = lib.KillPane(newPane)
	if err != nil {
		log.Fatal(err)
	}
}

func splitHalf(dst, src lib.Pane, dir string) {
	defaultArgs := map[string]string{
		"-t": dst.ID,
		"-s": src.ID,
	}

	extraArgs := make(map[string]string, 2)
	switch dir {
	case "up":
		extraArgs["-b"] = ""
		extraArgs["-v"] = ""
	case "down":
		extraArgs["-v"] = ""
	case "left":
		extraArgs["-b"] = ""
		extraArgs["-h"] = ""
	case "right":
		extraArgs["-h"] = ""
	}

	merged := mergeMaps(defaultArgs, extraArgs)

	o, e, err := lib.Tmux(lib.GlobalArgs, "join-pane", merged, "")
	if err != nil {
		log.Fatal(fmt.Errorf("cmd: moveWindowInDir: lib.Tmux: %s: command failed: err=%s, stdout=%s, err=%s", dir, err, o, e))
	}
}

func moveWindowInDir(dir string) {
	panesLen := lib.GetPanesLen()

	// If there's only one pane, or it errored, return with int
	if panesLen <= 0 {
		os.Exit(panesLen)
	}

	currPane, err := lib.GetCurrentPane()
	if err != nil {
		log.Fatalf("GetCurrentPane: %s", err)
	}

	neighbors, err := lib.GetNeighbors(currPane)
	if err != nil {
		log.Fatalf("GetNeighbors: %s", err)
	}

	// favor splitting to the right or bottom when merging panes
	switch dir {
	case "up":
		if neighbors.Panes["up"].Exists {
			splitHalf(neighbors.Panes["up"].Pane, currPane, "right")
		} else {
			if panesLen == 2 || neighbors.Panes["left"].Exists || neighbors.Panes["right"].Exists {
				splitFull(currPane, "up")
			}
		}
	case "down":
		if neighbors.Panes["down"].Exists {
			splitHalf(neighbors.Panes["down"].Pane, currPane, "right")
		} else {
			if panesLen == 2 || neighbors.Panes["left"].Exists || neighbors.Panes["right"].Exists {
				splitFull(currPane, "down")
			}
		}
	case "left":
		if neighbors.Panes["left"].Exists {
			splitHalf(neighbors.Panes["left"].Pane, currPane, "down")
		} else {
			if panesLen == 2 || neighbors.Panes["up"].Exists || neighbors.Panes["down"].Exists {
				splitFull(currPane, "left")
			}
		}
	case "right":
		if neighbors.Panes["right"].Exists {
			splitHalf(neighbors.Panes["right"].Pane, currPane, "down")
		} else {
			if panesLen == 2 || neighbors.Panes["up"].Exists || neighbors.Panes["down"].Exists {
				splitFull(currPane, "right")
			}
		}
	}

	_ = lib.SelectPane(currPane)

	lib.UsePaneCache = false

	// TODO: checking neighbors is expensive and this doesn't work.
	// if allColumns() && !allRows() {
	// 	_, _, _ = lib.Tmux(lib.GlobalArgs, "select-layout", map[string]string{
	// 		"even-vertical": "",
	// 	}, "")
	// } else if allRows() && !allColumns() {
	// 	_, _, _ = lib.Tmux(lib.GlobalArgs, "select-layout", map[string]string{
	// 		"even-horizontal": "",
	// 	}, "")
	// }
}

var wmLeftCmd = &cobra.Command{
	Use:   "left",
	Short: "Move window left",
	Run: func(cmd *cobra.Command, args []string) {
		moveWindowInDir("left")
	},
}

var wmDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Move window down",
	Run: func(cmd *cobra.Command, args []string) {
		moveWindowInDir("down")
	},
}

var wmUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Move window up",
	Run: func(cmd *cobra.Command, args []string) {
		moveWindowInDir("up")
	},
}

var wmRightCmd = &cobra.Command{
	Use:   "right",
	Short: "Move window right",
	Run: func(cmd *cobra.Command, args []string) {
		moveWindowInDir("right")
	},
}

func init() {
	rootCmd.AddCommand(wmCmd)
	wmCmd.AddCommand(wmLeftCmd)
	wmCmd.AddCommand(wmDownCmd)
	wmCmd.AddCommand(wmUpCmd)
	wmCmd.AddCommand(wmRightCmd)

	lib.UsePaneCache = false
}
