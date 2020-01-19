package server

/*
 * @Description:
 * @Author: chenwei
 * @Date: 2020-01-15 17:12:09
 */

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-ini/ini"
)

/*
 * @Description:
 * @Author: chenwei
 * @Date: 2020-01-15 17:12:09
 */
var p *Person
var cfg *ini.File
var filePath string

func LoadConfig(source string, serveAddr string) {
	var err error
	p = new(Person)
	filePath = source
	cfg, err = ini.Load(source)
	if err != nil {
		panic("load error:" + err.Error())
	}

	cfg.NameMapper = ini.TitleUnderscore
	if err := cfg.MapTo(p); err != nil {
		panic("Map config error:" + err.Error())
	}
	startApi(serveAddr)
}

func startApi(addr string) {
	g := gin.New()
	g.GET("/get", get)
	g.GET("/set", set)

	srv := &http.Server{
		Addr: addr,
		//ReadTimeout:       30 * time.Second,
		//ReadHeaderTimeout: 10 * time.Second,
		//IdleTimeout:       200 * time.Microsecond,
		//WriteTimeout:      5 * time.Second,
		Handler: g,
	}
	AddEventMap("Note.Tkc", onNoteTkcChange)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		panic("listen error:" + err.Error())
	}

}
func get(c *gin.Context) {
	sec := c.Query("sec")
	key := c.Query("key")
	if key == "" {
		val := cfg.Section(sec).KeysHash()
		c.JSON(200, val)
	} else {
		val := cfg.Section(sec).Key(key).String()
		log.Println("val:", val)
		c.JSON(200, val)
		// c.String(200, string(encode([]byte(val))))
	}
}
func encode(bs []byte) []byte {
	return bs

}
func set(c *gin.Context) {
	sec := c.Query("sec")
	key := c.Query("key")
	val := c.Query("val")
	sec = upperCaseFirstLetter(sec)
	key = upperCaseFirstLetter(key)
	valueOfRoot := reflect.ValueOf(p)
	valueOfSec := valueOfRoot.Elem().FieldByName(sec)
	valueOfKey := valueOfSec.FieldByName(key)

	typeOfRoot := reflect.TypeOf(p)
	typeOfSec, found := typeOfRoot.Elem().FieldByName(sec)
	if !found {
		c.JSON(400, "section not fount")
		return
	}
	typeOfKey, found := typeOfSec.Type.FieldByName(key)
	if !found {
		c.JSON(400, "key not fount")
		return
	}

	switch typeOfKey.Type.String() {
	case "string":
		currentVal := valueOfKey.String()
		if currentVal != val {
			valueOfKey.SetString(val)
		} else {
			c.JSON(200, "ok")
		}
	case "int":
		currentVal := valueOfKey.Int()
		newVal, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			panic("parse int error:" + err.Error())
		}
		if currentVal != newVal {
			valueOfKey.SetInt(newVal)
		} else {
			c.JSON(200, "ok")
		}
	}
	err := ini.ReflectFromWithMapper(cfg, p, ini.TitleUnderscore)
	if err != nil {
		log.Println("reflect error:", err.Error())
	}
	cfg.SaveTo(filePath)
	onChange(sec, key, val)
	c.JSON(200, "ok")
}

var changeEventMap = make(map[string]func(string))

func AddEventMap(key string, fn func(string)) {
	changeEventMap[key] = fn
}

func onChange(sec, key, val string) {
	secKey := fmt.Sprintf("%s.%s", sec, key)
	if fn, ok := changeEventMap[secKey]; !ok {
		fmt.Println("no event found")
		return
	} else {
		fn(val)
	}
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
func upperCaseFirstLetter(s string) string {
	if len(s) == 0 {
		return s
	}
	if s[0] < 96 {
		return s
	}
	return string(s[0]-32) + s[1:]
}

type ConfigNode struct {
	Name  string
	Token string
}
type Note struct {
	Tkc     int
	Content string
	Cities  []string
}
type Person struct {
	Name string
	Age  int `ini:"age"`
	Male bool
	Born time.Time
	Note
	Created time.Time `ini:"-"`
}
