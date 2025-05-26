package main

import (
	"fmt"
	"github.com/xxl6097/go-service/cmd/differ/core"
	"os"
)

func main() {
	if len(os.Args) < 4 {
		_, _ = fmt.Fprintln(os.Stderr, "Usage: differ <oldDir> <newDir> <version>")
		return
	}
	fmt.Println("Args", os.Args)
	err := core.Diff(os.Args[1], os.Args[2], os.Args[3])
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
}
