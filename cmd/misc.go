package cmd

import (
	"fmt"
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
	Use:       "focus-pane {left | down | up | right}",
	Short:     "Focus pane in a given direction (left, down, up, right)",
	ValidArgs: []string{"left", "down", "up", "right"},
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		dir := args[0]

		switch dir {
		case "left", "down", "up", "right":
		default:
			log.Printf("Provide only one of %v", cmd.ValidArgs)
			_ = cmd.Usage()
			os.Exit(1)
		}

		p, err := lib.GetCurrentPane()
		if err != nil {
			log.Fatal(err)
		}

		dstP, exists, err := lib.GetPaneInDir(p, dir)
		if err != nil {
			log.Fatal(err)
		}

		if !exists {
			var vimDir string
			switch dir {
			case "left":
				vimDir = "h"
			case "down":
				vimDir = "j"
			case "up":
				vimDir = "k"
			case "right":
				vimDir = "l"
			}

			_, _, err := lib.Tmux(lib.GlobalArgs, "send-keys", map[string]string{
				fmt.Sprintf("M-%s", vimDir): "",
			}, "")
			if err != nil {
				log.Fatal(err)
			}

			return
		}

		err = lib.FocusPane(dstP)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
	rootCmd.AddCommand(focusPaneCmd)
}
