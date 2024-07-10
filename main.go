package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ryancragun/terraform-plan-editor/internal/edit"
)

var config = &edit.Config{}

func init() {
	flag.StringVar(&config.TextEditorCmd, "editor", "", "the editor to use when editing text files")
	flag.StringVar(&config.BinEditorCmd, "bin-editor", "", "the editor to use when editing binary files")
}

func getEditorCmd(cmd string) (string, error) {
	if cmd != "" {
		fmt.Println("text editor: " + cmd)
		return cmd, nil
	}

	if editor := os.Getenv("EDITOR"); editor != "" {
		fmt.Println("editor: " + editor)
		return editor, nil
	}

	return "", fmt.Errorf("you must set the editor with the '-editor' flag or set $EDITOR")
}

func getBinEditorCmd(cmd string) (string, error) {
	if cmd != "" {
		fmt.Println("binary editor: " + cmd)
		return cmd, nil
	}

	if editor := os.Getenv("EDITOR"); editor != "" {
		fmt.Println("binary editor: " + editor)
		return editor, nil
	}

	return "", fmt.Errorf("you must set the editor with the '-bin-editor' flag or set $EDITOR")
}

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) != 2 {
		panic("terraform-plan-editor: <flags> <source-plan-path> <dest-plan-path>")
	}
	var err error
	config.PlanPath, err = filepath.Abs(args[0])
	if err != nil {
		panic(err)
	}
	config.DstPath, err = filepath.Abs(args[1])
	if err != nil {
		panic(err)
	}

	config.TextEditorCmd, err = getEditorCmd(config.TextEditorCmd)
	if err != nil {
		panic(err)
	}

	config.BinEditorCmd, err = getBinEditorCmd(config.BinEditorCmd)
	if err != nil {
		panic(err)
	}

	err = edit.New(config).Edit()
	if err != nil {
		panic(err)
	}
}
