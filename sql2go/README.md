# sql2go
这是一个生成golang数据库访问代码的工具。比如，
## 使用
1. 编译[main](./main)包，得到执行的程序。
2. main --config cfg.json

## 程序配置

下面是配置文件的例子，字段有解析。

```json
{
  "?": "数据库连接字符串，url.Schema指定driver，目前只有mysql",
  "dbUrl": "mysql://root:123456@tcp(192.168.1.66)/pro_rbac",
  "?": "生成代码文件路径，空则使用程序当前目录+数据库名.go",
  "file": "dao",
  "?": "db代码包名，空则使用文件名称",
  "pkg": "dao",
  "?": "生成函数",
  "func": [
    {
  		"?": "函数的名称",
      "name": "UserDeleteById",
  		"?": "sql，写多行，避免过长不好阅读",
      "sql": [
        "delete from user where id={id:int64}"
      ]
    },
    {
      "name": "UserInsert",
      "sql": [
        "insert into user(name,password,email,mobile,state)",
        "values({name:string},{password:string},{email:string},{mobile:string},1)"
      ]
    },
    {
      "name": "UserUpdate",
      "sql": [
        "update user set password={password:string} where id={id:int64}"
      ]
    },
    {
      "name": "UserCount",
      "sql": [
        "select count(id) from user where id>{id:int64}"
      ]
    },
    {
      "name": "UserList",
      "sql": [
        "select * from user where id>{id:int64}"
      ]
    },
    {
      "name": "UserSearchByNameLike",
      "sql": [
        "select * from user where name like {name:string}",
        "order by {order:id} limit {begin:int64}, {total:int64}"
      ]
    },
    {
      "name": "AppRoleAccessList",
      "sql": [
        "select id,name,access from",
        "(select id,name from res where app_id={appId:int64}) a",
        "left join",
        "(select res_id,access from role_res where role_id={roleId:int64}) b",
        "on a.id = b.res_id",
        "order by {order:id} {sort:desc} limit {begin:int64},{total:int64}"
      ]
    }
  ]
}

```

## sql例子
`delete from user where id={id:int64}`

```
// delete from user where id=?
func UserDeleteById(id interface{}) (sql.Result, error) {
	return stmt0.Exec(
		id,
	)
}
```
`insert into user(name,password,email,mobile,state) values({name:string},{password:string},{email:string},{mobile:string},1)`

```
type UserInsertModel struct {
	Name string `json:"name"`
	Password string `json:"password"`
	Email string `json:"email"`
	Mobile string `json:"mobile"`
}

// insert into user(name,password,email,mobile,state) values(?,?,?,?,1)
func UserInsert(model *UserInsertModel) (sql.Result, error) {
	return stmt1.Exec(
		model.Name,
		model.Password,
		model.Email,
		model.Mobile,
	)
}
```
`update user set password={password:string} where id={id:int64}`

```
type UserUpdateModel struct {
	Password string `json:"password"`
	Id int64 `json:"id"`
}

// update user set password=? where id=?
func UserUpdate(model *UserUpdateModel) (sql.Result, error) {
	return stmt2.Exec(
		model.Password,
		model.Id,
	)
}
```
`select count(id) from user where id>{id:int64}`

```
// select count(id) from user where id>?
func UserCount(id interface{}) ([]int64, error) {
	models := make([]int64, 0)
	rows, err := stmt3.Query(
		id,
	)
	if nil != err {
		if err != sql.ErrNoRows {
			return nil, err
		}
		return models, nil
	}
	var model int64
	for rows.Next() {
		err = rows.Scan(&model)
		if nil != err {
			return nil, err
		}
		models = append(models, model)
	}
	return models, nil
}
```
`select * from user where id>{id:int64}`

```
type UserListModel struct {
	Id int `json:"id"`
	Name string `json:"name"`
	Password string `json:"password"`
	State int8 `json:"state"`
	Mobile string `json:"mobile"`
	Email string `json:"email"`
}

// select * from user where id>?
func UserList(id interface{}) ([]*UserListModel, error) {
	models := make([]*UserListModel, 0)
	rows, err := stmt4.Query(
		id,
	)
	if nil != err {
		if err != sql.ErrNoRows {
			return nil, err
		}
		return models, nil
	}
	for rows.Next() {
		model := new(UserListModel)
		err = rows.Scan(
			&model.Id,
			&model.Name,
			&model.Password,
			&model.State,
			&model.Mobile,
			&model.Email,
		)
		if nil != err {
			return nil, err
		}
		models = append(models, model)
	}
	return models, nil
}
```
`select * from user where name like {name:string} order by {order:} {sort:} limit {begin:int64}, {total:int64}`

```
type UserSearchByNameLikeModel struct {
	Id int `json:"id"`
	Name string `json:"name"`
	Password string `json:"password"`
	State int8 `json:"state"`
	Mobile string `json:"mobile"`
	Email string `json:"email"`
}

// select * from user where name like ? order by id limit ?, ?
func UserSearchByNameLike(order string, name, begin, total interface{}) ([]*UserSearchByNameLikeModel, error) {
	var str strings.Builder
	str.WriteString("select * from user where name like ? order by ")
	str.WriteString(order)
	str.WriteString(" limit ?, ?")
	models := make([]*UserSearchByNameLikeModel, 0)
	rows, err := DB.Query(
		str.String(),
		name,
		begin,
		total,
	)
	if nil != err {
		if err != sql.ErrNoRows {
			return nil, err
		}
		return models, nil
	}
	for rows.Next() {
		model := new(UserSearchByNameLikeModel)
		err = rows.Scan(
			&model.Id,
			&model.Name,
			&model.Password,
			&model.State,
			&model.Mobile,
			&model.Email,
		)
		if nil != err {
			return nil, err
		}
		models = append(models, model)
	}
	return models, nil
}
```
## sql参数
1. 格式**{name:type}**或者**{name:}**。
2. name是函数参数名，或者结构体的字段。
3. type是golang基本数据类型，如果为空，代表的是分页的order或sort。  
4. 如果有分页，生成的是原始sql的代码，否则生成的事预编译代码。
## 下一步
实现http方式的在线生成
