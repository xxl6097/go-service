go get github.com/kardianos/service

go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest

CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "-linkmode internal" -o AAServiceApp.exe main.go

CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o AAATest1.exe main.go