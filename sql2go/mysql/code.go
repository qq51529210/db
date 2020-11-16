package mysql

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
)

func parseError(s string) error {
	return fmt.Errorf("parse error '%s'", s)
}

func goType(dataType string) (string, string, string) {
	dataType = strings.ToLower(dataType)
	switch dataType {
	case "tinyint":
		return "int8", "sql.NullInt32", "Int32"
	case "smallint":
		return "int16", "sql.NullInt32", "Int32"
	case "mediumint":
		return "int32", "sql.NullInt32", "Int32"
	case "int":
		return "int", "sql.NullInt64", "Int64"
	case "bigint":
		return "int64", "sql.NullInt64", "Int64"
	case "tinyint unsigned":
		return "uint8", "sql.NullInt32", "Int32"
	case "smallint unsigned":
		return "uint16", "sql.NullInt32", "Int32"
	case "mediumint unsigned":
		return "uint32", "sql.NullInt32", "Int32"
	case "int unsigned":
		return "uint", "sql.NullInt64", "Int64"
	case "bigint unsigned":
		return "uint64", "sql.NullInt64", "Int64"
	case "float":
		return "float32", "sql.NullFloat64", "Float64"
	case "double", "decimal":
		return "float64", "sql.NullFloat64", "Float64"
	case "tinyblob", "blob", "mediumblob", "longblob":
		return "[]byte", "", ""
	case "tinytext", "text", "mediumtext", "longtext":
		return "string", "sql.NullString", "String"
	case "date", "time", "year", "datetime", "timestamp":
		return "string", "sql.NullString", "String"
	default:
		if strings.HasPrefix(dataType, "binary") {
			return "[]byte", "", ""
		}
		if strings.HasPrefix(dataType, "decimal") {
			return "float64", "sql.NullFloat64", "Float64"
		}
		return "string", "sql.NullString", "String"
	}
}

type sqlSegment struct {
	string string
	value  string
	param  bool
	column bool
}

