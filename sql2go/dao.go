package dao

import (
	"database/sql"
	"strings"
	"time"
)

var (
	DB    *sql.DB
	stmt0 *sql.Stmt // delete from user where id=?
	stmt1 *sql.Stmt // insert into user(name,password,email,mobile,state) values(?,?,?,?,1)
	stmt2 *sql.Stmt // update user set password=? where id=?
	stmt3 *sql.Stmt // select count(id) from user where id>?
	stmt4 *sql.Stmt // select * from user where id>?
)

func Init(url string, maxOpen, maxIdle int, maxLifeTime, maxIdleTime time.Duration) (err error) {
	DB, err = sql.Open("mysql", url)
	if err != nil {
		return err
	}
	DB.SetMaxOpenConns(maxOpen)
	DB.SetMaxIdleConns(maxIdle)
	DB.SetConnMaxLifetime(maxLifeTime)
	DB.SetConnMaxIdleTime(maxIdleTime)
	return PrepareStmt(DB)
}

func UnInit() {
	if DB == nil {
		return
	}
	_ = DB.Close()
	CloseStmt()
}

func PrepareStmt(db *sql.DB) (err error) {
	stmt0, err = db.Prepare("delete from user where id=?")
	if err != nil {
		return
	}
	stmt1, err = db.Prepare("insert into user(name,password,email,mobile,state) values(?,?,?,?,1)")
	if err != nil {
		return
	}
	stmt2, err = db.Prepare("update user set password=? where id=?")
	if err != nil {
		return
	}
	stmt3, err = db.Prepare("select count(id) from user where id>?")
	if err != nil {
		return
	}
	stmt4, err = db.Prepare("select * from user where id>?")
	if err != nil {
		return
	}
	return
}

func CloseStmt() {
	if stmt0 != nil {
		_ = stmt0.Close()
	}
	if stmt1 != nil {
		_ = stmt1.Close()
	}
	if stmt2 != nil {
		_ = stmt2.Close()
	}
	if stmt3 != nil {
		_ = stmt3.Close()
	}
	if stmt4 != nil {
		_ = stmt4.Close()
	}
}

// delete from user where id=?
func UserDeleteById(id interface{}) (sql.Result, error) {
	return stmt0.Exec(
		id,
	)
}

type UserInsertModel struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Mobile   string `json:"mobile"`
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

type UserUpdateModel struct {
	Password string `json:"password"`
	Id       int64  `json:"id"`
}

// update user set password=? where id=?
func UserUpdate(model *UserUpdateModel) (sql.Result, error) {
	return stmt2.Exec(
		model.Password,
		model.Id,
	)
}

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
	var model sql.NullInt64
	for rows.Next() {
		err = rows.Scan(&model)
		if nil != err {
			return nil, err
		}
		if model.Valid {
			models = append(models, model.Int64)
		}
	}
	return models, nil
}

type UserListModel struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
	State    int8   `json:"state"`
	Mobile   string `json:"mobile"`
	Email    string `json:"email"`
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

type UserSearchByNameLikeModel struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
	State    int8   `json:"state"`
	Mobile   string `json:"mobile"`
	Email    string `json:"email"`
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

type AppRoleAccessListModel struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Access int8   `json:"access"`
}

// select id,name,access from (select id,name from res where app_id=?) a left join (select res_id,access from role_res where role_id=?) b on a.id = b.res_id order by id desc limit ?,?
func AppRoleAccessList(order, sort string, appId, roleId, begin, total interface{}) ([]*AppRoleAccessListModel, error) {
	var str strings.Builder
	str.WriteString("select id,name,access from (select id,name from res where app_id=?) a left join (select res_id,access from role_res where role_id=?) b on a.id = b.res_id order by ")
	str.WriteString(order)
	str.WriteString(" ")
	str.WriteString(sort)
	str.WriteString(" limit ?,?")
	models := make([]*AppRoleAccessListModel, 0)
	rows, err := DB.Query(
		str.String(),
		appId,
		roleId,
		begin,
		total,
	)
	if nil != err {
		if err != sql.ErrNoRows {
			return nil, err
		}
		return models, nil
	}
	var Access sql.NullInt32
	for rows.Next() {
		model := new(AppRoleAccessListModel)
		err = rows.Scan(
			&model.Id,
			&model.Name,
			&Access,
		)
		if nil != err {
			return nil, err
		}
		if Access.Valid {
			model.Access = int8(Access.Int32)
		}
		models = append(models, model)
	}
	return models, nil
}
