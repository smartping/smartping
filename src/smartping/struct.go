package main

import (
	"sync"
)

type Config struct {
	Name     string
	Ip       string
	Db       string
	Alertsound    	     string
	Thresholchecksec        string
	Thresholdavgdelay    string
	Thresholdloss        string
	Tline    string
	Tsymbolsize string
	Type     string
	Targets  []Target
}

type State struct {
	Localname string
	Localip string
	Showtype string
	Lock  sync.Mutex
	State map[*Target]TargetStatus
}

type Target struct {
	Name string
	Addr string
	Interval string
	Type string
}

type topo struct{
	From map[string]string
	To map[string]string
	Color string
}

type showlist struct {
	Tlist []*topo
	Nlist []map[string]string
	Status map[string]string
	AGraph []map[string]string
	Alert string
	Tline string
	Tsymbolsize string
}

type TargetStatus struct {
	Target    *Target
	MaxDelay  string
	MinDelay  string
	AvgDelay  string
	SendPk      string
	RevcPk      string
	LossPk	  string
	LastCheck string
}

type LogInfo struct{
	logtime string
	ip       string
	name     string
	maxdelay string
	mindelay string
	avgdelay string
	sendpk   string
	revcpk   string
	losspk   string
	lastcheck  string

}