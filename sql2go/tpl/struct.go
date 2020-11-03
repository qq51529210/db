package tpl

import (
	"text/template"
)

var (
	tplStruct = template.Must(template.New("Struct").Parse(`package {{.Pkg}}

import (
	"database/sql"
	{{- range $k,$v := .ImportPkg}}	
	"{{$k}}"
	{{- end}}
)

type {{.Name}} struct {
	{{- range .Field}}
	{{index . 0}} {{index . 1}} {{index . 2}}
	{{- end}}
}

func Read{{.Name}}List(query string, args... interface{}) ([]*{{.Name}}, error) {
	var models []*{{.Name}}
	rows, err := DB.Query(query, args...)
	if nil != err {
		if err != sql.ErrNoRows {
			return nil, err
		}
		return models, nil
	}
	for rows.Next() {
		model := new({{.Name}})
		err := rows.Scan(
			{{- range .Fields}}
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

{{if .Table -}}
func {{.Name}}Page(order, sort, begin, total string) ([]*{{.Name}}, error) {
	var str strings.Builder
	str.WriteString("select * from {{.Table}} order by ")
	str.WriteString(order)
	str.WriteString(" ")
	str.WriteString(sort)
	str.WriteString(" limit ")
	str.WriteString(begin)
	str.WriteString(", ")
	str.WriteString(total)
	return Read{{.Name}}List(str.String())
}
{{end -}}

{{range .FuncTPL -}}
{{.}}
{{end -}}
`))
)

type Struct interface {
	TPL
	Name() string
	AddField(name, _type, tag string)
	AddFunc(function Func)
	Save(file string) error
}

func NewStruct(pkg, table, name string) Struct {
	s := new(_struct)
	s.Pkg = pkg
	s.Table = table
	s.name = name
	s.ImportPkg = make(map[string]int)
	if table != "" {
		s.ImportPkg["strings"] = 1
	}
	return s
}

type _struct struct {
	Pkg       string
	ImportPkg map[string]int
	name      string
	Table     string
	Field     [][3]string
	FuncTPL   []Func
}

func (s *_struct) Name() string {
	return s.name
}

func (s *_struct) AddFunc(f Func) {
	s.FuncTPL = append(s.FuncTPL, f)
}

func (s *_struct) AddField(name, _type, tag string) {
	s.Field = append(s.Field, [3]string{name, _type, tag})
}

func (s *_struct) Save(file string) error {
	return saveFile(tplStruct, s, file)
}

func (s *_struct) TPL() *template.Template {
	return tplStruct
}

func (s *_struct) Fields() []string {
	var fs []string
	for _, f := range s.Field {
		fs = append(fs, f[0])
	}
	return fs
}
