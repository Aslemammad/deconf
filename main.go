package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/kirsle/configdir"
	"github.com/radovskyb/watcher"
	"github.com/spf13/cobra"
)

var localConfig = configdir.LocalConfig("deconf")
var configFilesPath = path.Join(localConfig, "files")

var rootCmd = &cobra.Command{
	Use:   "deconf",
	Short: "One config ro rule them all",
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Makes the config file usable by parsing, symlinking and adding to watch list to the daemon",
	Run: func(cmd *cobra.Command, args []string) {
		var customConfigFile string
		if len(args) == 0 {
			customConfigFile = ""
		} else {
			customConfigFile = args[0]
		}
		configFile := getConfigFile(customConfigFile)

		fds, err := configFile.Parse()
		if err != nil {
			log.Fatalln(err)
		}
		err = configFile.Write(fds)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println("Wrote down config files successfully.")
		err = configFile.Symlink(fds)
		if err != nil {
			log.Fatalln(err)
		}

		if configFile.gitignore {
			err = configFile.Gitignore(fds)
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Println("Changes applied to .gitignore")
		}

		if configFile.vscode {
			err = configFile.Vscode(fds)
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Println("Changes applied to .vscode/settings.json")
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
		var customConfigFile string
		if len(args) == 0 {
			customConfigFile = ""
		} else {
			customConfigFile = args[0]
		}
		configFile := getConfigFile(customConfigFile)
		w := watcher.New()

		w.SetMaxEvents(1)

		w.FilterOps(watcher.Write)

		go func() {
			for {
				select {
				case <-w.Event:
					if customConfigFile != "" {
						fmt.Println(configFile.path, "changed")
					} else {
						fmt.Println(customConfigFile, "changed")
					}
					initCmd.Run(cmd, args)
				case err := <-w.Error:
					log.Fatalln(err)
				case <-w.Closed:
					return
				}
			}
		}()

		fmt.Println("Watching for changes on", configFile.path)
		if err := w.Add(configFile.path); err != nil {
			log.Fatalln(err)
		}

		if err := w.Start(time.Millisecond * 100); err != nil {
			log.Fatalln(err)
		}
	}}

// var daemonCmd = &cobra.Command{
// 	Use:   "daemon",
// 	Short: "Similar to init, but watches for changes in the configuration file",
// 	Run: func(cmd *cobra.Command, args []string) {
// 		f := initDaemonFile()
// 		defer f.Close()
// 		err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
// 		if errors.Is(err, syscall.EAGAIN) {
// 			// Only allow one file to be run as daemon when injected in the bashrc file for instance
// 			os.Exit(0)
// 		}
// 		if err != nil {
// 			log.Fatalln(err)
// 		}
// 		defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
// 		var wg sync.WaitGroup

// 		s := bufio.NewScanner(f)
// 		for s.Scan() {
// 			wg.Add(1)
// 			line := s.Text()
// 			go func() {
// 				watchCmd.Run(watchCmd, []string{line})
// 				wg.Done()
// 			}()
// 		}
// 		if err := s.Err(); err != nil {
// 			log.Fatalln(err)
// 		}
// 		wg.Wait()
// 	}}

func initDaemonFile() *os.File {
	if err := configdir.MakePath(localConfig); err != nil {
		log.Fatalln(err)
	}

	f, err := os.OpenFile(configFilesPath, os.O_RDWR|os.O_CREATE, os.ModePerm)
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

func getConfigFile(file string) ConfigFile {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	var configFile ConfigFile
	for i := 0; i < len(ConfigFileVariations); i++ {
		variation := ConfigFileVariations[i]
		if len(file) == 0 {
			file = path.Join(wd, variation)
		} else {
			if !path.IsAbs(file) {
				file = path.Join(wd, file)
			}
		}

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
				true,
				false,
			}
			break
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
	rootCmd.AddCommand(GetDaemonCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println((err))
		os.Exit(1)
	}
}
