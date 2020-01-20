package main

/*
 * @Description:
 * @Author: chenwei
 * @Date: 2020-01-19 17:16:43
 */

import (
	"fmt"
	"strconv"
	"time"

	"github.com/silenceWe/configgo"
)

func main() {
	c := AllConfig{}
	configgo.LoadConfig(&c, "./cfg_base.ini", ":8080")
	configgo.AddWatcher("Note.Tkc", onNoteTkcChange)
}

type AllConfig struct {
	Configgo configgo.Configgo
}

func (p *AllConfig) GetConfiggo() *configgo.Configgo {
	return &p.Configgo
}
func onNoteTkcChange(val string) {
	tkc, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		panic(err)
	}
	printTk(tkc)
}

var tk *time.Ticker

func printTk(tkc int64) {
	fmt.Println("tkc:", tkc)
	if tk != nil {
		tk.Stop()
	}
	tk = time.NewTicker(time.Duration(tkc * int64(time.Second)))
	go func() {
		for {
			select {
			case <-tk.C:
				fmt.Println("tk:", time.Now().String())
			}
		}
	}()
}

type Note struct {
	Tkc     int
	Content string
	Cities  []string
}
