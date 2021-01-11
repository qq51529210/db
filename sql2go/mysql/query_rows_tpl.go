package mysql

import (
	"io"
	"text/template"
)

var _queryNullRowsTPL = template.Must(template.New("queryNullRowsTPL").Parse(`
// {{.Sql}}
func {{.Func}}({{.ParamTPL}}) ([]{{.Scan.NullType}}, error) {
	rows, err := {{.StmtTPL}}.Query(
		{{- range .Param}}
		{{.}},
		{{- end}}
	)
	if nil != err {
		return nil, err
	}
	var models []{{.Scan.NullType}}
	var model {{.Scan.NullType}}
	for rows.Next() {
		err = rows.Scan(&model)
		if nil != err {
			return nil, err
		}
		models = append(models, model)
	}
	return models, nil
}`))

var _queryRowsTPL = template.Must(template.New("queryRowsTPL").Parse(`
// {{.Sql}}
func {{.Func}}({{.ParamTPL}}) ([]{{.Scan.Type}}, error) {
	rows, err := {{.StmtTPL}}.Query(
		{{- range .Param}}
		{{.}},
		{{- end}}
	)
	if nil != err {
		return nil, err
	}
	var models []{{.Scan.Type}}
	{{- if .Scan.NullType}}
	var model {{.Scan.NullType}}
	{{- else}}
	var model {{.Scan.Type}}
	{{- end}}
	for rows.Next() {
		err = rows.Scan(&model)
		if nil != err {
			return nil, err
		}
		{{- if .Scan.NullType}}
		{{- if eq .Scan.Type .Scan.NullType2}}
		models = append(models, model.{{.Scan.NullValue}})
		{{- else}}
		models = append(models, {{.Scan.Type}}(model.{{.Scan.NullValue}}))
		{{- end}}
		{{- else}}
		models = append(models, model)
		{{- end}}
	}
	return models, nil
}`))

type queryRowsTPL struct {
	queryRowTPL
}

func (t *queryRowsTPL) Execute(w io.Writer) error {
	if t.NullField {
		return _queryNullRowsTPL.Execute(w, t)
	}
	return _queryRowsTPL.Execute(w, t)
}
