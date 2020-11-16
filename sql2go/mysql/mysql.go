package mysql

//type sqlSegment struct {
//	string string
//	_type  string
//	param  bool
//}
//
//type sqlParser struct {
//	seg []*sqlSegment
//	sql strings.Builder
//}
//
//func (p *sqlParser) Parse(s string) error {
//	// 解析sql片段
//	s = strings.TrimSpace(s)
//	if s == "" {
//		return nil
//	}
//	i := 0
//Loop:
//	for ; i < len(s); i++ {
//		switch s[i] {
//		case '\'':
//			j := i
//			i++
//			for ; i < len(s); i++ {
//				if s[i] == '\'' && s[i-1] != '\\' {
//					i++
//					continue Loop
//				}
//			}
//			return parseError(s[j:])
//		case '{':
//			// "{}"前的sql
//			if i != 0 {
//				p.seg = append(p.seg, &sqlSegment{string: s[:i]})
//				s = s[i:]
//			}
//			// "{}"
//			i = strings.IndexByte(s, '}')
//			if i < 0 {
//				return parseError(s)
//			}
//			j := i + 1
//			// "{string:type}"
//			ss := strings.Split(s[1:i], ":")
//			switch len(ss) {
//			case 1:
//				p.seg = append(p.seg, &sqlSegment{string: ss[0], param: true})
//			case 2:
//				p.seg = append(p.seg, &sqlSegment{string: ss[0], _type: ss[1], param: true})
//			default:
//				return parseError(s[:j])
//			}
//			s = s[j:]
//			i = 0
//		}
//	}
//	if i != 0 {
//		p.seg = append(p.seg, &sqlSegment{string: s})
//	}
//	return nil
//}

