package mysql

import (
	"database/sql"
	"io"
	"text/template"
)

var _queryNullRowTPL = template.Must(template.New("queryNullRowTPL").Parse(`
// {{.Sql}}
func {{.Func}}({{.ParamTPL}}) ({{.Scan.NullType}}, error) {
	var model {{.Scan.NullType}}
	return model, {{.StmtTPL}}.QueryRow(
		{{- range .Param}}
		{{.}},
		{{- end}}
	).Scan(&model)
}`))

var _queryRowTPL = template.Must(template.New("queryRowTPL").Parse(`
// {{.Sql}}
func {{.Func}}({{.ParamTPL}}) ({{.Scan.Type}}, error) {
	{{- if .Scan.NullType}}
	var model {{.Scan.NullType}}
	err := {{.StmtTPL}}.QueryRow(
		{{- range .Param}}
		{{.}},
		{{- end}}
	).Scan(&model)
	if err != nil {
		return model.{{.Scan.NullValue}}, err
	}
	{{- if eq .Scan.Type .Scan.NullType2}}
	return model.{{.Scan.NullValue}}, nil
	{{- else}}
	return {{.Scan.Type}}(model.{{.Scan.NullValue}}), nil
	{{- end}}
	{{- else}}
	var model {{.Scan.Type}}
	return model, {{.StmtTPL}}.QueryRow(
		{{- range .Param}}
		{{.}},
		{{- end}}
	).Scan(&model)
	{{- end}}
}`))

type queryRowTPL struct {
	queryTPL
	Scan *scanTPL
}

func (t *queryRowTPL) Execute(w io.Writer) error {
	if t.NullField {
		return _queryNullRowTPL.Execute(w, t)
	}
	return _queryRowTPL.Execute(w, t)
}

func (t *queryRowTPL) InitScan(column *sql.ColumnType) {
	t.Scan = dbColumnToScanTPL(column, t.NullField)
	// 如果数据库字段不为null，那么NullField无效
	if t.NullField {
		t.NullField = t.Scan.NullType != ""
	}
}
