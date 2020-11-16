package mysql

import (
	"bytes"
	"strings"
)

func pascalCaseToSnakeCase(s string) string {
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

func camelCaseToPascalCase(s string) string {
	return strings.ToUpper(s[:1]) + s[1:]
}

func pascalCaseToCamelCase(s string) string {
	return strings.ToLower(s[:1]) + s[1:]
}

func camelCaseToSnakeCase(s string) string {
	return pascalCaseToSnakeCase(camelCaseToPascalCase(s))
}

func snakeCaseToCamelCase(s string) string {
	return pascalCaseToCamelCase(snakeCaseToPascalCase(s))
}
