package main

import (
	"errors"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type Tweet struct {
	No       uint64     `json:"no" db:"no"`
	Author   uint64     `json:"author" db:"author"`
	Content  string     `json:"content" db:"content"`
	TweetAt  time.Time  `json:"tweet_at" db:"tweet_at"`
	DeleteAt *time.Time `json:"delete_at" db:"delete_at"`
}

func (app *App) GetTweets(w rest.ResponseWriter, r *rest.Request) {
	tweet, err := FindTweet{No: r.PathParam("no")}.Query(app)
	if err != nil {
		ResponseError(w, err)
		return
	}

	w.WriteJson(tweet)
}

func (app *App) PostTweets(w rest.ResponseWriter, r *rest.Request) {
	claims, err := app.ValidateAuthorization(r)
	if err != nil {
		ResponseError(w, err)
		return
	}

	body := map[string]interface{}{}
	r.DecodeJsonPayload(&body)

	res, err := app.DB.Exec("INSERT INTO tweets (author, content) VALUES (?, ?)", claims.UserNo, body["content"])
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

	tweet, err := FindTweet{No: no}.Query(app)
	if err != nil {
		ResponseError(w, err)
		return
	}

	w.WriteJson(tweet)
}

func (app *App) GetUsersTweets(w rest.ResponseWriter, r *rest.Request) {
	user, err := FindUser{No: r.PathParam("no")}.Query(app)
	if err != nil {
		ResponseError(w, err)
		return
	}

	tweets, err := FindTweets{Author: user.No}.Query(app)
	if err != nil {
		ResponseError(w, err)
		return
	}

	w.WriteJson(tweets)
}

func (app *App) GetUsersMeTweets(w rest.ResponseWriter, r *rest.Request) {
	claims, err := app.ValidateAuthorization(r)
	if err != nil {
		ResponseError(w, err)
		return
	}

	tweets, err := FindTweets{Author: claims.UserNo}.Query(app)
	if err != nil {
		ResponseError(w, err)
		return
	}

	w.WriteJson(tweets)
}

type FindTweet struct {
	No      interface{}
	Queryer sqlx.Queryer
}

func (find FindTweet) Query(app *App) (Tweet, error) {
	if find.Queryer == nil {
		find.Queryer = app.DB
	}

	tweet := Tweet{}

	var rows *sqlx.Rows
	var err error
	if find.No != nil {
		rows, err = find.Queryer.Queryx("SELECT * FROM tweets WHERE `no`=?", find.No)
		if err != nil {
			return tweet, err
		}
	} else {
		return tweet, errors.New("Invalid parameters error")
	}

	if rows.Next() {
		err = rows.StructScan(&tweet)
		return tweet, err
	} else {
		return tweet, ResourceNotFound
	}
}

type FindTweets struct {
	Author  interface{}
	Queryer sqlx.Queryer
}

func (find FindTweets) Query(app *App) ([]Tweet, error) {
	if find.Queryer == nil {
		find.Queryer = app.DB
	}

	tweets := []Tweet{}

	var rows *sqlx.Rows
	var err error
	if find.Author != nil {
		rows, err = find.Queryer.Queryx("SELECT * FROM tweets WHERE `author`=?", find.Author)
		if err != nil {
			return tweets, err
		}
	} else {
		return tweets, errors.New("Invalid parameters error")
	}

	for rows.Next() {
		tweet := Tweet{}
		err = rows.StructScan(&tweet)
		if err != nil {
			return nil, err
		}
		tweets = append(tweets, tweet)
	}

	return tweets, nil
}
