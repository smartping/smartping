package mn

import (
	"testing"
	"log"
)

func TestTest10Base(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	n := len(nums)
	indexs := zuheResult(n, 2)
	result := findNumsByIndexs(nums, indexs)

	log.Println(result)

}
