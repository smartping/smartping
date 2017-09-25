package g

import (
	"database/sql"
	"encoding/json"
	"github.com/gy-games-libs/file"
	"github.com/gy-games-libs/seelog"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var DLock sync.Mutex

var (
	Root string
)

// Opening (or creating) config file in JSON format
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

func ParseConfig(ver string) (Config, *sql.DB) {

	cfile := "config.json"
	if !file.IsExist(GetRoot() + "/conf/" + "config.json") {
		if !file.IsExist(GetRoot() + "/conf/" + "config-base.json") {
			log.Fatalln("[Fault]config file:", GetRoot()+"/conf/"+"config(-base).json", "both not existent.")
		}
		cfile = "config-base.json"
	}
	Root = GetRoot()
	logger, err := seelog.LoggerFromConfigAsFile(Root + "/conf/" + "seelog.xml")
	seelog.ReplaceLogger(logger)
	cfg := ReadConfig(GetRoot() + "/conf/" + cfile)
	if cfg.Name == "" {
		cfg.Name, _ = os.Hostname()
	}
	if cfg.Ip == "" {
		cfg.Ip = "127.0.0.1"
	}
	if cfg.Ping == "" {
		cfg.Ping = "sysping"
	}
	cfg.Ver = ver
	if !file.IsExist(GetRoot() + "/db/" + "database.db") {
		if !file.IsExist(GetRoot() + "/db/" + "database-base.db") {
			log.Fatalln("[Fault]db file:", GetRoot()+"/db/"+"database(-base).db", "both not existent.")
		}
		src, err := os.Open(GetRoot() + "/db/" + "database-base.db")
		if err != nil {
			log.Fatalln("[Fault]db-base file open error.")
		}
		defer src.Close()
		dst, err := os.OpenFile(GetRoot()+"/db/"+"database.db", os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			log.Fatalln("[Fault]db-base file copy error.")
		}
		defer dst.Close()
		io.Copy(dst, src)
	}
	cfg.Db = GetRoot() + "/db/database.db"
	seelog.Info("Config loaded")
	db, err := sql.Open("sqlite3", cfg.Db)
	if err != nil {
		log.Fatalln("[Fault]db open fail .", err)
	}
	for k, target := range cfg.Targets {
		if target.Thdavgdelay == 0 {
			cfg.Targets[k].Thdavgdelay = cfg.Thdavgdelay
		}
		if target.Thdchecksec == 0 {
			cfg.Targets[k].Thdchecksec = cfg.Thdchecksec
		}
		if target.Thdloss == 0 {
			cfg.Targets[k].Thdloss = cfg.Thdloss
		}
		if target.Thdoccnum == 0 {
			cfg.Targets[k].Thdoccnum = cfg.Thdoccnum
		}
	}

	return cfg, db

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
