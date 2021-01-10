package mysql

import (
	"io"
	"text/template"
)

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
	tpl
}

func (t *execTPL) Execute(w io.Writer) error {
	return _execTPL.Execute(w, t)
}
