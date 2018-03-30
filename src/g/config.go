package g

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/gy-games-libs/seelog"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	DLock sync.Mutex
	Root  string
	Db    *sql.DB
	Cfg   Config
)

func IsExist(fp string) bool {
	_, err := os.Stat(fp)
	return err == nil || os.IsExist(err)
}

// Opening config file in JSON format
func ReadConfig(filename string) Config {
	config := Config{}
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		log.Fatal("Config Not Found!")
	} else {
		err = json.NewDecoder(file).Decode(&config)
		if err != nil {
			log.Fatal(err)
		}
	}
	return config
}

func GetRoot() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal("Get Root Path Error:", err)
	}
	dirctory := strings.Replace(dir, "\\", "/", -1)
	runes := []rune(dirctory)
	l := 0 + strings.LastIndex(dirctory, "/")
	if l > len(runes) {
		l = len(runes)
	}
	return string(runes[0:l])
}

func ParseConfig(ver string) {
	Root = GetRoot()
	cfile := "config.json"
	if !IsExist(Root + "/conf/" + "config.json") {
		if !IsExist(Root + "/conf/" + "config-base.json") {
			log.Fatalln("[Fault]config file:", Root+"/conf/"+"config(-base).json", "both not existent.")
		}
		cfile = "config-base.json"
	}

	logger, err := seelog.LoggerFromConfigAsFile(Root + "/conf/" + "seelog.xml")
	seelog.ReplaceLogger(logger)
	Cfg = ReadConfig(Root + "/conf/" + cfile)
	if Cfg.Name == "" {
		Cfg.Name, _ = os.Hostname()
	}
	if Cfg.Ip == "" {
		Cfg.Ip = "127.0.0.1"
	}
	Cfg.Ver = ver
	if !IsExist(Root + "/db/" + "database.db") {
		if !IsExist(Root + "/db/" + "database-base.db") {
			log.Fatalln("[Fault]db file:", Root+"/db/"+"database(-base).db", "both not existent.")
		}
		src, err := os.Open(Root + "/db/" + "database-base.db")
		if err != nil {
			log.Fatalln("[Fault]db-base file open error.")
		}
		defer src.Close()
		dst, err := os.OpenFile(Root+"/db/"+"database.db", os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			log.Fatalln("[Fault]db-base file copy error.")
		}
		defer dst.Close()
		io.Copy(dst, src)
	}
	Cfg.Db = Root + "/db/database.db"
	seelog.Info("Config loaded")
	Db, err = sql.Open("sqlite3", Cfg.Db)
	if err != nil {
		log.Fatalln("[Fault]db open fail .", err)
	}
	for k, target := range Cfg.Targets {
		if target.Thdavgdelay == 0 {
			Cfg.Targets[k].Thdavgdelay = Cfg.Thdavgdelay
		}
		if target.Thdchecksec == 0 {
			Cfg.Targets[k].Thdchecksec = Cfg.Thdchecksec
		}
		if target.Thdloss == 0 {
			Cfg.Targets[k].Thdloss = Cfg.Thdloss
		}
		if target.Thdoccnum == 0 {
			Cfg.Targets[k].Thdoccnum = Cfg.Thdoccnum
		}
	}
}

func SaveConfig() error {
	rrs, _ := json.Marshal(Cfg)
	var out bytes.Buffer
	errjson := json.Indent(&out, rrs, "", "\t")
	if errjson != nil {
		seelog.Error("[func:SaveConfig] Json Parse ", errjson)
		return errjson
	}
	err := ioutil.WriteFile(Root+"/conf/"+"config.json", []byte(out.String()), 0644)
	if err != nil {
		seelog.Error("[func:SaveConfig] Config File Write", err)
		return err
	}
	return nil
}
