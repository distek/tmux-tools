package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/distek/tmux-tools/lib"
	"github.com/spf13/cobra"
)

var wmCmd = &cobra.Command{
	Use:       "wm {left | bottom | top | right}",
	Short:     "Window manager",
	ValidArgs: []string{"left", "bottom", "top", "right"},
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Printf("Provide only one of %v", cmd.ValidArgs)
			_ = cmd.Usage()
			os.Exit(1)
		}

		dir := args[0]

		moveWindowInDir(dir)
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
	case "top":
		extraArgs["-b"] = ""
		extraArgs["-v"] = ""
	case "bottom":
		extraArgs["-v"] = ""
	case "left":
		extraArgs["-b"] = ""
		extraArgs["-h"] = ""
	case "right":
		extraArgs["-h"] = ""
	}

	merged := mergeMaps(defaultArgs, extraArgs)

	_, e, err := lib.Tmux(lib.GlobalArgs, "split-window", merged, "cat")
	if err != nil {
		log.Println(e)
		log.Fatal(err)
	}

	newPane, err := lib.GetCurrentPane("")
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
	case "top":
		extraArgs["-b"] = ""
		extraArgs["-v"] = ""
	case "bottom":
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

	currPane, err := lib.GetCurrentPane("")
	if err != nil {
		log.Fatalf("GetCurrentPane: %s", err)
	}

	neighbors, err := lib.GetNeighbors(currPane)
	if err != nil {
		log.Fatalf("GetNeighbors: %s", err)
	}

	// favor splitting to the right or bottom when merging panes
	switch dir {
	case "top":
		if neighbors.Panes["top"].Exists {
			splitHalf(neighbors.Panes["top"].Pane, currPane, "right")
		} else {
			if panesLen == 2 || neighbors.Panes["left"].Exists || neighbors.Panes["right"].Exists {
				splitFull(currPane, "top")
			}
		}
	case "bottom":
		if neighbors.Panes["bottom"].Exists {
			splitHalf(neighbors.Panes["bottom"].Pane, currPane, "right")
		} else {
			if panesLen == 2 || neighbors.Panes["left"].Exists || neighbors.Panes["right"].Exists {
				splitFull(currPane, "bottom")
			}
		}
	case "left":
		if neighbors.Panes["left"].Exists {
			splitHalf(neighbors.Panes["left"].Pane, currPane, "bottom")
		} else {
			if panesLen == 2 || neighbors.Panes["top"].Exists || neighbors.Panes["bottom"].Exists {
				splitFull(currPane, "left")
			}
		}
	case "right":
		if neighbors.Panes["right"].Exists {
			splitHalf(neighbors.Panes["right"].Pane, currPane, "bottom")
		} else {
			if panesLen == 2 || neighbors.Panes["top"].Exists || neighbors.Panes["bottom"].Exists {
				splitFull(currPane, "right")
			}
		}
	}

	_ = lib.SelectPane(currPane)

	lib.UsePaneCache = false
}

func init() {
	rootCmd.AddCommand(wmCmd)

	initGlobalArgs()

	lib.UsePaneCache = false
}
