package ping

import (
	"../g"
	"bytes"
	"encoding/binary"
	"github.com/gy-games-libs/seelog"
	"log"
	"math/rand"
	"net"
	"time"
)

//timeOut ping请求超时时间
const timeOut = 3000

//ICMP 数据包结构体

//CheckSum 校验和计算
func CheckSum(data []byte) uint16 {
	var (
		sum    uint32
		length = len(data)
		index  int
	)
	for length > 1 {
		sum += uint32(data[index])<<8 + uint32(data[index+1])
		index += 2
		length -= 2
	}
	if length > 0 {
		sum += uint32(data[index])
	}
	sum += (sum >> 16)
	return uint16(^sum)
}

//sendICMP 向目的地址发送icmp包
func sendICMP(raddr *net.IPAddr, i int) (float64, error) {
	ts := time.Now().UnixNano()
	//构建发送的ICMP包
	icmp := g.ICMP{
		Type:      8,
		Code:      0,
		Checksum:  0, //默认校验和为0，后面计算再写入
		ID:        uint16(rand.Intn(65535)),
		Seq:       uint16(i),
		Timestamp: ts,
	}
	//新建buffer将包内数据写入，以计算校验和并将校验和并存入icmp结构体中
	var buffer bytes.Buffer
	binary.Write(&buffer, binary.BigEndian, icmp)
	icmp.Checksum = CheckSum(buffer.Bytes())
	buffer.Reset()
	//与目的ip地址建立连接，第二个参数为空则默认为本地ip，第三个参数为目的ip
	con, err := net.DialIP("ip4:icmp", nil, raddr)
	if err != nil {
		log.Fatal(err)
	}
	//函数结束后后关闭连接
	defer con.Close()
	//构建buffer将要发送的数据存入
	var sendBuffer bytes.Buffer
	binary.Write(&sendBuffer, binary.BigEndian, icmp)
	if _, err := con.Write(sendBuffer.Bytes()); err != nil {
		//log.Fatal(err)
		return 0, err
	}
	//设置读取超时时间为2s
	con.SetReadDeadline((time.Now().Add(time.Millisecond * timeOut)))
	//构建接受的比特数组
	for {
		rec := make([]byte, 1024)
		//读取连接返回的数据，将数据放入rec中
		recCnt, err := con.Read(rec)
		if err != nil {
			//fmt.Println("")
			return 0, err
		}
		timeEnd := time.Now().UnixNano()
		if icmp.Seq == uint16(binary.BigEndian.Uint16(rec[(recCnt-10):(recCnt-8)])) && icmp.ID == uint16(binary.BigEndian.Uint16(rec[(recCnt-12):(recCnt-10)])) {
			timeStart := int64(binary.BigEndian.Uint64(rec[(recCnt - 8):recCnt]))
			durationTime := float64((timeEnd - timeStart)) / 1e6
			seelog.Debug("[func:IcmpPing] ID:%d | %d bytes from %s: seq=%d time=%.2fms\n", i, recCnt, raddr.String(), icmp.Seq, durationTime)
			//seelog.Debug("[func:IcmpPing] Finish Addr:", ip, " MaxDelay:", stat.MaxDelay, " MinDelay:", stat.MinDelay, " AvgDelay:", stat.AvgDelay, " Revc:", stat.RevcPk, " LossPK:", stat.LossPk)
			return durationTime, nil
		}
	}
	return 0, nil
}
func IcmpPing(addr string) g.PingSt {
	stat := g.PingSt{}
	var ip, _ = net.ResolveIPAddr("ip", addr)
	if ip == nil {
		seelog.Error("[func:IcmpPing] Finish Addr:", ip, " Domain or Ip not valid!")
		return stat
	}
	stat.MinDelay = -1
	for i := 0; i < 5; i++ {
		delay, err := sendICMP(ip, i)
		if err == nil {
			stat.AvgDelay = stat.AvgDelay + delay
			if stat.MaxDelay < delay {
				stat.MaxDelay = delay
			}
			if stat.MinDelay == -1 || stat.MinDelay > delay {
				stat.MinDelay = delay
			}
			stat.RevcPk = stat.RevcPk + 1
		} else {
			log.Print(err)
			stat.LossPk = stat.LossPk + 1
		}
		stat.SendPk = stat.SendPk + 1
		//每间隔3s ping一次
		time.Sleep(3000 * time.Millisecond)
	}
	stat.AvgDelay = stat.AvgDelay / float64(stat.SendPk)
	seelog.Info("[func:IcmpPing] Finish Addr:", ip, " MaxDelay:", stat.MaxDelay, " MinDelay:", stat.MinDelay, " AvgDelay:", stat.AvgDelay, " Revc:", stat.RevcPk, " LossPK:", stat.LossPk)
	return stat
}
