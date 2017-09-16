package g

import (
	"sync"
)

//type HttpPort string

type Config struct {
	Ver          string
	Port         int
	Name         string
	Timeout      string
	Ip           string
	Db           string
	Password     string
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
}

type PingResultList struct {
	Lock       sync.Mutex
	PingResult map[string]PingResult
}

//PING RESULT
type PingResult struct {
	MaxDelay  string
	MinDelay  string
	AvgDelay  string
	SendPk    string
	RevcPk    string
	LossPk    string
	LastCheck string
}

type State struct {
	Conf      Config
	Localname string
	Localip   string
	Showtype  string
	Lock      sync.Mutex
	State     map[*Target]TargetStatus
}

type Target struct {
	Name string
	Addr string
	Type        string
	Thdchecksec int
	Thdoccnum   int
	Thdavgdelay int
	Thdloss     int
}

type Topo struct {
	From  map[string]string
	To    map[string]string
	Color string
}

type Showlist struct {
	Tlist       []*Topo
	Nlist       []map[string]string
	Status      map[string]string
	AGraph      []map[string]string
	Alert       string
	Tline       string
	Tsymbolsize string
}

type TargetStatus struct {
	Target    *Target
	MaxDelay  string
	MinDelay  string
	AvgDelay  string
	SendPk    string
	RevcPk    string
	LossPk    string
	LastCheck string
}

type LogInfo struct {
	Logtime   string
	Ip        string
	Name      string
	Maxdelay  string
	Mindelay  string
	Avgdelay  string
	Sendpk    string
	Revcpk    string
	Losspk    string
	Lastcheck string
}

type TopoLog struct {
	Ip          string
	Name        string
	Maxavgdelay string
	Maxlosspk   string
	Cnt         string
}

type ToitemStyle struct {
	Normal map[string]string `json:"normal"`
}

type Todataarea struct {
	Name      string      `json:"name"`
	ItemStyle ToitemStyle `json:"itemStyle"`
}

type Todataline struct {
	Source    string      `json:"source"`
	Target    string      `json:"target"`
	ItemStyle ToitemStyle `json:"itemStyle"`
}

type Todata struct {
	Area []Todataarea `json:"area"`
	Line []Todataline `json:"line"`
}

type Alterdata struct {
	Logtime   string
	Fromname  string
	Toname    string
	Alerttype int
}
