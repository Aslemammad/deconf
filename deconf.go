package main

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "deconf",
	Short: "One config ro rule them all",
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Makes the config file usable by parsing, symlinking and adding to watch list to the daemon",
	Run: func(cmd *cobra.Command, _ []string) {
		wd, err := os.Getwd()
		if err != nil {
			logErr(err)
		}
		base := "./"

		var configFile ConfigFile
		for i := 0; i < len(ConfigFileVariations); i++ {
			variation := ConfigFileVariations[i]
			file := path.Join(wd, variation)

			fi, _ := os.Stat(file)
			if fi != nil {
				b, err := os.ReadFile(file)

				format, err := parseFormat(file)
				if err != nil {
					logErr(err)
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
			logErr(errors.New(fmt.Sprint("No config file was found. Please create any of", ConfigFileVariations)))
		}

		fds, err := configFile.Parse(base)
		if err != nil {
			logErr(err)
		}
		err = configFile.Write(fds)
		if err != nil {
			logErr(err)
		}
		err = configFile.Symlink(fds)
		if err != nil {
			logErr(err)
		}
		err = configFile.Gitignore(fds)
		if err != nil {
			logErr(err)
		}
	},
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

func logErr(err error) {
	fmt.Printf("err: %v\n", err)
	os.Exit(1)
}

func main() {
	rootCmd.AddCommand(initCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println((err))
		os.Exit(1)
	}
}
