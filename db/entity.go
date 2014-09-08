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

type UserPostTimeView struct {
	Id             int
	YahooId        string
	DisplayName    sql.NullString
	Url            string
	PostTimeString sql.NullString
	PostTime       time.Time
	NewPostCount   int
}

type Brand struct {
	Id        int
	BrandName string
	Url       string
}

type BrandPostTimeView struct {
	Id                       int
	BrandName                string
	Url                      string
	PostTimeString           string
	PostTime                 time.Time
	NewPostCount             int
	BrandNotificationBrandId sql.NullInt64
}

type BrandNotification struct {
	BrandId int
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
	Id                     int
	UserId                 int
	BrandId                int
	CommentNo              string
	Title                  string
	Url                    string
	RefNo                  sql.NullString
	RefUrl                 sql.NullString
	Detail                 string
	PostTime               time.Time
	BrandName              string
	BrandUrl               string
	PostNotificationPostId sql.NullInt64
}

type PostNotification struct {
	PostId int
}
