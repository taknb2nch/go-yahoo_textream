package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"../db"
)

type UserJson struct {
	Id          int    `json:"Id"`
	YahooId     string `json:"YahooId"`
	DisplayName string `json:"DisplayName"`
	Url         string `json:"Url"`
}

type PostDto struct {
	BrandName string
	BrandUrl  string
	CommentNo string
	Title     string
	Url       string
	HasRef    bool
	RefNo     string
	RefUrl    string
	Detail    string
	PostTime  time.Time
}

const USER_JSON = "./users.json"

func main() {
	us := readUsersFromJson()

	var users []db.User

	container := db.NewTxContainer()
	container.Do(func(tc *db.TxContainer) error {
		var err error

		l := NewMyLogic(tc)

		l.addUsersIfNotExist(us)

		users, err = l.getUsers()
		if err != nil {
			return err
		}

		return nil
	})

	for _, user := range users {
		var lastPostTime time.Time
		var ok bool

		container.Do(func(tc *db.TxContainer) error {
			var err error

			l := NewMyLogic(tc)

			lastPostTime, ok, err = l.getLastPostTime(user.Id)
			if err != nil {
				return err
			}

			return nil
		})

		if !ok {
			lastPostTime = time.Now().AddDate(-1, 0, 0)
		}

		var displayName string
		if user.DisplayName.Valid {
			displayName = user.DisplayName.String
		} else {
			displayName = user.YahooId
		}

		fmt.Printf("%s(%s): %v\n", displayName, user.YahooId, lastPostTime)

		posts := getPage(user.Url, lastPostTime)

		if len(posts) > 0 {
			fmt.Printf("新規投稿 :　%d件\n-----\n", len(posts))

			for _, post := range posts {
				fmt.Printf("%s\n%s\n%s\n%v\n-----\n", post.BrandName, post.Title, post.Url, post.PostTime)
			}

			container.Do(func(tc *db.TxContainer) error {
				NewMyLogic(tc).savePosts(user.Id, posts)

				return nil
			})
		} else {
			fmt.Printf("投稿なし\n")
		}

		fmt.Println()
	}
}

func readUsersFromJson() []UserJson {
	f, err := os.Open(USER_JSON)
	if err != nil {
		exit(err)
	}

	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		exit(err)
	}

	var users []UserJson

	err = json.Unmarshal(data, &users)
	if err != nil {
		exit(err)
	}

	return users
}

func getPage(url string, lastPostTime time.Time) []PostDto {
	list := make([]PostDto, 0)

	skip := false

	for {
		doc, err := goquery.NewDocument(url)
		if err != nil {
			exit(err)
		}

		doc.Find("li.commentBox").EachWithBreak(func(_ int, sel *goquery.Selection) bool {
			post := PostDto{}

			anchor := sel.Find("div.breadcrumbs ul li a")
			post.BrandUrl, post.BrandName, _ = getHrefAndText(anchor)

			post.CommentNo, _ = trimCommentNo(sel.Find("div.commentHeaderInfo div").Text())
			anchor2 := sel.Find("div.commentHeaderInfo h2 a")
			post.Url, post.Title, _ = getHrefAndText(anchor2)

			uptime := sel.Find("div.ttlInfoDateNum p").Text()
			// 取得した日時は+09:00 JST
			post.PostTime, _ = time.ParseInLocation("2006/01/02 15:04", uptime, time.Local)

			if post.PostTime.Sub(lastPostTime) <= 0 {
				skip = true
				return false
			}

			detail := sel.Find("div.detail")

			anchor4 := detail.Find("span a")
			if isExist(anchor4) {
				post.HasRef = true
				post.RefUrl, post.RefNo, _ = getHrefAndText(anchor4)
				post.RefNo, _ = trimRefNo(post.RefNo)
			} else {
				post.HasRef = false
			}

			post.Detail = trim(detail.Find("p").Text())

			list = append(list, post)

			return true
		})

		if skip {
			break
		}

		next := doc.Find("a:contains(\"次のページ\")").First()

		if !isExist(next) {
			break
		}

		fmt.Println(".")
		time.Sleep(time.Millisecond * 1500)

		href4, _ := next.Attr("href")
		url = href4
	}

	return list
}

func trim(s string) string {
	return strings.Trim(s, " 　\n\r\t")
}

func isExist(sel *goquery.Selection) bool {
	return len(sel.Parent().Nodes) > 0
}

