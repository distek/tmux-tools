package cmd

import (
	"log"
	"os"
	"strings"

	"github.com/distek/tmux-tools/lib"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "cleanup unattached sessions",
	Run: func(cmd *cobra.Command, args []string) {
		initGlobalArgs()

		o, e, err := lib.Tmux(lib.GlobalArgs, "ls", nil, "")
		if err != nil {
			log.Println(e)
			log.Fatal(err)
		}

		for _, v := range strings.Split(o, "\n") {
			if v == "" {
				continue
			}

			if !strings.Contains(v, "(attached)") {
				id := strings.Split(v, ":")[0]

				_, e, err := lib.Tmux(lib.GlobalArgs, "kill-window", map[string]string{"-t": id}, "")
				if err != nil {
					log.Println(e)
					log.Fatal(err)
					return
				}
			}
		}

	},
}

var focusPaneCmd = &cobra.Command{
	Use:       "focus-pane {left | bottom | top | bottom}",
	Short:     "Focus pane in a given direction (left, bottom, top, right) (doesn't wrap)",
	ValidArgs: []string{"left", "bottom", "top", "right"},
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		dir := args[0]

		switch dir {
		case "left", "right", "top", "bottom":
		default:
			log.Printf("Provide only one of %v", cmd.ValidArgs)
			_ = cmd.Usage()
			os.Exit(1)
		}

		p, err := lib.GetCurrentPane()
		if err != nil {
			log.Fatal(err)
		}

		neighbors, err := lib.GetNeighbors(p)
		if err != nil {
			log.Fatal(err)
		}

		if !neighbors.Panes[dir].Exists {
			log.Fatalf("no pane in dir %s", dir)
		}

		err = lib.FocusPane(neighbors.Panes[dir].Pane)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
	rootCmd.AddCommand(focusPaneCmd)
}
