package mysql

import (
	"bytes"
	"fmt"
	"github.com/qq51529210/db/db2go"
	"github.com/qq51529210/db/sql2go/tpl"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	// 运算符名称
	operatorsNames = map[string]string{
		">":           "Bg",
		">=":          "BgEq",
		"<":           "Le",
		"<=":          "LeEq",
		"<>":          "NEq",
		"!=":          "NEq",
		"=":           "Eq",
		"in":          "In",
		"not in":      "NIn",
		"like":        "Lk",
		"not like":    "NLk",
		"between":     "Bet",
		"not between": "NBet",
		"exists":      "Ext",
		"not exists":  "NExt",
	}
)

// 把'UserId'转换成'userId'
func PascalCaseToCamelCase(s string) string {
	return strings.ToLower(s[:1]) + s[1:]
}

// 把'userId'转换成'UserId'
func CamelCaseToPascalCase(s string) string {
	return strings.ToUpper(s[:1]) + s[1:]
}

// 把'UserId'转换成'user_id'
func PascalCaseToSnakeCase(s string) string {
	if len(s) < 1 {
		return ""
	}
	var buf bytes.Buffer
	c1 := s[0]
	if c1 >= 'A' && c1 <= 'Z' {
		c1 = c1 + 'a' - 'A'
	}
	buf.WriteByte(c1)

	for i := 1; i < len(s); i++ {
		c2 := s[i]
		if c2 >= 'A' && c2 <= 'Z' {
			c2 = c2 + 'a' - 'A'
			c1 = s[i-1]
			if (c1 >= 'a' && c1 <= 'z') || (c1 >= '0' && c1 <= '9') {
				buf.WriteByte('_')
			}
		}
		buf.WriteByte(c2)
	}
	return string(buf.Bytes())
}

// 把'user_id'转换成'UserId'
func SnakeCaseToPascalCase(s string) string {
	if len(s) < 1 {
		return ""
	}
	var buf strings.Builder
	c1 := s[0]
	if c1 >= 'a' && c1 <= 'z' {
		c1 = c1 - 'a' + 'A'
	}
	buf.WriteByte(c1)
	for i := 1; i < len(s); i++ {
		c1 = s[i]
		if c1 == '_' {
			i++
			if i == len(s) {
				break
			}
			c1 = s[i]
			if c1 >= 'a' && c1 <= 'z' {
				c1 = c1 - 'a' + 'A'
			}
		}
		buf.WriteByte(c1)
	}
	return buf.String()
}

// 把'user_id'转换成'userId'
func SnakeCaseToCamelCase(s string) string {
	return PascalCaseToCamelCase(SnakeCaseToPascalCase(s))
}

// userId,roleId转换成UserIdRoleId，用于函数名称ByParams
func JoinCamelCaseToPascalCase(ss []string) string {
	var str strings.Builder
	for _, s := range ss {
		str.WriteString(CamelCaseToPascalCase(s))
	}
	return str.String()
}

// user_id,role_id转换成UserIdRoleId，用于函数名称(Select/Update/Insert)Field
func JoinSnakeCaseToPascalCase(ss []string) string {
	var str strings.Builder
	for _, s := range ss {
		str.WriteString(CamelCaseToPascalCase(s))
	}
	return str.String()
}

func recoverError() error {
	re := recover()
	if re != nil {
		e, ok := re.(error)
		if ok {
			return e
		} else {
			return fmt.Errorf("%v", e)
		}
	}
	return nil
}

func unknownTable(table string) error {
	return fmt.Errorf("unknown table '%s'", table)
}

type holder struct {
	table    *db2go.Table
	column   *db2go.Column
	name     string
	operator string
}

type holders struct {
	table      *db2go.Table
	tableAlias string
	join       *db2go.Table
	joinAlias  string
	holder     []*holder
	sub        *holders
}