func NewCode(pkg, dbUrl string) (*Code, error) {
	c := new(Code)
	c.dbUrl = dbUrl
	c.file = new(fileTPL)
	c.file.Pkg = pkg
	c.file.Ipt = make(map[string]int)
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
		t, err = c.genQuery(function, tx, segments)
	} else {
		t, err = c.genExec(function, tx, segments)
	}
	if err != nil {
		return nil, err
	}
	switch t.(type) {
	case *querySqlTPL:
		c.file.Ipt["strings"] = 1
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

func (c *Code) genQuery(function, tx string, segments []*sqlSegment) (TPL, error) {
	// sql和参数和分页条件
	var str strings.Builder
	var page []*sqlSegment
	var params []*sqlSegment
	var testArgs []interface{}
	{
		for _, seg := range segments {
			if seg.param {
				if seg.column {
					page = append(page, seg)
					str.WriteString(seg.value)
				} else {
					params = append(params, seg)
					if strings.Contains(seg.value, "int") {
						testArgs = append(testArgs, 0)
					} else if strings.Contains(seg.value, "float") {
						testArgs = append(testArgs, float64(0))
					} else {
						testArgs = append(testArgs, "''")
					}
					str.WriteByte('?')
				}
			} else {
				str.WriteString(seg.string)
			}
		}
	}
	// 测试sql，获取结果集的字段信息
	var columns []*sql.ColumnType
	{
		db, err := sql.Open("mysql", c.dbUrl)
		if err != nil {
			return nil, err
		}
		defer func() {
			_ = db.Close()
		}()
		// 测试sql
		rows, err := db.Query(str.String(), testArgs...)
		if err != nil {
			return nil, err
		}
		// 获取结果集的字段信息
		columns, err = rows.ColumnTypes()
		if err != nil {
			return nil, err
		}
	}
	// 公共模板
	tp := new(tpl)
	tp.Sql = str.String()
	tp.Func = function
	tp.Stmt = tx
	// 无法预编译的sql
	{
		if page != nil && len(page) > 0 {
			t := new(querySqlTPL)
			t.tpl = tp
			for _, p := range page {
				t.Column = append(t.Column, p.string)
			}
			for _, p := range params {
				t.Param = append(t.Param, p.string)
			}
			for _, c := range columns {
				s := new(scanField)
				s.Name = snakeCaseToPascalCase(strings.Replace(c.Name(), ".", "_", -1))
				s.Type, s.NullType, s.NullValue = goType(c.DatabaseTypeName())
				if nul, ok := c.Nullable(); ok && !nul {
					s.NullType = ""
					s.NullValue = ""
				}
				t.Scan = append(t.Scan, s)
				var field [3]string
				field[0] = s.Name
				field[1] = s.Type
				field[2] = fmt.Sprintf("`json:\"%s\"`", pascalCaseToCamelCase(field[0]))
				t.Field = append(t.Field, field)
			}
			var segment strings.Builder
			segment.WriteByte('"')
			for _, s := range segments {
				if s.column {
					segment.WriteByte('"')
					t.Segment = append(t.Segment, segment.String())
					segment.Reset()
					segment.WriteByte('"')
					t.Segment = append(t.Segment, s.string)
				} else {
					if s.param {
						segment.WriteByte('?')
					} else {
						segment.WriteString(s.string)
					}
				}
			}
			if segment.Len() > 0 {
				segment.WriteByte('"')
				t.Segment = append(t.Segment, segment.String())
			}
			return t, nil
		}
	}
	tp.Stmt = c.stmtName(tp.Sql)
	// 只有一个结果，不生成结构体
	{
		if len(columns) < 2 {
			t := new(queryTPL)
			t.tpl = tp
			t.Type, t.NullType, t.NullValue = goType(columns[0].DatabaseTypeName())
			for _, p := range params {
				t.Param = append(t.Param, p.string)
			}
			return t, nil
		}
	}
	// 有多个结果，生成结构体
	{
		t := new(queryStructTPL)
		t.tpl = tp
		for _, p := range params {
			t.Param = append(t.Param, p.string)
		}
		for _, c := range columns {
			s := new(scanField)
			s.Name = snakeCaseToPascalCase(strings.Replace(c.Name(), ".", "_", -1))
			s.Type, s.NullType, s.NullValue = goType(c.DatabaseTypeName())
			if nul, ok := c.Nullable(); ok && !nul {
				s.NullType = ""
				s.NullValue = ""
			}
			t.Scan = append(t.Scan, s)
			var field [3]string
			field[0] = s.Name
			field[1] = s.Type
			field[2] = fmt.Sprintf("`json:\"%s\"`", pascalCaseToCamelCase(field[0]))
			t.Field = append(t.Field, field)
		}
		return t, nil
	}
}

func (c *Code) genExec(function, tx string, segments []*sqlSegment) (TPL, error) {
	// sql和参数
	var str strings.Builder
	var params []*sqlSegment
	{
		for _, seg := range segments {
			if seg.param {
				params = append(params, seg)
				str.WriteByte('?')
			} else {
				str.WriteString(seg.string)
			}
		}
	}
	// 测试sql
	{
		db, err := sql.Open("mysql", c.dbUrl)
		if err != nil {
			return nil, err
		}
		defer func() {
			_ = db.Close()
		}()
		_, err = db.Prepare(str.String())
		if err != nil {
			return nil, err
		}
	}
	// 公共模板
	tp := new(tpl)
	tp.Func = function
	tp.Tx = tx
	tp.Sql = str.String()
	tp.Stmt = c.stmtName(tp.Sql)
	// 只有一个入参，不生成结构体
	{
		if len(params) < 2 {
			t := new(execTPL)
			t.tpl = tp
			for _, p := range params {
				t.Param = append(t.Param, p.string)
			}
			return t, nil
		}
	}
	// 有多个入参，生成结构体
	{
		t := new(execStructTPL)
		t.tpl = tp
		t.Model = function + "Model"
		for _, p := range params {
			var field [3]string
			field[0] = snakeCaseToPascalCase(p.string)
			field[1] = p.value
			field[2] = fmt.Sprintf("`json:\"%s\"`", pascalCaseToCamelCase(field[0]))
			t.Field = append(t.Field, field)
		}
		return t, nil
	}
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
			// "{string:type}"或者"{string:column}"
			ss := strings.Split(s[1:i], ":")
			if len(ss) != 2 {
				return nil, parseError(s[:j])
			}
			// 是否基本类型
			switch ss[1] {
			case "int", "int8", "int16", "int32", "int64",
				"uint", "uint8", "uint16", "uint32", "uint64",
				"float32", "float64", "string", "[]byte":
				segments = append(segments, &sqlSegment{string: ss[0], value: ss[1], param: true})
			default:
				segments = append(segments, &sqlSegment{string: ss[0], value: ss[1], param: true, column: true})
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
