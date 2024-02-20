package cmd

import (
	"os"

	"github.com/distek/tmux-tools/lib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	flagConfigFile   string
	flagTmuxSockName string
	flagTmuxSockPath string
)

var rootCmd = &cobra.Command{
	Use:   "tmux-tools",
	Short: "Tools for simplifying the manipulation of tmux",
	Long:  `For more info on flags denoted with "See tmux docs", check tmux's man page`,
	// Run:   func(cmd *cobra.Command, args []string) {},
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func initConfig() {
	if flagConfigFile != "" {
		viper.SetConfigFile(flagConfigFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		cfgPath := home + "/.config/tmux/tools"

		viper.AddConfigPath(cfgPath)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	_ = viper.ReadInConfig()
}

func initGlobalArgs() {
	if lib.GlobalArgs == nil {
		lib.GlobalArgs = make(map[string]string, 1)
	}

	if flagTmuxSockName != "" {
		lib.GlobalArgs["-L"] = flagTmuxSockName
	}

	if flagTmuxSockPath != "" {
		lib.GlobalArgs["-S"] = flagTmuxSockPath
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&flagConfigFile, "config", "", "config file (default: $HOME/.config/tmux/tools/tmux-tools.yaml)")

	rootCmd.PersistentFlags().StringVarP(&flagTmuxSockPath, "socket-path", "S", "", "tmux socket path (See tmux docs)")
	rootCmd.PersistentFlags().StringVarP(&flagTmuxSockName, "socket-name", "L", "", "tmux socket name (See tmux docs)")

	initGlobalArgs()
}