func (h *holders) addColumn(column, operator string) {
	hh := new(holder)
	hh.operator = operator
	if column == "" {
		hh.name = "arg"
		h.holder = append(h.holder, hh)
		return
	}
	p := strings.Split(column, ".")
	if len(p) == 2 {
		hh.name = p[1]
		if p[0] == h.tableAlias || p[0] == h.table.Name() {
			hh.table = h.table
			col := h.table.GetColumn(p[1])
			if col != nil {
				hh.column = col
				h.holder = append(h.holder, hh)
				return
			}
		}
		if h.join != nil {
			if p[0] == h.joinAlias || p[0] == h.join.Name() {
				hh.table = h.join
				col := h.join.GetColumn(p[1])
				if col != nil {
					hh.column = col
					h.holder = append(h.holder, hh)
					return
				}
			}
		}
		// 如果sql测试通过，那么不是table.column就是join.column，bug了
		panic(fmt.Errorf("invalid expression '%s'", column))
	}
	hh.name = p[0]
	col := h.table.GetColumn(p[0])
	if col != nil {
		hh.table = h.table
		hh.column = col
		h.holder = append(h.holder, hh)
		return
	}
	if h.join != nil {
		col = h.join.GetColumn(p[0])
		if col != nil {
			hh.table = h.join
			hh.column = col
			h.holder = append(h.holder, hh)
			return
		}
	}
	// 防止不能生成变量名的
	if hh.name[0] != '_' &&
		!(hh.name[0] >= 'a' && hh.name[0] <= 'z') &&
		!(hh.name[0] >= 'A' && hh.name[0] <= 'Z') {
		hh.name = "arg"
	}
	h.holder = append(h.holder, hh)
}

func (h *holders) toArgs() []*tpl.Arg {
	var args []*tpl.Arg
	for _, hh := range h.holder {
		arg := new(tpl.Arg)
		if hh.column != nil {
			arg.Name = SnakeCaseToPascalCase(hh.column.Name())
			arg.IsField = true
		} else {
			arg.Name = SnakeCaseToCamelCase(hh.name)
		}
		args = append(args, arg)
	}
	// 如果有相同的字段，两个都转成参数
	field := make(map[string]*tpl.Arg)
	for _, a := range args {
		if a.IsField {
			aa, ok := field[a.Name]
			if ok {
				a.IsField = false
				a.Name = PascalCaseToSnakeCase(a.Name)
				aa.IsField = false
				aa.Name = PascalCaseToSnakeCase(aa.Name)
			} else {
				field[a.Name] = a
			}
		}
	}
	// 如果有相同的参数，重命名
	param := make(map[string]int)
	for i, a := range args {
		if !a.IsField {
			a.Name += operatorsNames[h.holder[i].operator]
			n, ok := param[a.Name]
			if ok {
				n++
				a.Name = SnakeCaseToCamelCase(a.Name) + strconv.Itoa(n)
			} else {
				param[a.Name] = 0
			}
		}
	}
	return args
}

func (h *holders) toJoinArgs() []*tpl.Arg {
	var args []*tpl.Arg
	return args
}

func newFuncColumn(table, join *db2go.Table, query *SelectStmt, column string) *funcColumn {
	c := new(funcColumn)
	p := strings.Split(column, ".")
	if len(p) == 2 {
		if p[0] == query.Table || p[0] == query.TableAlias {
			c.table = table
			c.column = table.GetColumn(p[1])
			return c
		}
		if join != nil {
			if p[0] == query.Join || p[0] == query.JoinAlias {
				c.table = join
				c.column = join.GetColumn(p[1])
				return c
			}
		}
	}
	c.table = table
	c.column = table.GetColumn(column)
	return c
}

type funcColumn struct {
	table  *db2go.Table
	column *db2go.Column
	alias  string
}

