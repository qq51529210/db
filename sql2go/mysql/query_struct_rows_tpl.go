package mysql

import (
	"io"
	"text/template"
)

var _queryNullFieldStructRowsTPL = template.Must(template.New("queryNullFieldStructRowsTPL").Parse(`
type {{.Func}}Model struct {
	{{- range .Field}}
	{{index . 0}} {{index . 1}} {{index . 2}}
	{{- end}}
}

// {{.Sql}}
func {{.Func}}({{.ParamTPL}}) ([]*{{.Func}}Model, error) {
	rows, err := {{.StmtTPL}}.Query(
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

var _queryStructRowsTPL = template.Must(template.New("queryStructRowsTPL").Parse(`
type {{.Func}}Model struct {
	{{- range .Field}}
	{{index . 0}} {{index . 1}} {{index . 2}}
	{{- end}}
}

// {{.Sql}}
func {{.Func}}({{.ParamTPL}}) ([]*{{.Func}}Model, error) {
	rows, err := {{.StmtTPL}}.Query(
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

type queryStructRowsTPL struct {
	queryStructTPL
}

func (t *queryStructRowsTPL) Execute(w io.Writer) error {
	if t.NullField {
		return _queryNullFieldStructRowsTPL.Execute(w, t)
	}
	return _queryStructRowsTPL.Execute(w, t)
}
