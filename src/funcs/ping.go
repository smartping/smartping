package funcs

import (
	"fmt"
	"github.com/cihub/seelog"
	_ "github.com/mattn/go-sqlite3"
	"github.com/smartping/smartping/src/g"
	"github.com/smartping/smartping/src/nettools"
	"net"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	//"reflect"
)

func MonitorDomain() {
	seelog.Info("[func:Monitor_Domain] Start monitor domain...")
	alldomainslice := g.Ads.AllDomainslice //所有的域名
	if g.Ads.Size == 0 {
		seelog.Info("not have domain")
		time.Sleep(5 * time.Second)
	} else if g.Ads.Size > 0 {
		seelog.Info("[func:Monitor_Domain] Have domain")
		for _, domainstr := range alldomainslice {
			thisdomainipslice := g.Ads.Domainipslice[domainstr]
			nownetworkipslice := DnsResolve(domainstr, 0)
			seelog.Info(fmt.Sprintf("[func:Monitor_Domain] configfile %s %v", domainstr, thisdomainipslice))
			seelog.Info(fmt.Sprintf("[func:Monitor_Domain] resolve %s %v", domainstr, nownetworkipslice))

			addslice, delslice := CompareSlice(nownetworkipslice, thisdomainipslice)

			//add
			if len(addslice) != 0 {
				seelog.Info(fmt.Sprintf("[func:Monitor_Domain] Have add domain ip %v", addslice))
				for networkey, networkvalue := range g.Cfg.Network {
					networkMember := g.Cfg.Network[networkey]
					if networkvalue.Smartping {
						selfpingslice := g.Cfg.Network[networkey].Ping
				                selfpingtopology := g.Cfg.Network[networkey].Topology
                                                newpingslice := make([]string,len(selfpingslice))
                                                newpingtopology := make([]map[string]string,len(selfpingtopology))
                                                copy(newpingslice,selfpingslice)
                                                copy(newpingtopology,selfpingtopology)

						for _, ipline := range addslice { //增加的
							if !CheckIsIn(selfpingslice, ipline) {
								newpingslice = append(newpingslice, ipline)
							}
							seelog.Info(fmt.Sprintf("[func:Monitor_Domain] %s selfpingslice add %s", networkey, ipline))

							ipmap := Backmap(ipline)
							if !CheckIsInMap(&selfpingtopology, ipmap) {
								newpingtopology = append(newpingtopology, ipmap)
							}
							seelog.Info(fmt.Sprintf("[func:Monitor_Domain] %s selfpingtopology add %s", networkey, ipline))

						}
                                                networkMember.Ping = newpingslice
                                                networkMember.Topology = newpingtopology
                                                g.SelfCfg = networkMember
                                                g.Cfg.Network[networkey] = g.SelfCfg
					}
				}
				for _, ipline := range addslice { //增加的
					adddomainipstruct := g.NetworkMember{Name: ipline, Addr: ipline, Smartping: false, Ping: []string{}, Topology: []map[string]string{}}
					if !CheckIsInAllMap(&(g.Cfg), adddomainipstruct) { //返回true说明已经存在
						seelog.Info(fmt.Sprintf("[func:Monitor_Domain] allmap add %s", ipline))
						g.Cfg.Network[ipline] = adddomainipstruct
					}
				}
				g.Ads.Domainipslice[domainstr] = nownetworkipslice

				saveerr := g.SaveConfig()
				if saveerr != nil {
					seelog.Info(fmt.Sprintf("[func:Monitor_Domain] save config create %v", saveerr))
				} else {
					seelog.Info("[func:Monitor_Domain] save finshed!")
				}
			} else {
				seelog.Info("[func:Monitor_Domain] Not have add domain ip")
			}

			//del
			if len(delslice) != 0 {
				seelog.Info(fmt.Sprintf("[func:Monitor_Domain] Have del domain ip %v", delslice))
				for networkey, networkvalue := range g.Cfg.Network {
					networkMember := g.Cfg.Network[networkey]
					if networkvalue.Smartping {
						selfpingslice := g.Cfg.Network[networkey].Ping
						newselfpingslice := Delipfromslice(selfpingslice, delslice)
						networkMember.Ping = newselfpingslice
						seelog.Info(fmt.Sprintf("[func:Monitor_Domain] %s selfpingslice del %v", networkey, delslice))

						selfpingtopology := g.Cfg.Network[networkey].Topology
						newselfpingtopology := DelipfromMap(selfpingtopology, delslice)
						networkMember.Topology = newselfpingtopology
						seelog.Info(fmt.Sprintf("[func:Monitor_Domain] %s selfpingtopology del %v", networkey, delslice))
						g.SelfCfg = networkMember
						g.Cfg.Network[networkey] = g.SelfCfg
					}
				}
				g.Ads.Domainipslice[domainstr] = nownetworkipslice

				for _, deline := range delslice {
					deldomainipstruct := g.NetworkMember{Name: deline, Addr: deline, Smartping: false, Ping: []string{}, Topology: []map[string]string{}}
					if CheckIsInAllMap(&(g.Cfg), deldomainipstruct) { //返回true说明已经存在
						delete(g.Cfg.Network, deline)
						seelog.Info(fmt.Sprintf("[func:Monitor_Domain] allmap del %s", deline))
					}
				}
				saveerr := g.SaveConfig()
				if saveerr != nil {
					seelog.Info(fmt.Sprintf("[func:Monitor_Domain] save config create %v", saveerr))
				} else {
					seelog.Info("[func:Monitor_Domain] save finshed!")
				}
			} else {
				seelog.Info("[func:Monitor_Domain] Not have del domain ip")
			}
		}
	}
}

