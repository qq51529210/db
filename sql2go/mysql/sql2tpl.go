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

func JoinCamelCaseToPascalCase(ss []string) string {
	var str strings.Builder
	for _, s := range ss {
		str.WriteString(CamelCaseToPascalCase(s))
	}
	return str.String()
}

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

func pick0(ss [][2]string) []string {
	var a []string
	for _, s := range ss {
		a = append(a, s[0])
	}
	return a
}

type holders struct {
	arg       []*tpl.Arg
	field     [][2]string
	param     [][2]string
	holder    []*holder
	sameParam map[string]int
}

type holder struct {
	table    string    // 表名
	column   [2]string // 1表名，2列名
	operator string    // 运算符
}

func (h *holders) paramName(column, operator string) string {
	if h.sameParam == nil {
		h.sameParam = make(map[string]int)
	}
	var str strings.Builder
	if column == "" {
		str.WriteString("param")
	} else {
		if (column[0] >= 'a' && column[0] <= 'z') ||
			(column[0] >= 'A' && column[0] <= 'Z') ||
			column[0] == '_' {
			str.WriteString(SnakeCaseToCamelCase(column))
			str.WriteString(operatorsNames[operator])
		} else {
			str.WriteString("param")
		}
	}
	// 如果有同名
	n, o := h.sameParam[str.String()]
	if !o {
		h.sameParam[str.String()] = 1
	} else {
		h.sameParam[str.String()] = n + 1
		str.WriteString(strconv.Itoa(n))
	}
	return str.String()
}

func (h *holders) addParam(column, operator string) {
	name := h.paramName(column, operator)
	// 添加并返回
	h.param = append(h.param, [2]string{name, operator})
	h.arg = append(h.arg, &tpl.Arg{Name: name})
}

func (h *holders) addExecField(table *db2go.Table, column, operator string) {
	columnPart := strings.Split(column, ".")
	if len(columnPart) == 2 {
		column = columnPart[1]
	}
	if table.GetColumn(column) != nil {
		h.addField(column, operator)
	} else {
		h.addParam(column, operator)
	}
}

func (h *holders) addQueryField(table, join *db2go.Table, query *SelectStmt, column, operator string, subQuery bool) {
	isColumn := func(part []string) *db2go.Table {
		if (part[0] == table.Name() || part[0] == query.TableAlias) && table.GetColumn(part[1]) != nil {
			return table
		}
		if join != nil && (part[0] == join.Name() || part[0] == query.JoinAlias) && join.GetColumn(part[1]) != nil {
			return join
		}
		return nil
	}
	columnPart := strings.Split(column, ".")
	// join表情况
	if len(columnPart) == 2 {
		// 如果是子查询
		if subQuery {
			h.addParam(strings.Join(columnPart, "_"), operator)
			return
		}
		if t := isColumn(columnPart); t != nil {
			h.addJoinField(t.Name(), columnPart[1], operator)
		}
		h.addParam(strings.Join(columnPart, "_"), operator)
		return
	}
	// 单表情况
	if subQuery {
		if table.GetColumn(column) != nil {
			h.addParam(table.Name()+"_"+column, operator)
			return
		}
		if join != nil && join.GetColumn(column) != nil {
			h.addParam(join.Name()+"_"+column, operator)
			return
		}
		h.addParam(column, operator)
		return
	}
	if table.GetColumn(column) != nil {
		h.addField(column, operator)
		return
	}
	if join != nil && join.GetColumn(column) != nil {
		h.addField(column, operator)
		return
	}
	// column有可能是别名
	for _, col := range query.Column {
		if col.Alias == column {
			name, ok := col.Expression.(string)
			if ok {
				columnPart := strings.Split(name, ".")
				if t := isColumn(columnPart); t != nil {
					h.addJoinField(t.Name(), columnPart[1], operator)
					return
				}
			}
			break
		}
	}
	h.addParam(column, operator)
}

// 单表的情况
func (h *holders) addField(column, operator string) {
	field := SnakeCaseToPascalCase(column)
	// 如果有两个相同的field，那么全部转成param
	for i, s := range h.field {
		if s[0] == field {
			// 去掉原来的field，加入到param
			h.field = append(h.field[:i], h.field[i+1:]...)
			name := h.paramName(field, s[1])
			h.param = append(h.param, [2]string{name, s[1]})
			// 修改arg
			for _, a := range h.arg {
				if a.IsField && a.Name == field {
					a.IsField = false
					a.Name = name
					break
				}
			}
			// 再添加
			h.addParam(field, operator)
			return
		}
	}
	h.field = append(h.field, [2]string{field, operator})
	h.arg = append(h.arg, &tpl.Arg{Name: field, IsField: true})
}

