package kv

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// GetHostInfo 获取主机信息
func GetHostInfo() (string, string, error) {
	name, err := os.Hostname()
	if err != nil {
		panic(fmt.Sprintf("Cannot get hostname: %v", err))
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		panic(fmt.Sprintf("Cannot get net interfaces: %v", err))
	}
	for _, i := range ifaces {
		if strings.Contains(i.Name, "eth") || strings.Contains(i.Name, "en") || strings.Contains(i.Name, "bond") {
			if addrs, err := i.Addrs(); err == nil {
				for _, addr := range addrs {
					var ip net.IP
					switch v := addr.(type) {
					case *net.IPNet:
						ip = v.IP
					case *net.IPAddr:
						ip = v.IP
					}
					if ip.To4() != nil {
						return name, ip.String(), nil
					}
				}
			}
		}
	}
	return "", "", fmt.Errorf("No IP info")
}
