package ping

import (
	//"../g"
	"bytes"
	"encoding/binary"
	"github.com/gy-games-libs/seelog"
	"math/rand"
	"net"
	"time"
)

type ICMP struct {
	Type      uint8
	Code      uint8
	Checksum  uint16
	ID        uint16
	Seq       uint16
	Timestamp int64
}

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

func SendICMP(raddr *net.IPAddr, i int) (float64, error) {
	ts := time.Now().UnixNano()
	icmp := ICMP{
		Type:      8,
		Code:      0,
		Checksum:  0,
		ID:        uint16(rand.Intn(65535)),
		Seq:       uint16(i),
		Timestamp: ts,
	}
	var buffer bytes.Buffer
	binary.Write(&buffer, binary.BigEndian, icmp)
	icmp.Checksum = CheckSum(buffer.Bytes())
	buffer.Reset()
	con, err := net.DialIP("ip4:icmp", nil, raddr)
	con.SetReadDeadline((time.Now().Add(time.Millisecond * 40)))
	if err != nil {
		return 0, err
	}
	defer con.Close()
	var sendBuffer bytes.Buffer
	binary.Write(&sendBuffer, binary.BigEndian, icmp)
	if _, err := con.Write(sendBuffer.Bytes()); err != nil {
		return 0, err
	}
	for {
		rec := make([]byte, 1024)
		recCnt, err := con.Read(rec)
		if err != nil {
			return 0, err
		}
		timeEnd := time.Now().UnixNano()
		if icmp.Seq == uint16(binary.BigEndian.Uint16(rec[(recCnt-10):(recCnt-8)])) && icmp.ID == uint16(binary.BigEndian.Uint16(rec[(recCnt-12):(recCnt-10)])) {
			timeStart := int64(binary.BigEndian.Uint64(rec[(recCnt - 8):recCnt]))
			durationTime := float64((timeEnd - timeStart)) / 1e6
			seelog.Debug("[func:IcmpPing] ID:", i, " | ", recCnt, " bytes from ", raddr.String(), ": seq=", icmp.Seq, " time=", durationTime, "ms")
			return durationTime, nil
		}
	}
	return 0, nil
}

/*
func IcmpPing(addr string) g.PingSt {
	stat := g.PingSt{}
	var ip, _ = net.ResolveIPAddr("ip", addr)
	if ip == nil {
		seelog.Error("[func:IcmpPing] Finish Addr:", ip, " Domain or Ip not valid!")
		return stat
	}
	stat.MinDelay = -1
	lossPK := 0
	for i := 0; i < 20; i++ {
		starttime := time.Now().UnixNano()
		delay, err := SendICMP(ip, i)
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
			seelog.Debug("[func:IcmpPing] ID:", i, " | ", err)
			//stat.LossPk = stat.LossPk + 1
			lossPK = lossPK + 1
		}
		stat.SendPk = stat.SendPk + 1
		stat.LossPk = int((float64(lossPK)/float64(stat.SendPk)) * 100 )
		duringtime := time.Now().UnixNano()-starttime
		time.Sleep(time.Duration(3000*1000000-duringtime) * time.Nanosecond)
	}
	stat.AvgDelay = stat.AvgDelay / float64(stat.SendPk)
	seelog.Debug("[func:IcmpPing] Finish Addr:", ip, " MaxDelay:", stat.MaxDelay, " MinDelay:", stat.MinDelay, " AvgDelay:", stat.AvgDelay, " Revc:", stat.RevcPk, " LossPK:", stat.LossPk)
	return stat
}
*/