// join表的情况
func (h *holders) addJoinField(table, column, operator string) {
	field := SnakeCaseToPascalCase(table) + "." + SnakeCaseToPascalCase(column)
	// 如果有两个相同的field，那么全部转成param
	for i, s := range h.field {
		if s[0] == field {
			// 去掉原来的field，加入到param
			h.field = append(h.field[:i], h.field[i+1:]...)
			name := h.paramName(table+"_"+column, s[1])
			h.param = append(h.param, [2]string{name, s[1]})
			// 修改arg
			for _, a := range h.arg {
				if a.IsField && a.Name == field {
					a.IsField = false
					a.Name = name
					break
				}
			}
			// 再添加
			h.addParam(table+"_"+column, operator)
			return
		}
	}
	h.field = append(h.field, [2]string{field, operator})
	h.arg = append(h.arg, &tpl.Arg{Name: field, IsField: true})
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
	if q.Join != "" {
		return c.funcSelectJoin(q, sql, tx)
	}
	t := c.funcGetTable(q.Table)
	// 占位符
	var h holders
	c.parseSelectHolder(q, &h, false)
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
		tp.Param = pick0(h.param)
		tp.Struct = structName
		tp.StmtName = c.funcStmtName()
		tp.Query = h.arg
		for _, col := range columns {
			tp.Scan = append(tp.Scan, SnakeCaseToPascalCase(col))
		}
		var byFields []string
		for _, a := range h.arg {
			byFields = append(byFields, a.Name)
		}
		if len(byFields) > 0 || len(h.param) > 0 {
			funcName.WriteString("By")
			funcName.WriteString(strings.Join(byFields, ""))
		}
		tp.Func = funcName.String()
		return tp
	case "function":
		tp := new(tpl.QueryRow)
		tp.Tx = tx
		tp.Sql = sql
		tp.Param = pick0(h.param)
		tp.Struct = structName
		tp.StmtName = c.funcStmtName()
		for _, name := range columns {
			tp.Return = append(tp.Return, functions[name])
		}
		// 函数名称
		funcName.WriteString(structName)
		funcName.WriteString(JoinSnakeCaseToPascalCase(columns))
		if len(h.param) > 0 {
			funcName.WriteString("By")
			funcName.WriteString(JoinCamelCaseToPascalCase(tp.Param))
		}
		tp.Func = funcName.String()
		return tp
	default: // *
		tp := new(tpl.StructQueryRow)
		tp.Tx = tx
		tp.Sql = strings.Replace(sql, "*", strings.Join(columns, ","), 1)
		tp.Param = pick0(h.param)
		tp.Struct = structName
		tp.Query = h.arg
		tp.Scan = pick0(h.field)
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
	// 分析
	for _, col := range q.Column {
		switch v := col.Expression.(type) {
		case string:
			// 全选
			if v == "*" {
				checkColumnType("*")
				// 添加所有的列
				for _, col := range t.Columns() {
					columns = append(columns, col.Name())
				}
				continue
			}
			// 普通的列
			checkColumnType("column")
			columns = append(columns, v)
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
	structName := SnakeCaseToPascalCase(t.Name())
	switch q.Value.(type) {
	case *SelectStmt:
		value := q.Value.(*SelectStmt)
		// 解析占位符
		var h holders
		c.parseSelectHolder(value, &h, true)
		tp := new(tpl.Exec)
		tp.Tx = tx
		tp.Sql = sql
		tp.Param = pick0(h.param)
		tp.Struct = structName
		tp.StmtName = c.funcStmtName()
		// 函数名
		var funcName strings.Builder
		funcName.WriteString("Insert")
		funcName.WriteString(tp.Struct)
		funcName.WriteString(strings.Join(pick0(h.field), ""))
		funcName.WriteString("From")
		funcName.WriteString(value.Table)
		if len(h.param) > 0 {
			funcName.WriteString("By")
			funcName.WriteString(JoinCamelCaseToPascalCase(tp.Param))
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
		for i, v := range values {
			switch s := v.(type) {
			case string:
				if s == "?" {
					h.addField(columns[i], "=")
				}
			default:
				c.parseExecHolder(t, v, &h)
			}
		}
		tp := new(tpl.StructExec)
		tp.Tx = tx
		tp.Sql = sql
		tp.Param = pick0(h.param)
		tp.Struct = structName
		tp.StmtName = c.funcStmtName()
		tp.Arg = h.arg
		// 函数名
		var funcName strings.Builder
		funcName.WriteString("Insert")
		funcName.WriteString(strings.Join(pick0(h.param), ""))
		tp.Func = funcName.String()
		return tp
	}
}

func (c *code) funcDelete(q *DeleteStmt, sql string, tx bool) tpl.FuncTPL {
	t := c.funcGetTable(q.Table)
	var h holders
	// 解析where占位符
	c.parseExecHolder(t, q.Where, &h)
	// 模板
	tp := new(tpl.StructExec)
	tp.Tx = tx
	tp.Sql = sql
	tp.Param = pick0(h.param)
	tp.Struct = SnakeCaseToPascalCase(t.Name())
	tp.StmtName = c.funcStmtName()
	tp.Arg = h.arg
	// 函数名
	var funcName strings.Builder
	funcName.WriteString("Delete")
	if len(h.field) > 0 {
		funcName.WriteString("By")
		funcName.WriteString(strings.Join(pick0(h.field), ""))
	}
	tp.Func = funcName.String()
	return tp
}

func (c *code) funcUpdate(q *UpdateStmt, sql string, tx bool) tpl.FuncTPL {
	t := c.funcGetTable(q.Table)
	var h holders
	// 解析column=expression占位符
	for _, col := range q.Column {
		c.parseExecHolder(t, col, &h)
	}
	fieldsCount := len(h.field)
	// 解析where占位符
	c.parseExecHolder(t, q.Where, &h)
	// 模板
	tp := new(tpl.StructExec)
	tp.Tx = tx
	tp.Sql = sql
	tp.Param = pick0(h.param)
	tp.Struct = SnakeCaseToPascalCase(t.Name())
	tp.StmtName = c.funcStmtName()
	tp.Arg = h.arg
	// 函数名
	var funcName strings.Builder
	funcName.WriteString("Update")
	funcName.WriteString(strings.Join(pick0(h.field[:fieldsCount]), ""))
	if len(h.field[fieldsCount:]) > 0 || len(h.param) > 0 {
		funcName.WriteString("By")
		funcName.WriteString(strings.Join(pick0(h.field[fieldsCount:]), ""))
		funcName.WriteString(JoinCamelCaseToPascalCase(tp.Param))
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

// o为true表示子查询
func (c *code) parseExecHolder(t *db2go.Table, v interface{}, h *holders) {
	if v == nil {
		return
	}
	switch v := v.(type) {
	case string:
		// 单个，没有列名和操作符
		if v == "?" {
			h.addParam("", "")
		}
	case *ExpressionStmt:
		left, ok := v.Left.(string)
		if ok {
			if left == "?" {
				h.addParam("", "")
			}
		} else {
			c.parseExecHolder(t, v.Left, h)
		}
		right, ok := v.Right.(string)
		if ok {
			if right == "?" {
				if left != "" {
					h.addExecField(t, left, v.Operator)
				} else {
					h.addParam("", "")
				}
			}
		} else {
			c.parseExecHolder(t, v.Right, h)
		}
	case *BoolExpressionStmt:
		c.parseExecHolder(t, v.Value, h)
	case *FuncExpressionStmt:
		c.parseExecHolder(t, v.Value, h)
	case *SelectStmt:
		c.parseSelectHolder(v, h, true)
	case []interface{}:
		for _, i := range v {
			c.parseExecHolder(t, i, h)
		}
	}
}

// o为true表示子查询
func (c *code) parseSelectHolder(q *SelectStmt, h *holders, o bool) {
	if q == nil {
		return
	}
	t := c.funcGetTable(q.Table)
	var j *db2go.Table
	if q.Join != "" {
		j = c.funcGetTable(q.Join)
	}
	// distinct
	if q.Distinct == "?" {
		h.addParam(q.Table+"_distinct", "")
	}
	// columns
	for _, col := range q.Column {
		c.parseQueryHolder(t, j, q, col.Expression, h, o)
	}
	// on
	c.parseQueryHolder(t, j, q, q.On, h, o)
	// where
	c.parseQueryHolder(t, j, q, q.Where, h, o)
	// group by
	for _, s := range q.GroupBy {
		if s == "?" {
			h.addParam(q.Table+"_group", "")
		}
	}
	// having
	c.parseQueryHolder(t, j, q, q.Having, h, o)
	// union
	c.parseSelectHolder(q.Union, h, o)
	// order by
	for _, s := range q.OrderBy {
		if s == "?" {
			h.addParam(q.Table+"_order", "")
		}
	}
	// asc/desc
	if q.Order == "?" {
		h.addParam(q.Table+"_sort", "")
	}
	// limit
	for _, s := range q.Limit {
		if s == "?" {
			h.addParam(q.Table+"_limit", "")
		}
	}
}

// o为true表示子查询
func (c *code) parseQueryHolder(t, j *db2go.Table, q *SelectStmt, v interface{}, h *holders, o bool) {
	if v == nil {
		return
	}
	switch v := v.(type) {
	case string:
		// 单个，没有列名和操作符
		if v == "?" {
			h.addParam("", "")
		}
	case *ExpressionStmt:
		left, ok := v.Left.(string)
		if ok {
			if left == "?" {
				h.addParam("", "")
			}
		} else {
			c.parseQueryHolder(t, j, q, v.Left, h, true)
		}
		right, ok := v.Right.(string)
		if ok {
			if right == "?" {
				if left != "" {
					h.addQueryField(t, j, q, left, v.Operator, o)
				} else {
					h.addParam("", "")
				}
			}
		} else {
			c.parseQueryHolder(t, j, q, v.Left, h, true)
		}
	case *BoolExpressionStmt:
		c.parseQueryHolder(t, j, q, v.Value, h, true)
	case *FuncExpressionStmt:
		c.parseQueryHolder(t, j, q, v.Value, h, true)
	case *SelectStmt:
		c.parseSelectHolder(v, h, true)
	case []interface{}:
		for _, i := range v {
			c.parseQueryHolder(t, j, q, i, h, true)
		}
	}
}
