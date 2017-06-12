@ECHO OFF
IF "%1%"=="" GOTO FAIL
IF "%1%"=="start" GOTO START
IF "%1%"=="build" GOTO BUILD

:BUILD
go build -o %~dp0/bin/smartping.exe  %~dp0/src/smartping.go
GOTO EXIT

:START
cd %~dp0
%~dp0/bin/smartping.exe
GOTO EXIT


:FAIL
echo "build|start"

:EXIT