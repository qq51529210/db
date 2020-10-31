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

// 解析sql
func ParseSQL(sql string) (v interface{}, err error) {
	defer func() {
		re := recover()
		if re != nil {
			e, ok := re.(error)
			if ok {
				err = e
			} else {
				err = fmt.Errorf("%v", e)
			}
		}
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
	switch strings.ToLower(p.token[0]) {
	case "select":
		v = p.Select()
	case "insert into":
		v = p.Insert()
	case "update":
		v = p.Update()
	case "delete from":
		v = p.Delete()
	default:
		v = p.ExpressionStmt()
	}
	return
}

type parser struct {
	token  []string // 未解析的token
	parsed []string // 已解析的token
}

// 读取所有的token
func (p *parser) readTokens(sql string) {
	var token string
	for sql != "" {
		sql, token = p.readToken(sql)
		p.token = append(p.token, token)
		switch strings.ToLower(token) {
		case "insert":
			sql = p.checkMultiKeywordToken(sql, true, "into")
		case "delete":
			sql = p.checkMultiKeywordToken(sql, true, "from")
		case "is":
			sql = p.checkMultiKeywordToken(sql, false, "not")
		case "not":
			sql = p.checkMultiKeywordToken(sql, false, "like", "between", "in", "exists")
		case "union":
			sql = p.checkMultiKeywordToken(sql, false, "all")
		case "group", "order":
			sql = p.checkMultiKeywordToken(sql, true, "by")
		case "inner":
			sql = p.checkMultiKeywordToken(sql, true, "join")
		case "left", "right":
			sql = p.checkMultiKeywordToken(sql, false, "outer")
			sql = p.checkMultiKeywordToken(sql, true, "join")
		case "natural":
			sql, token = p.readToken(sql)
			if matchAny(token, "left", "right") {
				p.token[len(p.token)-1] += " " + token
				sql = p.checkMultiKeywordToken(sql, false, "outer")
			}
			sql = p.checkMultiKeywordToken(sql, true, "join")
		case "outer":
			panic(p.tokenError())
		}
	}
}

// 检查多关键字
func (p *parser) checkMultiKeywordToken(sql string, must bool, ss ...string) string {
	var token string
	sql, token = p.readToken(sql)
	if matchAny(token, ss...) {
		p.token[len(p.token)-1] += " " + token
	} else {
		if must {
			panic(p.tokenError())
		}
		p.token = append(p.token, token)
	}
	return sql
}

// 读取token并返回
func (p *parser) readToken(sql string) (string, string) {
	sql = strings.TrimSpace(sql)
	if sql == "" {
		panic(p.tokenError())
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
		return p.readStringToken(sql)
	case '(':
		// 表达式
		return p.readParenthesesToken(sql)
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

// 解析字符串类型的token
func (p *parser) readStringToken(sql string) (string, string) {
	c := sql[0]
	n := 1
	for n < len(sql) {
		if sql[n] == c && sql[n-1] != '\\' {
			n++
			return sql[n:], sql[:n]
		}
		n++
	}
	panic(p.tokenError())
}

// 解析圆括号类型的token
func (p *parser) readParenthesesToken(sql string) (string, string) {
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
	panic(p.tokenError())
}

// 当前解析token的错误
func (p *parser) tokenError() error {
	var str strings.Builder
	for _, t := range p.token {
		str.WriteString(t)
		str.WriteByte(' ')
	}
	return fmt.Errorf("parse error at '%s'", str.String())
}

// 当前解析stmt的错误
func (p *parser) stmtError() error {
	var str strings.Builder
	for _, t := range p.parsed {
		str.WriteString(t)
		str.WriteByte(' ')
	}
	if len(p.token) > 0 {
		return fmt.Errorf("parse '%s' error at '%s'", str.String(), p.token[0])
	}
	return fmt.Errorf("parse '%s' error 'eof'", str.String())
}

func (p *parser) record() string {
	t := p.token[0]
	p.parsed = append(p.parsed, t)
	p.token = p.token[1:]
	return t
}

func (p *parser) mustHasToken() {
	if len(p.token) < 1 {
		panic(p.stmtError())
	}
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
	// 只有单一值或者下一个不是运算符
	if len(p.token) < 1 || !isOperators(p.token[0]) {
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
	t := p.record()
	// (xxx)
	if isParentheses(t) {
		// 解析子表达式
		pp := new(parser)
		pp.readTokens(removeParentheses(t))
		// 是否select子查询
		if matchAny(pp.token[0], "select") {
			return pp.Select()
		}
		// 解析，如果不是整个expression，那就是(x,x,x)
		v := pp.ExpressionStmt()
		if len(pp.token) < 1 {
			return v
		}
		vv := []interface{}{v}
		for len(pp.token) > 0 {
			if pp.token[0] != "," {
				panic(pp.stmtError())
			}
			pp.record()
			vv = append(vv, pp.ExpressionStmt())
		}
		return vv
	}
	// exists(xxx)
	if matchAny(t, "exists", "not exists") {
		expr := new(BoolExpressionStmt)
		expr.Operator = t
		expr.Value = p.mustParenthesesExpressionValue()
		return expr
	}
	// func()
	if isFunction(t) && isParentheses(p.token[0]) {
		expr := new(FuncExpressionStmt)
		expr.Name = t
		expr.Value = p.mustParenthesesExpressionValue()
		return expr
	}
	// 单一值
	if !isIdentifier(t) {
		panic(p.stmtError())
	}
	return t
}

// 解析表达式的运算符和右值
func (p *parser) expressionOperatorsAndRight() *ExpressionStmt {
	t := p.record()
	p.mustHasToken()
	// 新的表达式
	expr := &ExpressionStmt{Operator: t}
	switch strings.ToLower(expr.Operator) {
	case "between", "not between":
		// ? and ?
		and := new(ExpressionStmt)
		pp := new(parser)
		// 左值
		for _, t := range p.token {
			if strings.ToLower(t) == "and" {
				break
			}
			pp.token = append(pp.token, t)
			p.record()
		}
		if len(p.token) < 2 {
			panic(p.stmtError())
		}
		and.Left = pp.ExpressionStmt()
		// and
		and.Operator = p.record()
		// 右值
		and.Right = p.ExpressionStmt()
		expr.Right = and
	case "in", "not in":
		// 右值必须是(xxx)
		expr.Right = p.mustParenthesesExpressionValue()
	case "is", "is not":
		if !matchAny(p.token[0], "null", "?") {
			panic(p.stmtError())
		}
		expr.Right = p.record()
	default:
		expr.Right = p.expressionValue()
	}
	return expr
}

// 当前的token必须是(x,x,x)，并解析
func (p *parser) mustParenthesesExpressionValue() interface{} {
	if len(p.token) < 1 || !isParentheses(p.token[0]) {
		panic(p.stmtError())
	}
	return p.expressionValue()
}

// 当前的token必须是普通标识符
func (p *parser) mustIdentifier() string {
	if len(p.token) < 1 || !isIdentifier(p.token[0]) {
		panic(p.stmtError())
	}
	return p.record()
}

// 当前的token必须是(x,x,x)
func (p *parser) mustParentheses() string {
	if len(p.token) < 1 || !isParentheses(p.token[0]) {
		panic(p.stmtError())
	}
	return p.record()
}

// 当前的token必须匹配
func (p *parser) mustMatch(ss ...string) string {
	if len(p.token) < 1 || !matchAny(p.token[0], ss...) {
		panic(p.stmtError())
	}
	return p.record()
}

// 当前的token如果匹配
func (p *parser) ifMatch(ss ...string) bool {
	if len(p.token) > 0 && matchAny(p.token[0], ss...) {
		p.record()
		return true
	}
	return false
}

type DeleteStmt struct {
	Table string
	Where interface{}
}

// delete from table [where condition]
func (p *parser) Delete() *DeleteStmt {
	p.record()
	//
	query := new(DeleteStmt)
	// table
	query.Table = p.mustIdentifier()
	// [where condition]
	if p.ifMatch("where") {
		query.Where = p.ExpressionStmt()
	}
	return query
}

type UpdateStmt struct {
	Table  string
	Column []*ExpressionStmt
	Where  interface{}
}

// update table set column=expression[, ...] [where condition]
func (p *parser) Update() *UpdateStmt {
	p.record()
	//
	query := new(UpdateStmt)
	// table
	query.Table = p.mustIdentifier()
	// set
	p.mustMatch("set")
	// column=expression[, ...]
	if len(p.token) < 1 {
		panic(p.stmtError())
	}
	for {
		expr := new(ExpressionStmt)
		// column必须是普通标识符
		expr.Left = p.mustIdentifier()
		// 运算符必须是'='
		expr.Operator = p.mustMatch("=")
		// 右值是表达式
		expr.Right = p.ExpressionStmt()
		query.Column = append(query.Column, expr)
		if len(p.token) < 1 || p.token[0] != "," {
			break
		}
		p.record()
	}
	// [where condition]
	if p.ifMatch("where") {
		query.Where = p.ExpressionStmt()
	}
	return query
}

type InsertStmt struct {
	Table  string      // 表名
	Column []string    // 列名
	Value  interface{} // 值
}

// insert into table [(column[, ...])] {values(expression[, ...])|select query}
func (p *parser) Insert() *InsertStmt {
	p.record()
	//
	query := new(InsertStmt)
	// table
	query.Table = p.mustIdentifier()
	// [(column[, ...])]
	if isParentheses(p.token[0]) {
		p.insertColumn(query)
	}
	// values(expression[, ...])
	if p.ifMatch("values") {
		p.insertValues(query)
		return query
	}
	// select query
	if p.ifMatch("select") {
		query.Value = p.Select()
		return query
	}
	panic(p.stmtError())
}

// [(column[, ...])]
func (p *parser) insertColumn(query *InsertStmt) {
	pp := new(parser)
	pp.readTokens(removeParentheses(p.record()))
	for {
		// column必须是标识符
		query.Column = append(query.Column, pp.mustIdentifier())
		if len(pp.token) < 1 {
			return
		}
		// 如果还有column，必须是','隔开
		pp.mustMatch(",")
	}
}

func (p *parser) insertValues(query *InsertStmt) {
	pp := new(parser)
	pp.readTokens(removeParentheses(p.mustParentheses()))
	var vv []interface{}
	for {
		v := pp.ExpressionStmt()
		vv = append(vv, v)
		if len(pp.token) < 1 {
			break
		}
		pp.mustMatch(",")
	}
	// column和values的个数必须相等
	if len(query.Column) > 0 && len(query.Column) != len(vv) {
		panic(fmt.Errorf("column count %d no equal value count %d", len(query.Column), len(vv)))
	}
	query.Value = vv
}

type SelectStmt struct {
	Distinct   string
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
	// select [all|distinct] expression[[as]alias][, ...]
	// from {table[[as]alias]|select [as]alias}
	p.selectBase(query)
	// [order by {column|int}[asc|desc]
	if p.ifMatch("order by") {
		for {
			// 必须是普通标识符
			query.OrderBy = append(query.OrderBy, p.mustIdentifier())
			// 没有token，或者不是','
			if len(p.token) < 1 || p.token[0] != "," {
				break
			}
		}
		// 是否有排序
		if matchAny(p.token[0], "asc", "desc", "?") {
			query.Order = p.record()
		}
	}
	// [limit{start}[total]]
	if p.ifMatch("limit") {
		// 必须有至少一个，必须是普通标识符
		query.Limit = append(query.Limit, p.mustIdentifier())
		// 如果是','，那么还有一个
		if len(p.token) > 0 && p.token[0] == "," {
			p.record()
			query.Limit = append(query.Limit, p.mustIdentifier())
		}
	}
	return query
}

func (p *parser) selectBase(query *SelectStmt) {
	p.record()
	// [all|distinct]
	if matchAny(p.token[0], "all", "distinct", "?") {
		query.Distinct = p.record()
	}
	// expression[[as]alias][, ...] from
	p.selectColumn(query)
	// table[[as]alias]
	query.Table = p.mustIdentifier()
	query.TableAlias = p.selectAlias()
	if len(p.token) < 1 {
		return
	}
	// [{{left|right}[outer]|natural|[full]outer}] join table alias {on condition}]
	if strings.Contains(strings.ToLower(p.token[0]), "join") {
		p.record()
		query.Join = p.mustIdentifier()
		query.JoinAlias = p.selectAlias()
		if p.ifMatch("on") {
			query.On = p.ExpressionStmt()
		}
	}
	// [where {condition}]
	if p.ifMatch("where") {
		query.Where = p.ExpressionStmt()
	}
	// [group by column [, ...]]
	if p.ifMatch("group by") {
		for {
			query.GroupBy = append(query.GroupBy, p.mustIdentifier())
			if len(p.token) < 1 || p.token[0] != "," {
				break
			}
		}
	}
	// [having condition]
	if p.ifMatch("having") {
		query.Having = p.ExpressionStmt()
	}
	// [{union[all]}select]
	if p.ifMatch("union", "union all") {
		query.Union = new(SelectStmt)
		p.selectBase(query.Union)
	}
}

func (p *parser) selectAlias() string {
	if p.ifMatch("as") {
		return p.mustIdentifier()
	}
	if len(p.token) > 0 && isIdentifier(p.token[0]) {
		return p.record()
	}
	return ""
}

var (
	errColumnType = errors.New("select field only support one of *|column|function")
)

func (p *parser) selectColumn(query *SelectStmt) {
	// 不能有相同的column，否则对应不了struct
	sameColumn := make(map[string]int)
	sameAlias := make(map[string]int)
	for len(p.token) > 0 {
		column := new(AliasStmt)
		if p.token[0] == "*" {
			if len(query.Column) > 0 {
				panic(errColumnType)
			}
			column.Expression = p.record()
			if !p.ifMatch("from") {
				panic(p.stmtError())
			}
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
		if column.Alias != "" {
			_, o := sameAlias[column.Alias]
			if o {
				panic(fmt.Errorf("unsupported same alias '%s'", column.Alias))
			}
			sameAlias[column.Alias] = 1
		}
		query.Column = append(query.Column, column)
		// 如果是from，退出
		if p.ifMatch("from") {
			return
		}
		p.mustMatch(",")
	}
}
