package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/distek/tmux-tools/lib"
	"github.com/spf13/cobra"
)

var (
	flagNotesX string
	flagNotesY string
	flagNotesW string
	flagNotesH string
)

var notesCmd = &cobra.Command{
	Use:   "notes",
	Short: "spawn a notes window for the current directory",
	Long:  "Notes are saved into ~/.local/share/tmux-tools/notes",
	Run: func(cmd *cobra.Command, args []string) {
		initGlobalArgs()

		globalSock, globalSockSpecified := lib.GlobalArgs["-S"]

		p, err := lib.GetCurrentPane()
		if err != nil {
			log.Fatal(err)
		}

		path := p.Cwd

		if du, err := lib.DirUp(path); err == nil {
			if lib.IsGitWorktree(du) {
				path = du
			}
		}

		home := os.Getenv("HOME")

		notesPath := fmt.Sprintf("%s/.local/share/tmux-tools/notes", home)
		notesFile := fmt.Sprintf("%s/%s.md", notesPath, strings.ReplaceAll(path, "/", "#"))
		sockPath := fmt.Sprintf("/tmp/tmux-notes_%s", strings.ReplaceAll(path, "/", "#"))

		if _, err := os.Stat(notesPath); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				log.Fatal(err)
			}

			err := os.MkdirAll(notesPath, 0o750)
			if err != nil {
				log.Fatal(err)
			}
		}

		if !lib.SockExists(sockPath) {
			if lib.SockActive(sockPath) {
				if globalSockSpecified {
					lib.GlobalArgs["-S"] = globalSock
				}

				if lib.SockHasAttached(sockPath) {
					o, e, err := lib.Tmux(map[string]string{"-S": sockPath}, "detach", nil, "")
					if err != nil {
						fmt.Println(o)
						fmt.Println(e)
						log.Fatal(err)
					}

					return
				}

				o, e, err := lib.Tmux(lib.GlobalArgs, "popup", map[string]string{
					"-E": "",
					"-x": flagNotesX,
					"-y": flagNotesY,
					"-w": flagNotesW,
					"-h": flagNotesH,
				}, fmt.Sprintf("tmux -S %s a -t 0", sockPath))
				if err != nil {
					fmt.Println(o)
					fmt.Println(e)
					log.Fatal(err)
				}

				return
			}
		}

		o, e, err := lib.Tmux(map[string]string{"-S": sockPath, "-f": "/dev/null"}, "new", map[string]string{"-d": ""},
			fmt.Sprintf("nvim %s", notesFile),
		)
		if err != nil {
			fmt.Println(o)
			fmt.Println(e)
			log.Fatal(err)
		}
		_, _, err = lib.Tmux(map[string]string{"-S": sockPath}, "set", map[string]string{"-g": "status-keys vi"}, "")
		if err != nil {
			log.Fatal(err)
		}
		_, _, err = lib.Tmux(map[string]string{"-S": sockPath}, "set", map[string]string{"-g": "mode-keys vi"}, "")
		if err != nil {
			log.Fatal(err)
		}
		_, _, err = lib.Tmux(map[string]string{"-S": sockPath}, "set", map[string]string{"-g": "status off"}, "")
		if err != nil {
			log.Fatal(err)
		}
		_, _, err = lib.Tmux(map[string]string{"-S": sockPath}, "bind", map[string]string{"-n": "'M-n' detach"}, "")
		if err != nil {
			log.Fatal(err)
		}

		if globalSockSpecified {
			lib.GlobalArgs["-S"] = globalSock
		}

		o, e, err = lib.Tmux(lib.GlobalArgs, "popup", map[string]string{
			"-E": "",
			"-x": flagNotesX,
			"-y": flagNotesY,
			"-w": flagNotesW,
			"-h": flagNotesH,
		}, fmt.Sprintf("tmux -S %s a -t 0", sockPath))
		if err != nil {
			fmt.Println(o)
			fmt.Println(e)
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(notesCmd)

	notesCmd.Flags().StringVarP(&flagNotesX, "X", "X", "50%", "X pos of notes window")
	notesCmd.Flags().StringVarP(&flagNotesY, "Y", "Y", "3", "Y pos of notes window")
	notesCmd.Flags().StringVarP(&flagNotesW, "width", "W", "50%", "width of notes window")
	notesCmd.Flags().StringVarP(&flagNotesH, "height", "H", "50%", "height of notes window")
}
