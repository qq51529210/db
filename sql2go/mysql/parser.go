package mysql

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// 函数
	functions = map[string]string{
		// 聚合
		"avg":   "float64",
		"count": "int64",
		"sum":   "float64",
		"max":   "float64",
		"min":   "float64",
		// 数学
		"abs":      "float64",
		"sqrt":     "float64",
		"mod":      "int64",
		"format":   "float64",
		"ceil":     "int64",
		"ceiling":  "int64",
		"floor":    "int64",
		"round":    "int64",
		"pi":       "float64",
		"sign":     "float64",
		"pow":      "float64",
		"power":    "float64",
		"rand":     "float64",
		"truncate": "int64",
		"sin":      "float64",
		"asin":     "float64",
		"cos":      "float64",
		"acos":     "float64",
		"tan":      "float64",
		"atan":     "float64",
		"cot":      "float64",
		// 字符串
		"length":    "int64",
		"concat":    "string",
		"insert":    "string",
		"lower":     "string",
		"upper":     "string",
		"left":      "string",
		"right":     "string",
		"trim":      "string",
		"ltrim":     "string",
		"rtrim":     "string",
		"replace":   "string",
		"substring": "string",
		"reverse":   "string",
		"repeat":    "string",
		"strcmp":    "int64",

		"curdate":        "string",
		"current_date":   "string",
		"curtime":        "string",
		"current_time":   "string",
		"now":            "string",
		"sysdate":        "string",
		"unix_timestamp": "int64",
		"from_unixtime":  "string",
		"month":          "int64",
		"monthname":      "string",
		"dayname":        "string",
		"dayofweek":      "int64",
		"week":           "int64",
		"dayofyear":      "int64",
		"dayofmonth":     "int64",
		"year":           "int64",
		"time_to_sec":    "int64",
		"sec_to_time":    "string",
	}

	// 运算符列表，值表示优先级，越小优先级越高
	operators = map[string]int{
		"*":           1,
		"/":           1,
		"%":           1,
		"+":           2,
		"-":           2,
		">":           3,
		">=":          3,
		"<":           3,
		"<>":          3,
		"<=":          3,
		"!=":          3,
		"=":           3,
		"&":           4,
		"|":           4,
		"!":           4,
		"&&":          5,
		"||":          5,
		"in":          6,
		"not in":      6,
		"like":        2,
		"not like":    2,
		"is":          2,
		"is not":      2,
		"and":         7,
		"or":          7,
		"between":     5,
		"not between": 5,
		"exists":      1,
		"not exists":  1,
	}

	// 关键字列表，解析是使用
	keywords = map[string]int{
		"select":                   1,
		"all":                      1,
		"distinct":                 1,
		"from":                     1,
		"as":                       1,
		"join":                     1,
		"inner join":               1,
		"left join":                1,
		"left outer join":          1,
		"right join":               1,
		"right outer join":         1,
		"natural join":             1,
		"natural left join":        1,
		"natural right join":       1,
		"natural left outer join":  1,
		"natural right outer join": 1,
		"on":                       1,
		"union":                    1,
		"union all":                1,
		"insert into":              1,
		"values":                   1,
		"update":                   1,
		"set":                      1,
		"delete from":              1,
		"where":                    1,
		"order by":                 1,
		"group by":                 1,
		"having":                   1,
		"limit":                    1,
		"desc":                     1,
		"asc":                      1,
	}

	noTokenFoundErr = errors.New("parse token error 'no token found'")
	parseSqlEOF     = errors.New("parse sql error 'eof'")
)

// 符号c是否运算符号
func isOperatorsSymbol(c byte) bool {
	switch c {
	case '+', '-', '*', '/', '%', '&', '|', '>', '<', '!', '=':
		return true
	default:
		return false
	}
}

// 符号c是否分隔符号
func isSeparatorSymbol(c byte) bool {
	switch c {
	case ' ', '\t', '\v', '\r', '\n', '\f', ',', '?', ';':
		return true
	default:
		return false
	}
}

// 是否分隔符
func isSeparator(t string) bool {
	return t == "," || t == ";"
}

// 是否运算符
func isOperators(t string) bool {
	_, o := operators[strings.ToLower(t)]
	return o
}

// 是否关键字
func isKeywords(t string) bool {
	_, o := keywords[strings.ToLower(t)]
	return o
}

// 是否函数
func isFunction(t string) bool {
	_, o := functions[strings.ToLower(t)]
	return o
}

