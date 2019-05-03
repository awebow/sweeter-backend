package main

import (
	"errors"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type Sweet struct {
	No       uint64     `json:"no" db:"no"`
	Author   uint64     `json:"author" db:"author"`
	Content  string     `json:"content" db:"content"`
	SweetAt  time.Time  `json:"sweet_at" db:"sweet_at"`
	DeleteAt *time.Time `json:"delete_at" db:"delete_at"`
}

func (app *App) GetSweets(w rest.ResponseWriter, r *rest.Request) {
	sweet, err := FindSweet{No: r.PathParam("no")}.Query(app)
	if err != nil {
		ResponseError(w, err)
		return
	}

	w.WriteJson(sweet)
}

func (app *App) PostSweets(w rest.ResponseWriter, r *rest.Request) {
	claims, err := app.ValidateAuthorization(r)
	if err != nil {
		ResponseError(w, err)
		return
	}

	body := map[string]interface{}{}
	r.DecodeJsonPayload(&body)

	res, err := app.DB.Exec("INSERT INTO sweets (author, content) VALUES (?, ?)", claims.UserNo, body["content"])
	if err != nil {
		if v, ok := err.(*mysql.MySQLError); ok {
			if v.Number == 1062 {
				err = ResourceConflicts
			}
		}

		ResponseError(w, err)
		return
	}

	no, err := res.LastInsertId()
	if err != nil {
		ResponseError(w, err)
		return
	}

	sweet, err := FindSweet{No: no}.Query(app)
	if err != nil {
		ResponseError(w, err)
		return
	}

	w.WriteJson(sweet)
}

func (app *App) GetUsersSweets(w rest.ResponseWriter, r *rest.Request) {
	user, err := FindUser{No: r.PathParam("no")}.Query(app)
	if err != nil {
		ResponseError(w, err)
		return
	}

	sweets, err := FindSweets{Author: user.No}.Query(app)
	if err != nil {
		ResponseError(w, err)
		return
	}

	w.WriteJson(sweets)
}

func (app *App) GetUsersMeSweets(w rest.ResponseWriter, r *rest.Request) {
	claims, err := app.ValidateAuthorization(r)
	if err != nil {
		ResponseError(w, err)
		return
	}

	sweets, err := FindSweets{Author: claims.UserNo}.Query(app)
	if err != nil {
		ResponseError(w, err)
		return
	}

	w.WriteJson(sweets)
}

func (app *App) GetUsersMeNewsfeeds(w rest.ResponseWriter, r *rest.Request) {
	claims, err := app.ValidateAuthorization(r)
	if err != nil {
		ResponseError(w, err)
		return
	}

	users := []interface{}{claims.UserNo}
	rows, err := app.DB.Queryx("SELECT target FROM followings WHERE `user`=?", claims.UserNo)
	if err != nil {
		ResponseError(w, err)
		return
	}

	for rows.Next() {
		var no uint64
		rows.Scan(&no)
		users = append(users, no)
	}

	sweets, err := FindSweets{Authors: users}.Query(app)
	if err != nil {
		ResponseError(w, err)
		return
	}

	w.WriteJson(sweets)
}

type FindSweet struct {
	No      interface{}
	Queryer sqlx.Queryer
}

func (find FindSweet) Query(app *App) (Sweet, error) {
	if find.Queryer == nil {
		find.Queryer = app.DB
	}

	sweet := Sweet{}

	var rows *sqlx.Rows
	var err error
	if find.No != nil {
		rows, err = find.Queryer.Queryx("SELECT * FROM sweets WHERE `no`=?", find.No)
		if err != nil {
			return sweet, err
		}
	} else {
		return sweet, errors.New("Invalid parameters error")
	}

	if rows.Next() {
		err = rows.StructScan(&sweet)
		return sweet, err
	} else {
		return sweet, ResourceNotFound
	}
}

type FindSweets struct {
	Author  interface{}
	Authors []interface{}
	Queryer sqlx.Queryer
}

func (find FindSweets) Query(app *App) ([]Sweet, error) {
	if find.Queryer == nil {
		find.Queryer = app.DB
	}

	sweets := []Sweet{}

	var rows *sqlx.Rows
	var err error
	if find.Author != nil {
		rows, err = find.Queryer.Queryx("SELECT * FROM sweets WHERE `author`=?", find.Author)
		if err != nil {
			return sweets, err
		}
	} else if len(find.Authors) > 0 {
		query, args, err := sqlx.In("SELECT * FROM sweets WHERE `author` IN (?)", find.Authors)
		if err != nil {
			return sweets, err
		}

		query = app.DB.Rebind(query)
		rows, err = find.Queryer.Queryx(query, args...)
		if err != nil {
			return sweets, err
		}
	} else {
		return sweets, errors.New("Invalid parameters error")
	}

	for rows.Next() {
		sweet := Sweet{}
		err = rows.StructScan(&sweet)
		if err != nil {
			return nil, err
		}
		sweets = append(sweets, sweet)
	}

	return sweets, nil
}
