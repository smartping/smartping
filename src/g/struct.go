package g

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
