package funcs

import (
	"../g"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func ValidIP4(ipAddress string) bool {
	ipAddress = strings.Trim(ipAddress, " ")
	re, _ := regexp.Compile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)
	if re.MatchString(ipAddress) {
		return true
	}
	return false
}

func GetRoot() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	dirctory := strings.Replace(dir, "\\", "/", -1)
	runes := []rune(dirctory)
	l := 0 + strings.LastIndex(dirctory, "/")
	if l > len(runes) {
		l = len(runes)
	}
	return string(runes[0:l])
}

func CurrentTimeMillis() int64 {
	return time.Now().UnixNano() / 1000000
}

func NewState() *g.State {
	s := new(g.State)
	s.State = make(map[*g.Target]g.TargetStatus)
	return s
}

func Compare(num string, nb int) bool {
	val, _ := strconv.Atoi(num)
	if val < nb {
		return false
	}
	return true
}

func Timestr(time string) string {
	return strings.Fields(time)[1]
}
