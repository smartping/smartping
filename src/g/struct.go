package g

import (
	"sync"
)

type Config struct {
	Ver      	string
	Name     	string
	Ip       	string
	Db       	string
	Alertsound    	string
	Thdchecksec     int
	Thdoccnum	int
	Thdavgdelay    	int
	Thdloss        	int
	Tline    	string
	Tsymbolsize 	string
	Topotimeout 	string
	Targets  	[]Target
}

type State struct {
	Conf 		Config
	Localname 	string
	Localip 	string
	Showtype 	string
	Lock  		sync.Mutex
	State 		map[*Target]TargetStatus
}

type Target struct {
	Name 		string
	Addr 		string
	Interval 	string
	Type 		string
	Thdchecksec     int
	Thdoccnum	int
	Thdavgdelay    	int
	Thdloss        	int
}

type Topo struct{
	From 		map[string]string
	To 		map[string]string
	Color 		string
}

type Showlist struct {
	Tlist 		[]*Topo
	Nlist 		[]map[string]string
	Status 		map[string]string
	AGraph 		[]map[string]string
	Alert 		string
	Tline 		string
	Tsymbolsize 	string
}

type TargetStatus struct {
	Target    	*Target
	MaxDelay  	string
	MinDelay  	string
	AvgDelay  	string
	SendPk      	string
	RevcPk      	string
	LossPk	  	string
	LastCheck 	string
}

type LogInfo struct{
	Logtime 	string
	Ip       	string
	Name     	string
	Maxdelay 	string
	Mindelay 	string
	Avgdelay 	string
	Sendpk   	string
	Revcpk   	string
	Losspk   	string
	Lastcheck  	string
}

type TopoLog struct{
	Ip       	string
	Name     	string
	Maxavgdelay	string
	Maxlosspk	string
	Cnt		string
}