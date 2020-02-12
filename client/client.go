package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/go-ini/ini"
	"github.com/silenceWe/configgo"
	"github.com/silenceWe/loggo"
)

const (
	KV_SEPARATOR        = "->"
	DEFAULT_CONFIG_FILE = "./template.ini"
)

var sec string
var o string
var f string
var param *Param
var configs []*NodeConfig
var nodeMap map[string]*NodeConfig

func main() {
	flag.StringVar(&f, "f", DEFAULT_CONFIG_FILE, "section")
	flag.StringVar(&o, "o", "get", "opertion")
	flag.StringVar(&sec, "sec", "", "section")
	flag.Parse()

	loggo.InitDefaultLog(&loggo.LoggerOption{StdOut: true, Level: loggo.ALL})
	switch o {
	case "init":
		buildTemplate(f)
		fmt.Println("Build template success!")
	case "get":
		printGet()
	case "set":
		printGet()
		if param.Operations.Set != nil && len(param.Operations.Set) != 0 {
			if printSet() {
				printGet()
			}
		}
	}
}
func fileExist(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			// file does not exist
			return false
		} else {
			// other error
			panic("check template file error:" + err.Error())
		}
	} else {
		//exist
		return true
	}
}
func buildTemplate(path string) {
	if fileExist(path) {
		t := time.Now().Format("2006-01-02T15-04-05")
		fmt.Printf("There is a file named %s,it will be moved to %s_%s\n", path, path, t)
		if err := os.Rename(path, path+"_"+t); err != nil {
			panic("Rename old file error:" + err.Error())
		}
	}

	p := new(Param)
	cfg := ini.Empty()
	ini.ReflectFromWithMapper(cfg, p, ini.TitleUnderscore)
	cfg.SaveTo(path)
}

func loadParam() {
	param = new(Param)
	cfg, err := ini.ShadowLoad(f)
	if err != nil {
		panic("load error:" + err.Error())
	}

	cfg.NameMapper = ini.TitleUnderscore
	if err := cfg.MapTo(param); err != nil {
		panic("Map config error:" + err.Error())
	}

}
func printGet() {
	loadParam()
	head := []string{"Name", "Addr"}
	rows := [][]string{}
	for _, v := range param.Nodes.Info {
		v = strings.Replace(v, " ", "", -1)
		parts := strings.Split(v, KV_SEPARATOR)
		if len(parts) != 2 {
			panic("node format error")
		}
		row := []string{parts[0], parts[1]}
		rows = append(rows, row)
	}

	loggo.Infoln("Nodes:")
	configgo.PrintTable(head, rows)
	fmt.Println("Section:", param.Operations.Sec)
	get(param.Operations.Sec, "")
}
func printSet() bool {
	fmt.Printf("Will set the following values:\n")

	head := []string{}
	headMap := make(map[string]bool)
	for _, v := range param.Operations.Set {
		v = strings.Replace(v, " ", "", -1)
		parts := strings.Split(v, KV_SEPARATOR)
		if len(parts) != 2 {
			loggo.Errorln("set format error. example : 'set = tkc -> 1'")
			os.Exit(0)
		}
		nodeKeyParts := strings.Split(parts[0], ".")
		ls := len(nodeKeyParts)
		if ls != 2 {
			fmt.Println("set format error")
			os.Exit(0)
		}
		node := nodeKeyParts[0]
		key := nodeKeyParts[1]
		val := parts[1]
		if _, ok := headMap[key]; !ok {
			headMap[key] = true
		}

		if node == "ALL" || node == "all" || node == "All" {
			for k := range nodeMap {
				config := nodeMap[k]
				if config.change == nil {
					config.change = make(map[string]string)
				}
				config.change[key] = val
				head = append(head, key)
			}
		} else {
			config, ok := nodeMap[node]
			if !ok {
				panic("node not found")
			}
			if config.change == nil {
				config.change = make(map[string]string)
			}
			config.change[key] = val
			head = append(head, key)
		}
	}

	sort.Sort(sort.StringSlice(head))
	head = append([]string{"NodeName"}, head...)

	change := false
	rows := [][]string{}
	for _, c := range configs {
		row := []string{c.name}
		for k, vv := range head {
			if k > 0 {
				oldVal, ok := c.data[vv]
				if !ok {
					row = append(row, "(Not Exists])")
					continue
				}
				newVal := c.change[vv]
				cell := fmt.Sprintf("%s => %s", oldVal, newVal)
				if c.data[vv] == newVal {
					cell += "(Not Change)"
				} else {
					if newVal != "" {
						change = true
					}
				}
				row = append(row, cell)
			}
		}
		rows = append(rows, row)
	}
	configgo.PrintTable(head, rows)
	if !change {
		fmt.Println("Nothing changed!")
		return false
	}
	if confirm() {
		do()
		return true
	}
	return false
}
func do() {
	for configIndex, _ := range configs {
		config := configs[configIndex]
		for k, v := range config.change {
			set(param.Operations.Sec, k, v)
		}
	}
}

