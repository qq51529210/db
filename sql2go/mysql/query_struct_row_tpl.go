package mysql

import (
	"io"
	"text/template"
)

var _queryNullFieldStructRowTPL = template.Must(template.New("queryNullFieldStructRowTPL").Parse(`
type {{.Func}}Model struct {
	{{- range .Field}}
	{{index . 0}} {{index . 1}} {{index . 2}}
	{{- end}}
}

// {{.Sql}}
func {{.Func}}({{.ParamTPL}}) (*{{.Func}}Model, error) {
	model := new({{.Func}}Model)
	return model, {{.StmtTPL}}.QueryRow(
		{{- range .Param}}
		{{.}},
		{{- end}}
	).Scan(
		{{- range .Scan}}
		&model.{{.Name}},
		{{- end}}
	)
}`))

var _queryStructRowTPL = template.Must(template.New("queryStructRowTPL").Parse(`
type {{.Func}}Model struct {
	{{- range .Field}}
	{{index . 0}} {{index . 1}} {{index . 2}}
	{{- end}}
}

// {{.Sql}}
func {{.Func}}({{.ParamTPL}}) (*{{.Func}}Model, error) {
	{{- range .Scan}}
	{{- if .NullType}}
	var {{.Name}} {{.NullType}}
	{{- end}}
	{{- end}}
	model := new({{.Func}}Model)
	err := {{.StmtTPL}}.QueryRow(
		{{- range .Param}}
		{{.}},
		{{- end}}
	).Scan(
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
		{{- if eq .Type .NullType2}}
		model.{{.Name}} = {{.Name}}.{{.NullValue}}
		{{- else}}
		model.{{.Name}} = {{.Type}}({{.Name}}.{{.NullValue}})
		{{- end}}
	}
	{{- end}}
	{{- end}}
	return model, nil
}`))

type queryStructRowTPL struct {
	queryStructTPL
}

func (t *queryStructRowTPL) Execute(w io.Writer) error {
	if t.NullField {
		return _queryNullFieldStructRowTPL.Execute(w, t)
	}
	return _queryStructRowTPL.Execute(w, t)
}
