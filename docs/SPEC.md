# SPEC тимлида → код

**Продукт:** 1020 · SAFEKID · [`catalog/safekid.json`](../catalog/safekid.json)  
**Review по шагам:** [SCENARIOS.md](./SCENARIOS.md)

## Транспорты и Viva

| SPEC | Код |
|------|-----|
| `intTransport` (NATS / SDS) | `viva.intTransport` |
| `extTransport` (HTTP / Viva) | `viva.extTransport` |
| `ussdTransport` (SMPP) | `viva.ussdTransport` |
| `vivaClient *vivaclient.Client` | `viva.vivaClient` |
| `catalog Catalog` | `catalog.Catalog` + `catalog.NewCatalog()` |
| `type Viva interface` + handlers | `Viva` + `viva` struct, [`viva.go`](../internal/app/viva.go), [`constructor.go`](../internal/app/constructor.go) |

## Catalog / Product

| SPEC | Код |
|------|-----|
| `Load`, `GetProductByShortNumber`, `GetProductByExternalId`¹, `SetDefaultLang` | [`internal/app/catalog/`](../internal/app/catalog/) |

¹ В черновике SPEC было `GetProductByExtermalId` — в коде исправлена опечатка.
| `GetNotify` + шаблоны | `GetNotify` + [`render.go`](../internal/app/catalog/render.go) |
| MO-меню в JSON `rules` | **нет** — команды в Go ([`inbound.go`](../internal/app/inbound.go)) |

## Методы по файлам SPEC

| SPEC | Файл |
|------|------|
| `InitHandlers`, подписки | [`instance.go`](../internal/app/instance.go) |
| `ussdHandler` | [`ussd.go`](../internal/app/ussd.go) → `smpp/inbound` |
| `orderCompleteHandler` | [`notify.go`](../internal/app/notify.go) |
| `webhookHandler` | [`webhook.go`](../internal/app/webhook.go) |
| `landingInit` / `landingConfirm` | [`landing.go`](../internal/app/landing.go) |
| `getOrderId`, `getOrder`, `isActiveOrder`, `createOrder`, `completeOrder` | [`order.go`](../internal/app/order.go) |
| `notify` | [`notify.go`](../internal/app/notify.go) |

## Сценарии подключения (SPEC)

| SPEC | Код |
|------|-----|
| Лендинг: Init + Confirm + `order.create` | `landing.go`, `viva_subscription.go`, `order.go` |
| Вебхуки Viva | `webhook.go` (+ renew: `scenario4.go`) |
| Короткий номер, SMS-меню | `ussd.go` + `activation` / `stop` / `lang` / `unknown` |

## Вне SPEC (есть в коде)

- `order/expires` → [`scenario3.go`](../internal/app/scenario3.go) (СМС №5)
- `POST /ExtAppPartneerProductActivation` → renew + №4 (`scenario4.go`, `paid_welcome_store.go`)
- `paidWelcomeSent`, `landingConfirmURL` в конфиге
- `GetNotifies` — **не реализовано**
