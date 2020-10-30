package tpl

import (
	"io"
	"strconv"
)

const tplStrQuery = `// {{.Sql}}
func {{.Func}} ({{.TPLParam}}) ({{.ReturnType}}, error) {
	rows, err := {{.TPLStmt}}.Query(
		{{- range .Params}}
		{{.}},
		{{- end}}
	)
	if nil != err {
		return {{range .Return}}nil, {{end}} err
	}
	{{- range $i,$s := .Return}}
	var ms{{$i}}{{$i}} []{{$s}}
	var m{{$i}}{{$i}} {{$s}}
	{{- end}}
	for rows.Next() {
		err = rows.Scan(
			{{- range $i,$s := .Return}}
			&m{{$i}}{{$i}},
			{{- end}}
		)
		if nil != err {
			return {{range .Return}}nil, {{end}} err
		}
		{{- range $i,$s := .Return}}
		ms{{$i}}{{$i}} = append(ms{{$i}}{{$i}}, m{{$i}}{{$i}})
		{{- end}}
	}
	return {{.ReturnVar}}, nil
}
`

type Query struct {
	funcTPL
	Return []string
}

func (t *Query) Execute(w io.Writer) error {
	return tplQuery.Execute(w, t)
}

func (t *Query) String() string {
	return tplString(t)
}

func (t *Query) ReturnType() string {
	s := "[]" + t.Return[0]
	for i := 1; i < len(t.Return); i++ {
		s += ", []" + t.Return[i]
	}
	return s
}

func (t *Query) ReturnVar() string {
	s := "ms00"
	for i := 1; i < len(t.Return); i++ {
		s += ", ms" + strconv.Itoa(i) + strconv.Itoa(i)
	}
	return s
}

const tplStrQueryRow = `// {{.Sql}}
func {{.Func}}({{.TPLParam}}) {{.ReturnString}} {
	{{- range $i,$s := .Return}}
	var m{{$i}} {{$s}}
	{{- end}}
	err := {{.TPLStmt}}.QueryRow(
		{{- range .Params}}
		{{.}},
		{{- end}}
	).Scan(
		{{- range $i,$s := .Return}}
		&m{{$i}},
		{{- end}}
	)
	return {{- range $i,$s := .Return}} m{{$i}}, {{- end}} err
}
`

type QueryRow struct {
	funcTPL
	Return []string
}

func (t *QueryRow) Execute(w io.Writer) error {
	return tplQueryRow.Execute(w, t)
}

func (t *QueryRow) String() string {
	return tplString(t)
}

func (t *QueryRow) ReturnString() string {
	if len(t.Return) < 1 {
		return "error"
	}
	str := "("
	for i := 0; i < len(t.Return); i++ {
		str += t.Return[i] + ", "
	}
	str += "error)"
	return str
}

const tplStrStructQuery = `// {{.Sql}}
func {{.Func}}({{.TPLParam}}) ([]*{{.Struct}}, error) {
	rows, err := {{.TPLStmt}}.Query(
		{{- range .Params}}
		{{.}},
		{{- end}}
	)
	if nil != err {
		return nil, err
	}
	var models []*{{.Struct}}
	for rows.Next() {
		model := new({{.Struct}})
		err = rows.Scan(
			{{- range .Scan}}
			&model.{{.}},
			{{- end}}
		)
		if nil != err {
			return nil, err
		}
		models = append(models, model)
	}
	return models, nil
}
`

type StructQuery struct {
	funcTPL
	Scan []string
}

func (t *StructQuery) Execute(w io.Writer) error {
	return tplStructQuery.Execute(w, t)
}

func (t *StructQuery) String() string {
	return tplString(t)
}

const tplStrStructQueryRow = `// {{.Sql}}
func (m *{{.Struct}}) {{.Func}}({{.TPLParam}}) error {
	return {{.TPLStmt}}.QueryRow(
		{{- range .Arg}}
		{{.}},
		{{- end}}
	).Scan(
		{{- range .Scan}}
		&m.{{.}},
		{{- end}}
	)
}
`

type StructQueryRow struct {
	funcTPL
	Scan []string
}

func (t *StructQueryRow) Execute(w io.Writer) error {
	return tplStructQueryRow.Execute(w, t)
}

func (t *StructQueryRow) String() string {
	return tplString(t)
}
