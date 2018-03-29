package nettools

import (
	"bytes"
	"encoding/binary"
	"errors"
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

func SendICMP(raddr *net.IPAddr, i int, timeout int) (float64, error) {
	timsStart := time.Now().UnixNano()
	icmp := ICMP{
		Type:     8,
		Code:     0,
		Checksum: 0,
		ID:       uint16(rand.Intn(65535)),
		Seq:      uint16(i),
		//Timestamp: ts,
	}
	var buffer bytes.Buffer
	binary.Write(&buffer, binary.BigEndian, icmp)
	icmp.Checksum = CheckSum(buffer.Bytes())
	buffer.Reset()
	con, err := net.DialIP("ip4:icmp", nil, raddr)
	con.SetReadDeadline((time.Now().Add(time.Second * time.Duration(timeout))))
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
		//ID=ID Seq=Seq
		if icmp.Seq == uint16(binary.BigEndian.Uint16(rec[(recCnt-10):(recCnt-8)])) && icmp.ID == uint16(binary.BigEndian.Uint16(rec[(recCnt-12):(recCnt-10)])) {
			//timeStart := int64(binary.BigEndian.Uint64(rec[(recCnt - 8):recCnt]))
			durationTime := float64((timeEnd - timsStart)) / 1e6
			seelog.Debug("[func:IcmpPing] ID:", i, " | ", recCnt, " bytes from ", raddr.String(), ": seq=", icmp.Seq, " time=", durationTime, "ms")
			return durationTime, nil
		}
	}
	return 0, errors.New(raddr.String() + " ICMP Timeout!")
}
