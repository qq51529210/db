package mysql

import (
	"io"
	"text/template"
)

var _querySqlNullRowsTPL = template.Must(template.New("querySqlNullRowsTPL").Parse(`
// {{.Sql}}
func {{.Func}}({{.ParamTPL}}) ([]{{.Scan.NullType}}, error) {
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

var _querySqlRowsTPL = template.Must(template.New("querySqlRowsTPL").Parse(`
// {{.Sql}}
func {{.Func}}({{.ParamTPL}}) ([]{{.Scan.Type}}, error) {
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
	var models []{{.Scan}}
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
}
`))

type querySqlRowsTPL struct {
	querySqlTPL
	Scan *scanTPL
}

func (t *querySqlRowsTPL) Execute(w io.Writer) error {
	if t.NullField {
		return _querySqlNullRowsTPL.Execute(w, t)
	}
	return _querySqlRowsTPL.Execute(w, t)
}
