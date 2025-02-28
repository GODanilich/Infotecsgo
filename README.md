# Infotecsgo
## Запуск

Можно запустить напрямую: go run ./main.go

### Для автоматизированной сборки и запуска Docker контейнера под Windows
Запустить CrateAndRunContainer.bat

### Для автоматизированной сборки и запуска Docker контейнера под Linux 
Выполнить команду chmod +x CreateAndRunContainer.sh и запустить CrateAndRunContainer.sh 

## Использование

### Просмотр последних N транзакций:
http://localhost:8080/api/transactions?count=N

### Просмотр баланса на кошельке:
http://localhost:8080/api/wallet/address/balance

Вместо address нужно использовать address кошелька из БД

#### Для прикрепленной БД актуальным является:
http://localhost:8080/api/wallet/62156e28841b1738e53a66e44e2c2e62164de66971ae79308d53c51ea9c0e3a4/balance

### Транзакции

curl -X POST http://localhost:8080/api/send -H "Content-Type: application/json" -d '{"from": "address1", "to": "address2", "amount": M}'

Перевод M средств с address1 на address2

#### SH
curl -X POST http://localhost:8080/api/send -H "Content-Type: application/json" -d '{"from": "c63cfd3ba7bfa7b3eb1738c0a08f855c076c12ab661438563cd9eb810874ff20", "to": "62156e28841b1738e53a66e44e2c2e62164de66971ae79308d53c51ea9c0e3a4", "amount": 3.50}'

#### Windows CMD
Необходимо экранирование внутренних ", так как cmd не поддерживает ', ' заменить на "

## Состояние БД
Состояние БД можно посмотреть через SQLite DB Browser
