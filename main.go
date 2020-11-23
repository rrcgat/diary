package main

import (
	"fmt"
	"os"
)

func main() {
	d, err := NewClient("diary.db", diaryBucket)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	d.command(os.Args...)
}
