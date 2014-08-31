package util

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

const CONFIG_FILE = "./config.json"

var Cfg *Config

type Config struct {
	DBFile string `json:"dbfile"`
}

func init() {
	LoadConfig()
}

func LoadConfig() {
	f, err := os.Open(CONFIG_FILE)
	if err != nil {
		log.Fatalln(err)
	}

	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalln(err)
	}

	err = json.Unmarshal(data, &Cfg)
	if err != nil {
		log.Fatalln(err)
	}
}
