package main

import (
	"fmt"
	"github.com/xxl6097/go-service/pkg/ukey"
	"github.com/xxl6097/go-service/pkg/utils"
)

func main() {
	file1 := "/Users/uuxia/Desktop/work/code/github/golang/go-service/release1"
	fmt.Println(len(ukey.GetBuffer()), utils.ResetDirector(file1))

}
