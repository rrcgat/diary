package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

func expand(path string) string {
	if strings.HasPrefix(path, "~/") {
		curr, err := user.Current()
		if err != nil {
			log.Fatal(err)
			return path
		}
		return filepath.Join(curr.HomeDir, path[2:])
	}
	return path
}

func main() {
	d, err := NewClient(expand(os.Getenv("DAIRY_DATABASE")), diaryBucket)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	d.command(os.Args...)
}
