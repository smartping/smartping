package g

import (
	"bytes"
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/cihub/seelog"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"sync"
)

var (
	Root           string
	Cfg            Config
	AlertStatus    map[string]bool
	AuthUserIpMap  map[string]bool
	AuthAgentIpMap map[string]bool
	DbMap          DbMapSt
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
	if Cfg.Ip == "" {
		Cfg.Ip = "127.0.0.1"
	}
	if Cfg.Mode == "" {
		Cfg.Mode = "local"
	}
	Cfg.Ver = ver
	DbMap = DbMapSt{}
	DbMap.Data = map[string]*bolt.DB{}
	DbMap.Lock = new(sync.Mutex)
	AlertStatus = map[string]bool{}
	for k, target := range Cfg.Targets {
		if target.Addr != Cfg.Ip {

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
	saveAuth()
}

func GetDb(t string, db string) *bolt.DB {
	dbname:=t+"_"+db
	boltdb,err:=DbMap.Get(dbname)
	if err!=nil{
		boltdb, err := bolt.Open(Root+"/db/"+t+"/"+db+".db", 0600, nil)
		if err != nil {
			seelog.Error("[Error] "+Root+"/db/"+t+"/"+db+".db open fail .", err)
		}
		DbMap.Set(dbname,boltdb)
		return boltdb
	}
	return boltdb
}

func SaveCloudConfig(url string, flag bool) (Config, error) {
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
	if flag == true {
		Cfg.Targets = config.Targets
		Cfg.Mode = "cloud"
		Cfg.Timeout = config.Timeout
		Cfg.Alertcycle = config.Alertcycle
		Cfg.Alerthistory = config.Alerthistory
		Cfg.Alertcycle = config.Alertcycle
		Cfg.Tsymbolsize = config.Tsymbolsize
		Cfg.Tline = config.Tline
		Cfg.Alertsound = config.Alertsound
		Cfg.Cendpoint = url
		Cfg.Authiplist = config.Authiplist
		saveAuth()
	} else {
		config.Mode = "cloud"
		config.Cendpoint = url
		config.Ip = Cfg.Ip
		config.Name = Cfg.Name
		config.Ver = Cfg.Ver
	}
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
	Cfg.Authiplist = strings.Replace(Cfg.Authiplist, " ", "", -1)
	if Cfg.Authiplist != "" {
		authiplist := strings.Split(Cfg.Authiplist, ",")
		for _, ip := range authiplist {
			AuthUserIpMap[ip] = true
		}
		for _, k := range Cfg.Targets {
			AuthAgentIpMap[k.Addr] = true
		}
	}
}