func getHrefAndText(sel *goquery.Selection) (string, string, error) {
	href, exist := sel.Attr("href")
	if !exist {
		return "", "", errors.New("href not found")
	}

	text := trim(sel.Text())

	return href, text, nil
}

func trimCommentNo(s string) (string, error) {
	str := strings.ToLower(trim(s))
	if !strings.HasPrefix(str, "no.") {
		return "", fmt.Errorf("not comment no. : %s", s)
	}

	return str[3:], nil
}

func trimRefNo(s string) (string, error) {
	i := strings.Index(s, " ")
	if i > -1 {
		return s[i+1:], nil
	} else {
		return "", fmt.Errorf("not ref no. : %s", s)
	}
}

func exit(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	os.Exit(1)
}

type MyLogic struct {
	tc *db.TxContainer
}

func NewMyLogic(tc *db.TxContainer) *MyLogic {
	return &MyLogic{tc: tc}
}

func (m *MyLogic) getLastPostTime(userId int) (time.Time, bool, error) {
	var v interface{}

	err := m.tc.Tx.SelectOne(&v, "select max(post_time) as last_post_time from post where user_id = ?", userId)
	if err != nil {
		return time.Now(), false, err
	}

	if v == nil {
		return time.Date(1900, 1, 1, 0, 0, 0, 0, time.Local), false, nil
	} else {
		switch t := v.(type) {
		case []byte:
			//fmt.Println(t, string(t))
			// SQLiteに保存した日時はUTCになるので、UTCで取得後+09:00JSTに変換
			s, err := time.ParseInLocation("2006-01-02 15:04:05", string(t), time.UTC)
			if err != nil {
				exit(err)
			}

			s = s.In(time.Local)

			return s, true, nil
		case time.Time:
			return t, true, nil
		default:
			exit(fmt.Errorf("%v を time.Timeに変換できません。", v))
			return time.Now(), false, nil
		}
	}
}

func (m *MyLogic) savePosts(userId int, posts []PostDto) error {
	for _, post := range posts {
		// 2回コネクションを取得することになる
		brand, err := m.getBrandByName(post.BrandName)
		if err != nil {
			return err
		}

		if brand == nil {
			brand = &db.Brand{
				BrandName: post.BrandName,
				Url:       post.BrandUrl,
			}

			brand, err = m.addBrand(brand)
			if err != nil {
				return err
			}
		}

		et := db.Post{
			UserId:    userId,
			BrandId:   brand.Id,
			CommentNo: post.CommentNo,
			Title:     post.Title,
			Url:       post.Url,
			Detail:    post.Detail,
			PostTime:  post.PostTime,
		}

		if post.HasRef {
			_ = et.RefNo.Scan(post.RefNo)
			_ = et.RefUrl.Scan(post.RefUrl)
		}

		err = m.tc.Tx.Insert(&et)
		if err != nil {
			m.tc.Err = err
			log.Println(err)
			return err
		}
	}

	return nil
}

func (m *MyLogic) addBrand(brand *db.Brand) (*db.Brand, error) {
	err := m.tc.Tx.Insert(brand)
	if err != nil {
		m.tc.Err = err
		log.Println(err)
		return nil, err
	}

	return brand, nil
}

func (m *MyLogic) getBrandByName(brandName string) (*db.Brand, error) {
	if brandName == "" {
		exit(errors.New("brand name is empty"))
	}

	var b db.Brand
	err := m.tc.Tx.SelectOne(&b, "select * from brand where brand_name=?", brandName)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		m.tc.Err = err
		log.Println(err)
		return nil, err
	}

	return &b, nil
}

func (m *MyLogic) addUsersIfNotExist(users []UserJson) error {
	for _, user := range users {
		var u db.User

		err := m.tc.Tx.SelectOne(&u, "select * from user where yahoo_id=?", user.YahooId)
		if err == sql.ErrNoRows {
			u = db.User{
				YahooId: user.YahooId,
				Url:     user.Url,
			}

			if user.DisplayName != "" {
				u.DisplayName.Scan(user.DisplayName)
			}

			err = m.tc.Tx.Insert(&u)
			if err != nil {
				m.tc.Err = err
				return err
			}
		} else if err != nil {
			m.tc.Err = err
			log.Println(err)
			return err
		}
	}

	return nil
}

func (m *MyLogic) getUsers() ([]db.User, error) {
	var users []db.User
	_, err := m.tc.Tx.Select(&users, "select * from user order by id")
	if err != nil {
		m.tc.Err = err
		log.Println(err)
		return nil, err
	}

	return users, nil
}