// 是否函数
func isIdentifier(t string) bool {
	return !isKeywords(t) && !isOperators(t) && !isSeparator(t)
}

// 是否是"(xx)"值
func isParentheses(t string) bool {
	if len(t) < 2 {
		return false
	}
	return t[0] == '(' && t[len(t)-1] == ')'
}

// 去掉"()"对
func removeParentheses(t string) string {
	for len(t) > 1 {
		if t[0] == '(' && t[len(t)-1] == ')' {
			t = t[1 : len(t)-1]
			continue
		}
		break
	}
	return t
}

// 小写，是否匹配其中一个
func matchAny(t string, ss ...string) bool {
	t = strings.ToLower(t)
	for _, s := range ss {
		if t == strings.ToLower(s) {
			return true
		}
	}
	return false
}

// 添加sql解析的函数，name是函数名称，_return是函数返回值
func AddFunction(name, _return string) {
	functions[name] = _return
}

func recoverError() error {
	re := recover()
	if re != nil {
		e, ok := re.(error)
		if ok {
			return e
		} else {
			return fmt.Errorf("%v", e)
		}
	}
	return nil
}

// 解析sql
func ParseSQL(sql string) (v interface{}, err error) {
	defer func() {
		err = recoverError()
	}()
	v = parseSQL(removeParentheses(sql))
	return
}

// 解析sql
func parseSQL(sql string) (v interface{}) {
	p := new(parser)
	// 解析读取所有token
	p.readTokens(sql)
	// 解析
	if p.Match("select") {
		v = p.Select()
	} else if p.Match("insert into") {
		v = p.Insert()
	} else if p.Match("update") {
		v = p.Update()
	} else if p.Match("delete from") {
		v = p.Delete()
	} else {
		v = p.ExpressionStmt()
	}
	return
}

type parser struct {
	token []string // token
	index int      // 正在解析的索引
}

// 读取所有的token
func (p *parser) readTokens(sql string) {
	var token string
	// 检查连续的关键字
	var checkKeywordToken = func(must bool, ss ...string) {
		var token string
		sql, token = p.readToken(sql)
		if matchAny(token, ss...) {
			p.token[len(p.token)-1] += " " + token
		} else {
			if must {
				panic(p.readTokenError())
			}
			p.token = append(p.token, token)
		}
	}
	for sql != "" {
		sql, token = p.readToken(sql)
		p.token = append(p.token, token)
		switch strings.ToLower(token) {
		case "insert":
			checkKeywordToken(true, "into")
		case "delete":
			checkKeywordToken(true, "from")
		case "is":
			checkKeywordToken(false, "not")
		case "not":
			checkKeywordToken(false, "like", "between", "in", "exists")
		case "union":
			checkKeywordToken(false, "all")
		case "group", "order":
			checkKeywordToken(true, "by")
		case "inner":
			checkKeywordToken(true, "join")
		case "left", "right":
			checkKeywordToken(false, "outer")
			checkKeywordToken(true, "join")
		case "natural":
			sql, token = p.readToken(sql)
			if matchAny(token, "left", "right") {
				p.token[len(p.token)-1] += " " + token
				checkKeywordToken(false, "outer")
			}
			checkKeywordToken(true, "join")
		case "outer":
			panic(p.readTokenError())
		}
	}
}

// 读取token并返回
func (p *parser) readToken(sql string) (string, string) {
	sql = strings.TrimSpace(sql)
	if sql == "" {
		panic(p.readTokenError())
	}
	// 运算符
	if isOperatorsSymbol(sql[0]) {
		n := 1
		for n < len(sql) && isOperatorsSymbol(sql[n]) {
			n++
		}
		return sql[n:], sql[:n]
	}
	// 分隔符
	if isSeparatorSymbol(sql[0]) {
		return sql[1:], sql[:1]
	}
	switch sql[0] {
	case '\'', '"', '`':
		// 字符串
		c := sql[0]
		n := 1
		for n < len(sql) {
			if sql[n] == c && sql[n-1] != '\\' {
				n++
				return sql[n:], sql[:n]
			}
			n++
		}
		panic(p.readTokenError())
	case '(':
		// 表达式
		n := 0
		m := 1
		for n < len(sql) {
			n++
			if sql[n] == '(' {
				m++
			}
			if sql[n] == ')' {
				m--
				if m == 0 {
					n++
					return sql[n:], sql[:n]
				}
			}
		}
		panic(p.readTokenError())
	default:
		// 标识符
		n := 1
		for n < len(sql) {
			// 直到遇到，运算符/分隔符/区间
			if isOperatorsSymbol(sql[n]) || isSeparatorSymbol(sql[n]) ||
				sql[n] == '\'' || sql[n] == '"' || sql[n] == '`' || sql[n] == '(' {
				break
			}
			n++
		}
		return sql[n:], sql[:n]
	}
}

