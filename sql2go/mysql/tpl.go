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
	"time"
	{{- range $k,$v := .Ipt}}
	"{{$k}}"
	{{- end}}
)

var (
	DB *sql.DB
	{{- range $i,$s := .Sql}}
	stmt{{$i}} *sql.Stmt // {{$s}}
	{{- end}}
)

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
	Pkg string
	Ipt map[string]int
	Sql []string
	TPL []TPL
}

func (t *fileTPL) Execute(w io.Writer) error {
	return _fileTPL.Execute(w, t)
}

func (t *fileTPL) String() string {
	var str strings.Builder
	_ = t.Execute(&str)
	return str.String()
}

//
//import (
//	"io"
//	"strconv"
//	"strings"
//	"text/template"
//)
//
//var (
//	tplInit = template.Must(template.New("Init").Parse(`package {{.Pkg}}
//import (
//	"database/sql"
//	{{if ne .DBPkg "github.com/go-sql-driver/mysql" -}}_ {{end}}"{{.DBPkg}}"
//	"time"
//)
//var (
//	DB *sql.DB
//	{{- range $i,$s := .Sql}}
//	stmt{{$i}} *sql.Stmt // {{$s}}
//	{{- end}}
//)
//{{- if eq .DBPkg "github.com/go-sql-driver/mysql"}}
//func IsUniqueError(err error) bool {
//	sqlErr, ok := err.(*mysql.MySQLError)
//	if ok {
//		ok = sqlErr.Number == 1062
//	}
//	return ok
//}
//{{- end}}
//func Init(url string, maxOpen, maxIdle int, maxLifeTime, maxIdleTime time.Duration) (err error){
//	DB, err = sql.Open("{{.DBType}}", url)
//	if err != nil {
//		return err
//	}
//	DB.SetMaxOpenConns(maxOpen)
//	DB.SetMaxIdleConns(maxIdle)
//	DB.SetConnMaxLifetime(maxLifeTime)
//	DB.SetConnMaxIdleTime(maxIdleTime)
//	return PrepareStmt(DB)
//}
//func UnInit() {
//	if DB == nil {
//		return
//	}
//	_ = DB.Close()
//	CloseStmt()
//}
//func PrepareStmt(db *sql.DB) (err error) {
//	{{- range $i,$s:= .Sql}}
//	stmt{{$i}}, err = db.Prepare("{{$s}}")
//	if err != nil {
//		return
//	}
//	{{- end}}
//	return
//}
//func CloseStmt() {
//	{{- range $i,$s:= .Sql}}
//	if stmt{{$i}} != nil {
//		_ = stmt{{$i}}.Close()
//	}
//	{{- end}}
//}
//`))
//
//	tplStruct = template.Must(template.New("Struct").Parse(`package {{.Pkg}}
//import (
//	"database/sql"
//	"strings"
//	{{- range $k,$v := .Import}}
//	"{{$k}}"
//	{{- end}}
//)
//type {{.Name}} struct {
//	{{- range .Field}}
//	{{index . 0}} {{index . 1}} {{index . 2}}
//	{{- end}}
//}
//func Read{{.Name}}List(query string, args... interface{}) ([]*{{.Name}}, error) {
//	models := make([]*{{.Name}}, 0)
//	rows, err := DB.Query(query, args...)
//	if nil != err {
//		if err != sql.ErrNoRows {
//			return nil, err
//		}
//		return models, nil
//	}
//	{{- range .Scan}}
//	{{- if .NullType}}
//	var {{.Name}} {{.NullType}}
//	{{- end}}
//	{{- end}}
//	for rows.Next() {
//		model := new({{.Name}})
//		err := rows.Scan(
//			{{- range .Scan}}
//			{{- end}}
//			{{- range .Scan}}
//			{{- if .NullType}}
//			&{{.Name}},
//			{{- else}}
//			&model.{{.Name}},
//			{{- end}}
//			{{- end}}
//		)
//		if nil != err {
//			return nil, err
//		}
//		{{- range .Scan}}
//		{{- if .NullType}}
//		if {{.Name}}.Valid {
//			model.{{.Name}} = {{.Name}}.{{.NullValue}}
//		}
//		{{- end}}
//		{{- end}}
//		models = append(models, model)
//	}
//	return models, nil
//}
//{{range .FuncTPL -}}
//{{.}}
//{{end -}}
//`))
//
//	tplExec = template.Must(template.New("Exec").Parse(`// {{.SQL}}
//func {{.FuncName}}({{.Params}}) (sql.Result, error) {
//	return {{.Stmt}}.Exec(
//		{{- range .Args}}
//		{{.}},
//		{{- end}}
//	)
//}
//`))
//
//	tplExecStruct = template.Must(template.New("ExecStruct").Parse(`// {{.SQL}}
//func (m *{{.Struct}}) {{.FuncName}}({{.Params}}) (sql.Result, error) {
//	return {{.Stmt}}.Exec(
//		{{- range .Args}}
//		{{.}},
//		{{- end}}
//	)
//}
//`))
//
//	tplQueryFunc = template.Must(template.New("QueryFunc").Parse(`
//{{- /**/ -}}
//// {{.SQL}}
//func {{.FuncName}}({{.Params}}) ({{- range $i,$s := .Type}}{{$s}}, {{- end}} error) {
//	{{- range $i,$s := .Type}}
//	var m{{$i}} {{$s}}
//	{{- end}}
//	err := {{.Stmt}}.QueryRow(
//		{{- range .Queries}}
//		{{.}},
//		{{- end}}
//	).Scan(
//		{{- range $i,$s := .Type}}
//		&m{{$i}},
//		{{- end}}
//	)
//	return {{- range $i,$s := .Type}} m{{$i}}, {{- end}} err
//}
//`))
//
//	tplQueryStructRow = template.Must(template.New("QueryStructRow").Parse(`
//{{- if .Model -}}
//type {{.Struct}}{{.FuncName}}Model struct {
//	{{- range .Scan}}
//	{{.Name}} {{.Type}} {{.Tag}}
//	{{- end}}
//}
//{{- end}}
//// {{.SQL}}
//{{if .Model -}}
//func (m *{{.Struct}}{{.FuncName}}Model) {{.FuncName}}({{.Params}}) error {
//{{- else -}}
//func (m *{{.Struct}}) {{.FuncName}}({{.Params}}) error {
//{{- end -}}
//	{{- range .Scan}}
//	{{- if .NullType}}
//	var {{.Name}} {{.NullType}}
//	{{- end}}
//	{{- end}}
//	{{- if .Scan.HasNull}}
//	err := {{.Stmt}}.QueryRow(
//		{{- range .Args}}
//		{{.}},
//		{{- end}}
//	).Scan(
//		{{- range .Scan}}
//		{{- if .NullType}}
//		&{{.Name}},
//		{{- else}}
//		&m.{{.Name}},
//		{{- end}}
//		{{- end}}
//	)
//	if err != nil {
//		return err
//	}
//	{{- range .Scan}}
//	{{- if .NullType}}
//	if {{.Name}}.Valid {
//		m.{{.Name}} = {{.Name}}.{{.NullValue}}
//	}
//	{{- end}}
//	{{- end}}
//	return nil
//	{{- else}}
//	return {{.Stmt}}.QueryRow(
//		{{- range .Args}}
//		{{.}},
//		{{- end}}
//	).Scan(
//		{{- range .Scan}}
//		&m.{{.Name}},
//		{{- end}}
//	)
//	{{- end}}
//}
//`))
//
//	tplQueryStructList = template.Must(template.New("QueryStructList").Parse(`
//{{- if .Model -}}
//type {{.FuncName}}Model struct {
//	{{- range .Scan}}
//	{{.Name}} {{.Type}} {{.Tag}}
//	{{- end}}
//}
//// {{.SQL}}
//func {{.FuncName}}({{.Params}}) ([]*{{.FuncName}}Model, error) {
//{{- else -}}
//// {{.SQL}}
//func {{.FuncName}}({{.Params}}) ([]*{{.Struct}}, error) {
//{{- end}}
//	{{- if .Scan}}
//	ms := make([]*{{.FuncName}}Model, 0)
//	rows, err := {{.Stmt}}.Query(
//		{{- range .Queries}}
//		{{.}},
//		{{- end}}
//	)
//	if nil != err {
//		if err != sql.ErrNoRows {
//			return nil, err
//		}
//		return ms, nil
//	}
//	{{- range .Scan}}
//	{{- if .NullType}}
//	var {{.Name}} {{.NullType}}
//	{{- end}}
//	{{- end}}
//	for rows.Next() {
//		{{- if .Model}}
//		m := new({{.FuncName}}Model)
//		{{- else}}
//		m := new({{.Struct}})
//		{{- end}}
//		err := rows.Scan(
//			{{- range .Scan}}
//			{{- if .NullType}}
//			&{{.Name}},
//			{{- else}}
//			&m.{{.Name}},
//			{{- end}}
//			{{- end}}
//		)
//		if nil != err {
//			return nil, err
//		}
//		{{- range .Scan}}
//		{{- if .NullType}}
//		if {{.Name}}.Valid {
//			m.{{.Name}} = {{.Name}}.{{.NullValue}}
//		}
//		{{- end}}
//		{{- end}}
//		ms = append(ms, m)
//	}
//	return ms, nil
//	{{- else}}
//	return Read{{.Struct}}List(str.String(), {{.Join .Args.Params}})
//	{{- end}}
//}
//`))
//
//	tplQueryStructPage = template.Must(template.New("QueryStructPage").Parse(`
//{{- if .Model -}}
//type {{.FuncName}}Model struct {
//	{{- range .Scan}}
//	{{.Name}} {{.Type}} {{.Tag}}
//	{{- end}}
//}
//func {{.FuncName}}({{.Params}}) ([]*{{.FuncName}}Model, error) {
//{{- else -}}
//func {{.FuncName}}({{.Params}}) ([]*{{.Struct}}, error) {
//{{- end}}
//	var str strings.Builder
//	{{- range .Segment}}
//	str.WriteString({{.}})
//	{{- end}}
//	{{- if .Scan}}
//	ms := make([]*{{.FuncName}}Model, 0)
//	rows, err := DB.Query(str.String(), {{.Join .Args.Params}})
//	if nil != err {
//		if err != sql.ErrNoRows {
//			return nil, err
//		}
//		return ms, nil
//	}
//	{{- range .Scan}}
//	{{- if .NullType}}
//	var {{.Name}} {{.NullType}}
//	{{- end}}
//	{{- end}}
//	for rows.Next() {
//		{{- if .Model}}
//		m := new({{.FuncName}}Model)
//		{{- else}}
//		m := new({{.Struct}})
//		{{- end}}
//		err := rows.Scan(
//			{{- range .Scan}}
//			{{- if .NullType}}
//			&{{.Name}},
//			{{- else}}
//			&m.{{.Name}},
//			{{- end}}
//			{{- end}}
//		)
//		if nil != err {
//			return nil, err
//		}
//		{{- range .Scan}}
//		{{- if .NullType}}
//		if {{.Name}}.Valid {
//			m.{{.Name}} = {{.Name}}.{{.NullValue}}
//		}
//		{{- end}}
//		{{- end}}
//		ms = append(ms, m)
//	}
//	return ms, nil
//	{{- else}}
//	return Read{{.Struct}}List(str.String(), {{.Join .Args.Params}})
//	{{- end}}
//}
//`))
//)
//
//type TPL interface {
//	Execute(writer io.Writer) error
//}
//
//type StructTPL interface {
//	TPL
//	Name() string
//	addFunc(tpl FuncTPL)
//}
//
//type FuncTPL interface {
//	TPL
//	SQL() string
//	FuncName() string
//	setStmt(stmt string)
//	tableName() string
//}
//
//type argTPL struct {
//	name  string
//	param bool
//}
//
//func (a *argTPL) String() string {
//	if a.param {
//		return a.name
//	}
//	return "m." + a.name
//}
//
//type argsTPL []*argTPL
//
//func (a argsTPL) HasField() bool {
//	for _, argTPL := range a {
//		if !argTPL.param {
//			return true
//		}
//	}
//	return false
//}
//
//func (a argsTPL) Field() []string {
//	var ss []string
//	for _, argTPL := range a {
//		if !argTPL.param {
//			ss = append(ss, argTPL.name)
//		}
//	}
//	return ss
//}
//
//func (a argsTPL) Params() []string {
//	var ss []string
//	for _, argTPL := range a {
//		if argTPL.param {
//			ss = append(ss, argTPL.name)
//		}
//	}
//	return ss
//}
//
//func (a argsTPL) Split() (field, param []string) {
//	for _, argTPL := range a {
//		if argTPL.param {
//			param = append(param, argTPL.name)
//		} else {
//			field = append(field, argTPL.name)
//		}
//	}
//	return
//}
//
//func (a argsTPL) ToParam() {
//	for _, aa := range a {
//		if !aa.param {
//			aa.name = pascalCaseToCamelCase(aa.name)
//			aa.param = true
//		}
//	}
//}
//
//type scanTPL struct {
//	Name      string
//	Type      string
//	Tag       string
//	NullType  string
//	NullValue string
//}
//
//type scansTPL []*scanTPL
//
//func (s scansTPL) HasNull() bool {
//	for _, t := range s {
//		if t.NullType != "" {
//			return true
//		}
//	}
//	return false
//}
//
//type initTPL struct {
//	Pkg    string
//	Sql    []string
//	DBType string
//	DBPkg  string
//}
//
//func (t *initTPL) Execute(writer io.Writer) error {
//	return tplInit.Execute(writer, t)
//}
//
//func (t *initTPL) addSql(sql string) string {
//	t.Sql = append(t.Sql, sql)
//	return "stmt" + strconv.Itoa(len(t.Sql)-1)
//}
//
//type structTPL struct {
//	pkg     string
//	Import  map[string]int
//	name    string
//	Field   [][3]string
//	FuncTPL []FuncTPL
//	Table   string
//	Scan    []*scanTPL
//}
//
//func (t *structTPL) Execute(writer io.Writer) error {
//	return tplStruct.Execute(writer, t)
//}
//
//func (t *structTPL) Pkg() string {
//	return t.pkg
//}
//
//func (t *structTPL) Name() string {
//	return t.name
//}
//
//func (t *structTPL) addFunc(tpl FuncTPL) {
//	t.FuncTPL = append(t.FuncTPL, tpl)
//}
//
//func (t *structTPL) addField(name, _type, tag string) {
//	t.Field = append(t.Field, [3]string{name, _type, tag})
//}
//
//type funcTPL struct {
//	table    string
//	sql      string
//	funcName string
//	tx       bool
//	stmt     string
//	args     argsTPL
//}
//
//func (t *funcTPL) SQL() string {
//	return t.sql
//}
//
//func (t *funcTPL) FuncName() string {
//	return t.funcName
//}
//
//func (t *funcTPL) Stmt() string {
//	if t.tx {
//		var s strings.Builder
//		s.WriteString("tx.Stmt(")
//		s.WriteString(t.stmt)
//		s.WriteString(")")
//		return s.String()
//	}
//	return t.stmt
//}
//
//func (t *funcTPL) setStmt(stmt string) {
//	t.stmt = stmt
//}
//
//func (t *funcTPL) Params() string {
//	var s strings.Builder
//	if t.tx {
//		s.WriteString("tx *sql.Tx")
//	}
//	params := t.args.Params()
//	if len(params) < 1 {
//		return s.String()
//	}
//	if t.tx {
//		s.WriteString(", ")
//	}
//	s.WriteString(strings.Join(params, ", "))
//	s.WriteString(" interface{}")
//	return s.String()
//}
//
//func (t *funcTPL) tableName() string {
//	return t.table
//}
//
//func (t *funcTPL) Args() argsTPL {
//	return t.args
//}
//
//func (t *funcTPL) toString(tpl TPL) string {
//	var str strings.Builder
//	_ = tpl.Execute(&str)
//	return str.String()
//}
//
//func (t *funcTPL) Join(ss []string) string {
//	return strings.Join(ss, ", ")
//}
//
//type execTPL struct {
//	funcTPL
//}
//
//func (t *execTPL) Execute(writer io.Writer) error {
//	return tplExec.Execute(writer, t)
//}
//
//func (t *execTPL) String() string {
//	return t.toString(t)
//}
//
//type execStructTPL struct {
//	funcTPL
//	Struct string
//}
//
//func (t *execStructTPL) Execute(writer io.Writer) error {
//	return tplExecStruct.Execute(writer, t)
//}
//
//func (t *execStructTPL) String() string {
//	return t.toString(t)
//}
//
//type queryFuncTPL struct {
//	funcTPL
//	Type []string
//}
//
//func (t *queryFuncTPL) Execute(writer io.Writer) error {
//	return tplQueryFunc.Execute(writer, t)
//}
//
//func (t *queryFuncTPL) String() string {
//	return t.toString(t)
//}
//
//func (t *queryFuncTPL) Queries() []string {
//	return t.args.Params()
//}
//
//type queryStructRowTPL struct {
//	funcTPL
//	Struct string
//	Scan   scansTPL
//	Model  bool
//}
//
//func (t *queryStructRowTPL) Execute(writer io.Writer) error {
//	return tplQueryStructRow.Execute(writer, t)
//}
//
//func (t *queryStructRowTPL) String() string {
//	return t.toString(t)
//}
//
//type queryStructListTPL struct {
//	funcTPL
//	Struct string
//	Scan   scansTPL
//	Model  bool
//}
//
//func (t *queryStructListTPL) Execute(writer io.Writer) error {
//	return tplQueryStructList.Execute(writer, t)
//}
//
//func (t *queryStructListTPL) String() string {
//	return t.toString(t)
//}
//
//func (t *queryStructListTPL) Queries() []string {
//	return t.args.Params()
//}
//
//type queryStructPageTPL struct {
//	funcTPL
//	ColumnParam []string
//	Segment     []string
//	Struct      string
//	Scan        scansTPL
//	Model       bool
//}
//
//func (t *queryStructPageTPL) Execute(writer io.Writer) error {
//	return tplQueryStructPage.Execute(writer, t)
//}
//
//func (t *queryStructPageTPL) String() string {
//	return t.toString(t)
//}
//
//func (t *queryStructPageTPL) Params() string {
//	var s strings.Builder
//	if t.tx {
//		s.WriteString("tx *sql.Tx")
//	}
//	if len(t.ColumnParam) > 0 {
//		s.WriteString(strings.Join(t.ColumnParam, ", "))
//		s.WriteString(" string")
//	}
//	params := t.args.Params()
//	if len(params) < 1 {
//		return s.String()
//	}
//	if s.Len() > 0 {
//		s.WriteString(", ")
//	}
//	s.WriteString(strings.Join(params, ", "))
//	s.WriteString(" interface{}")
//	return s.String()
//}
