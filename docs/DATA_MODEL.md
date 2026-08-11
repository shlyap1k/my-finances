# Data Model

Все суммы хранятся как `bigint` в minor units.

Например:
- 100000 рублей = `10000000`;
- 5000 рублей 50 копеек = `500050`.

Все даты денежных событий хранятся как `date`.

Все служебные даты создания и обновления хранятся как `timestamptz`.

Все пользовательские данные имеют `user_id`.

---

## users

Основная таблица пользователей.

| Поле | Тип | Ограничения |
|---|---|---|
| id | bigserial | PK |
| email | text | not null, unique |
| password_hash | text | not null |
| created_at | timestamptz | not null, default now() |
| updated_at | timestamptz | not null, default now() |

---

## user_settings

Настройки пользователя.

| Поле | Тип | Ограничения |
|---|---|---|
| user_id | bigint | PK, FK -> users.id, ON DELETE CASCADE |
| timezone | text | not null, default 'UTC' |
| avg_window_days | integer | not null, default 30, check 1..365 |
| default_include_avg | boolean | not null, default false |
| created_at | timestamptz | not null |
| updated_at | timestamptz | not null |

Пояснение:
- `timezone` — IANA timezone, например `Europe/Moscow`;
- `avg_window_days` — окно для расчета среднего расхода в день;
- `default_include_avg` — включать ли средний расход по умолчанию.

---

## income_rules

Регулярные доходы.

Например:
- зарплата 5-го числа;
- аванс 20-го числа;
- ежемесячная премия 10-го числа.

| Поле | Тип | Ограничения |
|---|---|---|
| id | bigserial | PK |
| user_id | bigint | not null, FK -> users.id, ON DELETE CASCADE |
| name | text | not null |
| amount_minor | bigint | not null, check > 0 |
| monthly_day | integer | not null, check 1..31 |
| start_date | date | not null |
| end_date | date | nullable |
| active | boolean | not null, default true |
| note | text | nullable |
| created_at | timestamptz | not null |
| updated_at | timestamptz | not null |

Правило считается активным для даты `d`, если:

```text
active = true
and start_date <= d
and (end_date is null or d <= end_date)
Если monthly_day больше числа дней в месяце, событие происходит в последний день месяца.
```

##expense_rules
###Обязательные регулярные расходы.
Например:

    аренда;
    кредит;
    связь;
    подписки.

Правило уменьшает фактический баланс только после подтверждения.

##savings_accounts
###Накопительные счета.
Проценты по накопительным счетам не учитываются.
Баланс накопительного счета считается по транзакциям:

    savings_deposit увеличивает баланс счета;
    savings_withdraw уменьшает баланс счета.

##savings_rules
###Правила перевода части дохода на накопительные счета.

Проценты храним в basis points:
Значение | percent
1%       | 100
12.5%    | 1250
20%      | 2000
100%     | 10000

Если income_rule_id указан, правило применяется только к этому доходу.
Если income_rule_id null, правило применяется ко всем регулярным доходам.

Процент всегда считается от исходной суммы дохода, а не от остатка после предыдущих правил.
Фиксированная сумма применяется от остатка дохода после предыдущих правил.
Общая сумма переводов по доходу не может превышать сумму дохода. Если правила требуют больше, 
переводы обрезаются по остатку и возвращается warning.

##transactions
###Транзакции — источник фактического баланса и плановых операций.
Возможные значения
income, expense, savings_deposit, savings_withdraw

status
planned, actual

origin
manual, regular_income, savings_rule, obligation_confirmation (подтвержденный пользователем обязательный расход)

source_rule_occurred_on
Используется для подтвержденных обязательных расходов.
Хранит дату планового события, за которое подтверждена оплата.
Например:

    аренда по правилу должна быть 5-го числа;
    пользователь подтвердил оплату 5-го числа;
    occurred_on = 2026-09-05;
    source_rule_occurred_on = 2026-09-05.

Если оплата произведена заранее:

    occurred_on = 2026-09-03;
    source_rule_occurred_on = 2026-09-05.

Это позволяет не списать обязательный расход дважды в прогнозе.

exclude_from_average
Если true, операция не участвует в расчете среднего расхода в день.

Пользователь может редактировать и удалять:

    origin = manual;
    origin = obligation_confirmation.

Пользователь не должен редактировать или удалять:

    origin = regular_income;
    origin = savings_rule.
