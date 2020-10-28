package db2go

import "testing"

func TestReadSchema(t *testing.T) {
	s, err := ReadSchema(MYSQL, "root:123456@tcp(192.168.1.66)/db2go_test")
	if err != nil {
		t.Fatal(err)
	}
	testT0(t, s, s.GetTable("t0"))
	testT1(t, s, s.GetTable("t1"))
	testT2(t, s, s.GetTable("t2"))
	testT3(t, s, s.GetTable("t3"))
	testT4(t, s, s.GetTable("t4"))
}

func testT0(t *testing.T, s *Schema, table *Table) {
	tc := new(testColumn)
	tc.isPK = true
	tc.test(t, s, table.GetColumn("id"), "int", "int", "")
	tc.isPK = false
	tc.isNull = true
	tc.test(t, s, table.GetColumn("c_tinyint"), "tinyint", "int8", "")
	tc.test(t, s, table.GetColumn("c_smallint"), "smallint", "int16", "")
	tc.test(t, s, table.GetColumn("c_mediumint"), "mediumint", "int32", "")
	tc.test(t, s, table.GetColumn("c_int"), "int", "int", "")
	tc.test(t, s, table.GetColumn("c_bigint"), "bigint", "int64", "")
	tc.test(t, s, table.GetColumn("c_tinyint_unsigned"), "tinyint unsigned", "uint8", "")
	tc.test(t, s, table.GetColumn("c_smallint_unsigned"), "smallint unsigned", "uint16", "")
	tc.test(t, s, table.GetColumn("c_mediumint_unsigned"), "mediumint unsigned", "uint32", "")
	tc.test(t, s, table.GetColumn("c_int_unsigned"), "int unsigned", "uint", "")
	tc.test(t, s, table.GetColumn("c_bigint_unsigned"), "bigint unsigned", "uint64", "")
	tc.test(t, s, table.GetColumn("c_float"), "float", "float32", "")
	tc.test(t, s, table.GetColumn("c_double"), "double", "float64", "")
	tc.test(t, s, table.GetColumn("c_decimal"), "decimal(10,5)", "string", "")
	tc.test(t, s, table.GetColumn("c_char"), "char(10)", "string", "")
	tc.test(t, s, table.GetColumn("c_varchar"), "varchar(20)", "string", "")
	tc.test(t, s, table.GetColumn("c_tinytext"), "tinytext", "string", "")
	tc.test(t, s, table.GetColumn("c_text"), "text", "string", "")
	tc.test(t, s, table.GetColumn("c_mediumtext"), "mediumtext", "string", "")
	tc.test(t, s, table.GetColumn("c_longtext"), "longtext", "string", "")
	tc.test(t, s, table.GetColumn("c_tinyblob"), "tinyblob", "[]byte", "")
	tc.test(t, s, table.GetColumn("c_blob"), "blob", "[]byte", "")
	tc.test(t, s, table.GetColumn("c_mediumblob"), "mediumblob", "[]byte", "")
	tc.test(t, s, table.GetColumn("c_longblob"), "longblob", "[]byte", "")
	tc.test(t, s, table.GetColumn("c_binary"), "binary(255)", "[]byte", "")
	tc.test(t, s, table.GetColumn("c_time"), "time", "string", "")
	tc.test(t, s, table.GetColumn("c_timestamp"), "timestamp", "string", "")
	tc.test(t, s, table.GetColumn("c_date"), "date", "string", "")
	tc.test(t, s, table.GetColumn("c_datetime"), "datetime", "string", "")
	tc.test(t, s, table.GetColumn("c_year"), "year", "string", "")
}

func testT1(t *testing.T, s *Schema, table *Table) {
	tc := new(testColumn)
	tc.isPK = true
	tc.isAI = true
	tc.test(t, s, table.GetColumn("id"), "int", "int", "")
	tc.isPK = false
	tc.isAI = false
	tc.isUni = true
	tc.test(t, s, table.GetColumn("name"), "varchar(32)", "string", "")
}

func testT2(t *testing.T, s *Schema, table *Table) {
	tc := new(testColumn)
	tc.isPK = true
	tc.isAI = true
	tc.test(t, s, table.GetColumn("id"), "int", "int", "")
	tc.isPK = false
	tc.isAI = false
	tc.isNull = true
	tc.test(t, s, table.GetColumn("name"), "varchar(32)", "string", "")
}

func testT3(t *testing.T, s *Schema, table *Table) {
	tc := new(testColumn)
	tc.isPK = true
	tc.isAI = true
	tc.test(t, s, table.GetColumn("id"), "int", "int", "")
	tc.isPK = false
	tc.isAI = false
	tc.isMul = true
	tc.isNull = true
	tc.test(t, s, table.GetColumn("t1_id"), "int", "int", "")
	tc.test(t, s, table.GetColumn("t2_id"), "int", "int", "")
}

func testT4(t *testing.T, s *Schema, table *Table) {
	tc := new(testColumn)
	tc.isPK = true
	tc.test(t, s, table.GetColumn("c1"), "int", "int", "")
	tc.test(t, s, table.GetColumn("c2"), "int", "int", "")
	tc.isPK = false
	tc.isNull = true
	tc.test(t, s, table.GetColumn("c3"), "int", "int", "123")
}

type testColumn struct {
	isPK   bool
	isAI   bool
	isUni  bool
	isMul  bool
	isNull bool
}

func (tc *testColumn) test(t *testing.T, s *Schema, c *Column, typ, goType, def string) {
	if tc.isPK && !c.IsPrimaryKey() {
		t.FailNow()
	}
	if tc.isAI && !c.IsAutoIncrement() {
		t.FailNow()
	}
	if tc.isUni && !c.IsUnique() {
		t.FailNow()
	}
	if tc.isMul && !c.IsMulUnique() {
		t.FailNow()
	}
	if tc.isNull && !c.IsNullable() {
		t.FailNow()
	}
	if c.Type() != typ || c.GoType() != goType || c.DefaultValue() != def {
		t.FailNow()
	}
	ft := c.ForeignTable()
	if ft != nil {
		st := s.GetTable(ft.Table().Name())
		if st != ft.Table() {
			t.FailNow()
		}
		sc := st.GetColumn(ft.Column().Name())
		if sc != ft.Column() {
			t.FailNow()
		}
	}
}
