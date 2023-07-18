package g

type Config struct {
	Ver        string
	Port       int
	Name       string
	Addr       string
	Mode       map[string]string
	Base       map[string]int
	Topology   map[string]string
	Alert      map[string]string
	Network    map[string]NetworkMember
	Chinamap   map[string]map[string][]string
	Toollimit  int
	Authiplist string
	Password   string
}

type NetworkMember struct {
	Name      string
	Addr      string
	Smartping bool
	Ping      []string
	//Tools map[string][]string
	Topology []map[string]string
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
	Logtime    string
	Targetip   string
	Targetname string
	Tracert    string
	Fromip     string
	Fromname   string
}

type ChinaMp struct {
	Text     string              `json:"text"`
	Subtext  string              `json:"subtext"`
	Avgdelay map[string][]MapVal `json:"avgdelay"`
}

type MapVal struct {
	Value float64 `json:"value"`
	Name  string  `json:"name"`
}

type ToolsRes struct {
	Status string `json:"status"`
	Error  string `json:"error"`
	Ip     string `json:"ip"`
	Ping   PingSt `json:"ping"`
}

type AllDomainstruct struct {
	Domainipslice  map[string][]string
	Domainipmap    map[string][]map[string]string
	Domainipstruct map[string][]NetworkMember
	AllDomainslice []string
	Size           int
}

type Domainstruct struct {
	Result []DomainInfostruct
}

type DomainInfostruct struct {
	ID            int
	Domaintitle   string
	Busline       string
	Principal     string
	On_off_status string
	Status        string
	Create_time   string
	Mod_time      string
}
