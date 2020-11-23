package main

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strings"

	bolt "go.etcd.io/bbolt"
)

func (d *Diary) ImportDiary(dirname string) int {
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		log.Fatal(err)
		return 0
	}

	imported := 0
	d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(d.bucket)
		re, _ := regexp.Compile(`# \d{4}-\d{2}-\d{2}`)
		for _, f := range files {
			splits := strings.Split(f.Name(), " ")
			if len(splits) != 2 || len(splits[0]) != 10 {
				log.Fatal(f.Name())
				continue
			}
			file, err := os.Open(path.Join(dirname, f.Name()))
			if err != nil {
				log.Fatal(err)
				continue
			}
			var buf [][]byte
			i := 0
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Bytes()
				if i == 0 && re.Match(line) || string(line) == "" {
					continue
				}
				buf = append(buf, line)
			}
			b.Put([]byte(splits[0]), bytes.TrimSpace(bytes.Join(buf, []byte("\n\n"))))
			imported += 1
		}
		return nil
	})

	return imported
}
