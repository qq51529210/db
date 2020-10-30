package db2go

import (
	"database/sql"
	"fmt"
)

const (
	MYSQL = "mysql"
)

var (
	schemaFunc = make(map[string]func(string) (*Schema, error))
	goTypeFunc = make(map[string]func(string) string)
	driver     = make(map[string]string)
)

// 驱动包，比如"github.com/go-sql-driver/mysql"
func DriverPkg(dbType string) string {
	return driver[dbType]
}

// 读取数据库结构
func ReadSchema(dbType, dbUrl string) (*Schema, error) {
	f, o := schemaFunc[dbType]
	if !o {
		return nil, fmt.Errorf("unsupported db '%s'", dbType)
	}
	return f(dbUrl)
}

// 返回go数据类型
func DBTypeToGo(dbType, dataType string) string {
	f, o := goTypeFunc[dbType]
	if !o {
		return ""
	}
	return f(dataType)
}

// 数据库结构
type Schema struct {
	dbUrl  string
	dbType string
	name   string   // 名称
	table  []*Table // 所有的表
}

func (s *Schema) GetTable(name string) *Table {
	for _, t := range s.table {
		if t.name == name {
			return t
		}
	}
	return nil
}

func (s *Schema) DBType() string {
	return s.dbType
}

func (s *Schema) Name() string {
	return s.name
}

func (s *Schema) Tables() []*Table {
	return s.table
}

// 测试
func (s *Schema) TestSQL(prepare string) error {
	db, err := sql.Open(s.dbType, s.dbUrl)
	if err != nil {
		return err
	}
	defer func() {
		_ = db.Close()
	}()
	_, err = db.Prepare(prepare)
	if err != nil {
		return err
	}
	return nil
}

// 数据库表
type Table struct {
	name   string
	column []*Column
}

func (t *Table) Name() string {
	return t.name
}

func (t *Table) Columns() []*Column {
	return t.column
}

func (t *Table) GetColumn(name string) *Column {
	for _, c := range t.column {
		if c.name == name {
			return c
		}
	}
	return nil
}

func (t *Table) PrimaryKeyColumns() (pk, npk []*Column) {
	for _, c := range t.column {
		if c.primaryKey {
			pk = append(pk, c)
		} else {
			npk = append(npk, c)
		}
	}
	return
}

func (t *Table) UniqueColumns() (un, nun []*Column) {
	for _, c := range t.column {
		if c.primaryKey {
			un = append(un, c)
		} else {
			nun = append(nun, c)
		}
	}
	return
}

func (t *Table) MulUniqueColumns() (mu, nmu []*Column) {
	for _, c := range t.column {
		if c.primaryKey {
			mu = append(mu, c)
		} else {
			nmu = append(nmu, c)
		}
	}
	return
}

// 数据库表字段
type Column struct {
	dbType        string        // db类型
	name          string        // 列名
	_type         string        // 数据库类型
	primaryKey    bool          // 主键
	autoIncrement bool          // 自增
	unique        bool          // 唯一
	mulUnique     bool          // 联合唯一
	nullable      bool          // NULL值
	defaultValue  string        // 默认值
	foreignTable  *ForeignTable // 引用表
}

func (c *Column) GoType() string {
	typ := goTypeFunc[c.dbType](c._type)
	if c.nullable {
		switch typ {
		case "int8", "int16", "int32", "uint8", "uint16", "uint32":
			return "sql.NullInt32"
		case "int", "int64", "uint", "uint64":
			return "sql.NullInt64"
		case "float32", "float64":
			return "sql.NullFloat64"
		default:
			return "sql.NullString"
		}
	}
	return typ
}

func (c *Column) Name() string {
	return c.name
}

func (c *Column) Type() string {
	return c._type
}

func (c *Column) IsPrimaryKey() bool {
	return c.primaryKey
}

func (c *Column) IsUnique() bool {
	return c.unique
}

func (c *Column) IsMulUnique() bool {
	return c.mulUnique
}

func (c *Column) IsNullable() bool {
	return c.nullable
}

func (c *Column) IsAutoIncrement() bool {
	return c.autoIncrement
}

func (c *Column) ForeignTable() *ForeignTable {
	return c.foreignTable
}

func (c *Column) DefaultValue() string {
	return c.defaultValue
}

type ForeignTable struct {
	table  *Table
	column *Column
}

func (t *ForeignTable) Table() *Table {
	return t.table
}

func (t *ForeignTable) Column() *Column {
	return t.column
}
