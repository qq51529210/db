package tpl

import (
	"os"
	"testing"
)

var (
	testStruct = NewStruct("test", "Test")
)

func Test_SelectFunc(t *testing.T) {
	tp := new(SelectFunc)
	tp.SetFunc("TestCount")
	tp.Sql = "select count(*) from test where id=?"
	tp.Type = append(tp.Type, "int64")
	tp.Type = append(tp.Type, "string")
	tp.Stmt = "stmt"
	testStruct.AddFunc(tp)
}

func Test_SelectColumn(t *testing.T) {
	tp := new(SelectColumn)
	tp.Sql = "select name from test where id=?"
	tp.Func = "SelectId"
	tp.Struct = testStruct.Name()
	tp.Arg = append(tp.Arg, &Arg{"id1", true})
	tp.Arg = append(tp.Arg, &Arg{"id2", false})
	tp.Stmt = "stmt"
	testStruct.AddFunc(tp)
}

func Test_SelectList(t *testing.T) {
	tp := new(SelectList)
	tp.Sql = "select * from test"
	tp.Func = "TestList"
	tp.Struct = testStruct.Name()
	tp.Arg = append(tp.Arg, &Arg{"id1", false})
	tp.Arg = append(tp.Arg, &Arg{"id2", false})
	tp.Column = append(tp.Column, "F1")
	tp.Column = append(tp.Column, "F2")
	tp.Stmt = "stmt"
	testStruct.AddFunc(tp)
}

func Test_Struct(t *testing.T) {
	testStruct.AddField("F1", "int64", `json:"id"`)
	testStruct.AddField("F2", "sql.NullString", "")
	_ = testStruct.TPL().Execute(os.Stderr, testStruct)
}

/*


 */
