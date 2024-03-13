package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"sync"
	"syscall"
	"time"

	"github.com/kirsle/configdir"
	"github.com/radovskyb/watcher"
	"github.com/spf13/cobra"
)

var localConfig = configdir.LocalConfig("deconf")
var configPath = path.Join(localConfig, "files")

var rootCmd = &cobra.Command{
	Use:   "deconf",
	Short: "One config ro rule them all",
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Makes the config file usable by parsing, symlinking and adding to watch list to the daemon",
	Run: func(cmd *cobra.Command, _ []string) {
		fmt.Printf("configPath: %v\n", configPath)
		configFile := getConfigFile()
		base := "./"

		fds, err := configFile.Parse(base)
		if err != nil {
			log.Fatalln(err)
		}
		err = configFile.Write(fds)
		if err != nil {
			log.Fatalln(err)
		}
		err = configFile.Symlink(fds)
		if err != nil {
			log.Fatalln(err)
		}
		modifyGitignore := rootCmd.PersistentFlags().Lookup("gitignore")
		if modifyGitignore.Value.String() == "true" {
			err = configFile.Gitignore(fds)
			if err != nil {
				log.Fatalln(err)
			}
		}
		modifyVscode := rootCmd.PersistentFlags().Lookup("vscode")

		if modifyVscode.Value.String() == "true" {
			err = configFile.Vscode(fds)
			if err != nil {
				log.Fatalln(err)
			}
		}

		f := initDaemonFile()
		defer f.Close()

		var written = false
		s := bufio.NewScanner(f)
		for s.Scan() {
			line := s.Text()
			if line == configFile.path {
				written = true
			}
			fmt.Printf("line: %v\n", line)
		}
		if err := s.Err(); err != nil {
			log.Fatalln(err)
		}
		if !written {
			_, err = f.WriteString(configFile.path + "\n")
		}
		if err != nil {
			log.Fatalln(err)
		}
	},
}

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Similar to init, but watches for changes in the configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		w := watcher.New()

		w.SetMaxEvents(1)

		w.FilterOps(watcher.Write)

		go func() {
			for {
				select {
				case <-w.Event:
					initCmd.Run(cmd, args)
				case err := <-w.Error:
					log.Fatalln(err)
				case <-w.Closed:
					return
				}
			}
		}()

		configFile := getConfigFile()
		fmt.Println("Watching for changes on", configFile.path)
		if err := w.Add(configFile.path); err != nil {
			log.Fatalln(err)
		}

		if err := w.Start(time.Millisecond * 100); err != nil {
			log.Fatalln(err)
		}
	}}

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Similar to init, but watches for changes in the configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		f := initDaemonFile()
		defer f.Close()
		err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
		fmt.Println("lock")
		if errors.Is(err, syscall.EAGAIN) {
			fmt.Println("exit")

			// Only allow one file to be run as daemon when injected in the bashrc file for instance
			os.Exit(0)
		}
		fmt.Println("success")
		if err != nil {
			log.Fatalln(err)
		}
		defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(4000)
		}()

		s := bufio.NewScanner(f)
		for s.Scan() {
			line := s.Text()
			fmt.Printf("line: %v\n", line)
		}
		if err := s.Err(); err != nil {
			log.Fatalln(err)
		}
		wg.Wait()
	}}

func initDaemonFile() *os.File {
	if err := configdir.MakePath(localConfig); err != nil {
		log.Fatalln(err)
	}

	f, err := os.OpenFile(configPath, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}

	return f
}

func parseFormat(file string) (Format, error) {
	switch path.Ext(file) {
	case ".md":
		return Markdown, nil
	case ".yml":
	case ".yaml":
		return Yaml, nil
	case ".json":
		return Json, nil
	}
	return 0, errors.New("Unsupported format")
}

func getConfigFile() ConfigFile {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	var configFile ConfigFile
	for i := 0; i < len(ConfigFileVariations); i++ {
		variation := ConfigFileVariations[i]
		file := path.Join(wd, variation)

		fi, _ := os.Stat(file)
		if fi != nil {
			b, err := os.ReadFile(file)

			format, err := parseFormat(file)
			if err != nil {
				log.Fatalln(err)
			}

			configFile = ConfigFile{
				file,
				b,
				format,
				parsers[format],
			}
		}
	}
	if configFile.path == "" {
		log.Fatalln(errors.New(fmt.Sprint("No config file was found. Please create any of", ConfigFileVariations)))
	}
	return configFile
}

func main() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(watchCmd)
	rootCmd.AddCommand(daemonCmd)

	// init flags
	rootCmd.PersistentFlags().Bool("gitignore", true, "modify .gitignore to include the symlinks for configuration files")
	rootCmd.PersistentFlags().Bool("vscode", false, "modify .vscode/settings.json to hide the symlinks for configuration files")
	// initCmd.PersistentFlags().BoolP("watch", "w", false, "watch for changes")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println((err))
		os.Exit(1)
	}
}
