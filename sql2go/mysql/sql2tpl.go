package mysql

import (
	"bytes"
	"errors"
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

	errColumnType = errors.New("select field only support either column or function")

	errUnknownTable = errors.New("unknown table name")
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
			if c1 >= 'a' && c1 <= 'z' {
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

func JoinCamelCaseToPascalCase(ss ...string) string {
	var str strings.Builder
	for _, s := range ss {
		str.WriteString(CamelCaseToPascalCase(s))
	}
	return str.String()
}

func JoinSnakeCaseToPascalCase(ss ...string) string {
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

type holders struct {
	arg    []*tpl.Arg
	field  []string
	param  []string
	holder []*holder
}

type holder struct {
	table    string // 表名
	column   string // 列名
	operator string // 运算符
}

func (h *holders) addParam(table *db2go.Table, column, operator string) {
	var str strings.Builder
	if column == "" {
		str.WriteString("param")
		str.WriteString(strconv.Itoa(len(h.param)))
	} else {
		if table != nil && table.GetColumn(column) != nil {
			str.WriteString(SnakeCaseToCamelCase(table.Name() + "_" + column))
		} else {
			str.WriteString(SnakeCaseToCamelCase(column))
		}
		str.WriteString(operatorsNames[operator])
	}
	h.param = append(h.param, str.String())
	h.arg = append(h.arg, &tpl.Arg{Name: str.String()})
}

func (h *holders) addField(table, join *db2go.Table, tableAlias, joinAlias, column, operator string) {
	columnPart := strings.Split(column, ".")
	if len(columnPart) == 2 {
		if columnPart[0] == table.Name() && table.GetColumn(columnPart[1]) != nil {
			name := SnakeCaseToPascalCase(columnPart[0]) + "." + SnakeCaseToPascalCase(columnPart[1])
			h.field = append(h.field, name)
			h.arg = append(h.arg, &tpl.Arg{Name: name, IsField: true})
		} else {
			if join == nil {
				h.addParam(table, columnPart[1], operator)
			} else {
				if columnPart[0] == join.Name() && join.GetColumn(columnPart[1]) != nil {
					name := SnakeCaseToPascalCase(columnPart[0]) + "." + SnakeCaseToPascalCase(columnPart[1])
					h.field = append(h.field, name)
					h.arg = append(h.arg, &tpl.Arg{Name: name, IsField: true})
				}
			}
		}
	}
}

func (h *holders) addExecField(table *db2go.Table, column, operator string, subQuery bool) {
	//columnPart := strings.Split(column, ".")
	//if len(columnPart) == 2 {
	//	if columnPart[0] == table.Name() && table.GetColumn(columnPart[1]) != nil {
	//		name := SnakeCaseToPascalCase(columnPart[0]) + "." + SnakeCaseToPascalCase(columnPart[1])
	//		h.field = append(h.field, name)
	//		h.arg = append(h.arg, &tpl.Arg{Name: name, IsField: true})
	//	} else {
	//		if join == nil {
	//			h.addParam(table, columnPart[1], operator)
	//		} else {
	//			if columnPart[0] == join.Name() && join.GetColumn(columnPart[1]) != nil {
	//				name := SnakeCaseToPascalCase(columnPart[0]) + "." + SnakeCaseToPascalCase(columnPart[1])
	//				h.field = append(h.field, name)
	//				h.arg = append(h.arg, &tpl.Arg{Name: name, IsField: true})
	//			}
	//		}
	//	}
	//}
}

type Code interface {
	// 获取所有的struct模板
	StructTPLs() []tpl.TPL
	// 获取名称为name的struct模板
	StructTPL(table string) (tpl.StructTPL, error)
	// 获取名称为name的struct模板
	FuncTPL(sql string, tx bool) (tpl.FuncTPL, error)
	// 获取table的默认函数模板
	DefaultFuncTPLs(table string) ([]tpl.FuncTPL, error)
	// 获取当前的schema
	Schema() *db2go.Schema
	// 保存当前的模板到文件
	SaveFiles(dir string) error
	// 修改tp的函数名称，同时修改init模板的stmt
	SetFuncName(tp tpl.FuncTPL, name string)
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

func (c *code) FuncTPL(sql string, tx bool) (tpl.FuncTPL, error) {
	// 直接用数据库测试sql是否正确
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
	return c.funcTPL(sql, tx, v)
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
		tp, err := c.FuncTPL(s, false)
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
	// 创建输出目录
	dir = filepath.Join(dir, c.pkg)
	err := os.MkdirAll(dir, os.ModePerm)
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

func (c *code) SetFuncName(tp tpl.FuncTPL, name string) {
	delete(c.tplInit.Stmt, tp.Stmt())
	tp.SetFuncName(name)
	c.tplInit.Stmt[tp.Stmt()] = tp.SQL()
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
	c.defaultFuncFields(&sql, condition)
	return sql.String()
}

func (c *code) defaultFuncDelete(table *db2go.Table, condition []*db2go.Column) string {
	var sql strings.Builder
	sql.Reset()
	sql.WriteString("delete from ")
	sql.WriteString(table.Name())
	sql.WriteString(" where ")
	c.defaultFuncFields(&sql, condition)
	return sql.String()
}

func (c *code) defaultFuncUpdate(table *db2go.Table, fields, condition []*db2go.Column) string {
	var sql strings.Builder
	sql.WriteString("update ")
	sql.WriteString(table.Name())
	sql.WriteString(" set ")
	if len(fields) > 0 {
		sql.WriteString(fields[0].Name())
		sql.WriteString("=?")
		for i := 1; i < len(fields); i++ {
			sql.WriteByte(',')
			sql.WriteString(fields[i].Name())
			sql.WriteString("=?")
		}
	}
	sql.WriteString(" where ")
	c.defaultFuncFields(&sql, condition)
	return sql.String()
}

func (c *code) defaultFuncFields(sql *strings.Builder, fields []*db2go.Column) {
	if len(fields) < 1 {
		return
	}
	sql.WriteString(fields[0].Name())
	for i := 1; i < len(fields); i++ {
		sql.WriteByte(',')
		sql.WriteString(fields[i].Name())
	}
}

func (c *code) defaultFuncConditions(sql *strings.Builder, condition []*db2go.Column) {
	if len(condition) < 1 {
		return
	}
	sql.WriteString(condition[0].Name())
	sql.WriteString(" =? ")
	for i := 1; i < len(condition); i++ {
		sql.WriteString(" and ")
		sql.WriteString(condition[i].Name())
		sql.WriteString(" =? ")
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

func (c *code) funcTPL(sql string, tx bool, v interface{}) (tp tpl.FuncTPL, err error) {
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
	// 添加到struct模板
	s := c.tplStruct[tp.StructName()]
	s.AddFuncTPL(tp)
	c.tplInit.Stmt[tp.Stmt()] = tp.SQL()
	// 返回
	return
}

func (c *code) funcSelect(q *SelectStmt, sql string, tx bool) tpl.FuncTPL {
	if q.Join != nil {
		return c.funcSelectJoin(q, sql, tx)
	}
	t := c.funcGetTable(c.funcSelectTableName(&q.Table))
	// 占位符
	var h holders
	c.parseQueryHolder(q, &h, false)
	// 函数名
	var funcName strings.Builder
	funcName.WriteString("Select")
	// 结构名
	structName := SnakeCaseToPascalCase(t.Name())
	// 区分column
	columns, columnType := c.funcSelectColumn(q, t)
	switch columnType {
	case "column":
		// 解析占位符
		tp := new(tpl.StructQueryRow)
		tp.Tx = tx
		tp.Sql = sql
		tp.Param = h.param
		tp.Struct = structName
		tp.Query = h.arg
		tp.Scan = h.field
		funcName.WriteString(strings.Join(columns, ","))
		tp.Func = funcName.String()
		return tp
	case "function":
		tp := new(tpl.QueryRow)
		tp.Tx = tx
		tp.Sql = sql
		tp.Param = h.param
		tp.Struct = structName
		for _, name := range columns {
			tp.Return = append(tp.Return, functions[name])
		}
		// 函数名称
		funcName.WriteString(structName)
		funcName.WriteString(strings.Join(columns, ""))
		tp.Func = funcName.String()
		return tp
	default: // *
		tp := new(tpl.StructQueryRow)
		tp.Tx = tx
		tp.Sql = strings.Replace(sql, "*", strings.Join(columns, ","), 1)
		tp.Param = h.param
		tp.Struct = structName
		tp.Query = h.arg
		tp.Scan = h.field
		// 函数名称
		funcName.WriteString("All")
		tp.Func = funcName.String()
		return tp
	}
}

// 检查select的column，返回columns和columnType
func (c *code) funcSelectColumn(q *SelectStmt, t *db2go.Table) ([]string, string) {
	var columns []string
	var columnType string
	// 确保column是同一种类型
	checkColumnType := func(s string) {
		if columnType == "" {
			columnType = s
			return
		}
		if columnType != s {
			panic(errColumnType)
		}
	}
	// 去掉t.c的"t."
	removeColumnTable := func(s string) string {
		p := strings.Split(s, ".")
		if len(p) == 1 {
			return s
		}
		return p[1]
	}
	// 分析
	for _, col := range q.Column {
		switch v := col.Expression.(type) {
		case string:
			// 全选
			if v == "*" {
				checkColumnType("*")
				// 添加所有的列
				for _, col := range t.Columns() {
					columns = append(columns, removeColumnTable(col.Name()))
				}
				continue
			}
			// 普通的列
			checkColumnType("column")
			columns = append(columns, removeColumnTable(v))
		case *FuncExpressionStmt:
			// 函数
			checkColumnType("function")
			columns = append(columns, v.Name)
		default:
			panic(errColumnType)
		}
	}
	return columns, columnType
}

func (c *code) funcSelectJoin(q *SelectStmt, sql string, tx bool) tpl.FuncTPL {
	return nil
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
		c.parseQueryHolder(value, &h, false)
		tp := new(tpl.Exec)
		tp.Tx = tx
		tp.Sql = sql
		tp.Param = h.param
		tp.Struct = SnakeCaseToPascalCase(t.Name())
		// 函数名
		var funcName strings.Builder
		funcName.WriteString("Insert")
		funcName.WriteString(tp.Struct)
		funcName.WriteString(strings.Join(h.field, ""))
		funcName.WriteString("From")
		funcName.WriteString(c.funcSelectTableName(&value.Table))
		if len(h.param) > 0 {
			funcName.WriteString("By")
			funcName.WriteString(JoinCamelCaseToPascalCase(h.param...))
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
		for _, v := range values {
			c.parseExecHolder(t, v, &h, false)
		}
		tp := new(tpl.StructExec)
		tp.Tx = tx
		tp.Sql = sql
		tp.Param = h.param
		tp.Struct = SnakeCaseToPascalCase(t.Name())
		tp.Arg = h.arg
		// 函数名
		var funcName strings.Builder
		funcName.WriteString("Insert")
		funcName.WriteString(strings.Join(h.field, ""))
		tp.Func = funcName.String()
		return tp
	}
}

func (c *code) funcDelete(q *DeleteStmt, sql string, tx bool) tpl.FuncTPL {
	t := c.funcGetTable(q.Table)
	var h holders
	// 解析where占位符
	c.parseExecHolder(t, q.Where, &h, false)
	// 模板
	tp := new(tpl.StructExec)
	tp.Tx = tx
	tp.Sql = sql
	tp.Param = h.param
	tp.Struct = SnakeCaseToPascalCase(t.Name())
	tp.Arg = h.arg
	// 函数名
	var funcName strings.Builder
	funcName.WriteString("Delete")
	if len(h.field) > 0 {
		funcName.WriteString("By")
		funcName.WriteString(strings.Join(h.field, ""))
	}
	tp.Func = funcName.String()
	return tp
}

func (c *code) funcUpdate(q *UpdateStmt, sql string, tx bool) tpl.FuncTPL {
	t := c.funcGetTable(q.Table)
	var h holders
	// 解析column=expression占位符
	for _, col := range q.Column {
		c.parseExecHolder(t, col, &h, false)
	}
	fieldsCount := len(h.field)
	// 解析where占位符
	c.parseExecHolder(t, q.Where, &h, false)
	// 模板
	tp := new(tpl.StructExec)
	tp.Tx = tx
	tp.Sql = sql
	tp.Param = h.param
	tp.Struct = SnakeCaseToPascalCase(t.Name())
	tp.Arg = h.arg
	// 函数名
	var funcName strings.Builder
	funcName.WriteString("Update")
	funcName.WriteString(strings.Join(h.field[:fieldsCount], ""))
	if len(h.field[fieldsCount:]) > 0 || len(h.param) > 0 {
		funcName.WriteString("By")
		funcName.WriteString(strings.Join(h.field[fieldsCount:], ""))
		funcName.WriteString(JoinCamelCaseToPascalCase(h.param...))
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

func (c *code) funcSelectTableName(table *SelectAliasStmt) string {
	switch name := table.Expression.(type) {
	case string:
		if c.schema.GetTable(name) == nil {
			panic(fmt.Errorf("unknown table '%s'", table))
		}
		return name
	case *SelectStmt:
		return c.funcSelectTableName(&name.Table)
	default:
		panic(errUnknownTable)
	}
}

// o为true表示子查询
func (c *code) parseExecHolder(t *db2go.Table, v interface{}, h *holders, o bool) {
	if v == nil {
		return
	}
	switch v := v.(type) {
	case string:
		// 单个，没有列名和操作符
		if v == "?" {
			h.addParam(t, "", "")
		}
	case *ExpressionStmt:
		left, ok := v.Left.(string)
		if ok {
			if left == "?" {
				h.addParam(t, "", "")
			}
		} else {
			c.parseExecHolder(t, v.Left, h, true)
		}
		right, ok := v.Right.(string)
		if ok {
			if right == "?" {
				if left != "" {
					h.addExecField(t, left, v.Operator, o)
				} else {
					h.addParam(t, "", "")
				}
			}
		} else {
			c.parseExecHolder(t, v.Right, h, true)
		}
	case *BoolExpressionStmt:
		c.parseExecHolder(t, v.Value, h, true)
	case *FuncExpressionStmt:
		c.parseExecHolder(t, v.Value, h, true)
	case *SelectStmt:
		c.parseQueryHolder(v, h, true)
	case []interface{}:
		for _, i := range v {
			c.parseExecHolder(t, i, h, true)
		}
	}
}

// o为true表示子查询
func (c *code) parseQueryHolder(q *SelectStmt, h *holders, o bool) {

}

//// 解析column=xxx的表达式或者占位符xx='?'
//func (c *code) parseExecHolder(table, join *db2go.Table, tableAlias, joinAlias string, v interface{}, h *holders) error {
//	if v == nil {
//		return nil
//	}
//	switch v := v.(type) {
//	case string:
//		// 单个，没有列名和操作符
//		if v == "?" {
//			h.addParam(table, "", "")
//		}
//	case *ExpressionStmt:
//		left, ok := v.Left.(string)
//		if ok {
//			if left == "?" {
//				h.addParam(table, "", "")
//			}
//		} else {
//			err := c.parseExecHolder(table, join, tableAlias, joinAlias, v.Left, h)
//			if err != nil {
//				return err
//			}
//		}
//		right, ok := v.Right.(string)
//		if ok {
//			if right == "?" {
//				if left != "" {
//					h.addField(table, join, tableAlias, joinAlias, left, v.Operator)
//				} else {
//					h.addParam(table, "", "")
//				}
//			}
//		} else {
//			return c.parseExecHolder(table, join, tableAlias, joinAlias, v.Right, h)
//		}
//	case *BoolExpressionStmt:
//		return c.parseExecHolder(table, join, tableAlias, joinAlias, v.Value, h)
//	case *FuncExpressionStmt:
//		return c.parseExecHolder(table, join, tableAlias, joinAlias, v.Value, h)
//	case *SelectStmt:
//		return c.parseSelectHolder(v, h)
//	case []interface{}:
//		for _, i := range v {
//			err := c.parseExecHolder(table, join, tableAlias, joinAlias, i, h)
//			if err != nil {
//				return err
//			}
//		}
//	}
//	return nil
//}
//
//// 解析select子查询中的表达式
//func (c *code) parseSelectHolder(q *SelectStmt, h *holders) error {
//	if q == nil {
//		return nil
//	}
//	tableName, err := c.tableName(&q.Table)
//	if err != nil {
//		return err
//	}
//	t := c.schema.GetTable(tableName)
//	if t == nil {
//		return c.errTableNotExists(tableName)
//	}
//	var j *db2go.Table
//	var joinAlias string
//	if q.Join != nil {
//		joinAlias = q.Join.Alias
//		tableName, err = c.tableName(q.Join)
//		if err != nil {
//			return err
//		}
//		j = c.schema.GetTable(tableName)
//	}
//	// distinct
//	if q.Distinct == "?" {
//		h.addParam(t, "distinct", "")
//	}
//	// columns
//	for _, col := range q.Column {
//		err := c.parseExecHolder(t, j, q.Table.Alias, joinAlias, col.Expression, h)
//		if err != nil {
//			return err
//		}
//	}
//	// on
//	err = c.parseExecHolder(t, j, q.Table.Alias, joinAlias, q.On, h)
//	if err != nil {
//		return err
//	}
//	// where
//	err = c.parseExecHolder(t, j, q.Table.Alias, joinAlias, q.Where, h)
//	if err != nil {
//		return err
//	}
//	// group by
//	n := 0
//	for _, s := range q.GroupBy {
//		if s == "?" {
//			h.addParam(t, "group"+strconv.Itoa(n), "")
//			n++
//		}
//	}
//	// having
//	err = c.parseExecHolder(t, j, q.Table.Alias, joinAlias, q.Having, h)
//	if err != nil {
//		return err
//	}
//	// union
//	err = c.parseSelectHolder(q.Union, h)
//	if err != nil {
//		return err
//	}
//	// order by
//	n = 0
//	for _, s := range q.OrderBy {
//		if s == "?" {
//			h.addParam(t, "order"+strconv.Itoa(n), "")
//			n++
//		}
//	}
//	// asc/desc
//	if q.Order == "?" {
//		h.addParam(t, "sort"+strconv.Itoa(n), "")
//	}
//	// limit
//	n = 0
//	for _, s := range q.Limit {
//		if s == "?" {
//			h.addParam(t, "limit"+strconv.Itoa(n), "")
//			n++
//		}
//	}
//	return nil
//}
//
//// 检查select的column
//func (c *code) funcSelectCheckColumn(q *SelectStmt, t, j *db2go.Table) ([]string, string, error) {
//	var columns []string
//	var columnType string
//	// 确保column是同一种类型
//	checkColumnType := func(s string) error {
//		if columnType == "" {
//			columnType = s
//		} else if columnType != s {
//			return errColumnType
//		}
//		return nil
//	}
//	// 去掉t.c的"t."，单表时使用
//	removeColumnTable := func(s string) string {
//		p := strings.Split(s, ".")
//		if len(p) == 1 {
//			return s
//		}
//		return p[1]
//	}
//	// 先分析列
//	for _, col := range q.Column {
//		switch v := col.Expression.(type) {
//		case string:
//			if v == "*" {
//				err := checkColumnType("*")
//				if err != nil {
//					return nil, "", err
//				}
//				// 添加所有的列
//				if j == nil {
//					// 单表查询
//					for _, col := range t.Columns() {
//						columns = append(columns, removeColumnTable(col.Name()))
//					}
//				} else {
//					// join查询
//					for _, col := range t.Columns() {
//						columns = append(columns, t.Name()+"."+col.Name())
//					}
//					for _, col := range j.Columns() {
//						columns = append(columns, j.Name()+"."+col.Name())
//					}
//				}
//			} else {
//				err := checkColumnType("column")
//				if err != nil {
//					return nil, "", err
//				}
//				if j == nil {
//					columns = append(columns, removeColumnTable(v))
//				} else {
//					// 必须是table|alias.field形式
//					p := strings.Split(v, ".")
//					if len(p) != 2 {
//						return nil, "", fmt.Errorf("column '%s' ambiguous in case of table join", v)
//					}
//					if p[0] == q.Table.Alias {
//						p[0] = t.Name()
//					} else if p[0] == q.Join.Alias {
//						p[0] = j.Name()
//					}
//					columns = append(columns, strings.Join(p, "."))
//				}
//			}
//		case *FuncExpressionStmt:
//			var name string
//			if col.Alias != "" {
//				// 别名是否是表的列，不是就当成函数处理
//				if j == nil {
//					if t.GetColumn(col.Alias) != nil {
//						name = col.Alias
//					}
//				} else {
//					cc := t.GetColumn(col.Alias)
//					if cc == nil {
//						cc = j.GetColumn(col.Alias)
//						if cc != nil {
//							name = j.Name() + "." + col.Alias
//						}
//					} else {
//						name = t.Name() + "." + col.Alias
//					}
//				}
//			}
//			if name != "" {
//				// 别名是字段
//				err := checkColumnType("column")
//				if err != nil {
//					return nil, "", err
//				}
//				columns = append(columns, name)
//			} else {
//				// 没有别名，就是函数
//				err := checkColumnType("function")
//				if err != nil {
//					return nil, "", err
//				}
//				columns = append(columns, v.Name)
//			}
//		default:
//			// 没有名称
//			if col.Alias == "" {
//				return nil, "", errColumnType
//			}
//			var name string
//			if j == nil {
//				if t.GetColumn(col.Alias) == nil {
//					return nil, "", errColumnType
//				}
//				name = col.Alias
//			} else {
//				if t.GetColumn(col.Alias) == nil {
//					if j.GetColumn(col.Alias) == nil {
//						return nil, "", errColumnType
//					}
//					name = j.Name() + "." + col.Alias
//				} else {
//					name = t.Name() + "." + col.Alias
//				}
//			}
//			err := checkColumnType("column")
//			if err != nil {
//				return nil, "", err
//			}
//			columns = append(columns, name)
//		}
//	}
//	return columns, columnType, nil
//}
//
//func (c *code) funcSelectTable(q *SelectStmt) (t, j *db2go.Table, tp tpl.StructTPL, err error) {
//	tableName, err := c.tableName(&q.Table)
//	if err != nil {
//		return
//	}
//	t = c.schema.GetTable(tableName)
//	if t == nil {
//		return nil, nil, nil, c.errTableNotExists(tableName)
//	}
//	if q.Join != nil {
//		tableName, err = c.tableName(&q.Table)
//		if err != nil {
//			return
//		}
//		j = c.schema.GetTable(tableName)
//		if j == nil {
//			return nil, nil, nil, c.errTableNotExists(tableName)
//		}
//		// 如果没有联合的名称，就新建一个
//		name1 := SnakeCaseToPascalCase(t.Name())
//		name2 := SnakeCaseToPascalCase(j.Name())
//		name := name1 + "Join" + name2
//		_, ok := c.tplStruct[name]
//		if !ok {
//			t := new(tpl.JoinStruct)
//			t.Pkg = c.pkg
//			t.Struct1 = c.tplStruct[name1].(*tpl.Struct)
//			t.Struct2 = c.tplStruct[name2].(*tpl.Struct)
//			c.tplStruct[name] = t
//			tp = t
//		}
//		return
//	}
//	tp, err = c.StructTPL(t.Name())
//	if err != nil {
//		return nil, nil, nil, err
//	}
//	return
//}

//// Deprecated: 没什么用
//func (c *code) checkColumn(table, join *db2go.Table, tableAlias, joinAlias, column string) (string, string, error) {
//	columnPart := strings.Split(column, ".")
//	if len(columnPart) == 2 {
//		if (columnPart[0] == table.Name() || columnPart[1] == tableAlias) &&
//			table.GetColumn(columnPart[1]) != nil {
//			return table.Name(), columnPart[1], nil
//		}
//		if join != nil &&
//			(columnPart[0] == join.Name() || columnPart[1] == joinAlias) &&
//			join.GetColumn(columnPart[1]) != nil {
//			return join.Name(), columnPart[1], nil
//		}
//	} else {
//		if table.GetColumn(column) != nil {
//			return table.Name(), column, nil
//		}
//		if join != nil && join.GetColumn(column) != nil {
//			return join.Name(), column, nil
//		}
//	}
//	return "", "", c.errColumnNotExists(column)
//}