func (c *funcColumn) Field() string {
	if c.table != nil {
		return SnakeCaseToPascalCase(c.table.Name()) + "." + SnakeCaseToPascalCase(c.column.Name())
	}
	return SnakeCaseToPascalCase(c.column.Name())
}

func (c *funcColumn) Param() string {
	if c.table != nil {
		return SnakeCaseToCamelCase(c.table.Name()) + SnakeCaseToPascalCase(c.column.Name())
	}
	return SnakeCaseToCamelCase(c.column.Name())
}

type Code interface {
	// 获取所有的struct模板
	StructTPLs() []tpl.TPL
	// 获取名称为name的struct模板
	StructTPL(table string) (tpl.StructTPL, error)
	// 生成sql函数模板，function和param为空自动生成，tx表示是否为事务
	FuncTPL(sql string, function string, tx bool, params []string) (tpl.FuncTPL, error)
	// 生成table的默认函数模板
	DefaultFuncTPLs(table string) ([]tpl.FuncTPL, error)
	// 获取当前的schema
	Schema() *db2go.Schema
	// 保存当前的模板到文件
	SaveFiles(dir string) error
}

func NewCode(dbUrl, pkg string) (Code, error) {
	schema, err := db2go.ReadSchema(db2go.MYSQL, dbUrl)
	if err != nil {
		return nil, err
	}
	if pkg == "" {
		pkg = PascalCaseToSnakeCase(schema.Name())
	}
	c := new(code)
	c.schema = schema
	c.pkg = pkg
	// init模板
	c.tplInit = tpl.NewInitTPL(pkg, db2go.MYSQL, db2go.DriverPkg(db2go.MYSQL))
	// struct模板
	c.tplStruct = make(map[string]tpl.StructTPL)
	for _, t := range schema.Tables() {
		tp, err := c.StructTPL(t.Name())
		if err != nil {
			return nil, err
		}
		c.tplStruct[tp.StructName()] = tp.(*tpl.Struct)
	}
	return c, nil
}

type code struct {
	pkg       string
	schema    *db2go.Schema
	tplStruct map[string]tpl.StructTPL
	tplInit   *tpl.Init
	stmt      map[string]int
}

func (c *code) StructTPLs() []tpl.TPL {
	var tps []tpl.TPL
	for _, tp := range c.tplStruct {
		tps = append(tps, tp)
	}
	return tps
}

func (c *code) StructTPL(table string) (tpl.StructTPL, error) {
	name := SnakeCaseToPascalCase(table)
	// 如果存在，就返回
	if tp, ok := c.tplStruct[name]; ok {
		return tp, nil
	}
	// 不存在
	t := c.schema.GetTable(table)
	if t == nil {
		return nil, unknownTable(table)
	}
	// 新建模板
	tp := new(tpl.Struct)
	tp.Pkg = c.pkg
	tp.Name = name
	for _, col := range t.Columns() {
		f := new(tpl.Field)
		f.Name = SnakeCaseToPascalCase(col.Name())
		f.Type = col.GoType()
		tp.Field = append(tp.Field, f)
	}
	return tp, nil
}

func (c *code) FuncTPL(sql string, function string, tx bool, params []string) (tpl.FuncTPL, error) {
	// 直接用数据库测试sql是否正确，如果正确，接下来解析和生成就少了很多不必要的判断了
	err := c.schema.TestSQL(sql)
	if err != nil {
		return nil, err
	}
	// 解析sql
	v, err := ParseSQL(sql)
	if err != nil {
		return nil, err
	}
	// 生成函数模板
	return c.funcTPL(v, sql, function, tx, params)
}

