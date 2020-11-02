package tpl

import (
	"strings"
	"text/template"
)

var (
	tplSelectFunc = template.Must(template.New("SelectFunc").Parse(`// {{.Sql}}
func {{.Func}} ({{.TPLParam}}) ({{.TPLReturn}}) {
	{{- range $i,$s := .Type}}
	var m{{$i}} {{$s}}
	{{- end}}
	err := {{.TPLStmt}}.QueryRow(
		{{- range .Args.Params}}
		{{.}},
		{{- end}}
	).Scan(
		{{- range $i,$s := .Type}}
		&m{{$i}},
		{{- end}}
	)
	return {{- range $i,$s := .Type}} m{{$i}}, {{- end}} err
}
`))

	tplSelectColumn = template.Must(template.New("SelectColumn").Parse(`// {{.Sql}}
func (m *{{.Struct}}) {{.Func}}({{.TPLParam}}) error {
	return {{.TPLStmt}}.QueryRow(
		{{- range .Arg}}
		{{.}},
		{{- end}}
	).Scan(
		{{- range .Column}}
		&m.{{.}},
		{{- end}}
	)
}
`))

	tplSelectList = template.Must(template.New("SelectList").Parse(`// {{.Sql}}
func {{.Func}}({{.TPLParam}}) ([]*{{.Struct}}, error) {
	var models []*{{.Struct}}
	rows, err := {{.TPLStmt}}.Query(
		{{- range .Args.Params}}
		{{.}},
		{{- end}}
	)
	if nil != err {
		if err != sql.ErrNoRows {
			return nil, err
		}
		return models, nil
	}
	for rows.Next() {
		model := new({{.Struct}})
		err = rows.Scan(
			{{- range .Column}}
			&model.{{.}},
			{{- end}}
		)
		if nil != err {
			return nil, err
		}
		models = append(models, model)
	}
	return models, nil
}
`))

	tplSelectPage = template.Must(template.New("SelectPage").Parse(`// {{.Sql}}
func {{.Func}}({{.TPLParam}}) ([]*{{.Struct}}, error) {
	var str strings.Builder
	{{- if .BeforeGroup}}
	str.WriteString("{{.BeforeGroup}} ")
	{{- end}}

	{{- range $i, $s := .Group}}

	{{- if $i}}
	str.WriteString(",")
	{{- end}}
	str.WriteString({{$s}})
	str.WriteString(" ")
	{{- end}}

	{{- if .Group2Order}}
	str.WriteString("{{.Group2Order}} ")
	{{- end}}

	{{- range $i, $s := .Order}}

	{{- if $i}}
	str.WriteString(",")
	{{- end}}
	str.WriteString({{$s}})	
	str.WriteString(" ")
	{{- end}}

	{{- if .Sort}}
	str.WriteString("{{.Sort}} ")
	{{- end}}

	{{- if .AfterOrder}}
	str.WriteString("{{.AfterOrder}}")
	{{- end}}

	var models []*{{.Struct}}
	rows, err := {{.TPLStmt}}.Query(str.String(),
		{{- range .Params}}
		{{.}},
		{{- end}}
	)
	if nil != err {
		if err != sql.ErrNoRows {
			return nil, err
		}
		return models, nil
	}
	for rows.Next() {
		model := new({{.Struct}})
		err = rows.Scan(
			{{- range .Column}}
			&model.{{.}},
			{{- end}}
		)
		if nil != err {
			return nil, err
		}
		models = append(models, model)
	}
	return models, nil
}

`))

	tplExec = template.Must(template.New("Exec").Parse(`// {{.Sql}}
func (m *{{.Struct}}) {{.Func}}({{.TPLParam}}) (sql.Result, error) {
	return {{.TPLStmt}}.Exec(
		{{- range .Arg}}
		{{.}},
		{{- end}}
	)
}
`))
)

type Func interface {
	TPL
	SQL() string
	StructName() string
	SetFunc(name string)
	SetParam(names []string)
	SetStmt(name string)
	Args() Args
}

type funcTPL struct {
	Tx     bool
	Sql    string
	Func   string
	Struct string
	Arg    []*Arg
	Stmt   string
}

func (t *funcTPL) SQL() string {
	return t.Sql
}

func (t *funcTPL) StructName() string {
	return t.Struct
}

func (t *funcTPL) Args() Args {
	return t.Arg
}

func (t *funcTPL) SetFunc(name string) {
	t.Func = name
}

func (t *funcTPL) SetParam(names []string) {
	for _, a := range t.Arg {
		if len(names) < 1 {
			return
		}
		if !a.IsField {
			a.Name = names[0]
			names = names[1:]
		}
	}
}

func (t *funcTPL) SetStmt(name string) {
	t.Stmt = name
}

func (t *funcTPL) TPLParam() string {
	s := ""
	if t.Tx {
		s = "tx *sql.Tx"
	}
	p := t.Args().Params()
	if len(p) < 1 {
		return s
	}
	if t.Tx {
		s += ", "
	}
	s += strings.Join(p, ", ") + " interface{}"
	return s
}

func (t *funcTPL) TPLStmt() string {
	return t.Stmt
}

func (t *funcTPL) toString(tp *template.Template, v interface{}) string {
	var w strings.Builder
	_ = tp.Execute(&w, v)
	return w.String()
}

type SelectFunc struct {
	funcTPL
	Type []string
}

func (t *SelectFunc) TPLReturn() string {
	return strings.Join(t.Type, ", ") + ", error"
}

func (t *SelectFunc) TPL() *template.Template {
	return tplSelectFunc
}

func (t *SelectFunc) String() string {
	return t.toString(tplSelectFunc, t)
}

type SelectColumn struct {
	funcTPL
	Column []string
}

func (t *SelectColumn) TPL() *template.Template {
	return tplSelectColumn
}

func (t *SelectColumn) String() string {
	return t.toString(tplSelectColumn, t)
}

type SelectList struct {
	funcTPL
	Column []string
}

func (t *SelectList) TPL() *template.Template {
	return tplSelectList
}

func (t *SelectList) String() string {
	return t.toString(tplSelectList, t)
}

type SelectPage struct {
	funcTPL
	Column      []string
	BeforeGroup string
	Group       []string
	Group2Order string
	Order       []string
	Sort        string
	AfterOrder  string
}

func (t *SelectPage) TPL() *template.Template {
	return tplSelectPage
}

func (t *SelectPage) String() string {
	return t.toString(tplSelectPage, t)
}

func (t *SelectPage) TPLStmt() string {
	if t.Tx {
		return "tx"
	}
	return "DB"
}

func (t *SelectPage) Params() []string {
	var p []string
	for _, a := range t.Arg {
		if !a.IsField &&
			!containString(t.Group, a.Name) &&
			!containString(t.Order, a.Name) {
			p = append(p, a.Name)
		}
	}
	return p
}

type Exec struct {
	funcTPL
}

func (t *Exec) TPL() *template.Template {
	return tplExec
}

func (t *Exec) String() string {
	return t.toString(tplExec, t)
}

func containString(ss []string, s string) bool {
	for i := 0; i < len(ss); i++ {
		if ss[i] == s {
			return true
		}
	}
	return false
}
