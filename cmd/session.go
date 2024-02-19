package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/adrg/xdg"
	"github.com/distek/tmux-tools/lib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type SessPane struct {
	Current bool   `json:"current"`
	Index   int    `json:"index"`
	Path    string `json:"path"`
	Command string `json:"command"`
}

type SessWin struct {
	Current bool       `json:"current"`
	Index   int        `json:"index"`
	Layout  string     `json:"layout"`
	Name    string     `json:"name"`
	Panes   []SessPane `json:"panes"`
}

type Session struct {
	Name    string    `json:"name"`
	Windows []SessWin `json:"windows"`
}

const (
	sessionWinLinesFmt     = "\"#{window_index}%#{window_name}%#{window_layout}%#{window_active}\""
	sessionEmtpyWinLineFmt = "%%%"

	sessionPaneLinesFmt     = "\"#{pane_index}%#{pane_pid}%#{pane_current_path}%#{pane_active}\""
	sessionEmtpyPaneLineFmt = "%%%"
)

var sessionName string

var sessionCmd = &cobra.Command{
	Use:   "sessions",
	Short: "manipulate session",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Usage()
	},
}

var sessionSaveCmd = &cobra.Command{
	Use:   "save",
	Short: "save a session",
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Print("Session name: ")
		fmt.Scanln(&sessionName)

		if sessionName == "" {
			// TODO: Be kinda cool to do a randomized name?
			log.Fatal("Give it a name.")
		}

		winLines, e, err := lib.Tmux(lib.GlobalArgs, "list-windows", map[string]string{
			"-F": sessionWinLinesFmt,
		}, "")
		if err != nil {
			log.Fatalf("got err: err=%s, out=%s", err, e)
		}

		var session Session

		for _, w := range strings.Split(winLines, "\n") {
			if w == "" || w == sessionEmtpyWinLineFmt {
				continue
			}
			var thisWin SessWin

			winSplit := strings.Split(w, "%")

			thisWin.Index, err = strconv.Atoi(winSplit[0])
			if err != nil {
				log.Fatal(err)
			}

			thisWin.Name = winSplit[1]

			thisWin.Layout = winSplit[2]

			thisWin.Current = lib.TmuxBool(winSplit[3])

			paneLines, e, err := lib.Tmux(lib.GlobalArgs, "list-panes", map[string]string{
				"-t": fmt.Sprint(thisWin.Index),
				"-F": sessionPaneLinesFmt,
			}, "")
			if err != nil {
				log.Fatalf("got err: err=%s, out=%s", err, e)
			}

			for _, p := range strings.Split(paneLines, "\n") {
				if p == "" || p == sessionEmtpyPaneLineFmt {
					continue
				}

				var thisPane SessPane

				paneSplit := strings.Split(p, "%")

				thisPane.Index, err = strconv.Atoi(paneSplit[0])
				if err != nil {
					log.Fatal(err)
				}

				pid, err := strconv.Atoi(paneSplit[1])
				if err != nil {
					log.Fatal(err)
				}

				thisPane.Command, err = lib.GetProcCmd(pid)
				if err != nil {
					thisPane.Command = ""
				}

				// if the current command is not within the allowed list, clear it
				if sessionsConfig := viper.GetStringMapStringSlice("sessions"); sessionsConfig != nil {
					if restoreCmds, ok := sessionsConfig["restore_cmds"]; ok {
						allowed := false
						for _, c := range restoreCmds {
							if strings.HasPrefix(thisPane.Command, c) {
								allowed = true
								break
							}
						}

						if !allowed {
							thisPane.Command = ""
						}
					}
				}

				thisPane.Path = paneSplit[2]

				thisPane.Current = lib.TmuxBool(paneSplit[3])

				thisWin.Panes = append(thisWin.Panes, thisPane)
			}

			session.Windows = append(session.Windows, thisWin)
		}

		s, err := json.Marshal(session)
		if err != nil {
			log.Fatal(err)
		}

		err = os.WriteFile(xdg.ConfigHome+"/tmux/sessions/"+sessionName+".json", s, 0640)
		if err != nil {
			log.Fatal(err)
		}
	},
}

var sessionLoadCmd = &cobra.Command{
	Use:   "load",
	Short: "load a session",
	Run: func(cmd *cobra.Command, args []string) {
		// if no user provided file or name, load all and start fzf

		// createWindows
		// inside createWindows, createPanes
		// finally, attach to session
	},
}

func init() {
	rootCmd.AddCommand(sessionCmd)

	sessionCmd.Flags().StringVarP(&sessionName, "name", "n", "", "name of session to save/load")

	sessionCmd.AddCommand(sessionSaveCmd)

	sessionCmd.AddCommand(sessionLoadCmd)
}
