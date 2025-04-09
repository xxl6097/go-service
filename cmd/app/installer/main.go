package main

import (
	"fmt"
	"github.com/xxl6097/go-service/cmd/app/installer/core"
	"github.com/xxl6097/go-service/gservice"
	"github.com/xxl6097/go-service/pkg"
	"os"
)

func main() {
	fmt.Println(os.Args)
	_ = gservice.Run(&core.SvrInstall{})
	pkg.Version()
}
