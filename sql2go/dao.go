package dao

import (
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"strings"
	"time"
)

var (
	DB    *sql.DB
	stmt0 *sql.Stmt // select count(id) sum from app
)

func IsUniqueKeyError(err error) bool {
	if e, o := err.(*mysql.MySQLError); o {
		return e.Number == 1169
	}
	return false
}

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
	stmt0, err = db.Prepare("select count(id) sum from app")
	if err != nil {
		return
	}
	return
}

func CloseStmt() {
	if stmt0 != nil {
		_ = stmt0.Close()
	}
}

// select count(id) sum from app
func GetAppCount() (int64, error) {
	var model int64
	return model, stmt0.QueryRow().Scan(&model)
}

type GetAppsModel struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Detail string `json:"detail"`
}

// select * from app order by id desc limit ?,?
func GetApps(tx *sql.Tx, order, sort string, begin, total interface{}) ([]*GetAppsModel, error) {
	var str strings.Builder
	str.WriteString("select * from app order by ")
	str.WriteString(order)
	str.WriteString(" ")
	str.WriteString(sort)
	str.WriteString(" limit ?,?")
	rows, err := tx.Query(
		str.String(),
		begin,
		total,
	)
	if nil != err {
		return nil, err
	}
	var models []*GetAppsModel
	var Detail sql.NullString
	for rows.Next() {
		model := new(GetAppsModel)
		err = rows.Scan(
			&model.Id,
			&model.Name,
			&Detail,
		)
		if nil != err {
			return nil, err
		}
		model.Detail = Detail.String
		models = append(models, model)
	}
	return models, nil
}
