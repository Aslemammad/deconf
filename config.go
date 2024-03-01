package main

import (
	"bufio"
	"reflect"

	"errors"
	"fmt"
	"os"
	"path"
	"slices"

	"github.com/yosuke-furukawa/json5/encoding/json5"
)

type Format = int8

const (
	Markdown Format = iota
	Yaml            // yet to be supported
	Json            // yet to be supported
)

var parsers = map[Format]Parser{
	Markdown: &MarkdownParser{},
}

type ConfigFile struct {
	path    string
	content []byte
	format  Format
	parser  Parser
}

var ConfigFileVariations = [...]string{"config.md"}

type FileData struct {
	name    string
	content []byte
}

type Parser interface {
	Parse(cf *ConfigFile, base string) ([]FileData, error)
}

func (cf *ConfigFile) Parse(base string) ([]FileData, error) {
	data, err := cf.parser.Parse(cf, base)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (cf *ConfigFile) Write(fds []FileData) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	nmDir := path.Join(wd, "node_modules")
	_, err = os.Stat(nmDir)
	if err != nil {
		return err
	}
	deconfDir := path.Join(nmDir, ".deconf")
	err = os.MkdirAll(deconfDir, os.ModePerm)
	if err != nil {
		return err
	}
	for _, fd2 := range fds {
		err = os.MkdirAll(path.Join(deconfDir, path.Dir(fd2.name)), os.ModePerm)
		if err != nil {
			return err
		}
		f, err := os.Create(path.Join(deconfDir, fd2.name))
		if err != nil {
			return err
		}
		_, err = f.Write(fd2.content)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cf *ConfigFile) Symlink(fds []FileData) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	nmDir := path.Join(wd, "node_modules")
	deconfDir := path.Join(nmDir, ".deconf")
	for _, fd := range fds {
		target := path.Join(deconfDir, fd.name)
		sym := path.Join(wd, fd.name)
		err = os.MkdirAll(path.Dir(sym), os.ModePerm)
		if err != nil {
			return err
		}
		os.Symlink(target, sym)
	}
	return nil
}

func (cf *ConfigFile) Gitignore(fds []FileData) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	gitignore := path.Join(wd, ".gitignore")
	f, err := os.OpenFile(gitignore, os.O_RDWR, os.ModeAppend)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var oldFiles []string
	for scanner.Scan() {
		s := scanner.Text()
		present := slices.ContainsFunc(fds, func(fd FileData) bool {
			return fd.name == s
		})
		if present {
			oldFiles = append(oldFiles, s)
		}
	}

	log := []byte("Adding ")
	for i, fd := range fds {
		if !slices.Contains(oldFiles, fd.name) {
			_, err = f.WriteString("\n# Added by deconf\n" + fd.name)
			if err != nil {
				return err
			}
			isLast := i == len(fds)-1
			if !isLast {
				log = append(log, []byte(fd.name+", ")...)
			} else {
				log = append(log, []byte("and "+fd.name+" to .gitignore")...)
			}
		}
	}
	fmt.Println(string(log))

	return nil
}

//	type Settings struct {
//		// FilesExclude map[string]bool `json:"files.exclude"`
//	}
type Settings map[string]interface{}

func (cf *ConfigFile) Vscode(fds []FileData) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	settingsPath := path.Join(wd, ".vscode", "settings.json")
	err = os.MkdirAll(path.Dir(settingsPath), os.ModePerm)
	if err != nil {
		return err
	}
	_, err = os.Stat(settingsPath)
	if errors.Is(err, os.ErrNotExist) {
		var f *os.File
		f, err = os.Create(settingsPath)
		f.Write([]byte("{}"))
	}
	if err != nil {
		return err
	}
	b, err := os.OpenFile(settingsPath, os.O_RDWR, os.ModeAppend)
	if err != nil {
		return err
	}
	defer b.Close()

	var settings Settings
	d := json5.NewDecoder(b)
	err = d.Decode(&settings)
	if err != nil {
		return err
	}
	err = b.Truncate(0)
	if err != nil {
		return err
	}
	_, err = b.Seek(0, 0)
	if err != nil {
		return err
	}
	fmt.Println("settings\n", settings["files.exclude"], reflect.TypeOf(settings["files.exclude"]))
	// filesExclude := settings["files.exclude"].(map[string]bool)
	var filesExclude map[string]bool = make(map[string]bool)
	untypedFilesExclude := settings["files.exclude"]

	switch t := untypedFilesExclude.(type) {
	case map[string]interface{}:
		for k, v := range t {
			switch tv := v.(type) {
			case bool:
				filesExclude[k] = tv
			}
		}
	}

	for _, fd := range fds {
		filesExclude[fd.name] = true
	}

	fmt.Printf("filesExclude: %v\n", filesExclude)

	settings["files.exclude"] = filesExclude
	b2, err := json5.MarshalIndent(settings, "", "\t")
	if err != nil {
		return err
	}
	_, err = b.Write(b2)
	if err != nil {
		return err
	}
	return nil
}
