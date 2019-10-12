# Сервер аукциона

Сервер [аукциона](http://auction.ncsd.ru). Документация по API находится в swagger.yaml

## Конфигурация сервера

Пример конфигурации сервера находится в `./cli/auction/.auction.example.yaml`, назначение полей описано в коментариях конфига.

## Создание пользователя на аукционе

### Авторизация с помощью root пользователя

```curl
curl -X POST \
  http://back.auction.prod.ncsd.ru/login \
  -F username=root \
  -F password=<ROOT_PASSWORD>
```

`ROOT_PASSWORD` - можно узнать в списке переменных окружения в prod среде в rancher, переменная называется `AUCTION_ROOT_PASSWORD`.

В результате выполнения получим token.

### Создание пользователя

```curl
curl -X POST \
  http://back.auction.prod.ncsd.ru/users \
  -H 'Accept: application/json' \
  -H 'Authorization: Bearer <TOKEN>' \
  -H 'Content-Type: application/json' \
  -d '{
  "username": "username",
  "object_type": "trip",
  "password": "password"
}'
```

`TOKEN` - token полученый на первом шаге.

В результате выполнения получим информацию о созданном пользователе.

### Добавления доступа к группе лотов

```curl
curl -X PUT \
  http://back.auction.prod.ncsd.ru/users/<USER_ID>/groups/<GROUP_KEY> \
  -H 'Accept: application/json' \
  -H 'Authorization: Bearer <TOKEN>' \
  -H 'Content-Type: application/json'
```

`USER_ID` - id пользователя полученый на втором шаге

`GROUP_KEY` - ключ группы (например nc_trip), можно посмотреть в таблице users_groups бд auction

`TOKEN` - token полученый на первом шаге.

В результате выполнения получим код 204.

## Описание взаимодействия через WebSocket

Обмен данными между клиентом и сервером происходит в формате `json`. Клиент
отправляет на сервер команды и в ответ получает событие с результатом 
или с ошибкой. Так же клиенту могут приходить сообщения от сервера без отправки
команды с клиента, например если на сервере произошли новые события.

### Команды

Пример команды:

```json
{
    "type": "command.place.bet",
    "payload": {
        "lot_id": 101,
        "value": 15500
    }
}
```

* `type` - тип отправляемой команды
* `payload` - данные необходимые для выполнений команды

#### command.get.lots

*Запрос:*

Получить список лотов

```json
{
    "type": "command.get.lots"
}
```

*Ответ:*

```json
{
    "type": "event.get.lots.success",
    "payload": []
}
```

#### command.get.lot

Получить подробную информацию о лоте

*Запрос:*

```json
{
    "type": "command.get.lot",
    "payload": {
        "lot_id": 101
    }
}
```

*Ответ:*

```json
{
    "type": "event.get.lot.success",
    "payload": {
    }
}
```

#### command.place.bet

Сделать ставку на лот

*Запрос:*

```json
{
    "type": "command.place.bet",
    "payload": {
        "lot_id": 101,
        "value": 15500
    }
}
```

*Ответ:*

```json
{
    "type": "event.place.bet.success",
}
```

#### command.cancel.bet

Отменить последнию сделанную ставку на лот

*Запрос:*

```json
{
    "type": "command.cancel.bet",
    "payload": {
        "lot_id": 101
    }
}
```

*Ответ:*

```json
{
    "type": "event.cancel.bet.success",
}
```

#### command.place.reservation

#### command.cancel.reservation

#### command.confirm.bet

Подтвердить сделанную ставку

*Запрос:*

```json
{
    "type": "command.confirm.bet",
    "payload": {
        "lot_id": 101
    }
}
```

### События

#### event.lot.added

Добавлен новый лот

```json
{
    "type": "event.lot.added",
    "payload": {
    }
}
```

#### event.lot.changed

Лот изменился

```json
{
    "type": "event.lot.changed",
    "payload": {
    }
}
```
