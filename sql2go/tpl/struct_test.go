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
	tp.StmtName = "stmt"
	testStruct.AddFunc(tp)
}

func Test_SelectColumn(t *testing.T) {
	tp := new(SelectColumn)
	tp.Sql = "select name from test where id=?"
	tp.Func = "SelectId"
	tp.Struct = testStruct.Name()
	tp.Arg = append(tp.Arg, &Arg{"id1", true})
	tp.Arg = append(tp.Arg, &Arg{"id2", false})
	tp.StmtName = "stmt"
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
	tp.StmtName = "stmt"
	testStruct.AddFunc(tp)
}

func Test_SelectPage(t *testing.T) {
	tp := new(SelectPage)
	tp.Sql = "select * from test"
	tp.Func = "TestPage"
	tp.Struct = testStruct.Name()
	tp.StmtName = "stmt1"
	tp.Column = append(tp.Column, "F1")
	tp.BeforeGroup = "select * from test where name like ? group by"
	tp.Arg = append(tp.Arg, &Arg{"name", false})
	tp.Group = append(tp.Group, "group1")
	tp.Group = append(tp.Group, "group2")
	tp.Arg = append(tp.Arg, &Arg{"group1", false})
	tp.Arg = append(tp.Arg, &Arg{"group2", false})
	tp.Group2Order = "order by"
	tp.Order = append(tp.Order, "order1")
	tp.Order = append(tp.Order, "order2")
	tp.Arg = append(tp.Arg, &Arg{"order1", false})
	tp.Arg = append(tp.Arg, &Arg{"order2", false})
	tp.Sort = "desc"
	tp.AfterOrder = "limit ?,?"
	tp.Arg = append(tp.Arg, &Arg{"limit1", false})
	tp.Arg = append(tp.Arg, &Arg{"limit2", false})
	testStruct.AddFunc(tp)
}

func Test_Struct(t *testing.T) {
	testStruct.AddField("F1", "int64", `json:"id"`)
	testStruct.AddField("F2", "sql.NullString", "")
	_ = testStruct.TPL().Execute(os.Stderr, testStruct)
}

/*


 */
