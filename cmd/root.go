package cmd

import (
	"fmt"
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

func init() {
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if flagConfigFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(flagConfigFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		cfgPath := home + "/.config/tmux/tools"

		// Search config in home directory with name ".tmux-tools" (without extension).
		viper.AddConfigPath(cfgPath)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagConfigFile, "config", "", "config file (default: $HOME/.config/tmux/tools/tmux-tools.yaml)")

	rootCmd.PersistentFlags().StringVarP(&flagTmuxSockPath, "socket-path", "S", "", "tmux socket path (See tmux docs)")
	rootCmd.PersistentFlags().StringVarP(&flagTmuxSockName, "socket-name", "L", "", "tmux socket name (See tmux docs)")

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
