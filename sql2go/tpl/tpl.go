package tpl

import (
	"io"
	"os"
	"strings"
	"text/template"
)

// 全部的*template.Template变量
var (
	tplExec,
	tplQuery,
	tplQueryRow,
	tplStructExec,
	tplStructQuery,
	tplStructQueryRow,
	tplInit,
	tplStruct,
	tplJoinStruct *template.Template
)

// 初始化全部的*template.Template变量
func init() {
	tplInit = template.Must(template.New("Init").Parse(tplStrInit))
	tplStruct = template.Must(template.New("Struct").Parse(tplStrStruct))
	tplJoinStruct = template.Must(template.New("JoinStruct").Parse(tplStrJoinStruct))
	tplExec = template.Must(template.New("Exec").Parse(tplStrExec))
	tplStructExec = template.Must(template.New("StructExec").Parse(tplStrStructExec))
	tplQuery = template.Must(template.New("Query").Parse(tplStrQuery))
	tplQueryRow = template.Must(template.New("QueryRow").Parse(tplStrQueryRow))
	tplStructQuery = template.Must(template.New("StructQuery").Parse(tplStrStructQuery))
	tplStructQueryRow = template.Must(template.New("StructQueryRow").Parse(tplStrStructQueryRow))
}

// 保存到文件
func SaveFile(tpl TPL, path string) error {
	// 打开文件写
	f, err := os.OpenFile(path, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	// 关闭文件
	defer func() { _ = f.Close() }()
	// 输出
	return tpl.Execute(f)
}

func tplString(tp TPL) string {
	var str strings.Builder
	_ = tp.Execute(&str)
	return str.String()
}

type TPL interface {
	StructName() string
	Execute(io.Writer) error
}

type FuncTPL interface {
	TPL
	SQL() string
	Stmt() string
	SetStmt(name string)
	IsTx() bool
	FuncName() string
	SetFuncName(name string)
	SetParamName(names []string)
}

type StructTPL interface {
	TPL
	AddFuncTPL(tpl FuncTPL)
}

type funcTPL struct {
	Tx       bool
	Sql      string
	Func     string
	Param    []string
	Struct   string
	StmtName string
}

func (t *funcTPL) StructName() string {
	return t.Struct
}

func (t *funcTPL) SQL() string {
	return t.Sql
}

func (t *funcTPL) SetStmt(name string) {
	t.StmtName = name
}

func (t *funcTPL) Stmt() string {
	return t.StmtName
}

func (t *funcTPL) IsTx() bool {
	return t.Tx
}

func (t *funcTPL) FuncName() string {
	return t.Func
}

func (t *funcTPL) SetFuncName(name string) {
	t.Func = name
}

func (t *funcTPL) SetParamName(names []string) {
	for i, name := range names {
		if i >= len(t.Param) {
			return
		}
		t.Param[i] = name
	}
}

func (t *funcTPL) TPLParam() string {
	s := ""
	if t.Tx {
		s = "tx *sql.Tx"
	}
	if len(t.Param) < 1 {
		return s
	}
	if t.Tx {
		s += ", "
	}
	s += strings.Join(t.Param, ", ") + " interface{}"
	return s
}

func (t *funcTPL) TPLStmt() string {
	if t.Tx {
		return "tx.Stmt(" + t.StmtName + ")"
	}
	return t.StmtName
}

type Arg struct {
	Name    string
	IsField bool
}

func (a Arg) String() string {
	if a.IsField {
		return "m." + a.Name
	}
	return a.Name
}
