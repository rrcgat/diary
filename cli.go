package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	bolt "go.etcd.io/bbolt"
)

func showUsage() {
	fmt.Println(`Useage: diary [command] [arg]
    - init
        初始化数据库
    - import <dir>
        从 <dir> 导入日记
    - history
        展示历史上的今天的日记
    - date <date>
        展示指定日期的日记
    - edit <date>
        编辑指定日期的日记
    - new
    - new <date>
        根据指定日期 <date> 创建日记，未指定时按当前日期
    -
    - random
        随机展示一篇日记`)
}

func (d Diary) newDiary(date string) {
	layout := "2006-01-02"
	if date == "" {
		date = time.Now().Format(layout)
	} else {
		t, err := time.Parse(layout, date)
		if err != nil {
			fmt.Fprintln(os.Stderr, "时间格式错误", err)
			return
		}
		if time.Now().After(t) {
			fmt.Fprintln(os.Stderr, "日记日期超过当前日期")
			return
		}
	}
	if err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(d.bucket)
		if b.Get([]byte(date)) != nil {
			return fmt.Errorf("该日期<%s>已有日记", date)
		}
		return nil
	}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	tmpfile, _ := ioutil.TempFile("", "diary."+date+".*.md")
	cmd := exec.Command("nvim", tmpfile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	buf, err := ioutil.ReadAll(tmpfile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	if len(buf) == 0 {
		fmt.Fprintln(os.Stderr, "未保存：日记不能为空")
		return
	}
	if err = d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(d.bucket)
		return b.Put([]byte(date), buf)
	}); err != nil {
		fmt.Fprintln(os.Stderr, "日记保存失败", err)
		return
	}
	fmt.Println("已保存")
}

func (d Diary) edit(date string) error {
	var newBuf []byte
	if err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(d.bucket)
		buf := b.Get([]byte(date))
		if buf != nil {
			tmpfile, _ := ioutil.TempFile("", "diary."+date+".*.md")
			if err := ioutil.WriteFile(tmpfile.Name(), buf, 0600); err != nil {
				return err
			}
			cmd := exec.Command("nvim", tmpfile.Name())
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return err
			}
			if buf, err := ioutil.ReadAll(tmpfile); err != nil {
				return err
			} else {
				newBuf = buf
			}
			return nil
		}
		return errors.New("未找到对应的日记")
	}); err != nil {
		return err
	}

	if len(newBuf) == 0 {
		return errors.New("编辑后文件内容为空，无效")
	}
	return d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(d.bucket)
		return b.Put([]byte(date), newBuf)
	})
}

func (d Diary) init() {
	err := d.InitBucket(diaryBucket)
	if err != nil {
		fmt.Fprintln(os.Stderr, "初始化失败", err)
	}
	fmt.Println("已初始化")
}

func (d Diary) importDiaryWithPath(arg string) {
	if arg == "" {
		fmt.Fprintln(os.Stderr, "需要指定导入目录")
		return
	}
	fmt.Printf("共导入 %d 篇日志\n", d.ImportDiary(arg))
}

func (d Diary) history() {
	hs := d.TodayInHistory()
	if hs != nil {
		for _, h := range hs {
			fmt.Println(h)
		}
	} else {
		fmt.Println("无")
	}
}

func (d Diary) locDate(date string) {
	if len(date) != 10 {
		fmt.Fprintln(os.Stderr, "请指定正确的日记格式（yyyy-MM-dd）")
		return
	}
	fmt.Println(d.Loc(date))
}

func (d Diary) editDiary(arg string) {
	if arg == "" {
		fmt.Fprintln(os.Stderr, "请指定要编辑的日记")
		return
	}
	if err := d.edit(arg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Printf("%s已修改\n", arg)
}

func (d Diary) command(args ...string) {
	var (
		cmd string
		arg string
	)
	switch len(args) {
	case 1:
		cmd = "random"
	case 2:
		cmd = args[1]
	default:
		cmd = args[1]
		arg = args[2]
	}

	switch cmd {
	case "-h":
		fallthrough
	case "--help":
		showUsage()
	case "import":
		d.importDiaryWithPath(arg)
	case "history":
		d.history()
	case "init":
		d.init()
	case "date":
		d.locDate(arg)
	case "edit":
		d.editDiary(arg)
	case "random":
		fmt.Println(d.Random())
	case "new":
		d.newDiary(arg)
	default:
		fmt.Println("Unknown command")
	}
}
