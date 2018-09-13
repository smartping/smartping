package g

//Main Config
type Config struct {
	Ver       string
	Mode      string //local,cloud
	Cendpoint string //cloud Endpoint
	Name      string
	Password  string
	Ip        string
	Port      int
	Timeout   string
	//	Db           string
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

//Ping Stuct
type PingSt struct {
	SendPk   int
	RevcPk   int
	LossPk   int
	MinDelay float64
	AvgDelay float64
	MaxDelay float64
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

//Topology
type Topo struct {
	From  map[string]string
	To    map[string]string
	Color string
}

type TopoLog struct {
	Maxavgdelay string
	Maxlosspk   string
	Cnt         string
}

type Todataarea struct {
	Name      string      `json:"name"`
	ItemStyle ToitemStyle `json:"itemStyle"`
}

type ToitemStyle struct {
	Normal map[string]string `json:"normal"`
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
	Logtime  string
	Fromname string
	Toname   string
	Tracert  string
}
