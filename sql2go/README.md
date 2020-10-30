# db2go
这是一个生成golang数据库访问代码的工具
## 运行
支持两种方式
- 作为工具，生成本地代码文件  
db2go --config
- 作为web服务，在线生成代码  
db2go --http 
## docker
docker包装了一个web服务
## 配置
```json
{
  "?": "数据库连接字符串，url.Schema指定driver，目前只有mysql",
  "dbUrl": "mysql://root:123456@tcp(192.168.1.66)/db2go_test",
  "?": "额外(不在mysql/parser.go定义functions中)的解析sql的函数，函数名:返回类型",
  "func": {
    "abs": "float64",
    "avg": "float64",
    "count": "int64"
  },
  "?": "生成代码根目录，空则使用程序当前目录",
  "dir": ".",
  "?": "db代码包名，目录名，空则使用数据库名称",
  "pkg": "dao",
  "?": "数据库表名，生成Struct.S/I/U/D，Count，List的默认代码",
  "default": [
  ],
  "?": "",
  "sql": [
    {
      "?": "sql，会调用prepare检查",
      "sql": "delete from t4 where c3=?",
      "?": "自定义函数名称，覆写自动生成的函数名",
      "func": "DeleteAll",
      "?": "自定义函数入参名称，覆写自动生成的函数名",
      "param": [
        "id"
      ]
    },
    {
      "?": "sql，会调用prepare检查，会在添加tx*sql.Tx到第一个参数",
      "tx": "delete from t2 where name=?"
    },
    {
      "sql": "delete from t2 where id=? or name='1'"
    },
    {
      "sql": "update t5 set c1=?, c2=1,c3=3+?, c4=(4*c1) where id=0"
    }
  ]
}
```
## sql
只支持下面格式的sql（mysql），因为要对应到struct，所以对sql的格式有要求。
- select  
1. "all|distinct"，可以是占位符'?'。
1. "*|column|function[[as]alias][, ...]"，必须是其中一种，不能是混合。
1. "table [[as]alias]"，table不能是子查询。
1. "group by column [, ...]"，column可以是占位符'?'。
1. "order by column [, ...]"，column可以是占位符'?'。
1. "limit{start}[total]"，start和total可以是占位符'?'。
```
select [all|distinct] *|column|function[[as]alias][, ...] from table [[as]alias]
[{[natural]{left|right}[outer]}] join table [[as]alias] {on condition}]
[where {condition}]
[group by column [, ...]]
[having condition]
[{union[all]}select]
[order by column[, ...][asc|desc]
[limit{start}[total]]
```
- insert  
1. table不能是子查询，不能有别名
```
insert into table [(column[, ...])] {values(expression[, ...])|select}
```
- update  
1. table不能是子查询，不能有别名
```
update table set column=expression[, ...] [where condition]
```
- delete  
1. table不能是子查询，不能有别名
```
delete from table [where condition]
```
