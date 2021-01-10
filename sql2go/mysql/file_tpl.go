package mysql

import (
	"fmt"
	"io"
	"strings"
	"text/template"
)

var _fileTPL = template.Must(template.New("fileTPL").Parse(`package {{.Pkg}}

import (
	"database/sql"
	{{if eq .Driver "github.com/go-sql-driver/mysql" -}}
	"{{.Driver}}"
	{{else -}}
	_ "{{.Driver}}"
	{{end -}}
	{{if .Strings -}}
	"strings"
	{{end -}}
	"time"
)

var (
	DB *sql.DB
	{{- range $i,$s := .Sql}}
	stmt{{$i}} *sql.Stmt // {{$s}}
	{{- end}}
)

{{if eq .Driver "github.com/go-sql-driver/mysql" -}}
func IsUniqueKeyError(err error) bool {
	if e, o := err.(*mysql.MySQLError); o {
		return e.Number == 1169
	}
	return false
}
{{- end}}

func Init(url string, maxOpen, maxIdle int, maxLifeTime, maxIdleTime time.Duration) (err error){
	DB, err = sql.Open("mysql", url)
	if err != nil {
		return err
	}
	DB.SetMaxOpenConns(maxOpen)
	DB.SetMaxIdleConns(maxIdle)
	DB.SetConnMaxLifetime(maxLifeTime)
	DB.SetConnMaxIdleTime(maxIdleTime)
	return PrepareStmt(DB)
}

func UnInit() {
	if DB == nil {
		return
	}
	_ = DB.Close()
	CloseStmt()
}

func PrepareStmt(db *sql.DB) (err error) {
	{{- range $i,$s:= .Sql}}
	stmt{{$i}}, err = db.Prepare("{{$s}}")
	if err != nil {
		return
	}
	{{- end}}
	return
}

func CloseStmt() {
	{{- range $i,$s:= .Sql}}
	if stmt{{$i}} != nil {
		_ = stmt{{$i}}.Close()
	}
	{{- end}}
}

{{- range .Func}}
{{$.FuncString .}}
{{- end}}
`))

type fileTPL struct {
	Pkg     string
	Driver  string // driver
	Strings bool   // import
	Sql     []string
	Func    []TPL
}

func (t *fileTPL) Execute(w io.Writer) error {
	return _fileTPL.Execute(w, t)
}

func (t *fileTPL) FuncString(tpl TPL) string {
	var str strings.Builder
	_ = tpl.Execute(&str)
	return str.String()
}

func (t *fileTPL) StmtName(s string) string {
	stmt := fmt.Sprintf("stmt%d", len(t.Sql))
	t.Sql = append(t.Sql, s)
	return stmt
}
