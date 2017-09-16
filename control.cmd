@ECHO OFF
IF "%1%"=="" GOTO FAIL
IF "%1%"=="run" GOTO RUN
IF "%1%"=="build" GOTO BUILD
GOTO FAIL
:BUILD
go get ./...
go build -o %~dp0/bin/smartping.exe  %~dp0/src/smartping.go
GOTO EXIT

:RUN
cd %~dp0
%~dp0/bin/smartping.exe
GOTO EXIT


:FAIL
echo "build|run"

:EXIT