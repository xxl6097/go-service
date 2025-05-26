package main

import (
	"fmt"
	"github.com/xxl6097/go-service/pkg/utils"
)

func main() {
	file1 := "/Users/uuxia/Desktop/work/code/github/golang/go-service/release1"
	fmt.Println(utils.ResetDirector(file1))
}
