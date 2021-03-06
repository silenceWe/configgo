package configgo

/*
 * @Description:
 * @Author: chenwei
 * @Date: 2020-01-15 17:12:09
 */

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-ini/ini"
)

var rootConfig interface{}
var baseConfig *Configgo
var cfg *ini.File
var filePath string

func LoadConfig(bc interface{}, source string, addr string) {
	var err error
	filePath = source
	cfg, err = ini.Load(source)
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			fmt.Println(err.Error())
			fmt.Println("Try to create default config file")
			initConfig(bc, source)
			fmt.Println("Create default success!")
			return
		}
	}

	cfg.NameMapper = ini.TitleUnderscore
	if err := cfg.MapTo(bc); err != nil {
		panic("Map config error:" + err.Error())
	}

	rootConfig = bc
	configgo := reflect.ValueOf(bc).Elem().FieldByName("Configgo")
	baseConfig = configgo.Interface().(*Configgo)
	checkConfig()
	fmt.Printf("config node name:%+v\n", baseConfig.Name)
	if addr != "" {
		go startAPI(addr)
	}
}
func initConfig(bc interface{}, path string) {
	cfg := ini.Empty()
	ini.ReflectFrom(cfg, bc)
	cfg.SaveTo(path)
}
func checkConfig() bool {
	if baseConfig.Password == "" {
		fmt.Println("Please set the config password")
		os.Exit(0)
	}
	if baseConfig.Token == "" {
		fmt.Println("Please set the config token")
		os.Exit(0)
	}
	return true
}

func startAPI(addr string) {
	gin.SetMode(gin.ReleaseMode)
	g := gin.New()
	g.Use(gin.Recovery(), gin.Logger(), validPassword())
	g.GET("/get", get)
	g.GET("/set", set)

	fmt.Println("Start API at :", addr)

	srv := &http.Server{
		Addr:              addr,
		ReadTimeout:       30 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       200 * time.Microsecond,
		WriteTimeout:      5 * time.Second,
		Handler:           g,
	}
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		panic("listen error:" + err.Error())
	}
}

func validPassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		password := c.Query("p")
		if password == "" || password != baseConfig.Password {
			c.JSON(403, "")
			c.Abort()
			return
		}
		c.Next()
	}
}

func encode(src, key []byte) []byte {
	r := make([]byte, len(src))
	keyLen := len(key)
	for k, v := range src {
		r[k] = v ^ key[k%keyLen]
	}
	return r
}

func get(c *gin.Context) {
	sec := c.Query("sec")
	key := c.Query("key")
	if key == "" {
		val := cfg.Section(sec).KeysHash()
		valJson, err := json.Marshal(val)
		if err != nil {
			panic("parse json error:" + err.Error())
		}
		encodeStr := string(encode([]byte(valJson), []byte(baseConfig.Token)))
		c.String(200, encodeStr)
	} else {
		val := cfg.Section(sec).Key(key).String()
		log.Println("val:", val)
		c.JSON(200, val)
		// c.String(200, string(encode([]byte(val))))
	}
}
func set(c *gin.Context) {
	sec := c.Query("sec")
	key := c.Query("key")
	val := c.Query("val")
	sec = upperCaseFirstLetter(snackToCamelWithHead(sec))
	key = upperCaseFirstLetter(snackToCamelWithHead(key))

	valueOfRoot := reflect.ValueOf(rootConfig)
	valueOfSec := valueOfRoot.Elem().FieldByName(sec)
	valueOfKey := valueOfSec.FieldByName(key)

	typeOfRoot := reflect.TypeOf(rootConfig)
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
	case "int", "Int8", "Int16", "Int32", "int64":
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
	case "Uint", "Uint8", "Uint16", "Uint32", "Uint64":
		currentVal := valueOfKey.Uint()
		newVal, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			panic("parse int error:" + err.Error())
		}
		if currentVal != newVal {
			valueOfKey.SetUint(newVal)
		} else {
			c.JSON(200, "ok")
		}
	case "bool":
		currentVal := valueOfKey.Bool()
		newVal := false
		if val == "true" {
			newVal = true
		}
		if currentVal != newVal {
			valueOfKey.SetBool(newVal)
		} else {
			c.JSON(200, "ok")
		}
	}
	err := ini.ReflectFromWithMapper(cfg, rootConfig, ini.TitleUnderscore)
	if err != nil {
		log.Println("reflect error:", err.Error())
	}
	cfg.SaveTo(filePath)
	onChange(sec, key, val)
	c.JSON(200, "ok")
}

var changeWatcherMap = make(map[string]func(string))

func AddWatcher(sec, key string, fn func(string)) {
	sec = upperCaseFirstLetter(sec)
	key = upperCaseFirstLetter(key)
	watcherKey := fmt.Sprintf("%s.%s", sec, key)
	changeWatcherMap[watcherKey] = fn
}

func onChange(sec, key, val string) {
	secKey := fmt.Sprintf("%s.%s", sec, key)
	if fn, ok := changeWatcherMap[secKey]; !ok {
		fmt.Println("no event found")
		return
	} else {
		fn(val)
	}
}
func snackToCamelWithHead(s string) string {
	if len(s) == 0 {
		return s
	}
	res := ""
	i := 0
	l := len(s)
	for i < l {
		v := s[i]
		if v == '_' {
			res += string(s[i+1] - 32)
			i += 2
			continue
		}
		res += string(s[i])
		i++
	}
	return res
}
func upperCaseFirstLetter(s string) string {
	if len(s) == 0 {
		return s
	}

	if s[0] > 90 {
		return string(s[0]-32) + s[1:]
	}
	return s
}

type Configgo struct {
	Name     string
	Token    string
	Password string
}
