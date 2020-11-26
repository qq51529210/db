# sql2go
这是一个生成golang数据库访问代码的工具。
## 编译
- [main](./main)包是可执行的程序。
## 使用
sql2go --config cfg.json  
## sql写法
#### delete from user where id={id:int64}  
```
// delete from user where id=?
func UserDeleteById(id interface{}) (sql.Result, error) {
	return stmt0.Exec(
		id,
	)
}
```
#### insert into user(name,password,email,mobile,state) values({name:string},{password:string},{email:string},{mobile:string},1)  
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
#### update user set password={password:string} where id={id:int64}  
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
#### select count(id) from user where id>{id:int64}  
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
#### select * from user where id>{id:int64}  
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
#### select * from user where name like {name:string} order by {order:} {sort:} limit {begin:int64}, {total:int64}  
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
## 解析
#### 参数:{name:type}
1. name是定义个名称
2. type是golang基本数据类型，否则，代表的是分页的order或sort。  
3. 会转换成与编译'?'，分页则是实时组成sql。
#### db.Exec函数:
1. 参数name会转换snake case to pascal case。
2. 如果只有一个，不会生成struct。  
#### db.Query函数:
1. 参数name会转换snake case to camel case。
2. 如果select字段只有一个，不会生成struct。  
3. 返回的总是[]。
## 下一步
实现http方式的在线生成
