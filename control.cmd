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
cls
echo ©°©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©´
echo ©¦                        SmartPing                            ©¦
echo ©À©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©È
echo ©¦                                                             ©¦
echo ©¦INS USE                                                      ©¦
echo ©¦        build      run go get and build                      ©¦
echo ©¦        run        run smartping                             ©¦
echo ©¦        install    install smartping as service (use nssm)   ©¦
echo ©¦        uninstall  uninstall smartping service               ©¦
echo ©¦        start      start smartping service (after install)   ©¦
echo ©¦        stop       stop smartping service                    ©¦
echo ©¦        restart    stop and start smartping                  ©¦
echo ©¦        version    show smartping version                    ©¦
echo ©¦                                                             ©¦
echo ©¸©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¤©¼

%~d0
cd %~dp0
::SET select=
SET /P select="Please Enter Instructions:"
IF "%select%"=="build" (
    go get -v ./...
    ::go build -o %~dp0\bin\goping.exe  %~dp0\src\ping\goping\goping.go
    go build -o %~dp0\bin\smartping.exe  %~dp0\src\smartping.go
    echo Build Finish.. 
    pause
    GOTO BG
) ELSE (
    IF "%select%"=="run" (
        %~dp0/bin/smartping.exe 
    ) ELSE ( 
        IF "%select%"=="install" (
            %~dp0\\bin\\nssm.exe install smartping %~dp0\\bin\\smartping.exe 
            pause
            GOTO BG
        ) ELSE ( 
            IF "%select%"=="start" (
                net start smartping 
                pause
                GOTO BG
            ) ELSE (
                IF "%select%"=="stop" (
                    net stop smartping 
                    pause
                    GOTO BG
                ) ELSE (
                    IF "%select%"=="restart" (
                        net stop smartping 
                        net start smartping 
                        pause
                        GOTO BG
                    ) ELSE (
                        IF "%select%"=="uninstall" (
                            sc delete smartping 
                            pause
                            GOTO BG
                        ) ELSE (
                             IF "%select%"=="version" (
                                %~dp0\bin\smartping.exe -v 
                                pause
                                GOTO BG
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
)

pause

exit