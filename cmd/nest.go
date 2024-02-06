package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/distek/tmux-tools/lib"
	"github.com/spf13/cobra"
)

var (
	flagNestConfig string
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

	err = os.MkdirAll(fmt.Sprintf(tempDir+"/tmux-tools/nested/%s", fileInfo.Name()), 0700)
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

func remapPrefix(pfx string) {

}

var nestCmd = &cobra.Command{
	Use:   "nest",
	Short: "Nest a tmux session",
	Long: `nest allows for nesting of a tmux session and remapping prefix to --prefix

    The default sock path will be "$XDG_CACHE_DIR/tmux-tools/nested/$(pwd)/socket"
`,
	Run: func(cmd *cobra.Command, args []string) {
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

		c := exec.Command("tmux", "-S", flagTmuxSockPath, "-f", flagNestConfig)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		os.Setenv("NEST_TMUX", "1")

		err = c.Start()
		if err != nil {
			log.Fatal(err)
		}

		remapPrefix(flagNestConfig)

		err = c.Wait()
		if err != nil {
			log.Println(err)
		}

		if !lib.SockHasAttached(flagTmuxSockPath) {
			lib.KillServer(flagTmuxSockPath)
		}
	},
}

func init() {
	rootCmd.AddCommand(nestCmd)

	nestCmd.Flags().StringVarP(&flagNestConfig, "tmux-config", "t", "", "Use this config file for tmux")
}
