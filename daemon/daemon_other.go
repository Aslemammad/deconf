//go:build !wasm

package daemon

import (
	"bufio"
	"errors"
	"log"
	"os"
	"sync"
	"syscall"

	"github.com/spf13/cobra"
)

var DaemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Similar to init, but watches for changes in the configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		f := initDaemonFile()
		defer f.Close()
		err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if errors.Is(err, syscall.EAGAIN) {
			// Only allow one file to be run as daemon when injected in the bashrc file for instance
			os.Exit(0)
		}
		if err != nil {
			log.Fatalln(err)
		}
		defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		var wg sync.WaitGroup

		s := bufio.NewScanner(f)
		for s.Scan() {
			wg.Add(1)
			line := s.Text()
			go func() {
				watchCmd.Run(watchCmd, []string{line})
				wg.Done()
			}()
		}
		if err := s.Err(); err != nil {
			log.Fatalln(err)
		}
		wg.Wait()
	}}
