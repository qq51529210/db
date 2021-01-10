package mysql

import (
	"io"
	"strings"
	"text/template"
)

var _execStructTPL = template.Must(template.New("execStructTPL").Parse(`
type {{.Struct}} struct {
	{{- range .Field}}
	{{index . 0}} {{index . 1}} {{index . 2}}
	{{- end}}
}

// {{.Sql}}
func {{.Func}}({{.ParamTPL}}) (sql.Result, error) {
	return {{.StmtTPL}}.Exec(
		{{- range .Field}}
		model.{{index . 0}},
		{{- end}}
	)
}
`))

type execStructTPL struct {
	tpl
	Struct string
	Field  [][3]string
}

func (t *execStructTPL) Execute(w io.Writer) error {
	return _execStructTPL.Execute(w, t)
}

func (t *execStructTPL) ParamTPL() string {
	var s strings.Builder
	// tx *sql.Tx
	if t.Tx != "" {
		s.WriteString(t.Tx)
		s.WriteString(" *sql.Tx")
	}
	// tx *sql.Tx, model *Struct
	s.WriteString(", model *")
	s.WriteString(t.Struct)
	return s.String()
}
