package daemon

import "github.com/spf13/cobra"

func GetDaemonCmd() *cobra.Command {
	return DaemonCmd
}
