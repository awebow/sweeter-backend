package main

import (
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/jmoiron/sqlx"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/sha3"
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
	} else if keyword := r.URL.Query()["keyword"]; len(keyword) > 0 {
		search := FindUsers{Keyword: keyword[0]}
		users, err := search.Query(app)
		if err != nil {
			ResponseError(w, err)
			return
		}

		w.WriteJson(users)
		return
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

func (app *App) GetUsersMe(w rest.ResponseWriter, r *rest.Request) {
	claims, err := app.ValidateAuthorization(r)
	if err != nil {
		ResponseError(w, err)
		return
	}

	user, err := FindUser{No: claims.UserNo}.Query(app)
	if err != nil {
		ResponseError(w, err)
		return
	}

	w.WriteJson(user)
}

func (app *App) PostUsers(w rest.ResponseWriter, r *rest.Request) {
	body := struct {
		ID       string `json:"id" db:"id"`
		Password string `json:"password" db:"password"`
		Name     string `json:"name" db:"name"`
	}{}
	r.DecodeJsonPayload(&body)

	passwordHash := sha3.Sum256([]byte(body.Password))

	res, err := app.DB.Exec("INSERT INTO users (`id`, `password`, `name`) VALUES (?, ?, ?)", body.ID, passwordHash[:], body.Name)
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

func (app *App) GetUsersFollowings(w rest.ResponseWriter, r *rest.Request) {
	find := FindUsers{FollowedBy: r.PathParam("no")}
	users, err := find.Query(app)
	if err != nil {
		ResponseError(w, err)
		return
	}

	w.WriteJson(users)
}

func (app *App) GetUsersFollowers(w rest.ResponseWriter, r *rest.Request) {
	find := FindUsers{Following: r.PathParam("no")}
	users, err := find.Query(app)
	if err != nil {
		ResponseError(w, err)
		return
	}

	w.WriteJson(users)
}

func (app *App) GetUsersMeFollowings(w rest.ResponseWriter, r *rest.Request) {
	claims, err := app.ValidateAuthorization(r)
	if err != nil {
		ResponseError(w, err)
		return
	}

	find := FindUsers{FollowedBy: claims.UserNo}
	users, err := find.Query(app)
	if err != nil {
		ResponseError(w, err)
		return
	}

	w.WriteJson(users)
}

func (app *App) GetUsersMeFollowers(w rest.ResponseWriter, r *rest.Request) {
	claims, err := app.ValidateAuthorization(r)
	if err != nil {
		ResponseError(w, err)
		return
	}

	find := FindUsers{Following: claims.UserNo}
	users, err := find.Query(app)
	if err != nil {
		ResponseError(w, err)
		return
	}

	w.WriteJson(users)
}

func (app *App) PostUsersMeFollowings(w rest.ResponseWriter, r *rest.Request) {
	claims, err := app.ValidateAuthorization(r)
	if err != nil {
		ResponseError(w, err)
		return
	}

	no := r.PathParam("no")
	_, err = app.DB.Exec("INSERT INTO followings (`user`, `target`) VALUES (?, ?)", claims.UserNo, no)
	if err != nil {
		ResponseError(w, err)
		return
	}

	w.WriteJson(map[string]interface{}{
		"message": "success",
	})
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

type FindUsers struct {
	Keyword    string
	Following  interface{}
	FollowedBy interface{}
	Queryer    sqlx.Queryer
}

func (find FindUsers) Query(app *App) ([]User, error) {
	if find.Queryer == nil {
		find.Queryer = app.DB
	}

	users := []User{}

	var rows *sqlx.Rows
	var err error
	if len(find.Keyword) > 0 {
		rows, err = find.Queryer.Queryx("SELECT * FROM users WHERE `id` LIKE ? OR `name` LIKE ?",
			find.Keyword+"%", find.Keyword+"%")
		if err != nil {
			return users, err
		}
	} else if find.Following != nil {
		rows, err = find.Queryer.Queryx("SELECT u.* FROM users u JOIN followings f ON u.`no`=f.`user` WHERE f.`target`=?", find.Following)
		if err != nil {
			return users, err
		}
	} else if find.FollowedBy != nil {
		rows, err = find.Queryer.Queryx("SELECT u.* FROM users u JOIN followings f ON u.`no`=f.`target` WHERE f.`user`=?", find.FollowedBy)
		if err != nil {
			return users, err
		}
	} else {
		return users, errors.New("Invalid parameters error")
	}

	for rows.Next() {
		user := User{}
		rows.StructScan(&user)
		users = append(users, user)
	}

	return users, nil
}

type AuthenticationClaims struct {
	UserNo uint64 `json:"user_no"`
	jwt.StandardClaims
}

func (app *App) PostTokens(w rest.ResponseWriter, r *rest.Request) {
	body := struct {
		ID       string `json:"id"`
		Password string `json:"password"`
	}{}

	r.DecodeJsonPayload(&body)

	passwordHash := sha3.Sum256([]byte(body.Password))

	rows, err := app.DB.Query("SELECT `no` FROM users WHERE `id`=? AND `password`=?", body.ID, passwordHash[:])
	if err != nil {
		ResponseError(w, err)
		return
	}

	if !rows.Next() {
		ResponseError(w, ResourceNotFound)
		return
	}

	claims := AuthenticationClaims{}
	rows.Scan(&claims.UserNo)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(app.SigningKey))
	if err != nil {
		ResponseError(w, err)
		return
	}

	w.WriteJson(map[string]interface{}{
		"authorization": tokenString,
	})
}

func (app *App) ValidateToken(tokenString string) (*AuthenticationClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AuthenticationClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(app.SigningKey), nil
	})

	if err != nil {
		return nil, Unauthorized
	}

	if claims, ok := token.Claims.(*AuthenticationClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, Unauthorized
	}
}

func (app *App) ValidateAuthorization(r *rest.Request) (*AuthenticationClaims, error) {
	if auth, ok := r.Header["Authorization"]; ok && len(auth) > 0 && len(auth[0]) > 6 {
		return app.ValidateToken(auth[0][7:])
	}

	return nil, Unauthorized
}