func (c *code) DefaultFuncTPLs(table string) ([]tpl.FuncTPL, error) {
	t := c.schema.GetTable(table)
	if t == nil {
		return nil, unknownTable(table)
	}
	// 区分列
	pk, npk, uni, nuni, mul, nmul := c.defaultFuncPickColumns(t)
	var tps []tpl.FuncTPL
	var sql []string
	// 生成
	sql = append(sql, c.defaultFuncInsert(t, t.Columns()))
	if len(pk) > 0 {
		sql = append(sql, c.defaultFuncSelect(t, npk, pk))
		sql = append(sql, c.defaultFuncUpdate(t, npk, pk))
		sql = append(sql, c.defaultFuncDelete(t, pk))
	}
	if len(uni) > 0 {
		sql = append(sql, c.defaultFuncSelect(t, nuni, uni))
		sql = append(sql, c.defaultFuncUpdate(t, nuni, uni))
		sql = append(sql, c.defaultFuncDelete(t, uni))
	}
	if len(mul) > 0 {
		sql = append(sql, c.defaultFuncSelect(t, nmul, mul))
		sql = append(sql, c.defaultFuncUpdate(t, nmul, mul))
		sql = append(sql, c.defaultFuncDelete(t, mul))
	}
	// 生成模板
	for _, s := range sql {
		if s == "" {
			continue
		}
		tp, err := c.FuncTPL(s, "", false, nil)
		if err != nil {
			return nil, err
		}
		tps = append(tps, tp)
	}
	// 返回
	return tps, nil
}

func (c *code) Schema() *db2go.Schema {
	return c.schema
}

func (c *code) SaveFiles(dir string) error {
	// 输出目录
	dir = filepath.Join(dir, c.pkg)
	// 先删除
	err := os.RemoveAll(dir)
	if err != nil {
		return err
	}
	// 再创建
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}
	// 输出struct模板
	for k, v := range c.tplStruct {
		err = tpl.SaveFile(v, filepath.Join(dir, PascalCaseToSnakeCase(k)+".go"))
		if err != nil {
			return err
		}
	}
	// 输出init模板
	return tpl.SaveFile(c.tplInit, filepath.Join(dir, c.pkg+".init.go"))
}

func (c *code) defaultFuncCount(table *db2go.Table, pk []*db2go.Column) string {
	var sql strings.Builder
	sql.Reset()
	sql.WriteString("select count(")
	c.defaultFuncFields(&sql, pk)
	sql.WriteString(")from ")
	sql.WriteString(table.Name())
	return sql.String()
}

func (c *code) defaultFuncList(table *db2go.Table) string {
	var sql strings.Builder
	sql.WriteString("select * from ")
	sql.WriteString(table.Name())
	sql.WriteString(" limit ?,?")
	return sql.String()
}

func (c *code) defaultFuncInsert(table *db2go.Table, columns []*db2go.Column) string {
	var sql strings.Builder
	sql.WriteString("insert into ")
	sql.WriteString(table.Name())
	sql.WriteByte('(')
	c.defaultFuncFields(&sql, columns)
	sql.WriteByte(')')
	sql.WriteString(" values(")
	sql.WriteByte('?')
	for i := 1; i < len(columns); i++ {
		sql.WriteString(",?")
	}
	sql.WriteByte(')')
	return sql.String()
}

func (c *code) defaultFuncSelect(table *db2go.Table, fields, condition []*db2go.Column) string {
	var sql strings.Builder
	sql.WriteString("select ")
	c.defaultFuncFields(&sql, fields)
	sql.WriteString(" from ")
	sql.WriteString(table.Name())
	sql.WriteString(" where ")
	c.defaultFuncConditions(&sql, condition)
	return sql.String()
}

func (c *code) defaultFuncDelete(table *db2go.Table, condition []*db2go.Column) string {
	var sql strings.Builder
	sql.Reset()
	sql.WriteString("delete from ")
	sql.WriteString(table.Name())
	sql.WriteString(" where ")
	c.defaultFuncConditions(&sql, condition)
	return sql.String()
}

