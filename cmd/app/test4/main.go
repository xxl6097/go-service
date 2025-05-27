package main

import (
	"fmt"
	"github.com/xxl6097/go-service/pkg/utils"
)

func main() {
	//file1 := "/Users/uuxia/Desktop/work/code/github/golang/go-service/release1"
	//fmt.Println(len(ukey.GetBuffer()), utils.ResetDirector(file1))
	newVersion := "v0.2.0"
	oldVersion := "v0.1.94"
	v := utils.CompareVersions(newVersion, oldVersion)
	fmt.Println(v)
}
