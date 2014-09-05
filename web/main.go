package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"../db"
)

type Page struct {
	Title     string
	Container template.HTML
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", IndexHandler)
	r.HandleFunc("/posts/", PostsHandler)
	r.HandleFunc("/posts/user/{id:[0-9]+}/", PostsByUserHandler)
	r.HandleFunc("/posts/brand/{id:[0-9]+}/", PostsByBrandHandler)
	r.HandleFunc("/users/", UsersHandler)
	r.HandleFunc("/brands/", BrandsHandler)

	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("js"))))
	http.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.Dir("fonts"))))

	http.Handle("/", r)

	log.Println("statring server at localhost:8080 ...")

	if err := http.ListenAndServe(":8080", LoggingServeMux(http.DefaultServeMux)); err != nil {
		log.Fatalln("Could not start server.", err)
	}
}

func LoggingServeMux(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	err := writeOutput(w, "インディックス", "./template/index.tmpl", nil)
	if err != nil {
		writeError(w, err)
		return
	}
}

func PostsHandler(w http.ResponseWriter, r *http.Request) {
	//v := mux.Vars(r)
	container := db.NewTxContainer()

	var posts []db.PostView
	err := container.Do(func(tc *db.TxContainer) error {
		var err error

		posts, err = NewMyLogic2(tc).getPostsAll()

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

func PostsByUserHandler(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	container := db.NewTxContainer()

	id, _ := strconv.Atoi(v["id"])

	var posts []PostDto
	err := container.Do(func(tc *db.TxContainer) error {
		var err error

		l := NewMyLogic2(tc)
		ps, err := l.getPostsByUserId(id)
		if err != nil {
			return err
		}

		posts = convertPostViewToPostDto(ps)

		ids := getNewPostIds(posts)

		err = l.deletePostNotificationByPostId(ids)

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

func PostsByBrandHandler(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	container := db.NewTxContainer()

	id, _ := strconv.Atoi(v["id"])

	var posts []PostDto
	err := container.Do(func(tc *db.TxContainer) error {
		var err error

		l := NewMyLogic2(tc)

		ps, err := l.getPostsByBrandId(id)
		if err != nil {
			return err
		}

		posts = convertPostViewToPostDto(ps)

		ids := getNewPostIds(posts)

		err = l.deletePostNotificationByPostId(ids)

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

func convertPostViewToPostDto(ps []db.PostView) []PostDto {
	posts := make([]PostDto, len(ps))
	for i, p := range ps {
		posts[i].Id = p.Id
		posts[i].UserId = p.UserId
		posts[i].BrandId = p.BrandId
		posts[i].CommentNo = p.CommentNo
		posts[i].Title = p.Title
		posts[i].Url = p.Url
		posts[i].RefNo = p.RefNo.String
		posts[i].RefUrl = p.RefUrl.String
		posts[i].Detail = p.Detail
		posts[i].PostTime = p.PostTime
		posts[i].BrandName = p.BrandName
		posts[i].BrandUrl = p.BrandUrl
		posts[i].IsNewPost = p.PostNotificationPostId.Valid
	}

	return posts
}

func getNewPostIds(ps []PostDto) []int {
	ids := make([]int, 0, len(ps))
	for _, p := range ps {
		if !p.IsNewPost {
			continue
		}

		ids = append(ids, p.Id)
	}

	return ids
}

func convertBrandPostTimeViewToBrandDto(bs []db.BrandPostTimeView) []BrandDto {
	brands := make([]BrandDto, len(bs))
	for i, b := range bs {
		brands[i].Id = b.Id
		brands[i].BrandName = b.BrandName
		brands[i].Url = b.Url
		brands[i].PostTime = b.PostTime
		brands[i].NewPostCount = b.NewPostCount
		brands[i].IsNewBrand = b.BrandNotificationBrandId.Valid
	}

	return brands
}

var baseTmpl = template.Must(template.ParseFiles("./template/base.tmpl"))

func UsersHandler(w http.ResponseWriter, r *http.Request) {
	container := db.NewTxContainer()

	var users []db.UserPostTimeView
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

func BrandsHandler(w http.ResponseWriter, r *http.Request) {
	container := db.NewTxContainer()

	var brands []BrandDto
	err := container.Do(func(tc *db.TxContainer) error {
		var err error

		l := NewMyLogic2(tc)
		bs, err := l.getBrands()

		brands = convertBrandPostTimeViewToBrandDto(bs)

		l.deleteBrandNotification()

		return err
	})

	if err != nil {
		writeError(w, err)
		return
	}

	err = writeOutput(w, "銘柄一覧", "./template/brands.tmpl", brands)
	if err != nil {
		writeError(w, err)
		return
	}
}

func writeOutput(w http.ResponseWriter, title string, templateName string, data interface{}) error {

	funcMap := template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.In(time.Local).Format("2006-01-02 15:04:05")
		},
	}

	c, err := ioutil.ReadFile(templateName)
	if err != nil {
		return err
	}

	//t := template.Must(template.ParseFiles(templateName))
	t := template.Must(template.New("template").Funcs(funcMap).Parse(string(c)))

	var b bytes.Buffer

	err = t.Execute(&b, data)
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

	_, err := m.tc.Tx.Select(&posts, "select A.id, A.user_id as UserId, A.brand_id as BrandId, A.comment_no as CommentNo, A.title as Title, A.url as Url, A.ref_no as RefNo, A.ref_url as RefUrl, A.detail as Detail, A.post_time as PostTime, B.brand_name as BrandName, B.url as BrandUrl, C.post_id as PostNotificationPostId from post A inner join brand B on A.brand_id = B.id left join post_notification C on A.id = C.post_id where A.user_id=? order by A.post_time desc", userId)
	if err != nil {
		m.tc.Err = err
		log.Println(err)
		return nil, err
	}

	return posts, nil
}

func (m *MyLogic2) getPostsByBrandId(brandId int) ([]db.PostView, error) {
	var posts []db.PostView

	_, err := m.tc.Tx.Select(&posts, "select A.id, A.user_id as UserId, A.brand_id as BrandId, A.comment_no as CommentNo, A.title as Title, A.url as Url, A.ref_no as RefNo, A.ref_url as RefUrl, A.detail as Detail, A.post_time as PostTime, B.brand_name as BrandName, B.url as BrandUrl, C.post_id as PostNotificationPostId from post A inner join brand B on A.brand_id = B.id left join post_notification C on A.id = C.post_id where A.brand_id=? order by A.post_time desc", brandId)
	if err != nil {
		m.tc.Err = err
		log.Println(err)
		return nil, err
	}

	return posts, nil
}

func (m *MyLogic2) deletePostNotificationByPostId(ids []int) error {
	// sqlでやるなら
	// // delete from post_notification where post_id in (select id from post A1 inner join post_notification B1 on A1.id = B1.post_id where A1.brand_id=?)
	pa := make([]string, 0, len(ids))
	args := make([]interface{}, 0, len(ids))
	//for i := 0; i < len(ids); i++ {
	for _, id := range ids {
		pa = append(pa, "?")
		args = append(args, id)
	}

	sql := "delete from post_notification where post_id in (" + strings.Join(pa, ",") + ")"

	_, err := m.tc.Tx.Exec(sql, args...)
	if err != nil {
		m.tc.Err = err
		log.Println(err)
		return err
	}

	return nil
}

func (m *MyLogic2) deleteBrandNotification() error {
	// 一定期間表示するには日時を持たせておく
	_, err := m.tc.Tx.Exec("delete from brand_notification")
	if err != nil {
		m.tc.Err = err
		log.Println(err)
		return err
	}

	return nil
}

func (m *MyLogic2) getUsers() ([]db.UserPostTimeView, error) {
	var users []db.UserPostTimeView

	_, err := m.tc.Tx.Select(&users, "select A.id as Id, A.yahoo_id as YahooId, A.display_name as DisplayName, A.url as Url, B.post_time as PostTimeString, B.new_post_count as NewPostCount from user A inner join (select user_id, max(post_time) as post_time, count(B1.post_id) as new_post_count from post A1 left join post_notification B1 on A1.id = B1.post_id group by user_id) B on A.id = B.user_id order by B.post_time desc")
	if err != nil {
		m.tc.Err = err
		log.Println(err)
		return nil, err
	}

	for i, _ := range users {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", users[i].PostTimeString, time.UTC)
		if err != nil {
			return nil, err
		} else {
			users[i].PostTime = t
		}
	}

	return users, nil
}

func (m *MyLogic2) getBrands() ([]db.BrandPostTimeView, error) {
	var bs []db.BrandPostTimeView

	_, err := m.tc.Tx.Select(&bs, "select A.id as Id, A.brand_name as BrandName, A.url as Url, B.post_time as PostTimeString, B.new_post_count as NewPostCount, C.brand_id as BrandNotificationBrandId from brand A inner join (select A1.brand_id, max(A1.post_time) as post_time, count(B1.post_id) as new_post_count from post A1 left join post_notification B1 on A1.id = B1.post_id group by brand_id) B on A.id = B.brand_id left join brand_notification C on A.id = C.brand_id order by B.post_time desc")
	if err != nil {
		m.tc.Err = err
		log.Println(err)
		return nil, err
	}

	for i, _ := range bs {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", bs[i].PostTimeString, time.UTC)
		if err != nil {
			return nil, err
		} else {
			bs[i].PostTime = t
		}
	}

	return bs, nil
}

type PostDto struct {
	Id        int
	UserId    int
	BrandId   int
	CommentNo string
	Title     string
	Url       string
	RefNo     string
	RefUrl    string
	Detail    string
	PostTime  time.Time
	BrandName string
	BrandUrl  string
	IsNewPost bool
}

type BrandDto struct {
	Id             int
	BrandName      string
	Url            string
	PostTimeString string
	PostTime       time.Time
	NewPostCount   int
	IsNewBrand     bool
}
