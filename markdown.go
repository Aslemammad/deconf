package main

import (
	"errors"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type MarkdownParser struct{}

func (m MarkdownParser) Parse(cfg *ConfigFile) ([]FileData, error) {
	md := goldmark.New(
		goldmark.WithExtensions(
			meta.New(
				meta.WithStoresInDocument(),
			),
		),
	)
	document := md.Parser().Parse(text.NewReader(cfg.content))
	metaData := document.OwnerDocument().Meta()

	switch t := metaData["gitignore"].(type) {
	case bool:
		cfg.gitignore = t
	}
	switch t := metaData["vscode"].(type) {
	case bool:
		cfg.vscode = t
	}

	var files []FileData
	err := ast.Walk(document, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if n.Kind() == ast.KindCodeSpan {
			pNode, _ := n.Parent().(*ast.Heading)
			if pNode != nil && pNode.Kind() == ast.KindHeading && pNode.Level == 2 {
				file := string(n.Text(cfg.content))

				if len(files) != 0 {
					last := files[len(files)-1]
					if last.content == nil && last.name != file {
						return ast.WalkStop, errors.New(last.name + " does not contain any code block")
					}
					if last.name == file {
						return ast.WalkContinue, nil
					}
				}
				files = append(files, FileData{file, nil})
			}
		}

		if n.Kind() == ast.KindFencedCodeBlock {
			fcb, _ := n.(*ast.FencedCodeBlock)
			s := fcb.Lines()
			r := text.NewReader(cfg.content)

			var content []byte
			for i := 0; i < s.Len(); i++ {
				content = append(content, r.Value(s.At(i))...)
			}

			if len(files) == 0 {
				return ast.WalkStop, errors.New("There's no CodeSpan level-2 heading (## `path`) for this code block:\n" + string(content))
			}

			last := &files[len(files)-1]
			if last.content == nil {
				last.content = append(last.content, content...)
			}
		}
		return ast.WalkContinue, nil
	})

	return files, err
}
