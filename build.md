```
docker run --rm -it \
  -v /Users/shenshouer/workspaces/ipalfish/smartping:/data/smartping \
  -w /data/smartping/bin \
  -e GOPROXY="http://goproxy.pri.ibanyu.com,direct" \
  -e GONOPROXY="code.ibanyu.com,gitlab.pri.ibanyu.com" \
  -e GONOSUMDB="code.ibanyu.com,gitlab.pri.ibanyu.com" \clear
  -e GOOS=linux \
  -e CGO_ENABLED=1 \
  hub.pri.ibanyu.com/devops/golang:1.13.0 \
    go build -mod=vendor -ldflags "-linkmode external -extldflags -static" -o smartping-linux ../src/smartping.go
```
