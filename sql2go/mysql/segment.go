package mysql

import "strings"

type sqlSegment struct {
	string string
	value  string
	param  bool
	column bool
}

type sqlSegments []*sqlSegment

func (ss sqlSegments) ToTPL() []string {
	var s []string
	var b strings.Builder
	i := 0
	for ; i < len(ss); i++ {
		if ss[i].column {
			b.WriteString(`" `)
			s = append(s, b.String())
			s = append(s, ss[i].string)
			b.Reset()
		} else {
			if b.Len() < 1 {
				b.WriteByte('"')
			}
			if ss[i].param {
				b.WriteByte('?')
			} else {
				b.WriteString(ss[i].string)
			}
		}
	}
	if b.Len() > 0 {
		seg := ss[len(ss)-1]
		if !seg.column {
			b.WriteByte('"')
		}
		s = append(s, b.String())
	}
	return s
}

func (ss sqlSegments) ToParam() []string {
	var s []string
	for i := 0; i < len(ss); i++ {
		s = append(s, pascalCaseToCamelCase(snakeCaseToPascalCase(ss[i].string)))
	}
	return s
}

func (ss sqlSegments) Join() []string {
	var s []string
	for i := 0; i < len(ss); i++ {
		s = append(s, ss[i].string)
	}
	return s
}

// 解析sql片段并返回
func parseSegments(s string) (segments, params, columns sqlSegments, err error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return
	}
	i := 0
	for i < len(s) {
		switch s[i] {
		case '\'':
			j := indexString(s[i:])
			if j < 0 {
				err = parseError(s[i:])
				return
			}
			i += j
		case '[':
			// 前面的sql
			if i != 0 {
				segments = addSegment(segments, &sqlSegment{string: s[:i]})
				s = s[i:]
			}
			// 字段[xxx]
			i = indexByte(s, ']')
			if i < 0 {
				err = parseError(s)
				return
			}
			// 拆分变量id:''
			ss := strings.Split(s[1:i], ":")
			if len(ss) != 2 {
				err = parseError(s[1:i])
				return
			}
			seg := &sqlSegment{string: ss[0], value: ss[1], param: true, column: true}
			columns = append(columns, seg)
			segments = addSegment(segments, seg)
			s = s[i+1:]
			i = 0
		case '{':
			// 前面的sql
			if i != 0 {
				segments = addSegment(segments, &sqlSegment{string: s[:i]})
				s = s[i:]
			}
			// {xxx}变量
			i = indexByte(s, '}')
			if i < 0 {
				err = parseError(s)
				return
			}
			// 拆分变量id:''
			ss := strings.Split(s[1:i], ":")
			if len(ss) != 2 {
				err = parseError(s[1:i])
				return
			}
			seg := &sqlSegment{string: ss[0], value: ss[1], param: true}
			params = append(params, seg)
			segments = addSegment(segments, seg)
			s = s[i+1:]
			i = 0
		default:
			i++
		}
	}
	if s != "" {
		segments = addSegment(segments, &sqlSegment{string: s})
	}
	return
}

// 添加seg到segments中
func addSegment(segments []*sqlSegment, seg *sqlSegment) []*sqlSegment {
	//seg.string = strings.TrimSpace(seg.string)
	//if seg.string == "" {
	//	return segments
	//}
	//if seg.string == "," || len(segments) < 1 {
	//	segments = append(segments, seg)
	//	return segments
	//}
	//last := segments[len(segments)-1]
	//if last.string != "," {
	//	if seg.param && !seg.column {
	//		c := last.string[len(last.string)-1]
	//		if c != '(' {
	//			segments = append(segments, &sqlSegment{string: " "})
	//		}
	//	} else {
	//		segments = append(segments, &sqlSegment{string: " "})
	//	}
	//}
	//segments = append(segments, seg)
	//if seg.column {
	//	segments = append(segments, &sqlSegment{string: " "})
	//}
	segments = append(segments, seg)
	return segments
}

// 找到完整的'xx'
func indexString(s string) int {
	for i := 1; i < len(s); i++ {
		if s[i] == '\'' && s[i-1] != '\\' {
			return i
		}
	}
	return -1
}

// 找到完整的下一个c
func indexByte(s string, c byte) int {
	i := 1
	for i < len(s) {
		if s[i] == '\'' {
			j := indexString(s[i:])
			if j < 0 {
				return j
			}
			i += j
		} else if s[i] == c {
			return i
		}
		i++
	}
	return -1
}
