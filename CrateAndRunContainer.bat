@echo off
chcp 1251 >nul
echo Сборка Docker-образа
docker build -t infotecs-payment-system .

if %ERRORLEVEL% NEQ 0 (
    echo Ошибка при сборке Docker-образа!
    pause
    exit /b %ERRORLEVEL%
)

echo Запуск контейнера
docker run -p 8080:8080 -v "%CD%/wallet.db:/app/wallet.db" infotecs-payment-system

if %ERRORLEVEL% NEQ 0 (
    echo Ошибка при запуске контейнера!
    pause
    exit /b %ERRORLEVEL%
)

echo Контейнер успешно запущен.
pause