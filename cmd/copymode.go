package cmd

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/distek/tmux-tools/lib"
	"github.com/spf13/cobra"
)

var (
	flagCopyNumbersShowLines  bool
	flagCopyNumbersTargetPane string
)

func getLineNumber(target string) (int, error) {
	o, e, err := lib.Tmux(lib.GlobalArgs, "display-message", map[string]string{
		"-t": target,
		"-p": "",
	}, "'#{copy_cursor_y}'")
	if err != nil {
		return -1, fmt.Errorf("%s: %s", err, e)
	}

	if o == "" {
		return -1, nil
	}

	i, err := strconv.Atoi(o)
	if err != nil {
		return -1, err
	}

	return i, nil
}

func relative(l, currentLine int) string {
	if l == currentLine {
		return fmt.Sprintf("\033[1;33m%d\033[0m", currentLine)
	}

	if l < currentLine {
		return fmt.Sprintf("\033[0;97m%d\033[0m", currentLine-l)
	}

	return fmt.Sprintf("\033[0;97m%d\033[0m", l-currentLine)
}

var copyNumbersCmd = &cobra.Command{
	Use:   "copy-numbers",
	Short: "Show relative line numbers next to pane in copy mode",
	Run: func(cmd *cobra.Command, args []string) {
		initGlobalArgs()

		if flagCopyNumbersShowLines {
			lastLine := 0
			for {
				p, err := lib.GetCurrentPane(flagCopyNumbersTargetPane)
				if err != nil {
					return
				}

				if p.CurrentMode != lib.PaneModeCopyMode {
					return
				}

				lineNumber, err := getLineNumber(flagCopyNumbersTargetPane)
				if lineNumber == -1 || err != nil {
					return
				}

				if lineNumber == lastLine {
					time.Sleep(time.Millisecond * 25)
					continue
				} else {
					lastLine = lineNumber
					fmt.Print("\033[H\033[2J")
				}

				for i := 0; i < p.Height; i++ {
					if i == p.Height-1 {
						fmt.Print(relative(i, lineNumber))

						break
					}

					fmt.Println(relative(i, lineNumber))
				}

				time.Sleep(time.Millisecond * 25)
			}
		}

		p, err := lib.GetCurrentPane("")
		if err != nil {
			os.Exit(1)
		}

		if p.CurrentMode != lib.PaneModeCopyMode {
			return
		}

		_, _, err = lib.Tmux(lib.GlobalArgs, "split-window", map[string]string{
			"-h": "",
			"-b": "",
			"-l": "2",
		}, fmt.Sprintf("%s copy-numbers -l -t %s", os.Args[0], p.ID))
		if err != nil {
			os.Exit(1)
		}

		_, _, err = lib.Tmux(lib.GlobalArgs, "last-pane", map[string]string{}, "")

		if err != nil {
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(copyNumbersCmd)

	copyNumbersCmd.Flags().BoolVarP(&flagCopyNumbersShowLines, "lines", "l", false, "Print the line numbers for the (-t) target pane")
	copyNumbersCmd.Flags().StringVarP(&flagCopyNumbersTargetPane, "target", "t", "", "Target pane to get line info from")

	copyNumbersCmd.MarkFlagsRequiredTogether("target", "lines")
}
