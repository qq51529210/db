package mysql

import (
	"fmt"
	"os"
	"strings"
)

func parseError(s string) error {
	return fmt.Errorf("parse error '%s'", s)
}

type sqlSegment struct {
	string string
	_type  string
	param  bool
}

func NewCode(pkg, dbUrl string) (*Code, error) {
	c := new(Code)
	c.dbUrl = dbUrl
	c.file = new(fileTPL)
	c.file.Pkg = pkg
	return c, nil
}

type Code struct {
	file  *fileTPL
	dbUrl string
}

func (c *Code) Gen(sql, function, tx string) (TPL, error) {
	// 是否query
	var isQuery bool
	{
		i := strings.IndexByte(sql, ' ')
		if i < 0 {
			return nil, parseError(sql)
		}
		switch strings.ToLower(sql[:i]) {
		case "select":
			isQuery = true
		case "insert", "update", "delete":
		default:
			return nil, parseError(sql)
		}
	}
	// sql片段
	segments, err := c.parseSegments(sql)
	if err != nil {
		return nil, err
	}
	// 模板
	var t TPL
	if isQuery {

	} else {
		t = c.genExec(function, tx, segments)
	}
	c.file.TPL = append(c.file.TPL, t)
	return t, nil
}

func (c *Code) SaveFile(file string) error {
	// 输出模板
	f, err := os.OpenFile(file, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	// 关闭文件
	defer func() { _ = f.Close() }()
	// 输出
	return c.file.Execute(f)
}

func (c *Code) genQuery(function, tx string, segments []*sqlSegment) TPL {
	return nil
}

func (c *Code) genExec(function, tx string, segments []*sqlSegment) TPL {
	//
	var params []*sqlSegment
	for _, seg := range segments {
		if seg.param {
			params = append(params, seg)
		}
	}
	//
	tp := new(tpl)
	tp.Func = function
	tp.Tx = tx
	tp.Sql = combineSegmentsTo(segments)
	tp.Stmt = c.stmtName(tp.Sql)
	//
	if len(params) < 2 {
		t := new(execTPL)
		t.tpl = tp
		for _, p := range params {
			t.Param = append(t.Param, snakeCaseToCamelCase(p.string))
		}
		return t
	}
	//
	t := new(execStructTPL)
	t.tpl = tp
	t.Struct = function + "Model"
	for _, p := range params {
		var field [3]string
		field[0] = snakeCaseToPascalCase(p.string)
		field[1] = p._type
		field[2] = fmt.Sprintf("`json:\"%s\"`", pascalCaseToCamelCase(field[0]))
		t.Field = append(t.Field, field)
	}
	return t
}

func (c *Code) parseSegments(s string) ([]*sqlSegment, error) {
	// 解析sql片段
	var segments []*sqlSegment
	s = strings.TrimSpace(s)
	if s == "" {
		return segments, nil
	}
	i := 0
Loop:
	for ; i < len(s); i++ {
		switch s[i] {
		case '\'':
			j := i
			i++
			for ; i < len(s); i++ {
				if s[i] == '\'' && s[i-1] != '\\' {
					i++
					continue Loop
				}
			}
			return nil, parseError(s[j:])
		case '{':
			// "{}"前的sql
			if i != 0 {
				segments = append(segments, &sqlSegment{string: s[:i]})
				s = s[i:]
			}
			// "{}"
			i = strings.IndexByte(s, '}')
			if i < 0 {
				return nil, parseError(s)
			}
			j := i + 1
			// "{string:type}"
			ss := strings.Split(s[1:i], ":")
			switch len(ss) {
			case 1:
				segments = append(segments, &sqlSegment{string: ss[0], param: true})
			case 2:
				segments = append(segments, &sqlSegment{string: ss[0], _type: ss[1], param: true})
			default:
				return nil, parseError(s[:j])
			}
			s = s[j:]
			i = 0
		}
	}
	if s != "" {
		segments = append(segments, &sqlSegment{string: s})
	}
	return segments, nil
}

func (c *Code) stmtName(sql string) string {
	s := fmt.Sprintf("stmt%d", len(c.file.Sql))
	c.file.Sql = append(c.file.Sql, sql)
	return s
}

func combineSegmentsTo(segments []*sqlSegment) string {
	var str strings.Builder
	for _, s := range segments {
		if s.param {
			str.WriteByte('?')
		} else {
			str.WriteString(s.string)
		}
	}
	//str.WriteByte(' ')
	//var last string
	//for _, seg := range segments {
	//	if seg.param {
	//		str.WriteByte('?')
	//		last = "?"
	//	} else {
	//		if last != "" {
	//			c1 := last[len(last)-1]
	//			if c1 == ')' {
	//				if seg.string[0] != ',' && seg.string[0] != ')' {
	//					str.WriteByte(' ')
	//				}
	//			} else if c1 == '?' {
	//				if seg.string[0] != ',' && seg.string[0] != ')' {
	//					str.WriteByte(' ')
	//				}
	//			}
	//		}
	//		str.WriteString(seg.string)
	//		last = seg.string
	//	}
	//}
	return str.String()
}

//type Code interface {
//	StructTPL(table string) StructTPL
//	GenFunc(f *Func) (FuncTPL, error)
//	GenDefault(table string) ([]FuncTPL, error)
//	SaveFiles(dir string, clean bool) error
//}
//
//func NewCode(pkg, dbUrl string) (Code, error) {
//	schema, err := db2go.ReadSchema(db2go.MYSQL, dbUrl)
//	if err != nil {
//		return nil, err
//	}
//	if pkg == "" {
//		pkg = pascalCaseToSnakeCase(schema.Name())
//	}
//	c := new(code)
//	c.schema = schema
//	c.pkg = pkg
//	// struct模板
//	c.tplStruct = make(map[string]StructTPL)
//	for _, t := range schema.Tables() {
//		tp := new(structTPL)
//		tp.pkg = pkg
//		tp.name = snakeCaseToPascalCase(t.Name())
//		tp.Import = make(map[string]int)
//		tp.Table = t.Name()
//		for _, c := range t.Columns() {
//			name := snakeCaseToPascalCase(c.Name())
//			tp.addField(name, c.GoType(),
//				fmt.Sprintf("`json:\"%s,omitempy\"`", snakeCaseToCamelCase(c.Name())))
//			scan := new(scanTPL)
//			scan.Name = name
//			if c.IsNullable() {
//				scan.NullType, scan.NullValue = sqlNullType(c.GoType())
//			}
//			tp.Scan = append(tp.Scan, scan)
//		}
//		c.tplStruct[t.Name()] = tp
//	}
//	// init模板
//	c.tplInit = new(initTPL)
//	c.tplInit.Pkg = pkg
//	c.tplInit.DBType = db2go.MYSQL
//	c.tplInit.DBPkg = db2go.DriverPkg(db2go.MYSQL)
//	return c, nil
//}
//
//type code struct {
//	pkg        string
//	schema     *db2go.Schema
//	tplStruct  map[string]StructTPL
//	tplInit    *initTPL
//	funcSelect map[string]*selectStmt
//}
//
//func (c *code) StructTPL(table string) StructTPL {
//	tp, _ := c.tplStruct[table]
//	return tp
//}
//
//func (c *code) SaveFiles(dir string, clean bool) error {
//	// 输出目录
//	dir = filepath.Join(dir, c.pkg)
//	// 先删除
//	if clean {
//		err := os.RemoveAll(dir)
//		if err != nil {
//			return err
//		}
//	}
//	// 再创建
//	err := os.MkdirAll(dir, os.ModePerm)
//	if err != nil {
//		return err
//	}
//	// 输出struct模板
//	for k, v := range c.tplStruct {
//		err = saveTemplate(v, filepath.Join(dir, pascalCaseToSnakeCase(k)+".go"))
//		if err != nil {
//			return err
//		}
//	}
//	// 输出init模板
//	return saveTemplate(c.tplInit, filepath.Join(dir, c.pkg+".init.go"))
//}
//
//func (c *code) GenFunc(f *Func) (FuncTPL, error) {
//	funcName := strings.TrimSpace(f.Name)
//	if funcName == "" {
//		return nil, errors.New("function name is required")
//	}
//	sql := strings.TrimSpace(f.Sql)
//	if sql == "" {
//		return nil, errors.New("sql is required")
//	}
//	// 解析单词
//	tokens := readTokens(f.Sql)
//	// 生成模板
//	var tp FuncTPL
//	var err error
//	switch strings.ToLower(tokens[0]) {
//	case "select":
//		tp, err = c.genSelect(tokens[1:], funcName, f.Tx)
//	case "insert":
//		tp, err = c.genInsert(tokens[1:], funcName, f.Tx)
//	case "update":
//		tp, err = c.genUpdate(tokens[1:], funcName, f.Tx)
//	case "delete":
//		tp, err = c.genDelete(tokens[1:], funcName, f.Tx)
//	default:
//		return nil, fmt.Errorf("unsupport sql '%s'", f.Sql)
//	}
//	// 生成模板出错
//	if err != nil {
//		return nil, err
//	}
//	// 如果是预定义的，添加到初始化中
//	if tp.SQL() != "" {
//		tp.setStmt(c.tplInit.addSql(tp.SQL()))
//	}
//	// 添加到struct模板
//	c.StructTPL(tp.tableName()).addFunc(tp)
//	return tp, nil
//}
//
//func (c *code) GenDefault(table string) ([]FuncTPL, error) {
//	t := c.schema.GetTable(table)
//	if t == nil {
//		return nil, fmt.Errorf("unknown table '%s'", table)
//	}
//	// 各种约束列
//	pk, _, uni, _, mul, nmul := c.genDefaultPickColumns(t)
//	var def []*Func
//	def = append(def, c.genDefaultCount(t))
//	def = append(def, c.genDefaultList(t))
//	def = append(def, c.genDefaultInsert(t))
//	// 单键
//	for _, col := range pk {
//		field := c.genDefaultPickDiffColumns(col, t.Columns())
//		condition := []*db2go.Column{col}
//		def = append(def, c.genDefaultSelect(t, field, condition))
//		def = append(def, c.genDefaultUpdate(t, field, condition))
//		def = append(def, c.genDefaultDelete(t, condition))
//	}
//	for _, col := range uni {
//		field := c.genDefaultPickDiffColumns(col, t.Columns())
//		condition := []*db2go.Column{col}
//		def = append(def, c.genDefaultSelect(t, field, condition))
//		def = append(def, c.genDefaultUpdate(t, field, condition))
//		def = append(def, c.genDefaultDelete(t, condition))
//	}
//	// 多键
//	def = append(def, c.genDefaultSelect(t, nmul, mul))
//	def = append(def, c.genDefaultUpdate(t, nmul, mul))
//	def = append(def, c.genDefaultDelete(t, mul))
//	// 生成
//	var tps []FuncTPL
//	for _, s := range def {
//		if s == nil {
//			continue
//		}
//		tp, err := c.GenFunc(s)
//		if err != nil {
//			return nil, err
//		}
//		tps = append(tps, tp)
//	}
//	return tps, nil
//}
//
//func (c *code) genSelect(tokens []string, funcName string, tx bool) (FuncTPL, error) {
//	var q selectStmt
//	err := parseSelect(&q, tokens, c.schema)
//	if err != nil {
//		return nil, err
//	}
//	// 查询分组/排序，无法预编译sql
//	if q.isQueryStructPage() {
//		// 全部转成param
//		q.args.ToParam()
//		tp := new(queryStructPageTPL)
//		tp.table = q.tableName()
//		//tp.sql = q.sql.String()
//		tp.funcName = funcName
//		tp.tx = tx
//		tp.args = q.args
//		tp.Struct = snakeCaseToPascalCase(tp.table)
//		if !q.selectAll {
//			tp.Scan = append(tp.Scan, q.scan...)
//		}
//		tp.ColumnParam, tp.Segment = selectSqlSegments(&q)
//		tp.Model = !q.selectAll
//		return tp, nil
//	}
//	// 预编译
//	q.prepareSQL()
//	// 查询函数
//	{
//		if len(q.funcReturn) > 0 {
//			// 全部转成param
//			q.args.ToParam()
//			// 模板
//			tp := new(queryFuncTPL)
//			tp.table = q.tableName()
//			tp.sql = q.sql.String()
//			tp.funcName = funcName
//			tp.tx = tx
//			tp.Type = append(tp.Type, q.funcReturn...)
//			tp.args = q.args
//			return tp, nil
//		}
//	}
//	// 查询列表
//	{
//		if q.table2 != nil {
//			// 生成table1_join_table2结构
//			name := q.tableName()
//			_, ok := c.tplStruct[name]
//			if !ok {
//				stp := new(structTPL)
//				stp.pkg = c.pkg
//				stp.name = snakeCaseToPascalCase(name)
//				stp.addField(snakeCaseToPascalCase(q.table1.Name()), "", "")
//				stp.addField(snakeCaseToPascalCase(q.table2.Name()), "", "")
//				stp.Import = make(map[string]int)
//				//stp.Table = t.Name()
//				c.tplStruct[name] = stp
//			}
//			// 全部转成param
//			q.args.ToParam()
//			// 模板
//			tp := new(queryStructListTPL)
//			tp.table = q.tableName()
//			tp.sql = q.sql.String()
//			tp.funcName = funcName
//			tp.tx = tx
//			tp.args = q.args
//			tp.Struct = snakeCaseToPascalCase(tp.table)
//			tp.Scan = append(tp.Scan, q.scan...)
//			tp.Model = !q.selectAll
//			return tp, nil
//		}
//	}
//	// 查询单行
//	{
//		if q.isQueryStruct() {
//			tp := new(queryStructRowTPL)
//			tp.table = q.tableName()
//			tp.sql = q.sql.String()
//			tp.funcName = funcName
//			tp.tx = tx
//			tp.args = q.args
//			tp.Struct = snakeCaseToPascalCase(tp.table)
//			tp.Scan = append(tp.Scan, q.scan...)
//			tp.Model = !q.selectAll
//			if tp.Model {
//				tp.args.ToParam()
//			}
//			return tp, nil
//		}
//	}
//	// 查询列表
//	{
//		// 全部转成param
//		q.args.ToParam()
//		// 模板
//		tp := new(queryStructListTPL)
//		tp.table = q.tableName()
//		tp.sql = q.sql.String()
//		tp.funcName = funcName
//		tp.tx = tx
//		tp.args = q.args
//		tp.Struct = snakeCaseToPascalCase(tp.table)
//		tp.Scan = append(tp.Scan, q.scan...)
//		tp.Model = !q.selectAll
//		return tp, nil
//	}
//}
//
//func (c *code) genInsert(tokens []string, funcName string, tx bool) (FuncTPL, error) {
//	var q insertStmt
//	err := parseInsert(&q, tokens, c.schema)
//	if err != nil {
//		return nil, err
//	}
//	return c.genExec(q.sql.String(), funcName, tx, q.table.Name(), q.args)
//}
//
//func (c *code) genUpdate(tokens []string, funcName string, tx bool) (FuncTPL, error) {
//	var q updateStmt
//	err := parseUpdate(&q, tokens, c.schema)
//	if err != nil {
//		return nil, err
//	}
//	return c.genExec(q.sql.String(), funcName, tx, q.table.Name(), q.args)
//}
//
//func (c *code) genDelete(tokens []string, funcName string, tx bool) (FuncTPL, error) {
//	var q deleteStmt
//	err := parseDelete(&q, tokens, c.schema)
//	if err != nil {
//		return nil, err
//	}
//	return c.genExec(q.sql.String(), funcName, tx, q.table.Name(), q.args)
//}
//
//func (c *code) genExec(sql, funcName string, tx bool, table string, args argsTPL) (FuncTPL, error) {
//	// 模板
//	if args == nil || !args.HasField() {
//		tp := new(execTPL)
//		tp.table = table
//		tp.sql = sql
//		tp.funcName = funcName
//		tp.tx = tx
//		tp.args = args
//		return tp, nil
//	}
//	tp := new(execStructTPL)
//	tp.table = table
//	tp.sql = sql
//	tp.funcName = funcName
//	tp.tx = tx
//	tp.args = args
//	tp.Struct = snakeCaseToPascalCase(tp.table)
//	return tp, nil
//}
//
//func (c *code) genDefaultCount(table *db2go.Table) *Func {
//	var sql strings.Builder
//	sql.WriteString("select count(*){int64} from ")
//	sql.WriteString(table.Name())
//	return &Func{
//		Name: snakeCaseToPascalCase(table.Name()) + "Count",
//		Sql:  sql.String(),
//	}
//}
//
//func (c *code) genDefaultList(table *db2go.Table) *Func {
//	var sql strings.Builder
//	sql.WriteString("select * from ")
//	sql.WriteString(table.Name())
//	sql.WriteString(" order by {order} {sort} limit {begin}, {total}")
//	return &Func{
//		Name: snakeCaseToPascalCase(table.Name()) + "List",
//		Sql:  sql.String(),
//	}
//}
//
//func (c *code) genDefaultInsert(table *db2go.Table) *Func {
//	var sql strings.Builder
//	sql.WriteString("insert into ")
//	sql.WriteString(table.Name())
//	sql.WriteByte('(')
//	c.genDefaultFields(&sql, table.Columns())
//	sql.WriteByte(')')
//	column := table.Columns()
//	sql.WriteString(" values({")
//	sql.WriteString(column[0].Name())
//	sql.WriteByte('}')
//	for i := 1; i < len(column); i++ {
//		sql.WriteString(",{")
//		sql.WriteString(column[i].Name())
//		sql.WriteByte('}')
//	}
//	sql.WriteByte(')')
//	return &Func{
//		Name: "Insert",
//		Sql:  sql.String(),
//	}
//}
//
//func (c *code) genDefaultSelect(table *db2go.Table, fields, condition []*db2go.Column) *Func {
//	if len(condition) < 1 {
//		return nil
//	}
//	f := new(Func)
//	var sql strings.Builder
//	sql.WriteString("SelectBy")
//	c.genDefaultJoinPascalCase(&sql, condition)
//	f.Name = sql.String()
//	sql.Reset()
//	sql.WriteString("select ")
//	c.genDefaultFields(&sql, fields)
//	sql.WriteString(" from ")
//	sql.WriteString(table.Name())
//	sql.WriteString(" where ")
//	c.genDefaultConditions(&sql, condition)
//	f.Sql = sql.String()
//	return f
//}
//
//func (c *code) genDefaultUpdate(table *db2go.Table, fields, condition []*db2go.Column) *Func {
//	if len(condition) < 1 {
//		return nil
//	}
//	column := make([]*db2go.Column, 0)
//	for _, col := range fields {
//		if !col.IsAutoIncrement() {
//			column = append(column, col)
//		}
//	}
//	if len(column) < 1 {
//		return nil
//	}
//	f := new(Func)
//	var sql strings.Builder
//	sql.WriteString("UpdateBy")
//	c.genDefaultJoinPascalCase(&sql, condition)
//	f.Name = sql.String()
//	sql.Reset()
//	sql.WriteString("update ")
//	sql.WriteString(table.Name())
//	sql.WriteString(" set ")
//	sql.WriteString(column[0].Name())
//	sql.WriteString("={")
//	sql.WriteString(column[0].Name())
//	sql.WriteString("}")
//	for i := 1; i < len(column); i++ {
//		sql.WriteByte(',')
//		sql.WriteString(column[i].Name())
//		sql.WriteString("={")
//		sql.WriteString(column[i].Name())
//		sql.WriteString("}")
//	}
//	sql.WriteString(" where ")
//	c.genDefaultConditions(&sql, condition)
//	f.Sql = sql.String()
//	return f
//}
//
//func (c *code) genDefaultDelete(table *db2go.Table, condition []*db2go.Column) *Func {
//	if len(condition) < 1 {
//		return nil
//	}
//	f := new(Func)
//	var sql strings.Builder
//	sql.WriteString("DeleteBy")
//	c.genDefaultJoinPascalCase(&sql, condition)
//	f.Name = sql.String()
//	sql.Reset()
//	sql.WriteString("delete from ")
//	sql.WriteString(table.Name())
//	sql.WriteString(" where ")
//	c.genDefaultConditions(&sql, condition)
//	f.Sql = sql.String()
//	return f
//}
//
//func (c *code) genDefaultJoinPascalCase(sql *strings.Builder, fields []*db2go.Column) {
//	if len(fields) < 1 {
//		return
//	}
//	sql.WriteString(snakeCaseToPascalCase(fields[0].Name()))
//	for i := 1; i < len(fields); i++ {
//		sql.WriteString(snakeCaseToPascalCase(fields[i].Name()))
//	}
//}
//
//func (c *code) genDefaultFields(sql *strings.Builder, fields []*db2go.Column) {
//	if len(fields) < 1 {
//		return
//	}
//	sql.WriteString(fields[0].Name())
//	for i := 1; i < len(fields); i++ {
//		sql.WriteString(",")
//		sql.WriteString(fields[i].Name())
//	}
//}
//
//func (c *code) genDefaultConditions(sql *strings.Builder, condition []*db2go.Column) {
//	if len(condition) < 1 {
//		return
//	}
//	sql.WriteString(condition[0].Name())
//	sql.WriteString("={")
//	sql.WriteString(condition[0].Name())
//	sql.WriteString("}")
//	for i := 1; i < len(condition); i++ {
//		sql.WriteString(" and ")
//		sql.WriteString(condition[i].Name())
//		sql.WriteString("={")
//		sql.WriteString(condition[i].Name())
//		sql.WriteString("}")
//	}
//}
//
//func (c *code) genDefaultPickColumns(table *db2go.Table) (pk, npk, uni, nuni, mul, nmul []*db2go.Column) {
//	for _, c := range table.Columns() {
//		if c.IsPrimaryKey() {
//			pk = append(pk, c)
//		} else {
//			npk = append(npk, c)
//		}
//		if c.IsUnique() {
//			uni = append(uni, c)
//		} else {
//			nuni = append(nuni, c)
//		}
//		if c.IsMulUnique() {
//			mul = append(mul, c)
//		} else {
//			nmul = append(nmul, c)
//		}
//	}
//	return
//}
//
//func (c *code) genDefaultPickDiffColumns(column *db2go.Column, columns []*db2go.Column) []*db2go.Column {
//	var diff []*db2go.Column
//	for _, c := range columns {
//		if c != column {
//			diff = append(diff, c)
//		}
//	}
//	return diff
//}