// 当前解析token的错误
func (p *parser) readTokenError() error {
	if len(p.token) == 0 {
		return noTokenFoundErr
	}
	return fmt.Errorf("parse token error at '%s'", p.token[len(p.token)-1])
}

// 当前解析的错误
func (p *parser) parseSqlError() error {
	return fmt.Errorf("parse sql error '%s'", p.token[p.index])
}

func (p *parser) Token() string {
	if p.index == len(p.token) {
		panic(parseSqlEOF)
	}
	return p.token[p.index]
}

func (p *parser) NextToken() string {
	if p.index >= len(p.token)-1 {
		return ""
	}
	return p.token[p.index+1]
}

func (p *parser) Next() {
	p.index++
}

func (p *parser) IsEmpty() bool {
	return p.index >= len(p.token)
}

func (p *parser) Match(ss ...string) bool {
	if p.IsEmpty() {
		return false
	}
	if matchAny(p.Token(), ss...) {
		p.Next()
		return true
	}
	return false
}

func (p *parser) MustMatch(ss ...string) string {
	if !matchAny(p.Token(), ss...) {
		panic(p.parseSqlError())
	}
	t := p.Token()
	p.Next()
	return t
}

type ExpressionStmt struct {
	Left     interface{}
	Operator string
	Right    interface{}
}

type BoolExpressionStmt struct {
	Operator string
	Value    interface{}
}

type FuncExpressionStmt struct {
	Name  string
	Value interface{}
}

// 解析表达式
func (p *parser) ExpressionStmt() interface{} {
	// 左值
	left := p.expressionValue()
	// 只有单一值或者不是运算符，返回
	if p.IsEmpty() || !isOperators(p.Token()) {
		return left
	}
	// 读取运算符和右值
	expr := p.expressionOperatorsAndRight()
	expr.Left = left
	for len(p.token) > 0 && isOperators(p.token[0]) {
		// 新的运算符和右值
		newExpr := p.expressionOperatorsAndRight()
		// 调整表达式的语法树
		if c1, o := operators[strings.ToLower(newExpr.Operator)]; o {
			if c2, o := operators[strings.ToLower(expr.Operator)]; o {
				if c1 < c2 {
					newExpr.Left = expr.Right
					expr.Right = newExpr
				} else {
					newExpr.Left = expr
					expr = newExpr
				}
			}
		}
	}
	return expr
}

// 解析表达式的单一个值
func (p *parser) expressionValue() interface{} {
	t := p.Token()
	// (xxx)
	if isParentheses(t) {
		p.Next()
		// 解析子表达式
		pp := new(parser)
		pp.readTokens(removeParentheses(t))
		// 是否select子查询
		if pp.Match("select") {
			return pp.Select()
		}
		// 解析，如果不是整个expression，那就是(x,x,x)
		v := pp.ExpressionStmt()
		if len(pp.token) < 1 {
			return v
		}
		vv := []interface{}{v}
		for !pp.IsEmpty() {
			pp.MustMatch(",")
			vv = append(vv, pp.ExpressionStmt())
		}
		return vv
	}
	// exists(xxx)
	if p.Match("exists", "not exists") {
		expr := new(BoolExpressionStmt)
		expr.Operator = t
		expr.Value = p.mustParenthesesExpressionValue()
		return expr
	}
	// func()
	if isFunction(t) && isParentheses(p.NextToken()) {
		p.Next()
		expr := new(FuncExpressionStmt)
		expr.Name = t
		pp := new(parser)
		pp.readTokens(removeParentheses(p.Token()))
		if len(pp.token) == 1 && pp.Token() == "*" {
			expr.Value = "*"
		} else {
			expr.Value = pp.ExpressionStmt()
		}
		p.Next()
		return expr
	}
	// 单一值
	return p.mustIdentifier()
}