func confirm() bool {
	f := bufio.NewReader(os.Stdin)
	var input string
	for {
		fmt.Print("Confirm change ?(Y/n):")
		input, _ = f.ReadString('\n')
		switch input {
		case "Y\n":
			loggo.Infoln("Confirmed")
			return true
		default:
			fmt.Println("cancel")
			os.Exit(0)
			return false
		}
	}
}

type Param struct {
	Nodes      Nodes      `comment:"The nodes you want to operate.example:info = server1 => 127.0.0.1:8080"`
	Operations Operations `comment:"The operations you want to do.example:set = note.content => hello"`
	Password   string
	Token      string
}
type Nodes struct {
	Info []string `ini:",,allowshadow"`
}
type Operations struct {
	Sec string
	Set []string `ini:",,allowshadow"`
}

var addrs []string

type NodeConfig struct {
	name   string
	addr   string
	data   map[string]string
	change map[string]string
}

func get(sec, key string) {
	configs = make([]*NodeConfig, len(param.Nodes.Info))
	nodeMap = make(map[string]*NodeConfig)
	for k, v := range param.Nodes.Info {
		v = strings.Replace(v, " ", "", -1)
		parts := strings.Split(v, KV_SEPARATOR)
		if len(parts) != 2 {
			panic("node format error")
		}
		configs[k] = &NodeConfig{name: parts[0], addr: parts[1]}
		nodeMap[parts[0]] = configs[k]
	}

	headMap := make(map[string]bool)

	for k, _ := range configs {
		configs[k].data = make(map[string]string)
		var url string
		if key == "" {
			url = fmt.Sprintf("http://%s/get?sec=%s", configs[k].addr, sec)
		} else {
			url = fmt.Sprintf("http://%s/get?sec=%s&key=%s", configs[k].addr, sec, key)
		}
		res := httpGet(url)
		json.Unmarshal(res, &configs[k].data)
		for k, _ := range configs[k].data {
			headMap[k] = true
		}
	}
	head := []string{}
	for k, _ := range headMap {
		head = append(head, k)
	}

	sort.Sort(sort.StringSlice(head))
	head = append([]string{"NodeName"}, head...)
	rows := [][]string{}

	for _, c := range configs {
		row := []string{c.name}
		for k, vv := range head {
			if k > 0 {
				row = append(row, c.data[vv])
			}
		}
		rows = append(rows, row)
	}
	configgo.PrintTable(head, rows)

}

func set(sec, key, val string) {
	for _, v := range configs {
		url := fmt.Sprintf("http://%s/set?sec=%s&key=%s&val=%s", v.addr, sec, key, val)
		res := httpGet(url)
		fmt.Println("res:", string(res))
	}
}

func httpGet(url string) []byte {
	url += "&p=" + param.Password
	resp, err := http.Get(url)
	if err != nil {
		loggo.Errorln("get error:", err.Error())
		os.Exit(0)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		loggo.Errorln("ioutil.ReadAll error:", err.Error())
	}
	res := encode(body, []byte(param.Token))
	fmt.Println("res:", string(res))
	return res
}

func encode(src, key []byte) []byte {
	r := make([]byte, len(src))
	keyLen := len(key)
	for k, v := range src {
		r[k] = v ^ key[k%keyLen]
	}
	return r
}
