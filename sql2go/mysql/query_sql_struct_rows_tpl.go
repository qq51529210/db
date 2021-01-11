package mysql

import (
	"io"
	"text/template"
)

var _querySqlNullFieldStructRowsTPL = template.Must(template.New("querySqlNullFieldStructRowsTPL").Parse(`
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
	rows, err := {{.StmtTPL}}.Query(
		str.String(),
		{{- range .Param}}
		{{.}},
		{{- end}}
	)
	if nil != err {
		return nil, err
	}
	var models []*{{.Func}}Model
	for rows.Next() {
		model := new({{.Func}}Model)
		err = rows.Scan(
			{{- range .Scan}}
			&model.{{.Name}},
			{{- end}}
		)
		if nil != err {
			return nil, err
		}
		models = append(models, model)
	}
	return models, nil
}`))

var _querySqlStructRowsTPL = template.Must(template.New("querySqlStructRowsTPL").Parse(`
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
	rows, err := {{.StmtTPL}}.Query(
		str.String(),
		{{- range .Param}}
		{{.}},
		{{- end}}
	)
	if nil != err {
		return nil, err
	}
	var models []*{{.Func}}Model
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
		{{- if eq .Type .NullType2}}
		model.{{.Name}} = {{.Name}}.{{.NullValue}}
		{{- else}}
		model.{{.Name}} = {{.Type}}({{.Name}}.{{.NullValue}})
		{{- end}}
		{{- end}}
		{{- end}}
		models = append(models, model)
	}
	return models, nil
}`))

type querySqlStructRowsTPL struct {
	querySqlStructTPL
}

func (t *querySqlStructRowsTPL) Execute(w io.Writer) error {
	if t.NullField {
		return _querySqlNullFieldStructRowsTPL.Execute(w, t)
	}
	return _querySqlStructRowsTPL.Execute(w, t)
}
