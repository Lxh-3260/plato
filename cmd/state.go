package cmd

import (
	"github.com/lxh-3260/plato/state"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(stateCmd)
}

var stateCmd = &cobra.Command{
	Use: "state",
	Run: StateHandle,
}

func StateHandle(cmd *cobra.Command, args []string) {
	state.RunMain(ConfigPath)
}
