package main

import (
	"os"
	"text/tabwriter"
	"text/template"

    "go.containerssh.io/libcontainerssh/auditlog/message"
    "go.containerssh.io/libcontainerssh/internal/auditlog/codec/binary"
)

type context struct {
	Version      uint64
	MagicLength  int
	MagicValue   string
	HeaderLength int

	message.Documentation
}

func main() {
	tpl, err := template.New("template.md").ParseFiles("generate/format/template.md")
	if err != nil {
		panic(err)
	}
	target, err := os.Create("FORMAT.md")
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = target.Close()
	}()

	ctx := context{
		Version:       binary.CurrentVersion,
		MagicLength:   binary.FileFormatLength,
		MagicValue:    binary.FileFormatMagic,
		HeaderLength:  binary.FileFormatLength + 8,
		Documentation: message.DocumentMessages(),
	}

	w := tabwriter.NewWriter(target, 2, 2, 2, ' ', 0)
	if err := tpl.Execute(w, ctx); err != nil {
		panic(err)
	}
}
