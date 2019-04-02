package main

import (
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/jmoiron/sqlx"

	"github.com/ant0ine/go-json-rest/rest"
)

type User struct {
	No         uint64     `json:"no" db:"no"`
	ID         string     `json:"id" db:"id"`
	Name       string     `json:"name" db:"name"`
	RegisterAt time.Time  `json:"register_at" db:"register_at"`
	WithdrawAt *time.Time `json:"withdraw_at" db:"withdraw_at"`
}

func (app *App) GetUsers(w rest.ResponseWriter, r *rest.Request) {
	find := FindUser{}
	if no := r.PathParam("no"); len(no) > 0 {
		find.No = no
	} else if id := r.URL.Query()["id"]; len(id) > 0 {
		find.ID = id[0]
	} else {
		ResponseError(w, BadRequest)
		return
	}

	user, err := find.Query(app)
	if err != nil {
		ResponseError(w, err)
		return
	}

	w.WriteJson(user)
}

func (app *App) GetUsersId(w rest.ResponseWriter, r *rest.Request) {
	user, err := FindUser{ID: r.URL.Query()["id"]}.Query(app)
	if err != nil {
		ResponseError(w, err)
		return
	}

	w.WriteJson(user)
}

func (app *App) PostUsers(w rest.ResponseWriter, r *rest.Request) {
	body := struct {
		ID       string `json:"id" db:"id"`
		Passwrod string `json:"password" db:"password"`
		Name     string `json:"name" db:"name"`
	}{}
	r.DecodeJsonPayload(&body)

	res, err := app.DB.NamedExec("INSERT INTO users (`id`, `password`, `name`) VALUES (:id, :password, :name)", body)
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

	user, err := FindUser{No: no}.Query(app)
	if err != nil {
		ResponseError(w, err)
		return
	}

	w.WriteJson(user)
}

type FindUser struct {
	No      interface{}
	ID      interface{}
	Queryer sqlx.Queryer
}

func (find FindUser) Query(app *App) (User, error) {
	if find.Queryer == nil {
		find.Queryer = app.DB
	}

	user := User{}

	var rows *sqlx.Rows
	var err error
	if find.No != nil {
		rows, err = find.Queryer.Queryx("SELECT * FROM users WHERE `no`=?", find.No)
		if err != nil {
			return user, err
		}
	} else if find.ID != nil {
		rows, err = find.Queryer.Queryx("SELECT * FROM users WHERE `id`=?", find.ID)
		if err != nil {
			return user, err
		}
	} else {
		return user, errors.New("Invalid parameters error")
	}

	if rows.Next() {
		err = rows.StructScan(&user)
		return user, err
	} else {
		return user, ResourceNotFound
	}
}