func Backmap(fline string) map[string]string {
	var tmpmap = make(map[string]string, 6)
	tmpmap["Addr"] = fline
	tmpmap["Name"] = fline
	tmpmap["Thdavgdelay"] = "100"
	tmpmap["Thdchecksec"] = "130"
	tmpmap["Thdloss"] = "30"
	tmpmap["Thdoccnum"] = "2"
	return tmpmap
}

func CheckIsIn(slicename []string, linename string) bool {
	for _, line := range slicename {
		if strings.TrimSpace(line) == strings.TrimSpace(linename) {
			return true
		}
	}
	return false
}

func CheckIsInMap(slicename *[]map[string]string, mname map[string]string) bool {
	for _, mapvalue := range *slicename {
		if reflect.DeepEqual(mapvalue, mname) {
			return true
		}
	}
	return false
}

func CheckIsInAllMap(gconfig *g.Config, mname g.NetworkMember) bool {
	for _, networkvalue := range (*gconfig).Network {
		if reflect.DeepEqual(networkvalue, mname) {
			return true
		}
	}
	return false
}

func Delipfromslice(selfpingslice, delslice []string) []string {
	newselfpingslice := []string{}
	for _, lline := range selfpingslice {
		if !CheckIsIn(delslice, lline) {
			newselfpingslice = append(newselfpingslice, lline)
		}
	}
	return newselfpingslice
}

func DelipfromMap(selfpingtopology []map[string]string, delslice []string) []map[string]string {
	newselfpingtopology := []map[string]string{}
	for _, lline := range selfpingtopology {
		if !CheckIsIn(delslice, lline["Addr"]) {
			newselfpingtopology = append(newselfpingtopology, lline)
		}
	}
	return newselfpingtopology
}

func CompareSlice(nowslice, originslice []string) ([]string, []string) {
	addslice, delslice := []string{}, []string{}

	for _, nowline := range nowslice {
		if !CheckIsIn(originslice, nowline) {
			addslice = append(addslice, nowline)
		}
	}
	for _, originline := range originslice {
		if !CheckIsIn(nowslice, originline) {
			delslice = append(delslice, originline)
		}
	}

	return addslice, delslice
}

