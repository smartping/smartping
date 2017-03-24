rm -rf idcping
go build idcping
ps -aux | grep idcping | grep -v grep | awk '{print $2}' | xargs kill -9
./idcping
