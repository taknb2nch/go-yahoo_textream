package db

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	"github.com/coopernurse/gorp"
	_ "github.com/mattn/go-sqlite3"

	"../util"
)

const DB_PATH = "./data.db"

func init() {
	//util.LoadConfig()
	initDb(true)
}

type TxContainer struct {
	Tx  *gorp.Transaction
	Err error
}

func NewTxContainer() *TxContainer {
	return &TxContainer{}
}

func (m *TxContainer) Do(function func(tc *TxContainer) error) error {
	var err error

	dbmap := initDb(false)

	defer dbmap.Db.Close()

	m.Err = nil

	m.Tx, err = dbmap.Begin()
	if err != nil {
		return err
	}

	err = function(m)

	if err != nil {
		m.Tx.Rollback()
		return err
	} else if m.Err != nil {
		m.Tx.Rollback()
		return m.Err
	} else {
		m.Tx.Commit()
	}

	return nil
}

func initDb(createTable bool) *gorp.DbMap {
	//db, err := sql.Open("sqlite3", DB_PATH)
	path := util.Cfg.DBFile
	if !filepath.IsAbs(path) {
		path = filepath.Join(os.Getenv("GOPATH"), path)
	}

	db, err := sql.Open("sqlite3", path)

	if err != nil {
		log.Fatalln(err)
		// return err
	}

	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}

	//dbmap.TraceOn("[gorp]", log.New(os.Stdout, "myapp:", log.Lmicroseconds))

	//
	t := dbmap.AddTableWithName(User{}, "user").SetKeys(true, "Id")
	t.ColMap("Id").Rename("id")
	t.ColMap("YahooId").Rename("yahoo_id").SetNotNull(true)
	t.ColMap("DisplayName").Rename("display_name").SetNotNull(false)
	t.ColMap("Url").Rename("url").SetNotNull(true)

	t = dbmap.AddTableWithName(Brand{}, "brand").SetKeys(true, "Id")
	t.ColMap("Id").Rename("id")
	t.ColMap("BrandName").Rename("brand_name").SetNotNull(true)
	t.ColMap("Url").Rename("url").SetNotNull(true)

	t = dbmap.AddTableWithName(Post{}, "post").SetKeys(true, "Id")
	t.ColMap("Id").Rename("id")
	t.ColMap("UserId").Rename("user_id").SetNotNull(true)
	t.ColMap("BrandId").Rename("brand_id").SetNotNull(true)
	t.ColMap("CommentNo").Rename("comment_no").SetNotNull(true)
	t.ColMap("Title").Rename("title").SetNotNull(true)
	t.ColMap("Url").Rename("url").SetNotNull(true)
	t.ColMap("RefNo").Rename("ref_no").SetNotNull(false)
	t.ColMap("RefUrl").Rename("ref_url").SetNotNull(false)
	t.ColMap("Detail").Rename("detail").SetNotNull(true).SetMaxSize(10000)
	t.ColMap("PostTime").Rename("post_time")

	if createTable {
		err = dbmap.CreateTablesIfNotExists()
		if err != nil {
			log.Fatalln(err)
			//return err
		}
	}

	return dbmap
}
