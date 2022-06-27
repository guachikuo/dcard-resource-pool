package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

func main() {
	dedup := map[string]struct{}{}
	if err := filepath.Walk(".",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !strings.HasSuffix(path, ".go") {
				return nil
			}

			temps := strings.Split(path, "/")
			dir := strings.Join(temps[:len(temps)-1], "/")

			if _, ok := dedup[dir]; ok {
				return nil
			}

			fmt.Println("--------")
			fmt.Println("found :", path)
			fmt.Println("start running :", dir)

			cmd := exec.Command("go", "test")
			cmd.Dir = dir

			var stdBuffer bytes.Buffer
			mw := io.MultiWriter(os.Stdout, &stdBuffer)
			cmd.Stderr = mw

			// Execute the command
			red := color.New(color.FgRed).SprintFunc()
			if err := cmd.Run(); err != nil {
				fmt.Println("result :", red("fail"))
				return err
			}

			dedup[dir] = struct{}{}

			green := color.New(color.FgGreen).SprintFunc()
			fmt.Println("result :", green("success"))

			return nil
		},
	); err != nil {
		panic(err)
	}
}
