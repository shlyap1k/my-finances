
# API URLs

Base path:

```text
/api

Формат данных: JSON.
Даты: ISO format. YYYY-MM-DD
```
Все суммы: integer minor units.
Пример:

```text
100000 рублей = 10000000
```

Проценты для savings rules: basis points.

```text
1% = 100
10% = 1000
12.5% = 1250
20% = 2000
```

Ошибки
Стандартный формат ошибки:

```json
{
  "error": {
    "code": "validation",
    "message": "invalid request",
    "fields": {
      "email": "email is required"
    }
  }
}
```
Health
GET /api/health
Проверка доступности приложения.
Без авторизации.
Response:

{
"status": "ok"
}

GET /api/balances
Возвращает фактические балансы.
Auth required.

Auth
POST /api/auth/register
Регистрация пользователя.

POST /api/auth/login
Вход.

GET /api/auth/me
Текущий пользователь и настройки.
Auth required.

User settings
PATCH /api/me/settings
Обновить настройки текущего пользователя.
Auth required.
Request:

```json
{
  "timezone": "Europe/Moscow",
  "avg_window_days": 30,
  "default_include_avg": false
}
```
response:
```json
{
  "settings": {
    "timezone": "Europe/Moscow",
    "avg_window_days": 30,
    "default_include_avg": false
  }
}
```

Income rules
GET /api/income-rules
Список регулярных доходов.
Auth required.
Query params:

    active — optional boolean
    page — optional
    per_page — optional

POST /api/income-rules
Создать регулярный доход.
Auth required.
Request
```json
{
  "name": "Зарплата",
  "amount_minor": 10000000,
  "monthly_day": 5,
  "start_date": "2026-08-01",
  "end_date": null,
  "active": true,
  "note": "основная работа"
}
```
GET /api/income-rules/{id}
Получить регулярный доход.
Auth required.
Если объект принадлежит другому пользователю, возвращать 404.

PUT /api/income-rules/{id}
Обновить регулярный доход.
Auth required.
Request — те же поля, что при создании.

DELETE /api/income-rules/{id}
Удалить регулярный доход.
Auth required.
Response: 204 No Content.
Исторические транзакции, созданные от этого правила, остаются.

Expense rules
GET /api/expense-rules
Список обязательных расходов.
Auth required.

POST /api/expense-rules
Создать обязательный расход.
Auth required.
Request

```json
{
  "name": "Аренда",
  "amount_minor": 3000000,
  "monthly_day": 5,
  "start_date": "2026-08-01",
  "end_date": null,
  "active": true,
  "note": null
}
```

GET /api/expense-rules/{id}
Получить обязательный расход.
Auth required.

PUT /api/expense-rules/{id}
Обновить обязательный расход.
Auth required.

DELETE /api/expense-rules/{id}
Удалить обязательный расход.
Auth required.
Response: 204 No Content.
Исторические подтверждения остаются.

POST /api/expense-rules/{id}/confirm
Подтвердить обязательный расход.
Auth required.
Это действие создает фактическую транзакцию расхода.
Request:
```json
{
  "occurred_on": "2026-08-11",
  "source_rule_occurred_on": "2026-08-05",
  "amount_minor": 3000000
}
```

Поля:

    occurred_on — дата фактической оплаты;
    source_rule_occurred_on — дата планового события; необязательное поле;
    amount_minor — сумма; если не указана, используется сумма правила.

Поведение:

    если source_rule_occurred_on не указан, по умолчанию используется occurred_on;
    если обязательный расход для этой плановой даты уже подтвержден, вернуть 409 conflict;
    после подтверждения проекция не должна снова списывать это правило на дату source_rule_occurred_on.

Response
```json
{
  "transaction": {
    "id": 123,
    "occurred_on": "2026-08-11",
    "amount_minor": 3000000,
    "kind": "expense",
    "status": "actual",
    "origin": "obligation_confirmation",
    "source_expense_rule_id": 10,
    "source_rule_occurred_on": "2026-08-05",
    "note": null
  }
}
```
Savings accounts
GET /api/savings-accounts
Список накопительных счетов.
Auth required.
Response:

