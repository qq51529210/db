package mysql

import "testing"

func fatalError(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func failNotError(t *testing.T, err error) {
	if err == nil {
		t.FailNow()
	}
}

func failNotString(t *testing.T, v interface{}, s string) {
	if ss, ok := v.(string); ok {
		if ss == s {
			return
		}
	}
	t.FailNow()
}

func failNotExpress(t *testing.T, v interface{}, l, o, r string) {
	if expr, ok := v.(*ExpressionStmt); ok {
		failNotString(t, expr.Left, l)
		failNotString(t, expr.Operator, o)
		failNotString(t, expr.Right, r)
		return
	}
	t.FailNow()
}

func Test_ParseExpression(t *testing.T) {
	v, err := ParseSQL("1")
	fatalError(t, err)
	failNotString(t, v, "1")
	//
	v, err = ParseSQL("(1)")
	fatalError(t, err)
	failNotString(t, v, "1")
	//
	v, err = ParseSQL("1+2")
	fatalError(t, err)
	failNotExpress(t, v, "1", "+", "2")
	//
	v, err = ParseSQL("1+2*")
	failNotError(t, err)
	v, err = ParseSQL("+2*3")
	failNotError(t, err)
	//
	v, err = ParseSQL("a + b * c and d")
	fatalError(t, err)
	exr := v.(*ExpressionStmt)
	failNotString(t, exr.Operator, "and")
	failNotString(t, exr.Right, "d")
	exr = exr.Left.(*ExpressionStmt)
	failNotString(t, exr.Left, "a")
	failNotString(t, exr.Operator, "+")
	exr = exr.Right.(*ExpressionStmt)
	failNotString(t, exr.Left, "b")
	failNotString(t, exr.Operator, "*")
	failNotString(t, exr.Right, "c")
	//
	v, err = ParseSQL("a between b+c and 1+2")
	fatalError(t, err)
	exr = v.(*ExpressionStmt)
	failNotString(t, exr.Left, "a")
	failNotString(t, exr.Operator, "between")
	exr = exr.Right.(*ExpressionStmt)
	failNotExpress(t, exr.Left, "b", "+", "c")
	failNotString(t, exr.Operator, "and")
	failNotExpress(t, exr.Right, "1", "+", "2")
	//
	v, err = ParseSQL("a not in (b+c,1,a, 1+2)")
	fatalError(t, err)
	exr = v.(*ExpressionStmt)
	failNotString(t, exr.Left, "a")
	failNotString(t, exr.Operator, "not in")
	vv := exr.Right.([]interface{})
	failNotExpress(t, vv[0], "b", "+", "c")
	failNotString(t, vv[1], "1")
	failNotString(t, vv[2], "a")
	failNotExpress(t, vv[3], "1", "+", "2")
}

func Test_ParseSelect(t *testing.T) {
	v, err := ParseSQL("select *,t1.c0, 1 as c1,t2.2 c2,count(id) from test t1 " +
		"natural left outer join test2 as t2 on t1.c0=t2.c0 " +
		"where c0>? " +
		"group by c1 " +
		"having count(c1)>100 " +
		"union all select * from test3 where c1=? " +
		"order by c2 desc " +
		"limit 1,?")
	fatalError(t, err)
	q := v.(*SelectStmt)
	failNotString(t, q.Column[0].Expression, "*")
	failNotString(t, q.Column[1].Expression, "t1.c0")
	failNotString(t, q.Column[2].Expression, "1")
	failNotString(t, q.Column[2].Alias, "c1")
	failNotString(t, q.Column[3].Expression, "t2.2")
	failNotString(t, q.Column[3].Alias, "c2")
	failNotString(t, q.Column[4].Expression.(*FuncExpressionStmt).Name, "count")
	failNotString(t, q.Column[4].Expression.(*FuncExpressionStmt).Value, "id")
	failNotString(t, q.Table, "test")
	failNotString(t, q.TableAlias, "t1")
	failNotString(t, q.Join, "test2")
	failNotString(t, q.JoinAlias, "t2")
	failNotExpress(t, q.On, "t1.c0", "=", "t2.c0")
	failNotExpress(t, q.Where, "c0", ">", "?")
	failNotString(t, q.GroupBy[0], "c1")
	failNotString(t, q.Having.(*ExpressionStmt).Left.(*FuncExpressionStmt).Name, "count")
	failNotString(t, q.Having.(*ExpressionStmt).Left.(*FuncExpressionStmt).Value, "c1")
	failNotString(t, q.Having.(*ExpressionStmt).Operator, ">")
	failNotString(t, q.Having.(*ExpressionStmt).Right, "100")
	failNotString(t, q.OrderBy[0], "c2")
	failNotString(t, q.Limit[0], "1")
	failNotString(t, q.Limit[1], "?")
}

func Test_ParseDelete(t *testing.T) {
	v, err := ParseSQL("delete from t1 where id>(select id from t2)")
	fatalError(t, err)
	q := v.(*DeleteStmt)
	failNotString(t, q.Table, "t1")
	exr := q.Where.(*ExpressionStmt)
	failNotString(t, exr.Left, "id")
	failNotString(t, exr.Operator, ">")
}

func Test_ParseUpdate(t *testing.T) {
}

func Test_ParseInsert(t *testing.T) {
}
