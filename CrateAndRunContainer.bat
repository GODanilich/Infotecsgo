@echo off
chcp 1251 >nul
echo ������ Docker-������
docker build -t infotecs-payment-system .

if %ERRORLEVEL% NEQ 0 (
    echo ������ ��� ������ Docker-������!
    pause
    exit /b %ERRORLEVEL%
)

echo ������ ����������
docker run -p 8080:8080 -v "%CD%/wallet.db:/app/wallet.db" infotecs-payment-system

if %ERRORLEVEL% NEQ 0 (
    echo ������ ��� ������� ����������!
    pause
    exit /b %ERRORLEVEL%
)

echo ��������� ������� �������.
pause