func (c *code) defaultFuncUpdate(table *db2go.Table, fields, condition []*db2go.Column) string {
	column := make([]*db2go.Column, 0)
	for _, col := range fields {
		if !col.IsAutoIncrement() {
			column = append(column, col)
		}
	}
	if len(column) < 1 {
		return ""
	}
	var sql strings.Builder
	sql.WriteString("update ")
	sql.WriteString(table.Name())
	sql.WriteString(" set ")
	sql.WriteString(column[0].Name())
	sql.WriteString("=?")
	for i := 1; i < len(column); i++ {
		sql.WriteByte(',')
		sql.WriteString(column[i].Name())
		sql.WriteString("=?")
	}
	sql.WriteString(" where ")
	c.defaultFuncConditions(&sql, condition)
	return sql.String()
}

func (c *code) defaultFuncFields(sql *strings.Builder, fields []*db2go.Column) {
	if len(fields) < 1 {
		return
	}
	sql.WriteString(fields[0].Name())
	for i := 1; i < len(fields); i++ {
		sql.WriteString(",")
		sql.WriteString(fields[i].Name())
	}
}

func (c *code) defaultFuncConditions(sql *strings.Builder, condition []*db2go.Column) {
	if len(condition) < 1 {
		return
	}
	sql.WriteString(condition[0].Name())
	sql.WriteString("=?")
	for i := 1; i < len(condition); i++ {
		sql.WriteString(" and ")
		sql.WriteString(condition[i].Name())
		sql.WriteString("=?")
	}
}

func (c *code) defaultFuncPickColumns(table *db2go.Table) (pk, npk, uni, nuni, mul, nmul []*db2go.Column) {
	for _, c := range table.Columns() {
		if c.IsPrimaryKey() {
			pk = append(pk, c)
		} else {
			npk = append(npk, c)
		}
		if c.IsUnique() {
			uni = append(uni, c)
		} else {
			nuni = append(nuni, c)
		}
		if c.IsMulUnique() {
			mul = append(mul, c)
		} else {
			nmul = append(nmul, c)
		}
	}
	return
}

func (c *code) funcTPL(v interface{}, sql string, function string, tx bool, params []string) (tp tpl.FuncTPL, err error) {
	defer func() {
		err = recoverError()
	}()
	// 生成func模板
	switch q := v.(type) {
	case *InsertStmt:
		tp = c.funcInsert(q, sql, tx)
	case *DeleteStmt:
		tp = c.funcDelete(q, sql, tx)
	case *UpdateStmt:
		tp = c.funcUpdate(q, sql, tx)
	case *SelectStmt:
		tp = c.funcSelect(q, sql, tx)
	default:
		return nil, fmt.Errorf("unsupported sql '%s'", sql)
	}
	// 替换自定义名称
	if function != "" {
		tp.SetFuncName(function)
	}
	if len(params) > 0 {
		tp.SetParamName(params)
	}
	// 添加到struct模板
	s := c.tplStruct[tp.StructName()]
	s.AddFuncTPL(tp)
	c.tplInit.Stmt[tp.Stmt()] = tp.SQL()
	// 返回
	return
}

