package main

import (
	"fmt"
	"github.com/xxl6097/go-service/pkg/utils"
)

func main() {
	//file1 := "/Users/uuxia/Desktop/work/code/github/golang/go-service/release1"
	//fmt.Println(len(ukey.GetBuffer()), utils.ResetDirector(file1))
	newVersion := "v0.4.19"
	oldVersion := "v0.4.18"
	v := utils.CompareVersions(newVersion, oldVersion)
	fmt.Println(v)

	//a := strings.ReplaceAll(newVersion, "v", "")
	//a = strings.ReplaceAll(a, ".", "")
	//num64, err := strconv.ParseInt(a, 10, 64)
	//
	//fmt.Println(num64, err)
}
