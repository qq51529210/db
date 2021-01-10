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
	DBUrl  string      `json:"dbUrl"`           // 数据库配置
	Pkg    string      `json:"pkg,omitempy"`    // 代码包名，空则使用数据库名称
	File   string      `json:"file,omitempy"`   // 生成代码根目录，空则使用程序当前目录
	Query  []*cfgQuery `json:"query,omitempy"`  // 函数
	Exec   []*cfgExec  `json:"exec,omitempy"`   // 函数
	Driver string      `json:"driver,omitempy"` // 数据库驱动
}

type cfgQuery struct {
	Name string   `json:"func,omitempy"` //
	Tx   string   `json:"tx,omitempy"`   //
	Row  bool     `json:"row,omitempy"`  //
	Null bool     `json:"null,omitempy"` //
	SQL  []string `json:"sql,omitempy"`  //
}

type cfgExec struct {
	Name string   `json:"func,omitempy"` //
	Tx   string   `json:"tx,omitempy"`   //
	SQL  []string `json:"sql,omitempy"`  //
}

func main() {
	log.Recover(nil)
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

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func loadCfg(path string) *cfg {
	f, err := os.Open(path)
	checkError(err)
	defer func() {
		_ = f.Close()
	}()
	c := new(cfg)
	err = json.NewDecoder(f).Decode(c)
	checkError(err)
	return c
}

func genCode(config string) {
	c := loadCfg(config)
	_url, err := url.Parse(c.DBUrl)
	checkError(err)
	dbUrl := strings.Replace(c.DBUrl, _url.Scheme+"://", "", 1)
	switch strings.ToLower(_url.Scheme) {
	case db2go.MYSQL:
		// 包名
		pkg := c.Pkg
		if pkg == "" {
			_, pkg = path.Split(_url.Path)
		}
		// 驱动
		driver := c.Driver
		if driver == "" {
			driver = "github.com/go-sql-driver/mysql"
		}
		// 生成路径
		file := c.File
		if file == "" {
			file = pkg + ".go"
		}
		if filepath.Ext(file) == "" {
			file += ".go"
		}
		code, err := mysql.NewCode(pkg, driver, dbUrl)
		checkError(err)
		// sql生成FuncTPL
		for i, f := range c.Query {
			_, err = code.Query(strings.Join(f.SQL, " "), f.Name, f.Tx, f.Row, f.Null)
			if err != nil {
				checkError(fmt.Errorf("sql[%d]: %v", i, err))
			}
		}
		for i, f := range c.Exec {
			_, err = code.Exec(strings.Join(f.SQL, " "), f.Name, f.Tx)
			if err != nil {
				checkError(fmt.Errorf("sql[%d]: %v", i, err))
			}
		}
		// 保存
		checkError(code.SaveFile(file))
	default:
		panic(fmt.Errorf("unsupported database '%s'", _url.Scheme))
	}
}
