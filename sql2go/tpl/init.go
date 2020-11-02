package tpl

import (
	"strconv"
	"text/template"
)

var (
	tplInit = template.Must(template.New("Init").Parse(`package {{.Pkg}}

import (
	"database/sql"
	{{if ne .DBPkg "github.com/go-sql-driver/mysql" -}}_ {{end}}"{{.DBPkg}}"
	"time"
)

var (
	DB *sql.DB
	{{- range $k,$v := .Stmt}}
	{{$k}} *sql.Stmt // {{$v}}
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
`))
)

type Init interface {
	TPL
	AddStmt(sql string) string
	Save(file string) error
}

type _init struct {
	Pkg    string
	Stmt   map[string]string
	DBType string
	DBPkg  string
}

func (i *_init) TPL() *template.Template {
	return tplInit
}

func (i *_init) AddStmt(sql string) string {
	name := "stmt" + strconv.Itoa(len(i.Stmt))
	i.Stmt[name] = sql
	return name
}

func (s *_init) Save(file string) error {
	return saveFile(tplInit, s, file)
}

func NewInitTPL(pkg, dbType, dbPkg string) Init {
	t := new(_init)
	t.Pkg = pkg
	t.DBType = dbType
	t.DBPkg = dbPkg
	t.Stmt = make(map[string]string)
	return t
}
