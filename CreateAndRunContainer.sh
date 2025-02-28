#!/bin/bash

echo "Сборка Docker-образа"
sudo docker build -t infotecs-payment-system .

if [ $? -ne 0 ]; then
    echo "Ошибка при сборке Docker-образа!" >&2
    exit 1
fi

echo "Запуск контейнера"
sudo docker run -p 8080:8080 -v "$(pwd)/wallet.db:/app/wallet.db" infotecs-payment-system

if [ $? -ne 0 ]; then
    echo "Ошибка при запуске контейнера!" >&2
    exit 1
fi

echo "Контейнер успешно запущен."