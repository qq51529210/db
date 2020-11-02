package tpl

//
//import (
//	"io"
//)
//
//const tplStrExec = `// {{.Sql}}
//func {{.Func}} ({{.TPLParam}}) (sql.Result, error) {
//	return {{.TPLStmt}}.Exec(
//		{{- range .Params}}
//		{{.}},
//		{{- end}}
//	)
//}
//`
//
//type Exec struct {
//	funcTPL
//}
//
//func (t *Exec) Execute(w io.Writer) error {
//	return tplExec.Execute(w, t)
//}
//
//func (t *Exec) String() string {
//	return tplString(t)
//}
//
//const tplStrStructExec = `// {{.Sql}}
//func (m *{{.Struct}}) {{.Func}}({{.TPLParam}}) (sql.Result, error) {
//	return {{.TPLStmt}}.Exec(
//		{{- range .Arg}}
//		{{.}},
//		{{- end}}
//	)
//}
//`
//
//type StructExec struct {
//	funcTPL
//}
//
//func (t *StructExec) Execute(w io.Writer) error {
//	return tplStructExec.Execute(w, t)
//}
//
//func (t *StructExec) String() string {
//	return tplString(t)
//}
