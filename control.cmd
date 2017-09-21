@echo off
title SmartPing

setlocal
set uac=~uac_permission_tmp_%random%
md "%SystemRoot%\system32\%uac%" 2>nul
if %errorlevel%==0 ( rd "%SystemRoot%\system32\%uac%" >nul 2>nul ) else (
    echo set uac = CreateObject^("Shell.Application"^)>"%temp%\%uac%.vbs"
    echo uac.ShellExecute "%~s0","","","runas",1 >>"%temp%\%uac%.vbs"
    echo WScript.Quit >>"%temp%\%uac%.vbs"
    "%temp%\%uac%.vbs" /f
    del /f /q "%temp%\%uac%.vbs" & exit )
endlocal  

:BG
echo ©°©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©´
echo ©¦                               SmartPing                              ©¦
echo ©À©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©È
echo ©¦                                                                      ©¦
echo ©¦Instruction USE                                                       ©¦
echo ©¦        build   run go get and build                                  ©¦
echo ©¦        run     run smartping                                         ©¦
echo ©¦        install install smartping as service (use nssm)               ©¦
echo ©¦        start   start smartping service                               ©¦
echo ©¦        stop    stop smartping service                                ©¦
echo ©¦        restart stop and start smartping                              ©¦
echo ©¦                                                                      ©¦
echo ©¸©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¼

%~d0
cd %~dp0
::SET select=
SET /P select="Please Enter Instructions:"
IF "%select%"=="build" (
    go get -v ./...
    go build -o %~dp0\bin\smartping.exe  %~dp0\src\smartping.go
    echo Build Finish.. 
) ELSE (
    IF "%select%"=="run" (
        %~dp0/bin/smartping.exe 
    ) ELSE ( 
        IF "%select%"=="install" (
            %~dp0\\bin\\nssm.exe install smartping %~dp0\\bin\\smartping.exe 
        ) ELSE ( 
            IF "%select%"=="start" (
                net start smartping 
            ) ELSE (
                IF "%select%"=="stop" (
                    net stop smartping 
                ) ELSE (
                    IF "%select%"=="restart" (
                        net stop smartping 
                        net start smartping 
                    ) ELSE (
                        IF "%select%"=="uninstall" (
                            sc delete smartping 
                        ) ELSE (
                             echo Param Error Try Again!
                             pause
                             GOTO BG
                        ) 
                    ) 
                ) 
            ) 
        ) 
    )
)

pause

exit