func (c *code) funcSelect(q *SelectStmt, sql string, tx bool) tpl.FuncTPL {
	t, j, structName := c.funcSelectTable(q)
	// 占位符
	var h holders
	c.parseSelectHolder(&h, q)
	// 函数名
	// 是column还是function
	switch q.Column[0].Expression.(type) {
	case string:
		c.funcSelectColumns(q, t, j)
		//columns := c.funcSelectColumns(q, t, j)
		//structQueryRow := false
		//var byFields []string
		//for _, a := range h.arg {
		//	byFields = append(byFields, a.Name)
		//	if j == nil {
		//		col := t.GetColumn(PascalCaseToSnakeCase(a.Name))
		//		structQueryRow =
		//	}
		//}
		//if structQueryRow {
		//	tp := new(tpl.StructQueryRow)
		//	tp.Tx = tx
		//	tp.Sql = sql
		//	tp.Param = pick0(h.param)
		//	tp.Struct = structName
		//	tp.StmtName = c.funcStmtName()
		//	tp.Query = h.arg
		//	for _, col := range columns {
		//		if j == nil {
		//			tp.Scan = append(tp.Scan, SnakeCaseToPascalCase(col.column.Name()))
		//		} else {
		//			tp.Scan = append(tp.Scan, SnakeCaseToPascalCase(col.table.Name())+"."+
		//				SnakeCaseToPascalCase(col.column.Name()))
		//		}
		//	}
		//	if len(byFields) > 0 || len(h.param) > 0 {
		//		funcName.WriteString("By")
		//		funcName.WriteString(strings.Join(byFields, ""))
		//	}
		//	tp.Func = funcName.String()
		//	return tp
		//}
		tp := new(tpl.StructQuery)
		//tp.Tx = tx
		//tp.Sql = sql
		//tp.Param = pick0(h.param)
		//tp.Struct = structName
		//tp.StmtName = c.funcStmtName()
		//for _, col := range columns {
		//	if j == nil {
		//		tp.Scan = append(tp.Scan, SnakeCaseToPascalCase(col.column.Name()))
		//	} else {
		//		tp.Scan = append(tp.Scan, SnakeCaseToPascalCase(col.table.Name())+"."+
		//			SnakeCaseToPascalCase(col.column.Name()))
		//	}
		//}
		//funcName.WriteString(structName)
		//var byFields []string
		//for _, a := range h.arg {
		//	byFields = append(byFields, a.Name)
		//}
		//if len(byFields) > 0 || len(h.param) > 0 {
		//	funcName.WriteString("By")
		//	funcName.WriteString(strings.Join(byFields, ""))
		//}
		//tp.Func = funcName.String()
		return tp
	default:
		// 函数
		//columns := c.funcSelectFuncColumns(q)
		tp := new(tpl.QueryRow)
		tp.Tx = tx
		tp.Sql = sql
		tp.Struct = structName
		tp.StmtName = c.funcStmtName()
		//tp.Param = h.tplParams()
		//for _, name := range columns {
		//	tp.Return = append(tp.Return, functions[name])
		//}
		//// 函数名称
		//var funcName strings.Builder
		//funcName.WriteString("Select")
		//funcName.WriteString(structName)
		//funcName.WriteString(JoinSnakeCaseToPascalCase(columns))
		//if len(h.param) > 0 {
		//	funcName.WriteString("By")
		//	funcName.WriteString(JoinCamelCaseToPascalCase(tp.Param))
		//}
		//tp.Func = funcName.String()
		return tp
	}
}

func (c *code) funcSelectTable(q *SelectStmt) (t, j *db2go.Table, structName string) {
	t = c.funcGetTable(q.Table)
	struct1Name := SnakeCaseToPascalCase(q.Table)
	structName = struct1Name
	// 生成join结构
	if q.Join != "" {
		j = c.funcGetTable(q.Join)
		struct2Name := SnakeCaseToPascalCase(q.Join)
		structName += "Join" + struct2Name
		_, ok := c.tplStruct[structName]
		if !ok {
			tp := new(tpl.JoinStruct)
			tp.Pkg = c.pkg
			tp.Struct1 = c.tplStruct[struct1Name].(*tpl.Struct)
			tp.Struct2 = c.tplStruct[struct2Name].(*tpl.Struct)
			c.tplStruct[structName] = tp
		}
	}
	return
}

func (c *code) funcSelectColumns(q *SelectStmt, t, j *db2go.Table) []*funcColumn {
	var columns []*funcColumn
	for _, col := range q.Column {
		v := col.Expression.(string)
		if v == "*" {
			// 全选
			if j != nil {
				// 添加所有的列
				for _, col := range t.Columns() {
					columns = append(columns, &funcColumn{table: t, column: col})
				}
				for _, col := range j.Columns() {
					columns = append(columns, &funcColumn{table: j, column: col})
				}
			} else {
				// 添加所有的列
				for _, col := range t.Columns() {
					columns = append(columns, &funcColumn{table: t, column: col})
				}
			}
			break
		}
		// 普通的列
		col := newFuncColumn(t, j, q, v)
		columns = append(columns, col)
	}
	return columns
}

