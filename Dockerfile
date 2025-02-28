# Базовый образ для сборки
FROM golang:1.23-alpine AS builder

# Установка зависимостей для CGO и SQLite
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Копирование go.mod и go.sum для загрузки зависимостей
COPY go.mod go.sum ./

# Загрузка зависимостей
RUN go mod download

# Копирование исходника
COPY main.go ./

# Сборка
RUN GOOS=linux go build -o infotecs-payment-system main.go

# Финальный образ
FROM alpine:3.18

# Установка SQLite и зависимостей для работы приложения
RUN apk add --no-cache sqlite-libs

WORKDIR /app

# Копирование скомпилированного бинарника из builder
COPY --from=builder /app/infotecs-payment-system .

# Установка прав на выполнение
RUN chmod +x ./infotecs-payment-system

# Открытие порта 8080
EXPOSE 8080

CMD ["./infotecs-payment-system"]