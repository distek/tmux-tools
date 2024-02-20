package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
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

var (
	flagSessionName string
	flagSessionsDir string
)

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

		if flagSessionName == "" {
			fmt.Print("Session name: ")
			fmt.Scanln(&flagSessionName)
		}

		if flagSessionName == "" {
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

		session.Name = flagSessionName

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

		err = os.WriteFile(xdg.ConfigHome+"/tmux/sessions/"+flagSessionName+".json", s, 0640)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func getSessions(dir string) []Session {
	var ret []Session

	errExt := filepath.WalkDir(dir, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(info.Name(), ".json") {
			return nil
		}

		var thisSession Session

		f, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		err = json.Unmarshal(f, &thisSession)
		if err != nil {
			return err
		}

		ret = append(ret, thisSession)

		return nil
	})

	if errExt != nil {
		log.Fatal(errExt)
	}

	return ret
}

func lsSessions(sessions []Session) []string {
	var ret []string

	for _, v := range sessions {
		ret = append(ret, v.Name)
	}

	return ret
}

func sessionCreatePanes(sessNameWin string, window SessWin) error {
	var err error

	first := true
	focus := 0

	for _, p := range window.Panes {
		if !first {
			// start pane with -c for the remainder
			o, e, err := lib.Tmux(lib.GlobalArgs, "split-window", map[string]string{
				"-t": sessNameWin,
				"-c": p.Path,
			}, "")
			if err != nil {
				log.Println(o)
				log.Println(e)
				log.Println(err)
				return err
			}
		} else {
			first = false
			// cd in first pane
			_, _, err = lib.Tmux(lib.GlobalArgs, "send-keys", map[string]string{
				"-t": fmt.Sprintf("%s.%d", sessNameWin, p.Index),
			}, fmt.Sprintf("\"cd %s\" Enter", p.Path))
			if err != nil {
				log.Println(err)
				return err
			}

			_, _, err = lib.Tmux(lib.GlobalArgs, "send-keys", map[string]string{
				"-t": fmt.Sprintf("%s.%d", sessNameWin, p.Index),
			}, "C-l")
			if err != nil {
				log.Println(err)
				return err
			}
		}

		if p.Command != "" {
			_, _, err := lib.Tmux(lib.GlobalArgs, "send-keys", map[string]string{
				"-t": fmt.Sprintf("%s.%d", sessNameWin, p.Index),
			}, fmt.Sprintf("\"cd %s\" Enter", p.Path))
			if err != nil {
				log.Println(err)
				return err
			}
		}

		if p.Current {
			focus = p.Index
		}
	}

	_, _, err = lib.Tmux(lib.GlobalArgs, "select-pane", map[string]string{
		"-t": fmt.Sprintf("%s.%d", sessNameWin, focus),
	}, "")

	log.Println(err)
	return err
}

func sessionCreateWindows(sessName string, windows []SessWin) error {
	first := true
	focus := 0

	for _, w := range windows {
		target := fmt.Sprintf("%s:%d", sessName, w.Index)

		if !first {
			_, _, err := lib.Tmux(lib.GlobalArgs, "new-window", map[string]string{
				"-t": target,
				"-n": w.Name,
			}, "")
			if err != nil {
				log.Println(err)
				return err
			}
		} else {
			first = false

			o, e, err := lib.Tmux(lib.GlobalArgs, "rename-window", map[string]string{
				"-t":   target,
				w.Name: "",
			}, "")
			if err != nil {
				log.Println(o)
				log.Println(e)
				log.Println(err)
				return err
			}
		}

		if w.Current {
			focus = w.Index
		}

		err := sessionCreatePanes(target, w)
		if err != nil {
			log.Println(err)
			return err
		}

		o, e, err := lib.Tmux(lib.GlobalArgs, "select-layout", map[string]string{
			"-t": target,
		}, fmt.Sprintf("\"%s\"", w.Layout))

		if err != nil {
			log.Println(o)
			log.Println(e)
			log.Println(err)
			return err
		}
	}

	_, _, err := lib.Tmux(lib.GlobalArgs, "select-window", map[string]string{
		"-t": fmt.Sprintf("%s:%d", sessName, focus),
	}, "")

	log.Println(err)
	return err
}

var sessionLoadCmd = &cobra.Command{
	Use:   "load",
	Short: "load a session",
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		var session Session

		sessions := getSessions(flagSessionsDir)

		// if no user provided file or name, load all and start fzf
		if flagSessionName == "" {
			flagSessionName, err = lib.Fzf(lsSessions(sessions))
			if err != nil {
				log.Fatal(err)
			}

			if flagSessionName == "" {
				return
			}
		}

		for _, v := range sessions {
			if v.Name == flagSessionName {
				session = v
			}
		}

		o, e, err := lib.Tmux(lib.GlobalArgs, "new-session", map[string]string{
			"-d": "",
			"-s": session.Name,
		}, "")
		if err != nil {
			log.Println(o)
			log.Println(e)
			log.Fatal(err)
		}

		err = sessionCreateWindows(session.Name, session.Windows)
		if err != nil {
			log.Fatal(err)
		}

		if os.Getenv("TMUX") != "" {
			_, _, err = lib.Tmux(lib.GlobalArgs, "switch-client", map[string]string{
				"-t": session.Name,
			}, "")
			if err != nil {
				log.Fatal(err)
			}
		} else {
			_, _, err = lib.Tmux(lib.GlobalArgs, "attach", map[string]string{
				"-t": session.Name,
			}, "")
			if err != nil {
				log.Fatal(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(sessionCmd)

	sessionCmd.PersistentFlags().StringVarP(&flagSessionName, "name", "n", "", "name of session to save/load")
	sessionCmd.PersistentFlags().StringVarP(&flagSessionsDir, "dir", "d", xdg.ConfigHome+"/tmux/sessions", "directory to save/load sessions from")

	sessionCmd.AddCommand(sessionSaveCmd)

	sessionCmd.AddCommand(sessionLoadCmd)
}