//func (p *sqlParser) readToken(s string) (string, string) {
//	s = strings.TrimSpace(s)
//	if s == "" {
//		return "", ""
//	}
//	switch s[0] {
//	case ',':
//		return s[:1], s[1:]
//	case '\'':
//		for i := 1; i < len(s); i++ {
//			if s[i] == '\'' && s[i-1] != '\\' {
//				i++
//				return s[:i], s[i:]
//			}
//		}
//	case '(':
//		return p.readParenthesesToken(s)
//	case '{':
//		return p.readBracesToken(s)
//	default:
//		for i := 1; i < len(s); i++ {
//			switch s[i] {
//			case ' ', ',', ';', '\'', '(', '{':
//				return s[:i], s[i:]
//			}
//		}
//	}
//	return "", s
//}
//
//func (p *sqlParser) readStringToken(s string) (string, string) {
//	i := 1
//	for ; i < len(s); i++ {
//		if s[i] == '\'' && s[i-1] != '\\' {
//			i++
//			break
//		}
//	}
//	return s[:i], s[i:]
//}
//
//func (p *sqlParser) readBracesToken(s string) (string, string) {
//	n, m := 1, 1
//	for n < len(s) {
//		if s[n] == '{' {
//			m++
//		}
//		if s[n] == '}' {
//			m--
//			if m == 0 {
//				n++
//				return s[0:n], s[n:]
//			}
//		}
//		n++
//	}
//	return s, ""
//}
//
//func (p *sqlParser) readParenthesesToken(s string) (string, string) {
//	n, m := 1, 1
//	for n < len(s) {
//		if s[n] == '(' {
//			m++
//		}
//		if s[n] == ')' {
//			m--
//			if m == 0 {
//				n++
//				return s[0:n], s[n:]
//			}
//		}
//		n++
//	}
//	return s, ""
//}
//
//func (p *sqlParser) readTokens(s string) []string {
//	var t string
//	var ts []string
//	for {
//		t, s = p.readToken(s)
//		if s == "" {
//			return ts
//		}
//		ts = append(ts, t)
//	}
//}
//
//func match(t string, ss ...string) bool {
//	t = strings.ToLower(t)
//	for _, s := range ss {
//		if t == strings.ToLower(s) {
//			return true
//		}
//	}
//	return false
//}
//
//func parseError(token string) error {
//	return fmt.Errorf("parse error at '%s'", token)
//}
//
//func saveTemplate(t TPL, path string) error {
//	// 打开文件写
//	f, err := os.OpenFile(path, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, os.ModePerm)
//	if err != nil {
//		return err
//	}
//	// 关闭文件
//	defer func() { _ = f.Close() }()
//	// 输出
//	return t.Execute(f)
//}
//
//// 返回token和剩下的s
//func readToken(sql string) (string, string) {
//	readTokenBetween := func(c1, c2 byte, s string) (string, string) {
//		n, m := 1, 1
//		for n < len(s) {
//			if s[n] == c1 {
//				m++
//			}
//			if s[n] == c2 {
//				m--
//				if m == 0 {
//					n++
//					return s[0:n], s[n:]
//				}
//			}
//			n++
//		}
//		return s, ""
//	}
//	sql = strings.TrimSpace(sql)
//	if sql == "" {
//		return "", ""
//	}
//	switch sql[0] {
//	case ',':
//		return sql[:1], sql[1:]
//	case '\'':
//		i := 1
//		for i < len(sql) {
//			if sql[i] == '\'' && sql[i-1] != '\\' {
//				i++
//				return sql[:i], sql[i:]
//			}
//			i++
//		}
//	case '(':
//		return readTokenBetween('(', ')', sql)
//	case '{':
//		return readTokenBetween('{', '}', sql)
//	default:
//		i := 1
//		for i < len(sql) {
//			switch sql[i] {
//			case ' ', ',', ';', '\'', '(', '{':
//				return sql[:i], sql[i:]
//			}
//			i++
//		}
//	}
//	return sql, ""
//}
//
//func readTokens(sql string) []string {
//	var tokens []string
//	var token string
//	for {
//		token, sql = readToken(sql)
//		if token == "" {
//			return tokens
//		}
//		tokens = append(tokens, token)
//	}
//}
//
//// sql的分片，区分普通字符串和占位符"{...}"。
//// 比如，"select * from user where id={id}"，
//// 分成"select * from user where id="和"{id}"
//type segment struct {
//	string string
//	holder bool
//	column *db2go.Column
//}
//
//// 解析tokens到segments
//func readSegments(tokens []string) ([]*segment, error) {
//	var segments []*segment
//	var str strings.Builder
//	var lastToken string
//	for len(tokens) > 0 {
//		lastIndex := len(tokens[0]) - 1
//		if tokens[0][0] == '{' && tokens[0][lastIndex] == '}' {
//			s := tokens[0][1:lastIndex]
//			if s == "" {
//				return nil, parseError(tokens[0])
//			}
//			if str.Len() > 0 {
//				segments = append(segments, &segment{string: strings.TrimSpace(str.String())})
//				str.Reset()
//			}
//			segments = append(segments, &segment{string: s, holder: true})
//		} else {
//			if tokens[0][0] == '(' && tokens[0][lastIndex] == ')' {
//				if str.Len() > 0 {
//					segments = append(segments, &segment{string: strings.TrimSpace(str.String())})
//					str.Reset()
//				}
//				segments = append(segments, &segment{string: "("})
//				sub, err := readSegments(readTokens(tokens[0][1:lastIndex]))
//				if err != nil {
//					return nil, err
//				}
//				segments = append(segments, sub...)
//				segments = append(segments, &segment{string: ")"})
//			} else {
//				if lastToken != "" &&
//					lastToken != "," &&
//					lastToken != ";" &&
//					tokens[0] != "," &&
//					tokens[0] != ";" &&
//					lastToken[0] != '(' &&
//					lastToken[len(lastToken)-1] != ')' {
//					str.WriteByte(' ')
//				}
//				str.WriteString(tokens[0])
//			}
//		}
//		lastToken = tokens[0]
//		tokens = tokens[1:]
//	}
//	if str.Len() > 0 {
//		segments = append(segments, &segment{string: strings.TrimSpace(str.String())})
//		str.Reset()
//	}
//	return segments, nil
//}
//
//// 解析tokens到segments，直到遇到keywords其中一个
//func readSegmentsUntilMatch(tokens []string, keywords ...string) ([]*segment, []string, error) {
//	i := 1
//	for {
//		if len(tokens) <= i ||
//			matchAny(tokens[i], keywords...) {
//			break
//		}
//		i++
//	}
//	segs, err := readSegments(tokens[1:i])
//	if err != nil {
//		return nil, tokens, err
//	}
//	tokens = tokens[i:]
//	return segs, tokens, nil
//}
//
//// 将segments写到str中，替换{holder}为'?'，
//func writeSegments(str *strings.Builder, segs []*segment) {
//	str.WriteByte(' ')
//	var last string
//	for _, seg := range segs {
//		if seg.holder {
//			str.WriteByte('?')
//			last = "?"
//		} else {
//			if last != "" {
//				c1 := last[len(last)-1]
//				if c1 == ')' {
//					if seg.string[0] != ',' && seg.string[0] != ')' {
//						str.WriteByte(' ')
//					}
//				} else if c1 == '?' {
//					if seg.string[0] != ',' && seg.string[0] != ')' {
//						str.WriteByte(' ')
//					}
//				}
//			}
//			str.WriteString(seg.string)
//			last = seg.string
//		}
//	}
//}
//
//// 将segments的holder转成ars
//func segmentsHolderToArgs(table *db2go.Table, segs []*segment) argsTPL {
//	var args argsTPL
//	for _, seg := range segs {
//		if seg.holder {
//			p := strings.Split(seg.string, ".")
//			if len(p) == 2 {
//				args = append(args, &argTPL{name: snakeCaseToPascalCase(p[0]) + "." + snakeCaseToPascalCase(p[1])})
//			} else {
//				if table.GetColumn(p[0]) != nil {
//					args = append(args, &argTPL{name: snakeCaseToPascalCase(p[0])})
//				} else {
//					args = append(args, &argTPL{name: p[0], param: true})
//				}
//			}
//		}
//	}
//	return args
//}
//
//// 提取segments中的holder
//func segmentsHolders(segs []*segment) []string {
//	var holder []string
//	for _, seg := range segs {
//		if seg.holder {
//			holder = append(holder, seg.string)
//		}
//	}
//	return holder
//}
//
//type selectStmt struct {
//	sql        strings.Builder // 最终生成的预编译sql
//	table1     *db2go.Table
//	distinct   string
//	column     []string
//	selectAll  bool   // 是否全选"*"
//	join       string // join的sql
//	table2     *db2go.Table
//	on         []*segment
//	where      []*segment
//	group      []*segment
//	having     []*segment
//	union      string
//	unionStmt  *selectStmt
//	order      []*segment
//	sort       *segment // order有值才有效
//	limit      []*segment
//	funcReturn []string
//	scan       []*scanTPL // 扫描的字段
//	args       argsTPL    // 预编译sql的参数，字段/入参
//	segs       []*segment // 所有的sql片段
//	holders    []string   // 所有的占位符"{}"
//}
//
//func (q *selectStmt) tableName() string {
//	if q.table2 == nil {
//		return q.table1.Name()
//	}
//	return q.table1.Name() + "_join_" + q.table2.Name()
//}
//
//func (q *selectStmt) isQueryStructPage() bool {
//	return len(q.group) > 0 || len(q.order) > 0
//}
//
//func (q *selectStmt) isQueryStruct() bool {
//	if q.table2 != nil || len(q.funcReturn) > 0 {
//		return false
//	}
//	var mul []*db2go.Column
//	for i, a := range q.args {
//		if !a.param {
//			col := q.table1.GetColumn(q.holders[i])
//			// 主键或唯一
//			if col.IsPrimaryKey() || col.IsUnique() {
//				return true
//			}
//			// 多唯一
//			if col.IsMulUnique() {
//				mul = append(mul, col)
//			}
//		}
//	}
//	// 联合唯一
//	if len(mul) < 1 {
//		return false
//	}
//	// 表的所有联合唯一列
//	mulColumn, _ := q.table1.MulUniqueColumns()
//	// 如果不相等，肯定不包含了
//	if len(mul) != len(mulColumn) {
//		return false
//	}
//	for len(mul) > 0 {
//		// mul的第一项是否在mulColumn中
//		ok := false
//		for i, mc := range mulColumn {
//			if mc == mul[0] {
//				mulColumn = append(mulColumn[:i], mulColumn[i+1:]...)
//				ok = true
//				break
//			}
//		}
//		if !ok {
//			break
//		}
//		// 在，去掉第一项，继续
//		mul = mul[1:]
//		// 完全匹配
//		if len(mulColumn) == 0 {
//			return true
//		}
//	}
//	return false
//}
//
//func (q *selectStmt) prepareSQL() {
//	q.sql.Reset()
//	q.sql.WriteString("select ")
//	if q.distinct != "" {
//		q.sql.WriteString(q.distinct)
//		q.sql.WriteString(" ")
//	}
//	q.sql.WriteString(strings.Join(q.column, ","))
//	q.sql.WriteString(" from ")
//	q.sql.WriteString(q.table1.Name())
//	if q.table2 != nil {
//		q.sql.WriteByte(' ')
//		q.sql.WriteString(q.join)
//		q.sql.WriteByte(' ')
//		q.sql.WriteString(q.table2.Name())
//	}
//	if len(q.on) > 0 {
//		q.sql.WriteString(" on")
//		writeSegments(&q.sql, q.on)
//	}
//	if len(q.where) > 0 {
//		q.sql.WriteString(" where")
//		writeSegments(&q.sql, q.where)
//	}
//	if len(q.having) > 0 {
//		q.sql.WriteString(" having")
//		writeSegments(&q.sql, q.having)
//	}
//	if q.unionStmt != nil {
//		q.sql.WriteByte(' ')
//		q.sql.WriteString(q.union)
//		q.sql.WriteByte(' ')
//		q.unionStmt.prepareSQL()
//		q.sql.WriteString(q.unionStmt.sql.String())
//	}
//	if len(q.limit) > 0 {
//		q.sql.WriteString(" limit")
//		writeSegments(&q.sql, q.limit)
//	}
//}
//
//func sqlNullType(colType string) (string, string) {
//	switch colType {
//	case "int8", "int16", "int32", "uint8", "uint16", "uint32":
//		return "sql.NullInt32", "Int32"
//	case "int", "int64", "uint", "uint64":
//		return "sql.NullInt64", "Int64"
//	case "float32", "float64":
//		return "sql.NullFloat64", "Float64"
//	default:
//		return "sql.NullString", "String"
//	}
//}
//
//// select [all|distinct] *|column|function [[as]alias][, ...] from table[[as]alias]
//// [{[natural]{left|right}[outer]}] join table[[as]alias] {on condition}]
//// [where {condition}]
//// [group by column [, ...]]
//// [having condition]
//// [{union[all]}select]
//// [order by column[, ...][asc|desc]
//// [limit{start}[total]]
//func parseSelect(q *selectStmt, tokens []string, schema *db2go.Schema) (err error) {
//	defer func() {
//		q.args = append(q.args, segmentsHolderToArgs(q.table1, q.limit)...)
//		err = parseSelectCheckScanColumn(q)
//	}()
//	tokens, err = parseSelectBase(q, tokens, schema)
//	if err != nil {
//		return err
//	}
//	// order
//	{
//		if len(tokens) < 1 {
//			return nil
//		}
//		if matchAny(tokens[0], "order") {
//			if len(tokens) < 1 {
//				return errParseEOF
//			}
//			if !matchAny(tokens[1], "by") {
//				return parseError(tokens[1])
//			}
//			tokens = tokens[2:]
//			i := 0
//			for {
//				// column
//				if len(tokens) < 0 {
//					return errParseEOF
//				}
//				i++
//				if len(tokens) <= i {
//					return nil
//				}
//				// order by column limit
//				if matchAny(tokens[i], "limit") {
//					break
//				}
//				// order by column sort
//				if tokens[i] != "," {
//					j := len(tokens[i]) - 1
//					if tokens[i][0] == '{' && tokens[i][j] == '}' {
//						s := tokens[i][1:j]
//						if s == "" {
//							return parseError(tokens[i])
//						}
//						q.sort = &segment{string: s, holder: true}
//					} else {
//						q.sort = &segment{string: tokens[i]}
//					}
//					i++
//					break
//				}
//				i++
//				// order by column1,column2
//			}
//			if q.sort != nil {
//				q.order, err = readSegments(tokens[:i-1])
//			} else {
//				q.order, err = readSegments(tokens[:i])
//			}
//			if err != nil {
//				return err
//			}
//			q.segs = append(q.segs, q.order...)
//			tokens = tokens[i:]
//		}
//	}
//	// limit
//	{
//		if len(tokens) < 1 {
//			return nil
//		}
//		if matchAny(tokens[0], "limit") {
//			var err error
//			q.limit, err = readSegments(tokens[1:])
//			if err != nil {
//				return err
//			}
//			q.segs = append(q.segs, q.limit...)
//			q.holders = append(q.holders, segmentsHolders(q.limit)...)
//		}
//	}
//	return nil
//}
//
//// select [all|distinct] *|column|function [[as]alias][, ...] from table[[as]alias]
//// [{[natural]{left|right}[outer]}] join table[[as]alias] {on condition}]
//// [where {condition}]
//// [group by column [, ...]]
//// [having condition]
//// [{union[all]}select]
//func parseSelectBase(q *selectStmt, tokens []string, schema *db2go.Schema) ([]string, error) {
//	var err error
//	if len(tokens) < 1 {
//		return nil, errParseEOF
//	}
//	// distinct
//	if matchAny(tokens[0], "distinct", "all") {
//		q.distinct = tokens[0]
//		tokens = tokens[1:]
//	}
//	// column
//	{
//		i := 1
//		for {
//			if len(tokens) <= i {
//				return nil, errParseEOF
//			}
//			// 一直到关键字"from"
//			if matchAny(tokens[i], "from") {
//				q.column = strings.Split(strings.Join(tokens[:i], ""), ",")
//				i++
//				tokens = tokens[i:]
//				break
//			}
//			i++
//		}
//	}
//	// 表名table1
//	{
//		if len(tokens) < 1 {
//			return nil, errParseEOF
//		}
//		q.table1 = schema.GetTable(tokens[0])
//		if q.table1 == nil {
//			return nil, parseError(tokens[0])
//		}
//		tokens = tokens[1:]
//	}
//	// join
//	{
//		if len(tokens) < 1 {
//			return tokens, nil
//		}
//		i := 0
//		if matchAny(tokens[i], "natural", "left", "right", "inner", "outer") {
//			i = 1
//			for {
//				if len(tokens) <= i {
//					return nil, errParseEOF
//				}
//				// 一直到关键字"join"
//				if matchAny(tokens[i], "join") {
//					break
//				}
//				i++
//			}
//		}
//		if matchAny(tokens[i], "join") {
//			i++
//			q.join = strings.Join(tokens[:i], " ")
//			tokens = tokens[i:]
//			// 表名table2
//			{
//				if len(tokens) < 1 {
//					return nil, errParseEOF
//				}
//				q.table2 = schema.GetTable(tokens[0])
//				if q.table2 == nil {
//					return nil, parseError(tokens[0])
//				}
//				tokens = tokens[1:]
//			}
//			// on
//			{
//				if len(tokens) < 1 {
//					return tokens, nil
//				}
//				if matchAny(tokens[0], "on") {
//					q.on, tokens, err = readSegmentsUntilMatch(tokens,
//						"where", "group", "having", "union", "order", "limit")
//					if err != nil {
//						return nil, err
//					}
//					q.segs = append(q.segs, q.on...)
//					q.args = append(q.args, segmentsHolderToArgs(q.table1, q.on)...)
//					q.holders = append(q.holders, segmentsHolders(q.on)...)
//				}
//			}
//		}
//	}
//	// where
//	{
//		if len(tokens) < 1 {
//			return tokens, nil
//		}
//		if matchAny(tokens[0], "where") {
//			q.where, tokens, err = readSegmentsUntilMatch(tokens,
//				"group", "having", "union", "order", "limit")
//			if err != nil {
//				return nil, err
//			}
//			q.args = append(q.args, segmentsHolderToArgs(q.table1, q.where)...)
//			q.segs = append(q.segs, q.where...)
//			q.holders = append(q.holders, segmentsHolders(q.where)...)
//		}
//	}
//	// group
//	{
//		if len(tokens) < 1 {
//			return tokens, nil
//		}
//		if matchAny(tokens[0], "group") {
//			if len(tokens) < 1 {
//				return nil, errParseEOF
//			}
//			if !matchAny(tokens[1], "by") {
//				return nil, parseError(tokens[1])
//			}
//			q.group, tokens, err = readSegmentsUntilMatch(tokens[1:],
//				"having", "union", "order", "limit")
//			if err != nil {
//				return nil, err
//			}
//			q.segs = append(q.segs, q.group...)
//		}
//	}
//	// having
//	{
//		if len(tokens) < 1 {
//			return tokens, nil
//		}
//		if matchAny(tokens[0], "having") {
//			q.having, tokens, err = readSegmentsUntilMatch(tokens, "union", "order", "limit")
//			if err != nil {
//				return nil, err
//			}
//			q.args = append(q.args, segmentsHolderToArgs(q.table1, q.having)...)
//			q.segs = append(q.segs, q.having...)
//			q.holders = append(q.holders, segmentsHolders(q.having)...)
//		}
//	}
//	// union
//	{
//		if len(tokens) < 1 {
//			return tokens, nil
//		}
//		if matchAny(tokens[0], "union") {
//			// 一直到关键字"select"
//			i := 1
//			for {
//				if len(tokens) <= i {
//					return nil, errParseEOF
//				}
//				if matchAny(tokens[i], "select") {
//					break
//				}
//				i++
//			}
//			q.union = strings.Join(tokens[:i], " ")
//			tokens = tokens[i+1:]
//			q.unionStmt = new(selectStmt)
//			tokens, err = parseSelectBase(q.unionStmt, tokens, schema)
//			if err != nil {
//				return nil, err
//			}
//			q.args = append(q.args, q.unionStmt.args...)
//			q.segs = append(q.segs, q.unionStmt.segs...)
//			q.holders = append(q.holders, segmentsHolders(q.unionStmt.segs)...)
//		}
//	}
//	return tokens, nil
//}
//
//func parseSelectCheckScanColumn(q *selectStmt) error {
//	// scan column
//	for i, col := range q.column {
//		if col == "*" {
//			if len(q.scan) > 1 || len(q.funcReturn) > 0 {
//				return parseError(col)
//			}
//			q.selectAll = true
//			if q.table2 != nil {
//				t1Name := snakeCaseToPascalCase(q.table1.Name())
//				t2Name := snakeCaseToPascalCase(q.table2.Name())
//				for _, cc := range q.table1.Columns() {
//					scan := new(scanTPL)
//					scan.Name = t1Name + "." + snakeCaseToPascalCase(cc.Name())
//					if cc.IsNullable() {
//						scan.NullType, scan.NullValue = sqlNullType(cc.GoType())
//					}
//					q.scan = append(q.scan, scan)
//				}
//				for _, cc := range q.table2.Columns() {
//					scan := new(scanTPL)
//					scan.Name = t2Name + "." + snakeCaseToPascalCase(cc.Name())
//					if cc.IsNullable() {
//						scan.NullType, scan.NullValue = sqlNullType(cc.GoType())
//					}
//					q.scan = append(q.scan, scan)
//				}
//			} else {
//				for _, cc := range q.table1.Columns() {
//					scan := new(scanTPL)
//					scan.Name = snakeCaseToPascalCase(cc.Name())
//					if cc.IsNullable() {
//						scan.NullType, scan.NullValue = sqlNullType(cc.GoType())
//					}
//					q.scan = append(q.scan, scan)
//				}
//			}
//		} else {
//			j := strings.IndexByte(col, '{')
//			if j > 0 {
//				// 函数
//				if len(q.scan) > 0 {
//					return parseError(col)
//				}
//				q.column[i] = col[:j]
//				q.funcReturn = append(q.funcReturn, col[j+1:len(col)-1])
//			} else {
//				// column
//				scan := new(scanTPL)
//				p := strings.Split(col, ".")
//				if len(p) == 1 {
//					scan.Name = snakeCaseToPascalCase(p[0])
//					cc := q.table1.GetColumn(p[0])
//					if cc != nil {
//						scan.Type = cc.GoType()
//						scan.Tag = fmt.Sprintf("`json:\"%s,omitempy\"`", snakeCaseToCamelCase(cc.Name()))
//						if cc.IsNullable() {
//							scan.NullType, scan.NullValue = sqlNullType(cc.GoType())
//						}
//					} else {
//						return parseError(col)
//					}
//				} else {
//					scan.Name = snakeCaseToPascalCase(snakeCaseToPascalCase(p[0]) + "." + snakeCaseToPascalCase(p[1]))
//					if p[0] == q.table1.Name() {
//						cc := q.table1.GetColumn(p[1])
//						if cc != nil {
//							scan.Type = cc.GoType()
//							scan.Tag = fmt.Sprintf("`json:\"%s,omitempy\"`", snakeCaseToCamelCase(cc.Name()))
//							if cc.IsNullable() {
//								scan.NullType, scan.NullValue = sqlNullType(cc.GoType())
//							}
//						} else {
//							return parseError(col)
//						}
//					} else {
//						if p[0] == q.table2.Name() {
//							cc := q.table2.GetColumn(p[1])
//							if cc != nil {
//								scan.Type = cc.GoType()
//								scan.Tag = fmt.Sprintf("`json:\"%s,omitempy\"`", snakeCaseToCamelCase(cc.Name()))
//								if cc.IsNullable() {
//									scan.NullType, scan.NullValue = sqlNullType(cc.GoType())
//								}
//							} else {
//								return parseError(col)
//							}
//						} else {
//							return parseError(col)
//						}
//					}
//				}
//				q.scan = append(q.scan, scan)
//			}
//		}
//	}
//	return nil
//}
//
//func selectSqlSegments(q *selectStmt) ([]string, []string) {
//	var params, segments []string
//	q.sql.WriteString("select ")
//	if q.distinct != "" {
//		q.sql.WriteString(q.distinct)
//		q.sql.WriteString(" ")
//	}
//	q.sql.WriteString(strings.Join(q.column, ","))
//	q.sql.WriteString(" from ")
//	q.sql.WriteString(q.table1.Name())
//	if q.table2 != nil {
//		q.sql.WriteByte(' ')
//		q.sql.WriteString(q.join)
//		q.sql.WriteByte(' ')
//		q.sql.WriteString(q.table2.Name())
//	}
//	if len(q.on) > 0 {
//		q.sql.WriteString(" on")
//		writeSegments(&q.sql, q.on)
//	}
//	if len(q.where) > 0 {
//		q.sql.WriteString(" where")
//		writeSegments(&q.sql, q.where)
//	}
//	// group
//	if len(q.group) > 0 {
//		q.sql.WriteString(" group by ")
//		segments = append(segments, fmt.Sprintf(`"%s"`, q.sql.String()))
//		q.sql.Reset()
//		for _, seg := range q.group {
//			if seg.holder {
//				segments = append(segments, seg.string)
//				params = append(params, seg.string)
//			} else {
//				segments = append(segments, fmt.Sprintf(`"%s"`, seg.string))
//			}
//		}
//	}
//	if len(q.having) > 0 {
//		q.sql.WriteString(" having")
//		writeSegments(&q.sql, q.having)
//	}
//	if q.unionStmt != nil {
//		q.sql.WriteByte(' ')
//		q.sql.WriteString(q.union)
//		q.sql.WriteByte(' ')
//		p, s := selectSqlSegments(q.unionStmt)
//		params = append(params, p...)
//		segments = append(segments, s...)
//		q.sql.Reset()
//	}
//	// order
//	if len(q.order) > 0 {
//		q.sql.WriteString(" order by ")
//		segments = append(segments, fmt.Sprintf(`"%s"`, q.sql.String()))
//		q.sql.Reset()
//		for _, seg := range q.order {
//			if seg.holder {
//				segments = append(segments, seg.string)
//				params = append(params, seg.string)
//			} else {
//				segments = append(segments, fmt.Sprintf(`"%s"`, seg.string))
//			}
//		}
//		if q.sort != nil {
//			segments = append(segments, `" "`)
//			if q.sort.holder {
//				segments = append(segments, q.sort.string)
//				params = append(params, q.sort.string)
//			} else {
//				segments = append(segments, fmt.Sprintf(`"%s"`, q.sort.string))
//			}
//		}
//	}
//	if len(q.limit) > 0 {
//		q.sql.WriteString(" limit")
//		writeSegments(&q.sql, q.limit)
//		segments = append(segments, fmt.Sprintf(`"%s"`, q.sql.String()))
//	}
//	return params, segments
//}
//
//type execStmt struct {
//	sql   strings.Builder
//	table *db2go.Table
//	segs  []*segment
//	args  argsTPL
//}
//
//type deleteStmt struct {
//	execStmt
//}
//
//// delete from table [where condition]
//func parseDelete(q *deleteStmt, tokens []string, schema *db2go.Schema) error {
//	// 关键字"from"
//	{
//		if len(tokens) == 0 {
//			return errParseEOF
//		}
//		if !matchAny(tokens[0], "from") {
//			return parseError(tokens[0])
//		}
//		tokens = tokens[1:]
//	}
//	// 表名"table"
//	{
//		if len(tokens) == 0 {
//			return errParseEOF
//		}
//		q.table = schema.GetTable(tokens[0])
//		if q.table == nil {
//			return parseError(tokens[0])
//		}
//		tokens = tokens[1:]
//	}
//	// 剩下的sql片段
//	{
//		if len(tokens) > 0 {
//			var err error
//			// sql片段
//			q.segs, err = readSegments(tokens)
//			if err != nil {
//				return err
//			}
//			// 模板参数
//			q.args = segmentsHolderToArgs(q.table, q.segs)
//		}
//	}
//	// 替换占位符{}，重新生成数据库sql
//	{
//		q.sql.WriteString("delete from ")
//		q.sql.WriteString(q.table.Name())
//		writeSegments(&q.sql, q.segs)
//	}
//	return nil
//}
//
//type updateStmt struct {
//	execStmt
//}
//
//// update table set column=expression[, ...] [where condition]
//func parseUpdate(q *updateStmt, tokens []string, schema *db2go.Schema) error {
//	// 表名"table"
//	{
//		if len(tokens) == 0 {
//			return errParseEOF
//		}
//		q.table = schema.GetTable(tokens[0])
//		if q.table == nil {
//			return parseError(tokens[0])
//		}
//		tokens = tokens[1:]
//	}
//	// 剩下的sql片段
//	{
//		if len(tokens) < 1 {
//			return errParseEOF
//		}
//		var err error
//		// sql片段
//		q.segs, err = readSegments(tokens)
//		if err != nil {
//			return err
//		}
//		// 模板参数
//		q.args = segmentsHolderToArgs(q.table, q.segs)
//	}
//	// 替换占位符{}，重新生成数据库sql
//	{
//		q.sql.WriteString("update ")
//		q.sql.WriteString(q.table.Name())
//		writeSegments(&q.sql, q.segs)
//	}
//	return nil
//}
//
//type insertStmt struct {
//	execStmt
//}
//
//// insert into table [(column[, ...])] {values(expression[, ...])}
//func parseInsert(q *insertStmt, tokens []string, schema *db2go.Schema) error {
//	// 关键字"into"
//	{
//		if len(tokens) == 0 {
//			return errParseEOF
//		}
//		if !matchAny(tokens[0], "into") {
//			return parseError(tokens[0])
//		}
//		tokens = tokens[1:]
//	}
//	// 表名"table"
//	{
//		if len(tokens) == 0 {
//			return errParseEOF
//		}
//		q.table = schema.GetTable(tokens[0])
//		if q.table == nil {
//			return parseError(tokens[0])
//		}
//		tokens = tokens[1:]
//	}
//	// 剩下的sql片段
//	{
//		if len(tokens) < 1 {
//			return errParseEOF
//		}
//		var err error
//		// sql片段
//		q.segs, err = readSegments(tokens)
//		if err != nil {
//			return err
//		}
//		// 模板参数
//		q.args = segmentsHolderToArgs(q.table, q.segs)
//	}
//	// 替换占位符{}，重新生成数据库sql
//	{
//		q.sql.WriteString("insert into ")
//		q.sql.WriteString(q.table.Name())
//		writeSegments(&q.sql, q.segs)
//	}
//	return nil
//}
