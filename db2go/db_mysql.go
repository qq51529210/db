/*
MYSQL数据库的一些实现
*/
package db2go

import (
	"database/sql"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"strings"
)

func init() {
	schemaFunc[MYSQL] = mysqlReadSchema
	goTypeFunc[MYSQL] = mysqlGoType
	driver[MYSQL] = "github.com/go-sql-driver/mysql"
}

var (
	errEmptyDBName   = errors.New("db name is empty")
	errInvalidColumn = errors.New("column name or type is invalid")
)

// 数据类型对应表
func mysqlGoType(dataType string) string {
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
	case "date", "time", "year", "datetime", "timestamp":
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

// 读取数据库结构
func mysqlReadSchema(dbUrl string) (*Schema, error) {
	var err error
	schema := new(Schema)
	schema.dbUrl = dbUrl
	schema.dbType = MYSQL
	schema.name, err = mysqlParseSchemaName(dbUrl)
	if err != nil {
		return nil, err
	}
	// 打开数据库
	db, err := sql.Open(MYSQL, dbUrl)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = db.Close()
	}()
	// 读取数据库所有表
	err = mysqlReadSchemaTable(db, schema)
	if err != nil {
		return nil, err
	}
	// 读取表所有列信息
	for _, table := range schema.table {
		err = mysqlReadSchemaTableColumn(db, schema, table)
		if err != nil {
			return nil, err
		}
		err = mysqlReadSchemaColumnMulAndReference(db, schema, table)
		if err != nil {
			return nil, err
		}
	}
	return schema, nil
}

// 从连接字符串中解析出数据库名称
func mysqlParseSchemaName(dbUrl string) (string, error) {
	i := strings.Index(dbUrl, "/")
	if i < 0 {
		return "", errEmptyDBName
	}
	schema := dbUrl[i+1:]
	i = strings.Index(schema, "?")
	if i > 0 {
		schema = schema[:i]
	}
	if schema == "" {
		return "", errEmptyDBName
	}
	return schema, nil
}

// 读取数据库所有表
func mysqlReadSchemaTable(db *sql.DB, schema *Schema) error {
	// sql
	var str strings.Builder
	str.WriteString("select ")
	str.WriteString("table_name ")
	str.WriteString("from ")
	str.WriteString("information_schema.tables ")
	str.WriteString("where ")
	str.WriteString("table_schema='")
	str.WriteString(schema.name)
	str.WriteString("'")
	// 查询
	rows, err := db.Query(str.String())
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return nil
	}
	// 循环读table
	for rows.Next() {
		table := new(Table)
		err = rows.Scan(&table.name)
		if err != nil {
			return err
		}
		schema.table = append(schema.table, table)
	}
	return nil
}

// 读取表的所有列信息
func mysqlReadSchemaTableColumn(db *sql.DB, schema *Schema, table *Table) error {
	// sql
	var str strings.Builder
	str.WriteString("select ")
	str.WriteString("column_name,")
	str.WriteString("column_type,")
	str.WriteString("column_key,")
	str.WriteString("column_default,")
	str.WriteString("is_nullable,")
	str.WriteString("extra ")
	str.WriteString("from ")
	str.WriteString("information_schema.columns ")
	str.WriteString("where ")
	str.WriteString("table_schema='")
	str.WriteString(schema.name)
	str.WriteString("' ")
	str.WriteString("and ")
	str.WriteString("table_name='")
	str.WriteString(table.name)
	str.WriteString("' ")
	// 查询
	rows, err := db.Query(str.String())
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return nil
	}
	// 循环
	var columnName, columnType, columnKey, columnDefault, isNullable, extra sql.NullString
	for rows.Next() {
		err = rows.Scan(&columnName, &columnType, &columnKey, &columnDefault, &isNullable, &extra)
		if err != nil {
			return err
		}
		// 没有列的基本信息，出错
		if !columnName.Valid || !columnType.Valid {
			return errInvalidColumn
		}
		column := &Column{
			dbType: schema.dbType,
			name:   columnName.String,
			_type:  columnType.String,
		}
		// key
		if columnKey.Valid {
			switch strings.ToLower(columnKey.String) {
			case "pri":
				column.primaryKey = true
			case "uni":
				column.unique = true
			}
		}
		// 默认
		if columnDefault.Valid {
			column.defaultValue = columnDefault.String
		}
		// 可以为null
		if isNullable.Valid {
			column.nullable = strings.ToLower(isNullable.String) == "yes"
		}
		// 自增
		if extra.Valid {
			column.autoIncrement = strings.ToLower(extra.String) == "auto_increment"
		}
		table.column = append(table.column, column)
	}
	return nil
}

// 读取表的，多唯一，外键
func mysqlReadSchemaColumnMulAndReference(db *sql.DB, schema *Schema, table *Table) error {
	// sql
	var str strings.Builder
	str.WriteString("select ")
	str.WriteString("constraint_name,")
	str.WriteString("column_name,")
	str.WriteString("referenced_table_name,")
	str.WriteString("referenced_column_name ")
	str.WriteString("from ")
	str.WriteString("information_schema.key_column_usage ")
	str.WriteString("where ")
	str.WriteString("table_schema='")
	str.WriteString(schema.name)
	str.WriteString("' ")
	str.WriteString("and ")
	str.WriteString("table_name='")
	str.WriteString(table.name)
	str.WriteString("' ")
	// 查询
	rows, err := db.Query(str.String())
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return nil
	}
	// 循环
	nameColumns := make(map[string][]string)
	refNames := make(map[string][2]string)
	var constraintName, columnName, referencedTableName, referencedColumnName sql.NullString
	for rows.Next() {
		err = rows.Scan(&constraintName, &columnName, &referencedTableName, &referencedColumnName)
		if err != nil {
			return err
		}
		if !constraintName.Valid || strings.ToLower(constraintName.String) == "primary" || !columnName.Valid {
			continue
		}
		columns, ok := nameColumns[constraintName.String]
		if ok {
			columns = append(columns, columnName.String)
		} else {
			columns = make([]string, 0)
			columns = append(columns, columnName.String)
			nameColumns[constraintName.String] = columns
		}
		if referencedTableName.Valid && referencedColumnName.Valid {
			refNames[columnName.String] = [2]string{referencedTableName.String, referencedColumnName.String}
		}
	}
	// 分析
	for _, v := range nameColumns {
		if len(v) > 1 {
			for _, s := range v {
				// 如果为nil，那么mysql的数据是有问题的
				c := table.GetColumn(s)
				c.mulUnique = true
			}
		}
	}
	for k, v := range refNames {
		// 如果为nil，那么mysql的数据是有问题的
		c := table.GetColumn(k)
		t := new(ForeignTable)
		t.table = schema.GetTable(v[0])
		t.column = t.table.GetColumn(v[1])
		c.foreignTable = t
	}

	return nil
}