func BackString(m map[string]string) (string, string, string, string) {
	var thdavgdelay, thdchecksec, thdloss, thdoccnum string
	for key, value := range m {
		switch key {
		case "Thdavgdelay":
			thdavgdelay = value
		case "Thdchecksec":
			thdchecksec = value
		case "Thdloss":
			thdloss = value
		case "Thdoccnum":
			thdoccnum = value
		}
	}
	return thdavgdelay, thdchecksec, thdloss, thdoccnum
}

func ValidDomain(domainName string) bool {
	domainName = strings.Trim(domainName, " ")
	domainre, _ := regexp.Compile(`[a-zA-Z][-a-zA-Z]{0,62}(\.[a-zA-Z][-a-zA-Z]{0,62})\.com`)
	if domainre.MatchString(domainName) {
		return true
	}
	return false
}

func DnsResolve(domain string, count int) []string {
	if count <= 2 {
		domainipslice := make([]string, 0)
		ns, err := net.LookupHost(domain)
		if err != nil {
			return DnsResolve(domain, count+1)
		}
		tmpdomainslice := strings.Split(domain, ".")
		domainbeforeslice := tmpdomainslice[:len(tmpdomainslice)-3]
		domainbeforestr := strings.Join(domainbeforeslice, ".")
		for _, eveip := range ns {
			if !StringIsInSlice(domainipslice, domainbeforestr+"-"+strings.TrimSpace(eveip)) {
				domainipslice = append(domainipslice, domainbeforestr+"-"+strings.TrimSpace(eveip))
			}
		}
		seelog.Info(fmt.Sprintf("domain %s ",domain), domainipslice)
		return domainipslice
	} else if count > 2 {
		seelog.Info(domain, "domain resolve failed a")
		return []string{}
	}
	return []string{}
}

//该函数的作用是去重
func RemoveIsIntSlice(sourceslice *[]string) []string {
	sort.Strings(*sourceslice)
	left, right := 0, 1
	for right < len(*sourceslice) {
		if (*sourceslice)[left] != (*sourceslice)[right] {
			left++
			(*sourceslice)[left] = (*sourceslice)[right]
		}
		right++
	}
	return (*sourceslice)[:left+1]
}

func Ping() {
	seelog.Info("[func:Ping] ", "starting run Ping")
	var wg sync.WaitGroup
	ipslice := make([]string, 0)
	for _, target := range g.SelfCfg.Ping {
		if !ValidDomain(target) {
			ipslice = append(ipslice, target)
		} else if ValidDomain(target) {
			domainipslice := DnsResolve(target, 0)
			for _, ipline := range domainipslice {
				ipslice = append(ipslice, ipline)
			}
		}
	}
	ipslice = RemoveIsIntSlice(&ipslice)
	seelog.Info(fmt.Sprintf("all ip: %v",ipslice))

	for _, iptarget := range ipslice {
		wg.Add(1)
		go PingTask(iptarget, &wg)
	}
	wg.Wait()
	go StartAlert()
	go MonitorDomain()
	seelog.Info("[func:Ping] ", "Ping Finish")
}

func StringIsInSlice(slicename []string, str string) bool {
        sort.Strings(slicename)
        left,right:=0,len(slicename)-1

        for left <= right {
            middle:=(left+right)/2
            if slicename[middle]==str{
                return true
            }else if slicename[middle]<str{
                left=middle+1
            }else if slicename[middle]>str{
                right=middle-1
            }
        }
        return false
}

func Backslicesum(slicename []float64) float64 {
	var fsum float64
	for _, fvalue := range slicename {
		fsum += fvalue
	}
	return fsum
}