// 解析表达式的运算符和右值
func (p *parser) expressionOperatorsAndRight() *ExpressionStmt {
	t := p.Token()
	p.Next()
	// 新的表达式
	expr := &ExpressionStmt{Operator: t}
	switch strings.ToLower(expr.Operator) {
	case "between", "not between":
		// ? and ?
		and := new(ExpressionStmt)
		pp := new(parser)
		// and 左值
		i1 := p.index
		for !p.Match("and") {
		}
		i2 := p.index - 1
		pp.token = p.token[i1:i2]
		and.Left = pp.ExpressionStmt()
		// and
		and.Operator = p.token[i2]
		// 右值
		and.Right = p.ExpressionStmt()
		expr.Right = and
	case "in", "not in":
		// 右值必须是(xxx)
		expr.Right = p.mustParenthesesExpressionValue()
	case "is", "is not":
		expr.Right = p.MustMatch("null", "?")
		p.Next()
	default:
		expr.Right = p.expressionValue()
	}
	return expr
}

// 当前的token必须是(x,x,x)，并解析
func (p *parser) mustParenthesesExpressionValue() interface{} {
	if !isParentheses(p.Token()) {
		panic(p.parseSqlError())
	}
	return p.expressionValue()
}

// 当前的token必须是普通标识符
func (p *parser) mustIdentifier() string {
	t := p.Token()
	if !isIdentifier(t) {
		panic(p.parseSqlError())
	}
	p.Next()
	return t
}

// 当前的token必须是(x,x,x)
func (p *parser) mustParentheses() string {
	t := p.Token()
	if !isParentheses(t) {
		panic(p.parseSqlError())
	}
	p.Next()
	return t
}

type DeleteStmt struct {
	Table string
	Where interface{}
	Token []string
}

// delete from table [where condition]
func (p *parser) Delete() *DeleteStmt {
	query := new(DeleteStmt)
	query.Token = p.token
	// table
	query.Table = p.mustIdentifier()
	// [where condition]
	if p.Match("where") {
		query.Where = p.ExpressionStmt()
	}
	return query
}

type UpdateStmt struct {
	Table  string
	Column []*ExpressionStmt
	Where  interface{}
	Token  []string
}

// update table set column=expression[, ...] [where condition]
func (p *parser) Update() *UpdateStmt {
	query := new(UpdateStmt)
	query.Token = p.token
	// table
	query.Table = p.mustIdentifier()
	// set
	p.MustMatch("set")
	// column=expression[, ...]
	for {
		expr := new(ExpressionStmt)
		// column必须是普通标识符
		expr.Left = p.mustIdentifier()
		// 运算符必须是'='
		expr.Operator = p.MustMatch("=")
		// 右值是表达式
		expr.Right = p.ExpressionStmt()
		query.Column = append(query.Column, expr)
		if p.IsEmpty() || !p.Match(",") {
			break
		}
	}
	// [where condition]
	if p.Match("where") {
		query.Where = p.ExpressionStmt()
	}
	return query
}

type InsertStmt struct {
	Table  string        // 表名
	Column []string      // 列名
	Values []interface{} // 值
	Token  []string
}

// insert into table [(column[, ...])] {values(expression[, ...])}
func (p *parser) Insert() *InsertStmt {
	query := new(InsertStmt)
	query.Token = p.token
	// table
	query.Table = p.mustIdentifier()
	// [(column[, ...])]
	if isParentheses(p.Token()) {
		pp := new(parser)
		pp.readTokens(removeParentheses(p.Token()))
		for {
			// column必须是标识符
			query.Column = append(query.Column, pp.mustIdentifier())
			if pp.IsEmpty() {
				break
			}
			// 如果还有column，必须是','隔开
			pp.MustMatch(",")
		}
		p.Next()
	}
	// values(expression[, ...])
	p.MustMatch("values")
	pp := new(parser)
	pp.readTokens(removeParentheses(p.mustParentheses()))
	for {
		v := pp.ExpressionStmt()
		query.Values = append(query.Values, v)
		if pp.IsEmpty() {
			break
		}
		pp.MustMatch(",")
	}
	// column和values的个数必须相等
	if len(query.Column) > 0 && len(query.Column) != len(query.Values) {
		panic(fmt.Errorf("column count %d no equal value count %d", len(query.Column), len(query.Values)))
	}
	return query
}

type SelectStmt struct {
	Column     []*AliasStmt
	Table      string
	TableAlias string
	Join       string
	JoinAlias  string
	On         interface{}
	Where      interface{}
	GroupBy    []string
	Having     interface{}
	Union      *SelectStmt
	OrderBy    []string
	Order      string
	Limit      []string
	Token      []string
}

