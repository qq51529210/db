package tpl

import (
	"bytes"
	"os"
	"strings"
	"text/template"
)

type TPL interface {
	TPL() *template.Template
}

type Arg struct {
	Name    string
	IsField bool
}

func (a Arg) String() string {
	if a.IsField {
		return "m." + a.Name
	}
	return a.Name
}

type Args []*Arg

func (a Args) Params() []string {
	var s []string
	for _, i := range a {
		if !i.IsField {
			s = append(s, i.Name)
		}
	}
	return s
}

func (a Args) Fields() []string {
	var s []string
	for _, i := range a {
		if i.IsField {
			s = append(s, i.Name)
		}
	}
	return s
}

func saveFile(t *template.Template, data interface{}, path string) error {
	// 打开文件写
	f, err := os.OpenFile(path, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	// 关闭文件
	defer func() { _ = f.Close() }()
	// 输出
	return t.Execute(f, data)
}

// 把'UserId'转换成'userId'
func PascalCaseToCamelCase(s string) string {
	return strings.ToLower(s[:1]) + s[1:]
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

// 把'userId'转换成'UserId'
func CamelCaseToPascalCase(s string) string {
	return strings.ToUpper(s[:1]) + s[1:]
}

// 把'userId'转换成'user_id'
func CamelCaseToSnakeCase(s string) string {
	return PascalCaseToSnakeCase(CamelCaseToPascalCase(s))
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
