#!/bin/bash

# Скрипт для запуска бота с переменными окружения

# Проверяем наличие .env файла
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Проверяем наличие обязательных переменных
if [ -z "$BOT_TOKEN" ]; then
    echo "Ошибка: Переменная BOT_TOKEN не установлена!"
    echo "Создайте файл .env или экспортируйте переменную BOT_TOKEN"
    exit 1
fi

if [ -z "$ADMIN_ID" ]; then
    echo "Ошибка: Переменная ADMIN_ID не установлена!"
    echo "Создайте файл .env или экспортируйте переменную ADMIN_ID"
    exit 1
fi

echo "Запуск бота..."

# Проверяем наличие скомпилированного бинарника
if [ -f "./whitelist-bot" ]; then
    echo "Используется скомпилированный бинарник..."
    ./whitelist-bot
else
    echo "Запуск через go run..."
    go run .
fi

