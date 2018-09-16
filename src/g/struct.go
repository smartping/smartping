package g

import (
	"sync"
	"github.com/boltdb/bolt"
	"fmt"
)

//Main Config
type Config struct {
	Ver       string
	Mode      string //local,cloud
	Cendpoint string //cloud Endpoint
	Cstatus   bool   //cloud status
	Name      string
	Password  string
	Ip        string
	Port      int
	Timeout   string
	Alerthistory int
	Alertcycle   int
	Alertsound   string
	Thdchecksec  int
	Thdoccnum    int
	Thdavgdelay  int
	Thdloss      int
	Tline        string
	Tsymbolsize  string
	Targets      []Target
	Authiplist   string
}

//Target Config
type Target struct {
	Name        string
	Addr        string
	Type        string
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
	Logtime   string
	Maxdelay  string
	Mindelay  string
	Avgdelay  string
	Losspk    string
}

type AlertLog struct {
	Logtime   string
	Fromname string
	Toname string
	Tracert string
}

type DbMapSt struct {
	Data map[string]*bolt.DB
	Lock *sync.Mutex
}

func (d DbMapSt) Get(k string) (*bolt.DB,error){
	d.Lock.Lock()
	defer d.Lock.Unlock()
	if _,ok := d.Data[k]; ok {
		return d.Data[k],nil
	}
	return nil,fmt.Errorf("NotFound")
}

func (d DbMapSt) Set(k string,v *bolt.DB) {
	d.Lock.Lock()
	d.Data[k]=v
	d.Lock.Unlock()
}