```json
{
  "items": [
    {
      "id": 1,
      "name": "Подушка",
      "active": true,
      "balance_minor": 1000000
    }
  ]
}
```

POST /api/savings-accounts
Создать накопительный счет.
Auth required.
Request:

```json
{
  "name": "Подушка",
  "active": true,
  "note": "финансовая подушка"
}
```
PUT /api/savings-accounts/{id}
Обновить накопительный счет.
Auth required

DELETE /api/savings-accounts/{id}
Удалить накопительный счет.
Auth required.
Если по счету есть транзакции, вернуть 409 conflict.
Если транзакций нет, вернуть 204 No Content.

Savings rules
GET /api/savings-rules
Список правил накопления.
Auth required.
Query params:

    income_rule_id — optional
    savings_account_id — optional
    active — optional


POST /api/savings-rules
Создать правило накопления.
Auth required.
Request для процента:
```json
{
  "name": "10% в подушку",
  "income_rule_id": 1,
  "savings_account_id": 1,
  "rule_type": "percent",
  "percent_bps": 1000,
  "priority": 10,
  "active": true
}
```

для фиксированной суммы
```json
{
  "name": "5000 в отпуск",
  "income_rule_id": 1,
  "savings_account_id": 2,
  "rule_type": "fixed",
  "fixed_amount_minor": 500000,
  "priority": 20,
  "active": true
}
```
Валидация:

    rule_type обязателен;
    если rule_type = percent, обязателен percent_bps;
    если rule_type = fixed, обязателен fixed_amount_minor;
    percent_bps должен быть от 0 до 10000;
    fixed_amount_minor должен быть >= 0;
    savings_account_id обязан существовать и принадлежать пользователю;
    income_rule_id, если указан, обязан существовать и принадлежать пользователю.

PUT /api/savings-rules/{id}
Обновить правило накопления.
Auth required.
DELETE /api/savings-rules/{id}
Удалить правило накопления.
Auth required.
Исторические транзакции, созданные правилом, остаются.
Transactions
GET /api/transactions
Список транзакций.
Auth required.
Query params:

    from — optional date
    to — optional date
    status — optional: planned / actual
    kind — optional: income / expense / savings_deposit / savings_withdraw
    origin — optional
    page — optional
    per_page — optional

Response:
```json
{
  "items": [
    {
      "id": 100,
      "occurred_on": "2026-08-11",
      "amount_minor": 150000,
      "kind": "expense",
      "status": "actual",
      "origin": "manual",
      "category": "fuel",
      "note": "бензин",
      "exclude_from_average": false
    }
  ]
}
```

POST /api/transactions
Создать транзакцию.
Auth required.
Система сохраняет транзакцию даже если проекция показывает, что денег не хватает.
Request:

```json
{
  "occurred_on": "2026-09-10",
  "amount_minor": 15000000,
  "kind": "expense",
  "status": "planned",
  "category": "iphone",
  "note": "покупка iPhone",
  "exclude_from_average": true
}
```
Поля:

    occurred_on — обязательная дата;
    amount_minor — обязательная сумма;
    kind — обязателен;
    status — обязателен;
    savings_account_id — обязателен для savings_deposit и savings_withdraw;
    exclude_from_average — optional, default false.

Ограничения:

    status = actual не может быть будущей датой;
    status = planned может быть будущей датой;
    origin всегда устанавливается сервером как manual для пользовательских транзакций.

Response: 201 Created.
Ответ сразу содержит транзакцию и проекцию для даты транзакции.

GET /api/transactions/{id}
Получить транзакцию.
Auth required.
PUT /api/transactions/{id}
Обновить транзакцию.
Auth required.
Разрешено обновлять:

    origin = manual;
    origin = obligation_confirmation.

