package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/qq51529210/db/db2go"
	"github.com/qq51529210/db/sql2go/main/web"
	"github.com/qq51529210/db/sql2go/mysql"
	"github.com/qq51529210/log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

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
		web.ListenAndServe(http)
		return
	}
	flag.PrintDefaults()
}

type cfg struct {
	DBUrl    string            `json:"dbUrl"`        // 数据库配置
	Dir      string            `json:"dir,omitempy"` // 生成代码根目录，空则使用程序当前目录
	CleanDir bool              `json:"cleanDir,omitempy"`
	Pkg      string            `json:"pkg,omitempy"`     // 代码包名，空则使用数据库名称
	Func     map[string]string `json:"func,omitempy"`    // 代码中没有涉及到的函数，值表示函数的返回值
	SQL      []*cfgSQL         `json:"sql,omitempy"`     // 自动sql解析
	Def      []string          `json:"default,omitempy"` // 需要生成默认代码的表名
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
	c.Check()
	return c
}

func (c *cfg) Check() {
	if c.DBUrl == "" {
		panic(fmt.Errorf("dbUrl is empty"))
	}
	if c.Dir == "" {
		c.Dir, _ = filepath.Abs(os.Args[0])
	}
	for _, cc := range c.SQL {
		cc.Check()
	}
}

type cfgSQL struct {
	Sql   string   `json:"sql"`   // 原始sql
	Tx    string   `json:"tx"`    // 也是sql，但是会使用"stmt"作为第一参数参数
	Func  string   `json:"func"`  // 自定义函数名
	Param []string `json:"param"` // 自定义参数名
	IsTx  bool     // 是否tx
}

func (c *cfgSQL) Check() {
	p := c.Param
	c.Param = c.Param[:0]
	for _, s := range p {
		s = strings.TrimSpace(s)
		if s != "" {
			c.Param = append(c.Param, s)
		}
	}
	if c.Tx != "" {
		c.Sql = c.Tx
		c.IsTx = true
	}
	if len(c.Param) > 0 && c.Param[0] == "tx" {
		c.Param = c.Param[1:]
		c.IsTx = true
	}
}

func genCode(config string) {
	c := loadCfg(config)
	_url, err := url.Parse(c.DBUrl)
	log.CheckError(err)
	dbUrl := strings.Replace(c.DBUrl, _url.Scheme+"://", "", 1)
	switch strings.ToLower(_url.Scheme) {
	case db2go.MYSQL:
		code, err := mysql.NewCode(dbUrl, c.Pkg)
		log.CheckError(err)
		// 添加函数
		for k, v := range c.Func {
			mysql.AddFunction(k, v)
		}
		// 默认FuncTPL
		same := make(map[string]int)
		for _, t := range c.Def {
			if _, ok := same[t]; ok {
				continue
			}
			same[t] = 1
			_, err := code.DefaultFuncTPLs(t)
			log.CheckError(err)
		}
		// sql生成FuncTPL
		for _, s := range c.SQL {
			_, err := code.FuncTPL(s.Sql, s.Func, s.IsTx, s.Param)
			log.CheckError(err)
		}
		// 保存
		log.CheckError(code.SaveFiles(c.Dir, c.CleanDir))
	default:
		panic(fmt.Errorf("dbUrl: unsupported database '%s'", _url.Scheme))
	}
}
