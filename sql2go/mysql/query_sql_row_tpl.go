package mysql

import (
	"database/sql"
	"io"
	"text/template"
)

var _querySqlNullRowTPL = template.Must(template.New("querySqlNullRowTPL").Parse(`
// {{.Sql}}
func {{.Func}}({{.ParamTPL}}) ({{.Scan.NullType}}, error) {
	var str strings.Builder
	{{- range .Segment}}
	str.WriteString({{.}})
	{{- end}}
	var model {{.Scan.NullType}}
	return model, {{.StmtTPL}}.QueryRow(
		{{- range .Param}}
		{{.}},
		{{- end}}
	).Scan(&model)
}`))

var _querySqlRowTPL = template.Must(template.New("querySqlRowTPL").Parse(`
// {{.Sql}}
func {{.Func}}({{.ParamTPL}}) ({{.Scan.Type}}, error) {
	var str strings.Builder
	{{- range .Segment}}
	str.WriteString({{.}})
	{{- end}}
	{{- if .Scan.NullType}}
	var model {{.Scan.NullType}}
	err := {{.StmtTPL}}.QueryRow(
		str.String(),
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

type querySqlRowTPL struct {
	querySqlTPL
	Scan *scanTPL
}

func (t *querySqlRowTPL) Execute(w io.Writer) error {
	if t.NullField {
		return _querySqlNullRowTPL.Execute(w, t)
	}
	return _querySqlRowTPL.Execute(w, t)
}

func (t *querySqlRowTPL) InitScan(column *sql.ColumnType) {
	t.Scan = dbColumnToScanTPL(column, t.NullField)
	// 如果数据库字段不为null，那么NullField无效
	if t.NullField {
		t.NullField = t.Scan.NullType != ""
	}
}
