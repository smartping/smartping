package g

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/cihub/seelog"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	Root string
	Cfg  Config
	//CLock	       sync.Mutex
	SelfCfg        NetworkMember
	AlertStatus    map[string]bool
	AuthUserIpMap  map[string]bool
	AuthAgentIpMap map[string]bool
	ToolLimit      map[string]int
	Db             *sql.DB
	DLock          sync.Mutex
)

func IsExist(fp string) bool {
	_, err := os.Stat(fp)
	return err == nil || os.IsExist(err)
}

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
	//return "D:\\gopath\\src\\github.com\\smartping\\smartping"
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
	if err != nil {
		log.Fatalln("[Fault]log config open fail .", err)
	}
	seelog.ReplaceLogger(logger)
	Cfg = ReadConfig(Root + "/conf/" + cfile)
	if Cfg.Name == "" {
		Cfg.Name, _ = os.Hostname()
	}
	if Cfg.Addr == "" {
		Cfg.Addr = "127.0.0.1"
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
	seelog.Info("Config loaded")
	Db, err = sql.Open("sqlite3", Root+"/db/database.db")
	if err != nil {
		log.Fatalln("[Fault]db open fail .", err)
	}
	SelfCfg = Cfg.Network[Cfg.Addr]
	AlertStatus = map[string]bool{}
	ToolLimit = map[string]int{}
	saveAuth()
}

func SaveCloudConfig(url string) (Config, error) {
	config := Config{}
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(url)
	if err != nil {
		return config, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &config)
	if err != nil {
		config.Name = string(body)
		return config, err
	}
	Name := Cfg.Name
	Addr := Cfg.Addr
	Ver := Cfg.Ver
	Password := Cfg.Password
	Port := Cfg.Port
	Endpoint := Cfg.Mode["Endpoint"]
	Cfg = config
	Cfg.Name = Name
	Cfg.Addr = Addr
	Cfg.Ver = Ver
	Cfg.Port = Port
	Cfg.Password = Password
	Cfg.Mode["LastSuccTime"] = time.Now().Format("2006-01-02 15:04:05")
	Cfg.Mode["Status"] = "true"
	Cfg.Mode["Endpoint"] = Endpoint
	Cfg.Mode["Type"] = "cloud"
	SelfCfg = Cfg.Network[Cfg.Addr]
	saveAuth()
	return config, nil
}

func SaveConfig() error {
	saveAuth()
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

func saveAuth() {
	AuthUserIpMap = map[string]bool{}
	AuthAgentIpMap = map[string]bool{}
	for _, k := range Cfg.Network {
		AuthAgentIpMap[k.Addr] = true
	}
	Cfg.Authiplist = strings.Replace(Cfg.Authiplist, " ", "", -1)
	if Cfg.Authiplist != "" {
		authiplist := strings.Split(Cfg.Authiplist, ",")
		for _, ip := range authiplist {
			AuthUserIpMap[ip] = true
		}
	}
}