type AliasStmt struct {
	Expression interface{}
	Alias      string
}

// select [all|distinct] *|column|function [[as]alias][, ...] from table[[as]alias]
// [{[natural]{left|right}[outer]}] join table[[as]alias] {on condition}]
// [where {condition}]
// [group by column [, ...]]
// [having condition]
// [{union[all]}select]
// [order by column[, ...][asc|desc]
// [limit{start}[total]]
func (p *parser) Select() *SelectStmt {
	query := new(SelectStmt)
	query.Token = p.token
	// select [all|distinct] expression[[as]alias][, ...]
	// from {table[[as]alias]|select [as]alias}
	p.selectBase(query)
	// [order by {column|int}[asc|desc]
	if p.Match("order by") {
		for {
			// 必须是普通标识符
			query.OrderBy = append(query.OrderBy, p.mustIdentifier())
			// 没有token，或者不是','
			if p.IsEmpty() || !p.Match(",") {
				break
			}
		}
		// 是否有排序
		t := p.Token()
		if p.Match("asc", "desc", "?") {
			query.Order = t
		}
	}
	// [limit{start}[total]]
	if p.Match("limit") {
		// 必须有至少一个，必须是普通标识符
		query.Limit = append(query.Limit, p.mustIdentifier())
		// 如果是','，那么还有一个
		if p.Match(",") {
			query.Limit = append(query.Limit, p.mustIdentifier())
		}
	}
	return query
}

func (p *parser) selectBase(query *SelectStmt) {
	// [all|distinct]
	p.Match("all", "distinct")
	// expression[[as]alias][, ...] from
	p.selectColumn(query)
	// table[[as]alias]
	query.Table = p.mustIdentifier()
	query.TableAlias = p.selectAlias()
	if p.IsEmpty() {
		return
	}
	// [{{left|right}[outer]|natural|[full]outer}] join table alias {on condition}]
	if strings.Contains(strings.ToLower(p.Token()), "join") {
		p.Next()
		query.Join = p.mustIdentifier()
		query.JoinAlias = p.selectAlias()
		if p.Match("on") {
			query.On = p.ExpressionStmt()
		}
	}
	// [where {condition}]
	if p.Match("where") {
		query.Where = p.ExpressionStmt()
	}
	// [group by column [, ...]]
	if p.Match("group by") {
		for {
			query.GroupBy = append(query.GroupBy, p.mustIdentifier())
			if len(p.token) < 1 || p.token[0] != "," {
				break
			}
		}
	}
	// [having condition]
	if p.Match("having") {
		query.Having = p.ExpressionStmt()
	}
	// [{union[all]}select]
	if p.Match("union", "union all") {
		query.Union = new(SelectStmt)
		p.selectBase(query.Union)
	}
}

func (p *parser) selectAlias() string {
	if p.Match("as") {
		return p.mustIdentifier()
	}
	if !p.IsEmpty() && isIdentifier(p.Token()) {
		t := p.Token()
		p.Next()
		return t
	}
	return ""
}

var (
	errColumnType = errors.New("select field only support one of *|column|function")
)

func (p *parser) selectColumn(query *SelectStmt) {
	// 不能有相同的column，否则对应不了struct
	sameColumn := make(map[string]int)
	for !p.IsEmpty() {
		column := new(AliasStmt)
		t := p.Token()
		if t == "*" {
			if len(query.Column) > 0 {
				panic(errColumnType)
			}
			column.Expression = t
			p.Next()
			p.MustMatch("from")
			query.Column = append(query.Column, column)
			return
		}
		column.Expression = p.ExpressionStmt()
		switch name := column.Expression.(type) {
		case string:
			_, o := sameColumn[name]
			if o {
				panic(fmt.Errorf("unsupported same column '%s'", name))
			}
			sameColumn[name] = 1
			if len(query.Column) > 0 {
				switch query.Column[0].Expression.(type) {
				case string:
				default:
					panic(errColumnType)
				}
			}
		case *FuncExpressionStmt:
			if len(query.Column) > 0 {
				switch query.Column[0].Expression.(type) {
				case *FuncExpressionStmt:
				default:
					panic(errColumnType)
				}
			}
		default:
			panic(errColumnType)
		}
		column.Alias = p.selectAlias()
		query.Column = append(query.Column, column)
		// 如果是from，退出
		if p.Match("from") {
			return
		}
		p.MustMatch(",")
	}
}
