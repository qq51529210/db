package tpl

import (
	"strings"
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
	{{- range $i,$v := .Field}}
	{{$.TPLField $v}}
	{{- end}}
}

func Read{{.Name}}List(rows *sql.Rows) ([]*{{.Name}}, error) {
	var models []*{{.Name}}
	for rows.Next() {
		model := new({{.Name}})
		err = rows.Scan(
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

func NewStruct(pkg, name string) Struct {
	s := new(_struct)
	s.Pkg = pkg
	s.name = name
	s.ImportPkg = make(map[string]int)
	return s
}

type _struct struct {
	Pkg       string
	ImportPkg map[string]int
	name      string
	Field     [][3]string
	FuncTPL   []Func
}

func (s *_struct) Name() string {
	return s.name
}

func (s *_struct) AddFunc(f Func) {
	//switch f.(type) {
	//case *Exec:
	//	s.ImportPkg["database/sql"] = 1
	//}
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

func (s *_struct) TPLField(field [3]string) string {
	var str strings.Builder
	str.WriteString(field[0])
	str.WriteByte(' ')
	str.WriteString(field[1])
	if field[2] != "" {
		str.WriteByte(' ')
		str.WriteByte('`')
		str.WriteString(field[2])
		str.WriteByte('`')
	}
	return str.String()
}

func (s *_struct) Fields() []string {
	var fs []string
	for _, f := range s.Field {
		fs = append(fs, f[0])
	}
	return fs
}
