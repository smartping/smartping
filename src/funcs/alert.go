package funcs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cihub/seelog"
	_ "github.com/mattn/go-sqlite3"
	"github.com/smartping/smartping/src/g"
	"github.com/smartping/smartping/src/nettools"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

func StartAlert() {
	seelog.Info("[func:StartAlert] ", "starting run AlertCheck ")
	for _, v := range g.SelfCfg.Topology {
		if v["Addr"] != g.SelfCfg.Addr {
			sFlag := CheckAlertStatus(v)
			//if sFlag {
			//	g.AlertStatus[v["Addr"]] = true
			//}
			//_, haskey := g.AlertStatus[v["Addr"]]
			//if (!haskey && !sFlag) || (!sFlag && g.AlertStatus[v["Addr"]]) {
			seelog.Info(v, sFlag)
			if !sFlag {
				seelog.Debug("[func:StartAlert] ", v["Addr"]+" Alert!")
				g.AlertStatus[v["Addr"]] = false
				l := g.AlertLog{}
				l.Fromname = g.SelfCfg.Name
				l.Fromip = g.SelfCfg.Addr
				l.Logtime = time.Unix(time.Now().Unix(), 0).Format("2006-01-02 15:04")
				l.Targetname = v["Name"]
				l.Targetip = v["Addr"]
				mtrString := ""
				hops, err := nettools.RunMtr(v["Addr"], time.Second, 64, 6)
				if nil != err {
					seelog.Error("[func:StartAlert] Traceroute error ", err)
					mtrString = err.Error()
				} else {
					jHops, err := json.Marshal(hops)
					if err != nil {
						mtrString = err.Error()
					} else {
						mtrString = string(jHops)
					}
				}
				l.Tracert = mtrString
				go AlertStorage(l)
				go SendAlarm(v["Addr"])
				//if g.Cfg.Alert["SendEmailAccount"] != "" && g.Cfg.Alert["SendEmailPassword"] != "" && g.Cfg.Alert["EmailHost"] != "" && g.Cfg.Alert["RevcEmailList"] != "" {
				//	go AlertSendMail(l)
				//}
			}

		}
	}
	seelog.Info("[func:StartAlert] ", "AlertCheck finish ")
}

func CheckAlertStatus(v map[string]string) bool {
	type Cnt struct {
		Cnt int
	}
	Thdchecksec, _ := strconv.Atoi(v["Thdchecksec"])
	timeStartStr := time.Unix((time.Now().Unix() - int64(Thdchecksec)), 0).Format("2006-01-02 15:04")
	querysql := "SELECT count(1) cnt FROM  `pinglog` where logtime > '" + timeStartStr + "' and target = '" + v["Addr"] + "' and (cast(avgdelay as double) > " + v["Thdavgdelay"] + " or cast(losspk as double) > " + v["Thdloss"] + ") "
	rows, err := g.Db.Query(querysql)
	defer rows.Close()
	seelog.Debug("[func:StartAlert] ", querysql)
	if err != nil {
		seelog.Error("[func:StartAlert] Query Error ", err)
		return false
	}
	for rows.Next() {
		l := new(Cnt)
		err := rows.Scan(&l.Cnt)
		if err != nil {
			seelog.Error("[func:StartAlert]", err)
			return false
		}
		Thdoccnum, _ := strconv.Atoi(v["Thdoccnum"])
		seelog.Info(*l, Thdoccnum)
		if l.Cnt < Thdoccnum {
			return true
		} else {
			return false
		}
	}
	return false
}

func AlertStorage(t g.AlertLog) {
	seelog.Info("[func:AlertStorage] ", "(", t.Logtime, ")Starting AlertStorage ", t.Targetname)
	sql := "INSERT INTO [alertlog] (logtime, targetip, targetname, tracert) values('" + t.Logtime + "','" + t.Targetip + "','" + t.Targetname + "','" + t.Tracert + "')"
	seelog.Info(sql)
	g.DLock.Lock()
	g.Db.Exec(sql)
	_, err := g.Db.Exec(sql)
	if err != nil {
		seelog.Error("[func:StartPing] Sql Error ", err)
	}
	g.DLock.Unlock()
	seelog.Info("[func:AlertStorage] ", "(", t.Logtime, ") AlertStorage on ", t.Targetname, " finish!")
}

func AlertSendMail(t g.AlertLog) {
	hops := []nettools.Mtr{}
	err := json.Unmarshal([]byte(t.Tracert), &hops)
	if err != nil {
		seelog.Error("[func:AlertSendMail] json Error ", err)
		return
	}
	mtrstr := bytes.NewBufferString("")
	fmt.Fprintf(mtrstr, "<table>")
	fmt.Fprintf(mtrstr, "<tr><td>Host</td><td>Loss</td><td>Snt</td><td>Last</td><td>Avg</td><td>Best</td><td>Wrst</td><td>StDev</td></tr>")
	for i, hop := range hops {
		fmt.Fprintf(mtrstr, "<tr><td>%d %s</td><td>%.2f</td><td>%d</td><td>%v</td><td>%v</td><td>%v</td><td>%v</td><td>%.2f</td></tr>", i+1, hop.Host, ((float64(hop.Loss) / float64(hop.Send)) * 100), hop.Send, hop.Last, hop.Avg, hop.Best, hop.Wrst, hop.StDev)
	}
	fmt.Fprintf(mtrstr, "</table>")
	title := "【" + t.Fromname + "->" + t.Targetname + "】网络异常报警（" + t.Logtime + "）- SmartPing"
	content := "报警时间：" + t.Logtime + " <br> 来路：" + t.Fromname + "(" + t.Fromip + ") <br>  目的：" + t.Targetname + "(" + t.Targetip + ") <br> "
	SendEmailAccount := g.Cfg.Alert["SendEmailAccount"]
	SendEmailPassword := g.Cfg.Alert["SendEmailPassword"]
	EmailHost := g.Cfg.Alert["EmailHost"]
	RevcEmailList := g.Cfg.Alert["RevcEmailList"]
	err = SendMail(SendEmailAccount, SendEmailPassword, EmailHost, RevcEmailList, title, content+mtrstr.String())
	if err != nil {
		seelog.Error("[func:AlertSendMail] SendMail Error ", err)
	}
}

func SendMail(user, pwd, host, to, subject, body string) error {
	if len(strings.Split(host, ":")) == 1 {
		host = host + ":25"
	}
	auth := smtp.PlainAuth("", user, pwd, strings.Split(host, ":")[0])
	content_type := "Content-Type: text/html" + "; charset=UTF-8"
	msg := []byte("To: " + to + "\r\nFrom: " + user + "\r\nSubject: " + subject + "\r\n" + content_type + "\r\n\r\n" + body)
	send_to := strings.Split(to, ";")
	err := smtp.SendMail(host, auth, user, send_to, msg)
	if err != nil {
		return err
	}
	return nil
}

func SendAlarm(ipvalue string) {
	revicestring, ok := g.Cfg.Alert["RevcEmailList"]
	if !ok {
		revicestring = "zhiping7"
	}
	ret := SendAlert(fmt.Sprintf("smartping-(%s)", ipvalue), fmt.Sprintf("smartping-(%s)", ipvalue), revicestring)
	seelog.Info(ret)
}
