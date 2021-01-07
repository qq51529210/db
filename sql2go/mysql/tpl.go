package mysql

import (
	"io"
	"strings"
	"text/template"
)

type TPL interface {
	Execute(io.Writer) error
}

type tpl struct {
	Sql  string
	Func string
	Tx   string
	Stmt string
}

func (t *tpl) StmtTPL() string {
	if t.Tx != "" {
		var s strings.Builder
		s.WriteString(t.Tx)
		s.WriteString(".Stmt(")
		s.WriteString(t.Stmt)
		s.WriteByte(')')
		return s.String()
	}
	return t.Stmt
}

var _execTPL = template.Must(template.New("execTPL").Parse(`
{{- if .Sql}}
// {{.Sql}}
{{end -}}
func {{.Func}}({{.ParamTPL}}) (sql.Result, error) {
	return {{.StmtTPL}}.Exec(
		{{- range .Param}}
		{{.}},
		{{- end}}
	)
}
`))

type execTPL struct {
	*tpl
	Param []string
}

func (t *execTPL) Execute(w io.Writer) error {
	return _execTPL.Execute(w, t)
}

func (t *execTPL) ParamTPL() string {
	var s strings.Builder
	if t.Tx != "" {
		s.WriteString(t.Tx)
		s.WriteString(" *sql.Tx")
	}
	if len(t.Param) < 1 {
		return s.String()
	}
	if s.Len() > 0 {
		s.WriteString(", ")
	}
	s.WriteString(strings.Join(t.Param, ", "))
	s.WriteString(" interface{}")
	return s.String()
}

func (t *execTPL) String() string {
	var str strings.Builder
	_ = t.Execute(&str)
	return str.String()
}

var _queryTPL = template.Must(template.New("queryTPL").Parse(`
{{- if .Sql}}
// {{.Sql}}
{{end -}}
func {{.Func}}({{.ParamTPL}}) ([]{{.Type}}, error) {
	models := make([]{{.Type}}, 0)
	rows, err := {{.StmtTPL}}.Query(
		{{- range .Param}}
		{{.}},
		{{- end}}
	)
	if nil != err {
		if err != sql.ErrNoRows {
			return nil, err
		}
		return models, nil
	}
	{{- if .NullType}}
	var model {{.NullType}}
	{{- else}}
	var model {{.Type}}
	{{- end}}
	for rows.Next() {
		err = rows.Scan(&model)
		if nil != err {
			return nil, err
		}
		{{- if .NullType}}
		if model.Valid {
			models = append(models, {{.Type}}(model.{{.NullValue}}))
		}
		{{- else}}
		models = append(models, model)
		{{- end}}
	}
	return models, nil
}
`))

type queryTPL struct {
	*tpl
	Param []string
	*scanFieldTPL
}

func (t *queryTPL) Execute(w io.Writer) error {
	return _queryTPL.Execute(w, t)
}

func (t *queryTPL) String() string {
	var str strings.Builder
	_ = t.Execute(&str)
	return str.String()
}

func (t *queryTPL) ParamTPL() string {
	var s strings.Builder
	if t.Tx != "" {
		s.WriteString(t.Tx)
		s.WriteString(" *sql.Tx")
	}
	if len(t.Param) < 1 {
		return s.String()
	}
	if s.Len() > 0 {
		s.WriteString(", ")
	}
	s.WriteString(strings.Join(t.Param, ", "))
	s.WriteString(" interface{}")
	return s.String()
}

var _execStructTPL = template.Must(template.New("execStructTPL").Parse(`
type {{.Model}} struct {
	{{- range .Field}}
	{{index . 0}} {{index . 1}} {{index . 2}}
	{{- end}}
}

// {{.Sql}}
func {{.Func}}({{.ParamTPL}}) (sql.Result, error) {
	return {{.StmtTPL}}.Exec(
		{{- range .Field}}
		model.{{index . 0}},
		{{- end}}
	)
}
`))

type scanFieldTPL struct {
	Name      string
	Type      string
	NullType  string
	NullValue string
}

type execStructTPL struct {
	*tpl
	Model string
	Field [][3]string
}

func (t *execStructTPL) Execute(w io.Writer) error {
	return _execStructTPL.Execute(w, t)
}

func (t *execStructTPL) ParamTPL() string {
	var s strings.Builder
	if t.Tx != "" {
		s.WriteString(t.Tx)
		s.WriteString(" *sql.Tx, ")
	}
	s.WriteString("model *")
	s.WriteString(t.Model)
	return s.String()
}

func (t *execStructTPL) String() string {
	var str strings.Builder
	_ = t.Execute(&str)
	return str.String()
}

