//go:build wasm

package daemon

import (
	"log"

	"github.com/spf13/cobra"
)

var DaemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Similar to init, but watches for changes in the configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		log.Fatalln("daemon is not available in the wasm/wasi build")
	}}
