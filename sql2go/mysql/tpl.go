package mysql

import (
	"io"
	"strings"
)

type TPL interface {
	Execute(io.Writer) error
}

type tpl struct {
	Sql   string
	Func  string
	Tx    string
	Param []string
	Stmt  string
}

func (t *tpl) ParamTPL() string {
	var s strings.Builder
	// tx *sql.Tx
	if t.Tx != "" {
		s.WriteString(t.Tx)
		s.WriteString(" *sql.Tx")
	}
	if len(t.Param) < 1 {
		return s.String()
	}
	// tx *sql.Tx, param1,param2 ...interface{}
	if s.Len() > 0 {
		s.WriteString(", ")
	}
	s.WriteString(strings.Join(t.Param, ", "))
	s.WriteString(" interface{}")
	return s.String()
}

func (t *tpl) StmtTPL() string {
	if t.Tx != "" {
		var s strings.Builder
		s.WriteString(t.Tx)
		s.WriteString(".Stmt(")
		s.WriteString(t.Stmt)
		s.WriteByte(')')
		return s.String()
	}
	return t.Stmt
}

type scanTPL struct {
	Name      string
	Type      string
	NullType  string
	NullValue string
	NullType2 string // 强制转换的类型
}

type queryTPL struct {
	tpl
	Field     [][3]string
	NullField bool // 字段是否sql.NullXxx
}

type querySqlTPL struct {
	queryTPL
	Column  []string
	Segment []string
}

func (t *querySqlTPL) ParamTPL() string {
	var s strings.Builder
	if t.Tx != "" {
		s.WriteString(t.Tx)
		s.WriteString(" *sql.Tx")
	}
	if len(t.Param) < 1 && len(t.Column) < 1 {
		return s.String()
	}
	if s.Len() > 0 {
		s.WriteString(", ")
	}
	if len(t.Column) > 0 {
		s.WriteString(strings.Join(t.Column, ", "))
		s.WriteString(" string")
	}
	if s.Len() > 0 {
		s.WriteString(", ")
	}
	s.WriteString(strings.Join(t.Param, ", "))
	s.WriteString(" interface{}")
	return s.String()
}

func (t *querySqlTPL) StmtTPL() string {
	if t.Tx != "" {
		return t.Tx
	}
	return "DB"
}
