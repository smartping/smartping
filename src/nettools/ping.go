package nettools

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"math/rand"
	"net"
	"time"
)

type pkg struct {
	conn     net.PacketConn
	ipv4conn *ipv4.PacketConn
	msg      icmp.Message
	netmsg   []byte
	id       int
	seq      int
	maxrtt   time.Duration
	dest     net.Addr
}

type ICMP struct {
	Addr    net.Addr
	RTT     time.Duration
	MaxRTT  time.Duration
	MinRTT  time.Duration
	AvgRTT  time.Duration
	Final   bool
	Timeout bool
	Down    bool
	Error   error
}

func (t *pkg) Send(ttl int) ICMP {
	var hop ICMP
	var err error
	t.conn, hop.Error = net.ListenPacket("ip4:icmp", "0.0.0.0")
	if nil != err {
		return hop
	}
	defer t.conn.Close()
	t.ipv4conn = ipv4.NewPacketConn(t.conn)
	defer t.ipv4conn.Close()
	hop.Error = t.conn.SetReadDeadline(time.Now().Add(t.maxrtt))
	if nil != hop.Error {
		return hop
	}
	if nil != t.ipv4conn {
		hop.Error = t.ipv4conn.SetTTL(ttl)
	}
	if nil != hop.Error {
		return hop
	}
	sendOn := time.Now()
	if nil != t.ipv4conn {
		_, hop.Error = t.conn.WriteTo(t.netmsg, t.dest)
	}
	if nil != hop.Error {
		return hop
	}
	buf := make([]byte, 1500)
	for {
		var readLen int
		readLen, hop.Addr, hop.Error = t.conn.ReadFrom(buf)
		if nerr, ok := hop.Error.(net.Error); ok && nerr.Timeout() {
			hop.Timeout = true
			return hop
		}
		if nil != hop.Error {
			return hop
		}
		var result *icmp.Message
		if nil != t.ipv4conn {
			result, hop.Error = icmp.ParseMessage(1, buf[:readLen])
		}
		if nil != hop.Error {
			return hop
		}
		switch result.Type {
		case ipv4.ICMPTypeEchoReply:
			if rply, ok := result.Body.(*icmp.Echo); ok {
				if t.id == rply.ID && t.seq == rply.Seq {
					hop.Final = true
					hop.RTT = time.Since(sendOn)
					return hop
				}

			}
		case ipv4.ICMPTypeTimeExceeded:
			if rply, ok := result.Body.(*icmp.TimeExceeded); ok {
				if len(rply.Data) > 24 {
					if uint16(t.id) == binary.BigEndian.Uint16(rply.Data[24:26]) {
						hop.RTT = time.Since(sendOn)
						return hop
					}
				}
			}
		case ipv4.ICMPTypeDestinationUnreachable:
			if rply, ok := result.Body.(*icmp.Echo); ok {
				if t.id == rply.ID && t.seq == rply.Seq {
					hop.Down = true
					hop.RTT = time.Since(sendOn)
					return hop
				}

			}
		}
	}
}

func RunPing(IpAddr *net.IPAddr, maxrtt time.Duration, maxttl int, seq int) (float64, error) {
	var res pkg
	var err error
	res.dest = IpAddr
	res.maxrtt = maxrtt
	res.id = rand.Intn(65535)
	res.seq = seq
	res.msg = icmp.Message{Type: ipv4.ICMPTypeEcho, Code: 0, Body: &icmp.Echo{ID: res.id, Seq: res.seq}}
	res.netmsg, err = res.msg.Marshal(nil)
	if nil != err {
		return 0, err
	}
	pingRsult := res.Send(maxttl)
	return float64(pingRsult.RTT.Nanoseconds()) / 1e6, pingRsult.Error
}

func CheckSum(data []byte) (rt uint16) {
	var (
		sum    uint32
		length int = len(data)
		index  int
	)
	for length > 1 {
		sum += uint32(data[index])<<8 + uint32(data[index+1])
		index += 2
		length -= 2
	}
	if length > 0 {
		sum += uint32(data[index]) << 8
	}
	rt = uint16(sum) + uint16(sum>>16)

	return ^rt
}

type SICMP struct {
	Type        uint8
	Code        uint8
	Checksum    uint16
	Identifier  uint16
	SequenceNum uint16
}

var (
	originBytes []byte
)

func ShellRunPing(raddr *net.IPAddr, maxrtt time.Duration, PS int, seq uint16) (float64, error) {
	var (
		sicmp SICMP
		laddr = net.IPAddr{IP: net.ParseIP("0.0.0.0")} // 得到本机的IP地址结构
	)

	// 返回一个 ip socket
	conn, err := net.DialIP("ip4:icmp", &laddr, raddr)

	if err != nil {
		fmt.Println(err.Error())
		return 0.0, err
	}

	defer conn.Close()

	// 初始化 icmp 报文
	sicmp = SICMP{8, 0, 0, 0, seq}

	var buffer bytes.Buffer
        fmt.Println(raddr,originBytes)
	binary.Write(&buffer, binary.BigEndian, sicmp)
	binary.Write(&buffer, binary.BigEndian, originBytes[0:PS])
	b := buffer.Bytes()
	binary.BigEndian.PutUint16(b[2:], CheckSum(b))

	recv := make([]byte, 1024)

	if _, err := conn.Write(buffer.Bytes()); err != nil {
		return 0.0, err
	}
	// 否则记录当前得时间
	t_start := time.Now()
	conn.SetReadDeadline((time.Now().Add(maxrtt)))
	_, err = conn.Read(recv)
	if err != nil {
		return 0.0, err
	}
	t_end := time.Now()
	dur := float64(t_end.Sub(t_start).Nanoseconds()) / 1e6
	return dur, nil
}
