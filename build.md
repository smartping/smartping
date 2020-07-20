```
# 注意修改 -v的目录为本地源代码目录
docker run --rm -it \
  -v /Users/shenshouer/workspaces/ipalfish/smartping:/data/smartping \
  -w /data/smartping/bin \
  -e GOPROXY="https://goproxy.cn,direct" \
  -e GOOS=linux \
  -e CGO_ENABLED=1 \
  golang:1.13.0 \
    go build -mod=vendor -ldflags "-linkmode external -extldflags -static" -o smartping-linux ../src/smartping.go
```
