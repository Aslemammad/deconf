package main

import (
	"fmt"
	"os"
	"path"
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
	fmt.Printf("fds: %v\n", fds)
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
