package g

import (
	"fmt"
	"github.com/boltdb/bolt"
	"sync"
)

//Main Config
type Config struct {
	Ver         string
	Mode        string //local,cloud
	Endpoint    string //cloud Endpoint
	Status      bool   //cloud status
	Name        string
	Password    string
	Ip          string
	Port        int
	Timeout     string
	Archive     int
	Refresh     int
	Tsound      string
	Tline       string
	Tsymbolsize string
	Toollimit   int
	Targets     []Target
	Chinamap    map[string]map[string][]string
	Authiplist  string
}

//Target Config
type Target struct {
	Agent       bool
	Name        string
	Addr        string
	Revs        bool
	Topo        bool
	Thdchecksec int
	Thdoccnum   int
	Thdavgdelay int
	Thdloss     int
}

//Ping Struct
type PingSt struct {
	SendPk   int
	RevcPk   int
	LossPk   int
	MinDelay float64
	AvgDelay float64
	MaxDelay float64
}

//Ping mini graph Struct
type PingStMini struct {
	Lastcheck []string `json:"lastcheck"`
	LossPk    []string `json:"losspk"`
	AvgDelay  []string `json:"avgdelay"`
}

type PingLog struct {
	Logtime  string
	Maxdelay string
	Mindelay string
	Avgdelay string
	Losspk   string
}

type AlertLog struct {
	Logtime  string
	Fromname string
	Toname   string
	Tracert  string
}

type DbMapSt struct {
	Data map[string]*bolt.DB
	Lock *sync.Mutex
}

func (d DbMapSt) Get(k string) (*bolt.DB, error) {
	d.Lock.Lock()
	defer d.Lock.Unlock()
	if _, ok := d.Data[k]; ok {
		return d.Data[k], nil
	}
	return nil, fmt.Errorf("NotFound")
}

func (d DbMapSt) Set(k string, v *bolt.DB) {
	d.Lock.Lock()
	d.Data[k] = v
	d.Lock.Unlock()
}

type MapVal struct {
	Value float64 `json:"value"`
	Name  string  `json:"name"`
}

type ChinaMp struct {
	Text     string              `json:"text"`
	Subtext  string              `json:"subtext"`
	Avgdelay map[string][]MapVal `json:"avgdelay"`
}

type ToolsRes struct {
	Status string `json:"status"`
	Error  string `json:"error"`
	Ip     string `json:"ip"`
	Ping   PingSt `json:"ping"`
}