Для автоматически созданных транзакций:

    origin = regular_income;
    origin = savings_rule

возвращать 409 conflict.
Request:
```json
{
  "occurred_on": "2026-09-12",
  "amount_minor": 14000000,
  "kind": "expense",
  "status": "planned",
  "category": "iphone",
  "note": "покупка iPhone",
  "exclude_from_average": true
}
```

Response содержит обновленную транзакцию и проекцию.
DELETE /api/transactions/{id}
Удалить транзакцию.
Auth required.
Разрешено удалять:

    origin = manual;
    origin = obligation_confirmation.

Для автоматически созданных транзакций возвращать 409 conflict.
Response: 204 No Content.
POST /api/transactions/{id}/mark-actual
Отметить плановую транзакцию как фактическую.
Auth required.
Request:
```json
{
  "occurred_on": "2026-08-11"
}
```
Если occurred_on не указан, используется сегодняшняя дата пользователя.
Поведение:

    меняет status на actual;
    обновляет occurred_on;
    сохраняет origin как есть;
    если транзакция была плановой покупкой и не должна влиять на средний расход, клиент должен заранее или одновременно выставить exclude_from_average = true через PUT.

Response содержит транзакцию и проекцию.
Projection
GET /api/projection
Посчитать проекцию без привязки к новой транзакции.
Auth required.
Query params
target_date: "date"
include_avg: ?bool - необязательный параметр

ответ:
```json
{
  "as_of_date": "2026-08-11",
  "target_date": "2026-09-10",
  "include_avg": true,
  "average_daily_spend_minor": 120000,
  "projected_balance_before": 1500000,
  "projected_balance_after": 980000,
  "min_balance_in_period": 900000,
  "shortfall": 0,
  "enough": true,
  "warnings": []
}
```

POST /api/transactions/check
Проверить гипотетическую транзакцию без сохранения.
Auth required.
Request

```json
{
  "target_date": "2026-09-10",
  "amount_minor": 15000000,
  "kind": "expense",
  "include_avg": true
}
```

Поля:

    target_date — обязательная дата;
    amount_minor — обязательная сумма;
    kind — обязателен;
    include_avg — optional.

Response:

```json

{
  "as_of_date": "2026-08-11",
  "target_date": "2026-09-10",
  "include_avg": true,
  "average_daily_spend_minor": 120000,
  "projected_balance_before": 1500000,
  "projected_balance_after": -13500000,
  "min_balance_in_period": -13500000,
  "shortfall": 13500000,
  "enough": false,
  "warnings": [
    "negative balance expected on 2026-09-10"
  ]
}
```

Этот endpoint ничего не сохраняет.


итоговая карта эндпоинтов

GET    /api/health

GET    /api/balances

POST   /api/auth/register
POST   /api/auth/login
POST   /api/auth/logout
GET    /api/auth/me
PATCH  /api/me/settings

GET    /api/income-rules
POST   /api/income-rules
GET    /api/income-rules/{id}
PUT    /api/income-rules/{id}
DELETE /api/income-rules/{id}

GET    /api/expense-rules
POST   /api/expense-rules
GET    /api/expense-rules/{id}
PUT    /api/expense-rules/{id}
DELETE /api/expense-rules/{id}
POST   /api/expense-rules/{id}/confirm

GET    /api/savings-accounts
POST   /api/savings-accounts
PUT    /api/savings-accounts/{id}
DELETE /api/savings-accounts/{id}

GET    /api/savings-rules
POST   /api/savings-rules
PUT    /api/savings-rules/{id}
DELETE /api/savings-rules/{id}

GET    /api/transactions
POST   /api/transactions
GET    /api/transactions/{id}
PUT    /api/transactions/{id}
DELETE /api/transactions/{id}
POST   /api/transactions/{id}/mark-actual

GET    /api/projection
POST   /api/transactions/check


