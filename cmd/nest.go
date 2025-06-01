package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/distek/tmux-tools/lib"
	"github.com/spf13/cobra"
)

var (
	flagNestConfig string
	flagNestPID    int
	flagNestWatch  bool = false
)

func nestGetDefault() string {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Use the current dir's basename as the name of the socket
	fileInfo, err := os.Stat(pwd)
	if err != nil {
		log.Fatal(err)
	}

	tempDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatal(err)
	}

	err = os.MkdirAll(fmt.Sprintf(tempDir+"/tmux-tools/nested/%s", fileInfo.Name()), 0o700)
	if err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf(tempDir+"/tmux-tools/nested/%s/socket", fileInfo.Name())
}

func dirPath(sock string) string {
	var bld strings.Builder

	split := strings.Split(sock, "/")

	for i, v := range split {
		if i == len(split)-1 {
			break
		}
		bld.WriteString(v)
		bld.WriteRune('/')
	}

	return bld.String()
}

func spawnWatcher(pid int) {
	c := exec.Command(os.Args[0], "nest", "-S", flagTmuxSockPath, "--watch", "--pid", strconv.Itoa(pid))

	err := c.Start()
	if err != nil {
		log.Fatal(err)
	}

	// err = c.Process.Release()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	p, err := os.FindProcess(c.Process.Pid)
	if err != nil {
		log.Fatal(err)
		return
	}

	err = p.Release()
	if err != nil {
		log.Fatal(err)
	}
}

func watcher(pid int) {
	// Give time for the nested socket to start
	time.Sleep(time.Second * 1)

	for {
		_, err := os.FindProcess(pid)
		if err != nil || !lib.SockHasAttached(flagTmuxSockPath) {
			lib.KillServer(flagTmuxSockPath)
			time.Sleep(time.Second * 5)
			lib.KillServer(flagTmuxSockPath)
			return
		}

		time.Sleep(time.Second * 5)
	}
}

var nestCmd = &cobra.Command{
	Use:   "nest",
	Short: "Nest a tmux session",
	Long: `nest allows for nesting of a tmux session and remapping prefix to --prefix

    The default sock path will be "$XDG_CACHE_DIR/tmux-tools/nested/$(pwd)/socket"
`,
	Run: func(cmd *cobra.Command, args []string) {
		initGlobalArgs()

		if flagNestWatch {
			if flagNestPID == 0 {
				log.Fatal("PID is not set")
			}

			watcher(flagNestPID)

			return
		}

		// If user did not provide a socket address, make one
		if flagTmuxSockPath == "" {
			flagTmuxSockPath = nestGetDefault()
			lib.GlobalArgs["-S"] = flagTmuxSockPath
		}

		f, err := os.Create(dirPath(flagTmuxSockPath) + "/nest-output.log")
		if err != nil {
			log.Fatal(err)
		}

		log.SetOutput(f)

		if lib.SockActive(flagTmuxSockPath) {
			cmdArgs := []string{"-S", flagTmuxSockPath, "new-session"}
			if len(args) != 0 {
				cmdArgs = append(cmdArgs, args...)
			}
			c := exec.Command("tmux", cmdArgs...)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr

			err := c.Run()
			if err != nil {
				log.Println(err)
			}

			return
		}

		cmdArgs := []string{"-S", flagTmuxSockPath, "-f", flagNestConfig}
		if len(args) != 0 {
			cmdArgs = append(cmdArgs, args...)
		}
		c := exec.Command("tmux", cmdArgs...)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		os.Setenv("NEST_TMUX", "1")

		var wg1 sync.WaitGroup
		var wg2 sync.WaitGroup
		wg1.Add(1)
		go func() {
			err = c.Start()
			if err != nil {
				log.Println(err)
				wg1.Done()
				return
			}

			wg1.Done()

			wg2.Add(1)

			err = c.Wait()
			if err != nil {
				log.Println(err)
				wg2.Done()
				return
			}

			wg2.Done()
		}()

		wg1.Wait()

		go spawnWatcher(c.Process.Pid)

		wg2.Wait()
		if !lib.SockHasAttached(flagTmuxSockPath) {
			lib.KillServer(flagTmuxSockPath)
		}
	},
}

func init() {
	rootCmd.AddCommand(nestCmd)

	nestCmd.Flags().StringVarP(&flagNestConfig, "tmux-config", "t", "", "Use this config file for tmux")

	nestCmd.Flags().IntVarP(&flagNestPID, "pid", "[", -1, "PID to watch")
	nestCmd.Flags().BoolVarP(&flagNestWatch, "watch", "w", false, "enable watch")

	_ = nestCmd.Flags().MarkHidden("pid")
	_ = nestCmd.Flags().MarkHidden("watch")
}