func (c *code) funcSelectFuncColumns(q *SelectStmt) []string {
	var functions []string
	for _, col := range q.Column {
		v := col.Expression.(*FuncExpressionStmt)
		functions = append(functions, v.Name)
	}
	return functions
}

func (c *code) funcInsert(q *InsertStmt, sql string, tx bool) tpl.FuncTPL {
	t := c.funcGetTable(q.Table)
	// 如果sql中省略了(column[, ...])
	columns := q.Column
	if len(columns) < 1 {
		for _, col := range t.Columns() {
			columns = append(columns, col.Name())
		}
	}
	// 模板
	switch q.Value.(type) {
	case *SelectStmt:
		value := q.Value.(*SelectStmt)
		// 解析占位符
		var h holders
		h.table = t
		c.parseSelectHolder(&h, value)
		tp := new(tpl.Exec)
		tp.Tx = tx
		tp.Sql = sql
		tp.Struct = SnakeCaseToCamelCase(t.Name())
		tp.StmtName = c.funcStmtName()
		tp.Arg = h.toArgs()
		// 函数名
		var funcName strings.Builder
		funcName.WriteString("Insert")
		fields, params := tpl.ClassifyArgs(tp.Arg)
		// InsertTable1FromTable2ByParam
		funcName.WriteString(tp.Struct)
		funcName.WriteString(strings.Join(fields, ""))
		funcName.WriteString("From")
		funcName.WriteString(value.Table)
		if len(params) > 0 {
			funcName.WriteString("By")
			funcName.WriteString(JoinSnakeCaseToPascalCase(params))
		}
		tp.Func = funcName.String()
		return tp
	default:
		values := q.Value.([]interface{})
		if len(q.Column) < 1 && len(columns) != len(values) {
			panic(fmt.Errorf("values count '%d' no equal columns count '%d'", len(values), len(columns)))
		}
		// 解析占位符
		var h holders
		h.table = t
		for i, v := range values {
			switch s := v.(type) {
			case string:
				if s == "?" {
					h.addColumn(columns[i], "=")
				}
			default:
				c.parseHolders(&h, v)
			}
		}
		tp := new(tpl.StructExec)
		tp.Tx = tx
		tp.Sql = sql
		tp.Struct = SnakeCaseToPascalCase(t.Name())
		tp.StmtName = c.funcStmtName()
		tp.Arg = h.toArgs()
		// 函数名
		var funcName strings.Builder
		funcName.WriteString("Insert")
		// InsertField
		funcName.WriteString(strings.Join(tpl.PickFields(tp.Arg), ""))
		tp.Func = funcName.String()
		return tp
	}
}

func (c *code) funcDelete(q *DeleteStmt, sql string, tx bool) tpl.FuncTPL {
	t := c.funcGetTable(q.Table)
	var h holders
	h.table = t
	// 解析where占位符
	c.parseHolders(&h, q.Where)
	// 模板
	tp := new(tpl.StructExec)
	tp.Tx = tx
	tp.Sql = sql
	tp.Struct = SnakeCaseToPascalCase(t.Name())
	tp.StmtName = c.funcStmtName()
	tp.Arg = h.toArgs()
	// 函数名
	var funcName strings.Builder
	funcName.WriteString("Delete")
	// DeleteBy
	fields := tpl.PickFields(tp.Arg)
	if len(fields) > 0 {
		funcName.WriteString("By")
		funcName.WriteString(strings.Join(fields, ""))
	}
	tp.Func = funcName.String()
	return tp
}

