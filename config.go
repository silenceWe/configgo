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

	"github.com/gin-gonic/gin"
	"github.com/go-ini/ini"
)

/*
 * @Description:
 * @Author: chenwei
 * @Date: 2020-01-15 17:12:09
 */
var baseConfig BaseConfig
var cfg *ini.File
var filePath string

type BaseConfig interface {
	GetConfiggo() *Configgo
}

func LoadConfig(bc BaseConfig, source string, serveAddr string) {
	var err error
	baseConfig = bc
	filePath = source
	cfg, err = ini.Load(source)
	if err != nil {
		panic("load error:" + err.Error())
	}

	cfg.NameMapper = ini.TitleUnderscore
	if err := cfg.MapTo(baseConfig); err != nil {
		panic("Map config error:" + err.Error())
	}
	checkConfig()
	fmt.Printf("config:%+v\n", baseConfig.GetConfiggo().Name)
	startApi(serveAddr)
}

func checkConfig() bool {
	if baseConfig.GetConfiggo().Password == "" {
		fmt.Println("Please set the config password")
		os.Exit(0)
	}
	if baseConfig.GetConfiggo().Token == "" {
		fmt.Println("Please set the config token")
		os.Exit(0)
	}
	return true
}

func startApi(addr string) {
	g := gin.New()
	g.Use(gin.Recovery(), gin.Logger(), validPassword())
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
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		panic("listen error:" + err.Error())
	}

}

func validPassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		password := c.Query("p")
		if password == "" || password != baseConfig.GetConfiggo().Password {
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
		encodeStr := string(encode([]byte(valJson), []byte(baseConfig.GetConfiggo().Token)))
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
	sec = upperCaseFirstLetter(sec)
	key = upperCaseFirstLetter(key)
	valueOfRoot := reflect.ValueOf(baseConfig)
	valueOfSec := valueOfRoot.Elem().FieldByName(sec)
	valueOfKey := valueOfSec.FieldByName(key)

	typeOfRoot := reflect.TypeOf(baseConfig)
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
	err := ini.ReflectFromWithMapper(cfg, baseConfig, ini.TitleUnderscore)
	if err != nil {
		log.Println("reflect error:", err.Error())
	}
	cfg.SaveTo(filePath)
	onChange(sec, key, val)
	c.JSON(200, "ok")
}

var changeWatcherMap = make(map[string]func(string))

func AddEventMap(key string, fn func(string)) {
	changeWatcherMap[key] = fn
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

func upperCaseFirstLetter(s string) string {
	if len(s) == 0 {
		return s
	}
	if s[0] < 96 {
		return s
	}
	return string(s[0]-32) + s[1:]
}

type Configgo struct {
	Name     string
	Token    string
	Password string
}
