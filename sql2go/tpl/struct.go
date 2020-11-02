// 结构体模板
package tpl

import (
	"io"
	"strings"
)

const tplStrStruct = `package {{.Pkg}}

{{$ip := .Import}}
{{- if $ip -}}
import (
{{- range $ip}}
	"{{.}}"
{{- end}}
)
{{- end}}

type {{.Name}} struct {
	{{- range .Field}}
	{{.Name}} {{.Type}} {{.Tag}}
	{{- end}}
}

{{range .TPL -}}
{{.}}
{{end -}}
`

type Field struct {
	Name string
	Type string
	Tag  string
}

type Struct struct {
	Pkg   string
	Name  string
	Field []*Field
	TPL   []FuncTPL
}

func (t *Struct) Import() []string {
	var s []string
	for _, f := range t.Field {
		if strings.Contains(f.Type, "sql") {
			s = append(s, "database/sql")
			break
		}
	}
	for _, tp := range t.TPL {
		ft, ok := tp.(FuncTPL)
		if ok || ft.IsTx() {
			if len(s) == 0 {
				s = append(s, "database/sql")
				break
			}
		}
	}
	for _, tp := range t.TPL {
		_, ok := tp.(*SortStructQuery)
		if ok {
			s = append(s, "strings")
			break
		}
	}
	return s
}

func (t *Struct) Execute(w io.Writer) error {
	return tplStruct.Execute(w, t)
}

func (t *Struct) StructName() string {
	return t.Name
}

func (t *Struct) AddFuncTPL(tpl FuncTPL) {
	t.TPL = append(t.TPL, tpl)
}

const tplStrJoinStruct = `package {{.Pkg}}

{{$ip := .Import}}
{{- if $ip}}
import (
{{- range $ip}}
	"{{.}}"
{{- end}}
}
{{- end}}

type {{.StructName}} struct {
	{{.Struct1.Name}}
	{{.Struct2.Name}}
}

{{range .TPL -}}
{{.}}
{{end -}}
`

type JoinStruct struct {
	Pkg     string
	Struct1 *Struct
	Struct2 *Struct
	TPL     []FuncTPL
}

func (t *JoinStruct) Import() []string {
	var s []string
	for _, tp := range t.TPL {
		ft, ok := tp.(FuncTPL)
		if ok || ft.IsTx() {
			s = append(s, "database/sql")
			break
		}
	}
	for _, tp := range t.TPL {
		_, ok := tp.(*SortStructQuery)
		if ok {
			s = append(s, "strings")
			break
		}
	}
	return s
}

func (t *JoinStruct) Execute(w io.Writer) error {
	return tplJoinStruct.Execute(w, t)
}

func (t *JoinStruct) StructName() string {
	return t.Struct1.Name + "Join" + t.Struct2.Name
}

func (t *JoinStruct) AddFuncTPL(tpl FuncTPL) {
	t.TPL = append(t.TPL, tpl)
}