//ping main function
func PingTask(ipvalue string, wg *sync.WaitGroup) {
	defer wg.Done()
	var iptarget string
	if strings.HasPrefix(ipvalue, "10.") {
		iptarget = ipvalue
	} else if !strings.HasPrefix(ipvalue, "10.") {
		iptarget = strings.Split(ipvalue, "-")[1]
	}
	seelog.Info("Start Ping " + ipvalue + "..")
	stat := g.PingSt{}
	stat.MinDelay = -1
	lossPK := 0
	ipaddr, err := net.ResolveIPAddr("ip", iptarget) //判断是否是ip
	if err == nil {
		avedelayslice := []float64{}
		for i := 0; i < 20; i++ {
			starttime := time.Now().UnixNano()
			delay, err := nettools.RunPing(ipaddr, 3*time.Second, 48, i)
			if err == nil {
				stat.AvgDelay = stat.AvgDelay + delay
				avedelayslice = append(avedelayslice, delay)
				if stat.MaxDelay < delay {
					stat.MaxDelay = delay
				}
				if stat.MinDelay == -1 || stat.MinDelay > delay {
					stat.MinDelay = delay
				}
				stat.RevcPk = stat.RevcPk + 1
				seelog.Debug("[func:StartPing IcmpPing] ID:", i, " IP:", ipvalue)
			} else {
				seelog.Debug("[func:StartPing IcmpPing] ID:", i, " IP:", ipvalue, "| err:", err)
				lossPK = lossPK + 1
			}
			stat.SendPk = stat.SendPk + 1
			stat.LossPk = int((float64(lossPK) / float64(stat.SendPk)) * 100)
			duringtime := time.Now().UnixNano() - starttime
			time.Sleep(time.Duration(3000*1000000-duringtime) * time.Nanosecond)
		}
        if len(avedelayslice) > 3{
            sort.Float64s(avedelayslice)
            newavedelayslice := avedelayslice[2 : len(avedelayslice)-3]
            if stat.RevcPk > 0 {
                stat.AvgDelay = Backslicesum(newavedelayslice) / float64(len(newavedelayslice))
            }else{
                stat.AvgDelay = 0.0
            }
        }else{
            if stat.RevcPk > 0 {
                stat.AvgDelay = stat.AvgDelay / float64(stat.RevcPk)
            }else{
                stat.AvgDelay = 0.0
            }
        }
		seelog.Debug("[func:IcmpPing] Finish Addr:", ipvalue, " MaxDelay:", stat.MaxDelay, " MinDelay:", stat.MinDelay, " AvgDelay:", stat.AvgDelay, " Revc:", stat.RevcPk, " LossPK:", stat.LossPk)
	} else {
		stat.AvgDelay = 0.00
		stat.MinDelay = 0.00
		stat.MaxDelay = 0.00
		stat.SendPk = 0
		stat.RevcPk = 0
		stat.LossPk = 100
		seelog.Debug("[func:IcmpPing] Finish Addr:", ipvalue, " Unable to resolve destination host")
	}
	PingStorage(stat, ipvalue)
	seelog.Info("Finish Ping " + ipvalue + "..")
}

//storage ping data
func PingStorage(pingres g.PingSt, Addr string) {
	logtime := time.Now().Format("2006-01-02 15:04")
	seelog.Info("[func:StartPing] ", "(", ")Starting PingStorage ", Addr)
	sql := "INSERT INTO [pinglog] (logtime, target, maxdelay, mindelay, avgdelay, sendpk, revcpk, losspk) values('" + logtime + "','" + Addr + "','" + strconv.FormatFloat(pingres.MaxDelay, 'f', 2, 64) + "','" + strconv.FormatFloat(pingres.MinDelay, 'f', 2, 64) + "','" + strconv.FormatFloat(pingres.AvgDelay, 'f', 2, 64) + "','" + strconv.Itoa(pingres.SendPk) + "','" + strconv.Itoa(pingres.RevcPk) + "','" + strconv.Itoa(pingres.LossPk) + "')"
	seelog.Debug("[func:StartPing] ", sql)
	g.DLock.Lock()
	_, err := g.Db.Exec(sql)
	if err != nil {
		seelog.Error("[func:StartPing] Sql Error ", err, sql)
	}
	g.DLock.Unlock()
	seelog.Info("[func:StartPing] ", "(", logtime, ") Finish PingStorage  ", Addr)
}
