package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/qq51529210/db/db2go"
	"github.com/qq51529210/db/sql2go/mysql"
	"github.com/qq51529210/log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type cfg struct {
	DBUrl string     `json:"dbUrl"`         // 数据库配置
	Dir   string     `json:"dir,omitempy"`  // 生成代码根目录，空则使用程序当前目录
	Pkg   string     `json:"pkg,omitempy"`  // 代码包名，空则使用数据库名称
	Func  []*cfgFunc `json:"func,omitempy"` // 函数
}

type cfgFunc struct {
	Name string   `json:"name,omitempy"` //
	Tx   string   `json:"tx,omitempy"`   //
	SQL  []string `json:"sql,omitempy"`  //
}

func main() {
	defer func() {
		re := recover()
		if re != nil {
			log.Recover(re, false)
		}
	}()
	var config, http string
	flag.StringVar(&config, "config", "", "config file path")
	flag.StringVar(&http, "http", "", "http listen address")
	flag.Parse()
	if config != "" {
		genCode(config)
		return
	}
	if http != "" {
		return
	}
	flag.PrintDefaults()
}

func loadCfg(path string) *cfg {
	f, err := os.Open(path)
	log.CheckError(err)
	defer func() {
		_ = f.Close()
	}()
	c := new(cfg)
	err = json.NewDecoder(f).Decode(c)
	log.CheckError(err)
	return c
}

func genCode(config string) {
	c := loadCfg(config)
	_url, err := url.Parse(c.DBUrl)
	log.CheckError(err)
	dbUrl := strings.Replace(c.DBUrl, _url.Scheme+"://", "", 1)
	switch strings.ToLower(_url.Scheme) {
	case db2go.MYSQL:
		pkg := c.Pkg
		if pkg == "" {
			_, pkg = path.Split(_url.Path)
		}
		code, err := mysql.NewCode(pkg, dbUrl)
		log.CheckError(err)
		// sql生成FuncTPL
		for _, s := range c.Func {
			_, err = code.Gen(strings.Join(s.SQL, " "), s.Name, s.Tx)
			log.CheckError(err)
		}
		// 保存
		log.CheckError(os.MkdirAll(c.Dir, os.ModePerm))
		log.CheckError(code.SaveFile(filepath.Join(c.Dir, pkg+".go")))
	default:
		panic(fmt.Errorf("unsupported database '%s'", _url.Scheme))
	}
}
