package main

import (
	"fmt"
	"github.com/xxl6097/go-service/cmd/app/installer/core"
	"github.com/xxl6097/go-service/pkg"
	"github.com/xxl6097/go-service/pkg/gs"
	"os"
)

func main() {
	fmt.Println(os.Args)
	_ = gs.Run(&core.SvrInstall{})
	pkg.Version()
}
