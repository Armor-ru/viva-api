# Review — Safe Kids (1020)

**SPEC → код:** [SPEC.md](./SPEC.md) · **Каталог:** [`catalog/safekid.json`](../catalog/safekid.json) (только `notifications`, MO в Go)

```bash
go test ./internal/app/... -count=1
```

**Порядок:** база (ниже) → **8 → 7 → 6 → 5 → 3 → 4 → 2 → 1**

---

## База (один раз)

| Файлы | Зачем |
|-------|--------|
| [`instance.go`](../internal/app/instance.go) | `InitHandlers`: `smpp/inbound`, `order/completed`, `order/expires`, webhook, landing |
| [`ussd.go`](../internal/app/ussd.go) | Роутер MO: `1` → STOP → RUS/ARM/ENG → unknown |
| [`inbound.go`](../internal/app/inbound.go), [`inbound_sms.go`](../internal/app/inbound_sms.go) | Условия на 1020, парсинг SMPP |
| [`catalog/store.go`](../internal/app/catalog/store.go), [`catalog_access.go`](../internal/app/catalog_access.go) | Продукт 1020 / SAFEKID, `GetNotify` |
| [`order.go`](../internal/app/order.go), [`notify_product.go`](../internal/app/notify_product.go) | `order/get`, `order/create`, исходящие СМС |

**Вердикт:** ☐ OK · ☐ Замечания · ☐ Блокер

**Замечание / вопрос:**



---

## Сценарий 8 — неизвестный текст → СМС №13

**Триггер:** MO на 1020, текст не `1` / `STOP` / `RUS` / `ARM` / `ENG`.

**Файлы:** `ussd.go` → [`unknown.go`](../internal/app/unknown.go) → `safekid.json` (`unknown_command`)

**Тест:** `scenario8_test.go`

**Вердикт:** ☐ OK · ☐ Замечания · ☐ Блокер

**Замечание / вопрос:**



---

## Сценарий 7 — RUS / ARM / ENG → СМС №10–12

**Триггер:** MO `RUS` | `ARM` | `ENG` на 1020.

**Файлы:** `ussd.go` → [`lang.go`](../internal/app/lang.go) → [`lang_store.go`](../internal/app/lang_store.go) → `safekid.json` (`language_changed`)

**Тест:** `scenario7_test.go`

**Вердикт:** ☐ OK · ☐ Замечания · ☐ Блокер

**Замечание / вопрос:**



---

## Сценарий 6 — STOP, уже отключён → СМС №9

**Триггер:** `STOP` на 1020, в SDS заказ есть, **не** активен.

**Файлы:** `ussd.go` → [`stop.go`](../internal/app/stop.go) (`stopOrderAlreadyDeactivated`, **без** `removeVivaSubscription`)

**Тест:** `scenario6_test.go`

**Вердикт:** ☐ OK · ☐ Замечания · ☐ Блокер

**Замечание / вопрос:**



---

## Сценарий 5 — STOP, отключение → Viva Remove + СМС №6

**Триггер:** `STOP` на 1020, активный заказ (или нет активного — всё равно deactivate).

**Файлы:** `stop.go` → [`viva_subscription.go`](../internal/app/viva_subscription.go) → `notify_product.go` · `paid_welcome_store.Clear`

**Тест:** `scenario5_test.go`

**Вопрос к ТЗ:** grace period до `EndTime` — в коде **сразу** `RemoveSubscription`.

**Вердикт:** ☐ OK · ☐ Замечания · ☐ Блокер

**Замечание / вопрос:**



---

## Сценарий 3 — order.expires → СМС №5

**Триггер:** NATS `order/expires` (шаг 1 — SDS).

**Файлы:** `instance.go` → [`scenario3.go`](../internal/app/scenario3.go) → `safekid.json` (`trial_expires`, `{{.ExpiresAt}}`)

**Тест:** `scenario3_test.go`

**Вердикт:** ☐ OK · ☐ Замечания · ☐ Блокер

**Замечание / вопрос:**



---

## Сценарий 4 — trial → paid → СМС №4

**Триггер:** webhook `ExtAppPartneerProductActivation` → `order/create` **renew** → `order/completed`.

**Файлы (по порядку):** [`scenario4.go`](../internal/app/scenario4.go) → [`notify.go`](../internal/app/notify.go) (`isScenario4OrderCompleted`) → [`paid_welcome_store.go`](../internal/app/paid_welcome_store.go)

**Не путать:** `ExtAppPartneerProductActivationRequest` → `webhook.go` → **new** (не этот сценарий).

**Тест:** `scenario4_test.go`

**Вопрос к ТЗ:** №4 один раз на phone+product (in-memory; после рестарта может повториться).

**Вердикт:** ☐ OK · ☐ Замечания · ☐ Блокер

**Замечание / вопрос:**



---

## Сценарий 2 — MO `1`, уже активен → СМС №8

**Триггер:** MO `1` на 1020, `order/get` — заказ **активен**.

**Файлы:** [`activation.go`](../internal/app/activation.go) → `order.go` (`isActiveOrder`) → `safekid.json` (`already_active`)

**Тест:** `scenario2_test.go`

**Вердикт:** ☐ OK · ☐ Замечания · ☐ Блокер

**Замечание / вопрос:**



---

## Сценарий 1 — MO `1`, новая подписка → лендинг → №2–3

**Триггер:** MO `1`, заказа **нет** → Viva Init → SMS с лендингом → Confirm на лендинге → `order/create` → `order/completed` → №2 + №3.

**Файлы (по порядку):** `activation.go` → `viva_subscription.go` → [`landing_url.go`](../internal/app/landing_url.go) → `landing.go` → `order.go` → `notify.go` + [`order_artifacts.go`](../internal/app/order_artifacts.go)

**Тесты:** `ussd_test.go`, `viva_subscription_test.go`, `order_complete_test.go`, `activation_notify_test.go`, `landing_*_test.go`

**Вопрос к ТЗ:** заказ **есть**, но **не активен** — тихий выход, без SMS и без повторного Init (`activation.go`).

**Вердикт:** ☐ OK · ☐ Замечания · ☐ Блокер

**Замечание / вопрос:**



---

## Прочее (не MO)

| Endpoint | Файл |
|----------|------|
| `POST /ExtAppPartneerProductActivationRequest` | `webhook.go` (`order/create` new) |
| `POST /ExtAppPartneerProductRemove` | `webhook.go` |
| `POST /landing/init-subscription`, `confirm-subscription` | `landing.go` |

**Вердикт:** ☐ OK · ☐ Замечания · ☐ Блокер

**Замечание / вопрос:**


