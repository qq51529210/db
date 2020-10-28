package tpl

import (
	"io"
)

const tplStrInit = `package {{.Pkg}}

import (
	"database/sql"
	{{if ne .DBPkg "github.com/go-sql-driver/mysql" -}}_ {{end}}"{{.DBPkg}}"
	"time"
)

var (
	DB *sql.DB
	{{- range $k,$v := .Stmt}}
	{{$k}} *sql.Stmt
	{{- end}}
)

{{- if eq .DBPkg "github.com/go-sql-driver/mysql"}}
func IsUniqueError(err error) bool {
	sqlErr, ok := err.(*mysql.MySQLError)
	if ok {
		ok = sqlErr.Number == 1062
	}
	return ok
}
{{- end}}

func Init(url string, maxOpen, maxIdle int, maxLifeTime, maxIdleTime time.Duration) (err error){
	DB, err = sql.Open("{{.DBType}}", url)
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
	{{- range $k,$v:= .Stmt}}
	{{$k}}, err = db.Prepare("{{$v}}")
	if err != nil {
		return
	}
	{{- end}}
	return
}

func CloseStmt() {
	{{- range $k,$v:= .Stmt}}
	if {{$k}} != nil {
		_ = {{$k}}.Close()
	}
	{{- end}}
}
`

type Init struct {
	Pkg    string
	Stmt   map[string]string
	DBType string
	DBPkg  string
}

func NewInitTPL(pkg, dbType, dbPkg string) *Init {
	t := new(Init)
	t.Pkg = pkg
	t.DBType = dbType
	t.DBPkg = dbPkg
	t.Stmt = make(map[string]string)
	return t
}

func (t *Init) StructName() string {
	return ""
}

func (t *Init) Execute(w io.Writer) error {
	return tplInit.Execute(w, t)
}
