package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	errEmptyFunctionName = errors.New("empty function name")
	errEmptySQL          = errors.New("empty sql")
)

func snakeCaseToPascalCase(s string) string {
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

func pascalCaseToCamelCase(s string) string {
	return strings.ToLower(s[:1]) + s[1:]
}

func parseError(s string) error {
	return fmt.Errorf("parse error '%s'", s)
}

func goType(dataType string) string {
	dataType = strings.ToLower(dataType)
	switch dataType {
	case "tinyint":
		return "int8"
	case "smallint":
		return "int16"
	case "mediumint":
		return "int32"
	case "int":
		return "int"
	case "bigint":
		return "int64"
	case "tinyint unsigned":
		return "uint8"
	case "smallint unsigned":
		return "uint16"
	case "mediumint unsigned":
		return "uint32"
	case "int unsigned":
		return "uint"
	case "bigint unsigned":
		return "uint64"
	case "float":
		return "float32"
	case "double", "decimal":
		return "float64"
	case "tinyblob", "blob", "mediumblob", "longblob":
		return "[]byte"
	case "tinytext", "text", "mediumtext", "longtext":
		return "string"
	case "year":
		return "uint16"
	case "date", "time", "datetime", "timestamp":
		return "string"
	default:
		if strings.HasPrefix(dataType, "binary") {
			return "[]byte"
		}
		if strings.HasPrefix(dataType, "decimal") {
			return "float64"
		}
		return "string"
	}
}

func goNullType(typ string) string {
	if strings.Contains(typ, "int") {
		return "sql.NullInt64"
	} else if strings.Contains(typ, "float") {
		return "sql.NullFloat64"
	} else {
		return "sql.NullString"
	}
}

func dbColumnToScanTPL(c *sql.ColumnType, nullField bool) *scanTPL {
	t := new(scanTPL)
	t.Name = snakeCaseToPascalCase(strings.Replace(c.Name(), ".", "_", -1))
	t.Type = goType(c.DatabaseTypeName())
	if na, ok := c.Nullable(); nullField && ok && na {
		if strings.Contains(t.Type, "int") {
			t.NullType = "sql.NullInt64"
			t.NullValue = "Int64"
			t.NullType2 = "int64"
		} else if strings.Contains(t.Type, "float") {
			t.NullType = "sql.NullFloat64"
			t.NullValue = "Float64"
			t.NullType2 = "float64"
		} else {
			t.NullType = "sql.NullString"
			t.NullValue = "String"
			t.NullType2 = "string"
		}
	}
	return t
}

func dbColumnToField(c *sql.ColumnType, nullField bool) [3]string {
	var field [3]string
	field[0] = snakeCaseToPascalCase(c.Name())
	field[1] = dbColumnToFieldType(c, nullField)
	field[2] = "`" + fmt.Sprintf(`json:"%s"`, pascalCaseToCamelCase(field[0])) + "`"
	return field
}

func dbColumnToFieldType(c *sql.ColumnType, nullField bool) string {
	s := goType(c.DatabaseTypeName())
	if nullField {
		if na, ok := c.Nullable(); ok && na {
			s = goNullType(s)
		}
	}
	return s
}

func NewCode(pkg, driver, dbUrl string) (*Code, error) {
	c := new(Code)
	c.dbUrl = dbUrl
	c.file = new(fileTPL)
	c.file.Pkg = pkg
	c.file.Driver = driver
	return c, nil
}

type Code struct {
	file  *fileTPL
	dbUrl string
}

func (c *Code) SaveFile(file string) error {
	// 创建目录
	dir := filepath.Dir(file)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}
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

func (c *Code) Exec(originalSql, function, tx string) (TPL, error) {
	if originalSql == "" {
		return nil, errEmptySQL
	}
	if function == "" {
		return nil, errEmptyFunctionName
	}
	// 解析sql
	segments, paramSegments, _, err := parseSegments(originalSql)
	if err != nil {
		return nil, err
	}
	// 重新写sql
	var _sql strings.Builder
	{
		for _, s := range segments {
			if s.param {
				_sql.WriteByte('?')
			} else {
				_sql.WriteString(s.string)
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
		_, err = db.Prepare(_sql.String())
		if err != nil {
			return nil, err
		}
	}
	// 公共模板
	var tp tpl
	tp.Func = function
	tp.Tx = tx
	tp.Sql = _sql.String()
	tp.Stmt = c.file.StmtName(tp.Sql)
	// 只有一个入参，不生成结构体
	if paramSegments != nil && len(paramSegments) < 2 {
		t := new(execTPL)
		t.tpl = tp
		t.Param = append(t.Param, paramSegments[0].string)
		c.file.Func = append(c.file.Func, t)
		return t, nil
	}
	// 有多个入参，生成结构体
	t := new(execStructTPL)
	t.tpl = tp
	t.Struct = function + "Model"
	for _, s := range paramSegments {
		var field [3]string
		field[0] = snakeCaseToPascalCase(s.string)
		field[1] = s.value
		field[2] = fmt.Sprintf("`json:\"%s\"`", pascalCaseToCamelCase(field[0]))
		t.Field = append(t.Field, field)
	}
	c.file.Func = append(c.file.Func, t)
	return t, nil
}

func (c *Code) Query(originalSql, function, tx string, isRow, nullField bool) (TPL, error) {
	if originalSql == "" {
		return nil, errEmptySQL
	}
	if function == "" {
		return nil, errEmptyFunctionName
	}
	// 解析sql
	segments, paramSegments, columnSegments, err := parseSegments(originalSql)
	if err != nil {
		return nil, err
	}
	// 重新写sql
	var _sql strings.Builder
	var testArgs []interface{}
	{
		for _, seg := range segments {
			if seg.param {
				if seg.column {
					_sql.WriteString(seg.value)
				} else {
					testArgs = append(testArgs, seg.value)
					_sql.WriteByte('?')
				}
			} else {
				_sql.WriteString(seg.string)
			}
		}
	}
	// 测试sql，获取结果集的字段信息
	var results []*sql.ColumnType
	{
		db, err := sql.Open("mysql", c.dbUrl)
		if err != nil {
			return nil, err
		}
		defer func() {
			_ = db.Close()
		}()
		// 测试sql
		rows, err := db.Query(_sql.String(), testArgs...)
		if err != nil {
			return nil, err
		}
		// 获取结果集的字段信息
		results, err = rows.ColumnTypes()
		if err != nil {
			return nil, err
		}
	}
	// 公共模板
	var qt queryTPL
	qt.Sql = _sql.String()
	qt.Func = function
	qt.Tx = tx
	qt.NullField = nullField
	qt.Param = paramSegments.ToParam()
	// 无法预编译的sql
	if len(columnSegments) > 0 {
		var qst querySqlTPL
		qst.queryTPL = qt
		qst.Column = columnSegments.ToParam()
		qst.Segment = segments.ToTPL()
		c.file.Strings = true
		if isRow {
			// 查询一行
			if len(results) > 1 {
				// 查询一行结构
				t := new(querySqlStructRowTPL)
				t.querySqlStructTPL.querySqlTPL = qst
				t.querySqlStructTPL.InitFieldAndScan(results)
				c.file.Func = append(c.file.Func, t)
				return t, nil
			}
			// 查询一行单值
			t := new(querySqlRowTPL)
			t.querySqlTPL = qst
			t.InitScan(results[0])
			c.file.Func = append(c.file.Func, t)
			return t, nil
		}
		// 查询多行
		if len(results) > 1 {
			// 查询多行结构
			t := new(querySqlStructRowsTPL)
			t.querySqlStructTPL.querySqlTPL = qst
			t.querySqlStructTPL.InitFieldAndScan(results)
			c.file.Func = append(c.file.Func, t)
			return t, nil
		}
		// 查询多行单值
		t := new(querySqlRowsTPL)
		t.querySqlTPL = qst
		t.InitScan(results[0])
		c.file.Func = append(c.file.Func, t)
		return t, nil
	}
	// 可以预编译的sql
	qt.Stmt = c.file.StmtName(qt.Sql)
	// 只有一个结果，不生成结构体
	if len(results) < 2 {
		// 单行
		if isRow {
			t := new(queryRowTPL)
			t.queryTPL = qt
			t.InitScan(results[0])
			c.file.Func = append(c.file.Func, t)
			return t, nil
		}
		// 多行
		t := new(queryRowsTPL)
		t.queryTPL = qt
		t.InitScan(results[0])
		c.file.Func = append(c.file.Func, t)
		return t, nil
	}
	// 有多个结果，生成结构体
	if isRow {
		// 查询单行
		t := new(queryStructRowTPL)
		t.queryTPL = qt
		t.InitFieldAndScan(results)
		c.file.Func = append(c.file.Func, t)
		return t, nil
	}
	// 查询多行
	t := new(queryStructRowsTPL)
	t.queryTPL = qt
	t.InitFieldAndScan(results)
	c.file.Func = append(c.file.Func, t)
	return t, nil
}
