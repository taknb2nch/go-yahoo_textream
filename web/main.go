package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	// "time"

	"github.com/gorilla/mux"

	"../db"
)

type Page struct {
	Title     string
	Container template.HTML
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/posts/", PostsHandler)
	r.HandleFunc("/posts/{id:[0-9]+}/", PostsHandler)
	r.HandleFunc("/users/", UsersHandler)

	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("js"))))
	http.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.Dir("fonts"))))

	http.Handle("/", r)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalln("Could not start server.", err)
	}
}

func PostsHandler(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	container := db.NewTxContainer()

	var posts []db.PostView
	err := container.Do(func(tc *db.TxContainer) error {
		var err error

		if s, exists := v["id"]; exists {
			id, _ := strconv.Atoi(s)
			posts, err = NewMyLogic2(tc).getPostsByUserId(id)
		} else {
			posts, err = NewMyLogic2(tc).getPostsAll()
		}

		return err
	})

	if err != nil {
		writeError(w, err)
		return
	}

	err = writeOutput(w, "投稿一覧", "./template/posts.tmpl", posts)
	if err != nil {
		writeError(w, err)
		return
	}
}

var baseTmpl = template.Must(template.ParseFiles("./template/base.tmpl"))

func UsersHandler(w http.ResponseWriter, r *http.Request) {
	container := db.NewTxContainer()

	var users []db.User
	err := container.Do(func(tc *db.TxContainer) error {
		var err error
		users, err = NewMyLogic2(tc).getUsers()

		return err
	})

	if err != nil {
		writeError(w, err)
		return
	}

	err = writeOutput(w, "ユーザ一覧", "./template/users.tmpl", users)
	if err != nil {
		writeError(w, err)
		return
	}
}

func writeOutput(w http.ResponseWriter, title string, templateName string, data interface{}) error {

	// funcMap := template.FuncMap{
	// 	"formatTime": func(t time.Time) string {
	// 		return t.In(time.Local).Format("2006-01-02 15/04/05")
	// 	},
	// }

	t := template.Must(template.ParseFiles(templateName))

	var b bytes.Buffer

	err := t.Execute(&b, data)
	if err != nil {
		return err
	}

	p := &Page{
		Title:     title,
		Container: template.HTML(b.String()),
	}

	err = baseTmpl.Execute(w, p)
	if err != nil {
		return err
	}

	return nil
}

func writeError(w http.ResponseWriter, err error) {
	http.Error(
		w,
		fmt.Sprintf("%s\n\n%v", http.StatusText(http.StatusInternalServerError), err),
		http.StatusInternalServerError)
}

type MyLogic2 struct {
	tc *db.TxContainer
}

func NewMyLogic2(tc *db.TxContainer) *MyLogic2 {
	return &MyLogic2{tc: tc}
}

func (m *MyLogic2) getPostsAll() ([]db.PostView, error) {
	var posts []db.PostView

	_, err := m.tc.Tx.Select(&posts, "select A.id, A.user_id as UserId, A.brand_id as BrandId, A.comment_no as CommentNo, A.title as Title, A.url as Url, A.ref_no as RefNo, A.ref_url as RefUrl, A.detail as Detail, A.post_time as PostTime, B.brand_name as BrandName, B.url as BrandUrl from post A inner join brand B on A.brand_id = B.id order by A.post_time desc")
	if err != nil {
		m.tc.Err = err
		log.Println(err)
		return nil, err
	}

	return posts, nil
}

func (m *MyLogic2) getPostsByUserId(userId int) ([]db.PostView, error) {
	var posts []db.PostView

	_, err := m.tc.Tx.Select(&posts, "select A.id, A.user_id as UserId, A.brand_id as BrandId, A.comment_no as CommentNo, A.title as Title, A.url as Url, A.ref_no as RefNo, A.ref_url as RefUrl, A.detail as Detail, A.post_time as PostTime, B.brand_name as BrandName, B.url as BrandUrl from post A inner join brand B on A.brand_id = B.id where A.user_id=? order by A.post_time desc ", userId)
	if err != nil {
		m.tc.Err = err
		log.Println(err)
		return nil, err
	}

	return posts, nil
}

func (m *MyLogic2) getUsers() ([]db.User, error) {
	var users []db.User

	_, err := m.tc.Tx.Select(&users, "select * from user order by id")
	if err != nil {
		m.tc.Err = err
		log.Println(err)
		return nil, err
	}

	return users, nil
}
