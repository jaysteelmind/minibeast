@echo off
REM MiniBeast USB Agent - Windows Launcher
REM Double-click this file to run MiniBeast

echo ========================================
echo    MiniBeast USB Agent - Starting
echo ========================================
echo.

REM Get the drive letter of the USB stick
set USB_ROOT=%~dp0..
cd /d "%USB_ROOT%"

REM Run MiniBeast
echo Running system analysis...
bin\minibeast-win.exe --root "%USB_ROOT%" --config "%USB_ROOT%\config\default.yaml"

REM Show results
echo.
echo ========================================
echo    Analysis Complete!
echo ========================================
echo.
echo Results saved to: %USB_ROOT%\out\
echo.

REM Open output directory
start "" "%USB_ROOT%\out"

echo.
echo Press any key to exit...
pause >nul