func (c *code) funcUpdate(q *UpdateStmt, sql string, tx bool) tpl.FuncTPL {
	t := c.funcGetTable(q.Table)
	// 解析占位符
	var h1, h2 holders
	h1.table = t
	h2.table = t
	for _, col := range q.Column {
		c.parseHolders(&h1, col)
	}
	c.parseHolders(&h2, q.Where)
	// 模板
	tp := new(tpl.StructExec)
	tp.Tx = tx
	tp.Sql = sql
	tp.Struct = SnakeCaseToPascalCase(t.Name())
	tp.StmtName = c.funcStmtName()
	columnArgs := h1.toArgs() // 用于函数名
	tp.Arg = append(columnArgs, h2.toArgs()...)
	// 函数名
	var funcName strings.Builder
	funcName.WriteString("Update")
	// UpdateFieldByFieldParam
	fields := tpl.PickFields(columnArgs)
	funcName.WriteString(strings.Join(fields, ""))
	fields, params := tpl.ClassifyArgs(tp.Arg)
	if len(fields) > 0 || len(params) > 0 {
		funcName.WriteString("By")
		funcName.WriteString(strings.Join(fields, ""))
		funcName.WriteString(JoinSnakeCaseToPascalCase(params))
	}
	tp.Func = funcName.String()
	return tp
}

func (c *code) funcGetTable(table string) *db2go.Table {
	t := c.schema.GetTable(table)
	if t == nil {
		panic(unknownTable(table))
	}
	return t
}

func (c *code) funcStmtName() string {
	return "stmt" + strconv.Itoa(len(c.tplInit.Stmt))
}

func (c *code) parseHolders(h *holders, v interface{}) {
	if v == nil {
		return
	}
	switch v := v.(type) {
	case string:
		// 单个，没有列名和操作符
		if v == "?" {
			h.addColumn("", "")
		}
	case *ExpressionStmt:
		left, ok := v.Left.(string)
		if ok {
			if left == "?" {
				h.addColumn("", "")
			}
		} else {
			c.parseHolders(h, v.Left)
		}
		right, ok := v.Right.(string)
		if ok {
			if right == "?" {
				if left != "" {
					h.addColumn(left, v.Operator)
				} else {
					h.addColumn("", "")
				}
			}
		} else {
			c.parseHolders(h, v.Right)
		}
	case *BoolExpressionStmt:
		c.parseHolders(h, v.Value)
	case *FuncExpressionStmt:
		c.parseHolders(h, v.Value)
	case *SelectStmt:
		h.sub = new(holders)
		c.parseSelectHolder(h.sub, v)
	case []interface{}:
		for _, i := range v {
			c.parseHolders(h, i)
		}
	}
}

func (c *code) parseSelectHolder(h *holders, q *SelectStmt) {
	h.table = c.funcGetTable(q.Table)
	if q.Join != "" {
		h.join = c.funcGetTable(q.Join)
	}
	// distinct
	if q.Distinct == "?" {
		h.addColumn(q.Table+"_distinct", "")
	}
	// columns
	for _, col := range q.Column {
		c.parseHolders(h, col.Expression)
	}
	// on
	c.parseHolders(h, q.On)
	// where
	c.parseHolders(h, q.Where)
	// group by
	for _, s := range q.GroupBy {
		if s == "?" {
			h.addColumn(q.Table+"_group", "")
		}
	}
	// having
	c.parseHolders(h, q.Having)
	// union
	if q.Union != nil {
		h.sub = new(holders)
		c.parseSelectHolder(h.sub, q.Union)
	}
	// order by
	for _, s := range q.OrderBy {
		if s == "?" {
			h.addColumn(q.Table+"_order", "")
		}
	}
	// asc/desc
	if q.Order == "?" {
		h.addColumn(q.Table+"_sort", "")
	}
	// limit
	for _, s := range q.Limit {
		if s == "?" {
			h.addColumn(q.Table+"_limit", "")
		}
	}
}
