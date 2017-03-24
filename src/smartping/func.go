package main

import (
	"strings"
	"strconv"
)

func NewState() *State {
	s := new(State)
	s.State = make(map[*Target]TargetStatus)
	return s
}

func compare(num string,nb int) bool{
	val,_:=strconv.Atoi(num)
	if val < nb{
		return false
	}
	return true
}

func timestr(time string) string {
	return strings.Fields(time)[1]
}