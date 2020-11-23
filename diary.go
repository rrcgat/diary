package main

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	bolt "go.etcd.io/bbolt"
)

type Diary struct {
	db     *bolt.DB
	bucket []byte
}

const diaryBucket = "diary"

// NewClient 创建日记客户端，用于交互.
func NewClient(path, bucket string) (*Diary, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	return &Diary{db, []byte(bucket)}, nil
}

func (d Diary) InitBucket(bucket string) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		return err
	})
}

// Random 随机获取数据库中的一篇日记.
func (d Diary) Random() string {
	rand.Seed(time.Now().Unix())

	ret := ""
	_ = d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(d.bucket)
		if b != nil {
			end := rand.Intn(b.Stats().KeyN)
			c := b.Cursor()
			var (
				k []byte
				v []byte
			)
			if end == 0 {
				k, v = c.First()
			} else {
				c.First()
				for i := 0; i < end; i++ {
					c.Next()
				}
				k, v = c.Next()
			}
			ret = fmt.Sprintf("# %s\n\n%s\n", string(k), string(v))
		}
		return nil
	})

	return ret
}

// TodayInHistory 获取历史上的今天的日记.
func (d Diary) TodayInHistory() (ret []string) {
	t := time.Now()
	_ = d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(d.bucket)
		for y := 2016; y < t.Year(); y++ {
			date := fmt.Sprintf("%d-%02d-%02d", y, t.Month(), t.Day())
			content := b.Get([]byte(date))
			if content != nil {
				ret = append(ret, fmt.Sprintf("# %s\n\n%s\n\n", date, content))
			}
		}
		return nil
	})

	return ret
}

func (d Diary) Loc(date string) (ret string) {
	_ = d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(d.bucket)
		ret = string(b.Get([]byte(date)))
		return nil
	})

	return ret
}

func (d Diary) NewDiary(date, text []byte) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(d.bucket)
		if b != nil {
			return b.Put(date, text)
		}
		return errors.New("failed to retrieve bucket")
	})
}