var _queryStructTPL = template.Must(template.New("queryStructTPL").Parse(`
type {{.Func}}Model struct {
	{{- range .Field}}
	{{index . 0}} {{index . 1}} {{index . 2}}
	{{- end}}
}

// {{.Sql}}
func {{.Func}}({{.ParamTPL}}) ([]*{{.Func}}Model, error) {
	models := make([]*{{.Func}}Model, 0)
	rows, err := {{.StmtTPL}}.Query(
		{{- range .Param}}
		{{.}},
		{{- end}}
	)
	if nil != err {
		if err != sql.ErrNoRows {
			return nil, err
		}
		return models, nil
	}
	{{- range .Scan}}
	{{- if .NullType}}
	var {{.Name}} {{.NullType}}
	{{- end}}
	{{- end}}
	for rows.Next() {
		model := new({{.Func}}Model)
		err = rows.Scan(
			{{- range .Scan}}
			{{- if .NullType}}
			&{{.Name}},
			{{- else}}
			&model.{{.Name}},
			{{- end}}
			{{- end}}
		)
		if nil != err {
			return nil, err
		}
		{{- range .Scan}}
		{{- if .NullType}}
		if {{.Name}}.Valid {
			model.{{.Name}} = {{.Type}}({{.Name}}.{{.NullValue}})
		}
		{{- end}}
		{{- end}}
		models = append(models, model)
	}
	return models, nil
}
`))

type queryStructTPL struct {
	*tpl
	Param []string
	Field [][3]string
	Scan  []*scanFieldTPL
}

func (t *queryStructTPL) Execute(w io.Writer) error {
	return _queryStructTPL.Execute(w, t)
}

func (t *queryStructTPL) ParamTPL() string {
	var s strings.Builder
	if t.Tx != "" {
		s.WriteString(t.Tx)
		s.WriteString(" *sql.Tx, ")
	}
	if len(t.Param) < 1 {
		return s.String()
	}
	if s.Len() > 0 {
		s.WriteString(", ")
	}
	s.WriteString(strings.Join(t.Param, ", "))
	s.WriteString(" interface{}")
	return s.String()
}

func (t *queryStructTPL) String() string {
	var str strings.Builder
	_ = t.Execute(&str)
	return str.String()
}

var _querySqlTPL = template.Must(template.New("querySqlTPL").Parse(`
type {{.Func}}Model struct {
	{{- range .Field}}
	{{index . 0}} {{index . 1}} {{index . 2}}
	{{- end}}
}

// {{.Sql}}
func {{.Func}}({{.ParamTPL}}) ([]*{{.Func}}Model, error) {
	var str strings.Builder
	{{- range .Segment}}
	str.WriteString({{.}})
	{{- end}}
	models := make([]*{{.Func}}Model, 0)
	rows, err := DB.Query(
		str.String(),
		{{- range .Param}}
		{{.}},
		{{- end}}
	)
	if nil != err {
		if err != sql.ErrNoRows {
			return nil, err
		}
		return models, nil
	}
	{{- range .Scan}}
	{{- if .NullType}}
	var {{.Name}} {{.NullType}}
	{{- end}}
	{{- end}}
	for rows.Next() {
		model := new({{.Func}}Model)
		err = rows.Scan(
			{{- range .Scan}}
			{{- if .NullType}}
			&{{.Name}},
			{{- else}}
			&model.{{.Name}},
			{{- end}}
			{{- end}}
		)
		if nil != err {
			return nil, err
		}
		{{- range .Scan}}
		{{- if .NullType}}
		if {{.Name}}.Valid {
			model.{{.Name}} = {{.Type}}({{.Name}}.{{.NullValue}})
		}
		{{- end}}
		{{- end}}
		models = append(models, model)
	}
	return models, nil
}
`))

type querySqlTPL struct {
	*tpl
	Param   []string
	Column  []string
	Field   [][3]string
	Scan    []*scanFieldTPL
	Segment []string
}

func (t *querySqlTPL) Execute(w io.Writer) error {
	return _querySqlTPL.Execute(w, t)
}

func (t *querySqlTPL) ParamTPL() string {
	var s strings.Builder
	if t.Tx != "" {
		s.WriteString(t.Tx)
		s.WriteString(" *sql.Tx, ")
	}
	if len(t.Param) < 1 && len(t.Column) < 1 {
		return s.String()
	}
	if s.Len() > 0 {
		s.WriteString(", ")
	}
	if len(t.Column) > 0 {
		s.WriteString(strings.Join(t.Column, ", "))
		s.WriteString(" string")
	}
	if s.Len() > 0 {
		s.WriteString(", ")
	}
	s.WriteString(strings.Join(t.Param, ", "))
	s.WriteString(" interface{}")
	return s.String()
}

func (t *querySqlTPL) String() string {
	var str strings.Builder
	_ = t.Execute(&str)
	return str.String()
}

var _fileTPL = template.Must(template.New("fileTPL").Parse(`package {{.Pkg}}

import (
	"database/sql"
	{{if eq .Driver "github.com/go-sql-driver/mysql" -}}
	"{{.Driver}}"
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

{{- if eq .Driver "github.com/go-sql-driver/mysql"}}
func IsUniqueKeyError(err error) bool {
	if e, o := err.(*mysql.MySQLError); o {
		return e.Number == 1169
	}
	return false
}
{{end -}}

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

{{range .TPL -}}
{{.}}
{{- end}}
`))

type fileTPL struct {
	Pkg     string
	Driver  string // mysql driver
	Strings bool   // import
	Sql     []string
	TPL     []TPL
}

func (t *fileTPL) Execute(w io.Writer) error {
	return _fileTPL.Execute(w, t)
}

func (t *fileTPL) String() string {
	var str strings.Builder
	_ = t.Execute(&str)
	return str.String()
}
