package db

import (
	"database/sql"
	"time"
)

type User struct {
	Id          int
	YahooId     string
	DisplayName sql.NullString
	Url         string
}

type Brand struct {
	Id        int
	BrandName string
	Url       string
}

type Post struct {
	Id        int
	UserId    int
	BrandId   int
	CommentNo string
	Title     string
	Url       string
	RefNo     sql.NullString
	RefUrl    sql.NullString
	Detail    string
	PostTime  time.Time
}

type PostView struct {
	Id        int
	UserId    int
	BrandId   int
	CommentNo string
	Title     string
	Url       string
	RefNo     sql.NullString
	RefUrl    sql.NullString
	Detail    string
	PostTime  time.Time
	BrandName string
	BrandUrl  string
}
