package mysql

import (
	"io"
	"text/template"
)

var _querySqlNullFieldStructRowTPL = template.Must(template.New("querySqlNullFieldStructRowTPL").Parse(`
type {{.Func}}Model struct {
	{{- range .Field}}
	{{index . 0}} {{index . 1}} {{index . 2}}
	{{- end}}
}

// {{.Sql}}
func {{.Func}}({{.ParamTPL}}) (*{{.Func}}Model, error) {
	var str strings.Builder
	{{- range .Segment}}
	str.WriteString({{.}})
	{{- end}}
	model := new({{.Func}}Model)
	return model, {{.StmtTPL}}.QueryRow(
		str.String(),
		{{- range .Param}}
		{{.}},
		{{- end}}
	).Scan(
		{{- range .Scan}}
		&model.{{.Name}},
		{{- end}}
	)
}`))

var _querySqlStructRowTPL = template.Must(template.New("querySqlStructRowTPL").Parse(`
type {{.Func}}Model struct {
	{{- range .Field}}
	{{index . 0}} {{index . 1}} {{index . 2}}
	{{- end}}
}

// {{.Sql}}
func {{.Func}}({{.ParamTPL}}) (*{{.Func}}Model, error) {
	var str strings.Builder
	{{- range .Segment}}
	str.WriteString({{.}})
	{{- end}}
	{{- range .Scan}}
	{{- if .NullType}}
	var {{.Name}} {{.NullType}}
	{{- end}}
	{{- end}}
	model := new({{.Func}}Model)
	err := {{.StmtTPL}}.QueryRow(
		str.String(),
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

type querySqlStructRowTPL struct {
	querySqlTPL
	Scan []*scanTPL
}

func (t *querySqlStructRowTPL) Execute(w io.Writer) error {
	if t.NullField {
		return _querySqlNullFieldStructRowTPL.Execute(w, t)
	}
	return _querySqlStructRowTPL.Execute(w, t)
}
