# External app Integration API

## (Software Description)

 

14 | PAGE

  Software Development Unit 

 Ext-app Integrations API

## Table of Contents

1. Introduction 4

2. Authorization with client credentials 5

2.1 Getting an Access Token 5

2.2 Using an Access Token 5

2.3 Token Expiry 5

2.4* Pinging API for availability 5

3. API Endpoints 6

3.1 Message: GetSubscriberInfo 6

3.2 Message: AddTariffPlanProduct 6

3.3 Message: AddChargeableProduct 7

3.4 Message: GetProductsByPhoneNum 7

3.5 Message: InitSubscription 8

3.6 Message: ConfirmSubscription 8

3.7 Message: RemoveSubscription 9

4. Elements 9

4.1 Element: GetSubInfoResponse 9

4.2 Element: AddProductResponse 9

4.3 Element: TpProductDTO 9

4.4 Element: AddChargeableProductResponse 9

4.5 Element: ChargeableProductDTO 9

4.6 Element: ResponseModel 10

4.7 Element: ExtAppProductsRespDTO 10

4.8 Element: ExtAppProductDTO 10

5. Result code 11

6. ExtApp Subscription workflow detailed description 12

7. Partner BFF (viva-api) — Landing HTTP and Postman 12


14 | PAGE

  Software Development Unit 

 Ext-app Integrations API

## 1. Introduction

This API is intended to provide a possibility to partners to integrate their products into our tariff plans or activate chargeable products by deducting balance from Viva subscribers account.

For adding tariff plan product to subscriber, partners need to do two steps: 

Get subscriber information in first provisioning

Add appropriate product along with providing subscriber information (subNo)

For adding chargeable product to subscriber, partners need to do two steps:

Get subscriber information in first provisioning

Add appropriate product along with providing subscriber information (subNo)

All of the products should be pre-registered for given tariff plan before providing to a subscriber.

It is mandatory to employ a sim verification via SMS or via Header Enrichment procedure by Viva before doing above mentioned actions to subscriber.


14 | PAGE

  Software Development Unit 

 Ext-app Integrations API

## 2. Authorization with client credentials

This API uses the credentials flow for authentication and authorization

### 2.1 Getting an Access Token

To authenticate, partners need to make a POST request to the /auth/token endpoint with their clientId and clientSecret in the body. Upon successful authentication, they will receive an access token.

Example:
Request
POST /auth/token
Content-Type: application/json
```
{
  "userName": "partnerUserName",
  "password": "partnerPassword"
}
```

Response
A successful response will include an access_token in the body
```
{
  "access_token": "your_access_token",
  "token_type": "Bearer",
  "expires_in": 86399 
}
```

### 2.2 Using an Access Token

Include the access token in the Authorization header of all requests of the API
Authorization: Bearer your_access_token
Replace your_access_token with the actual access token that you received from the /auth/token endpoint. 

If the access token is invalid or expired, you will receive a 401 Unauthorized response.

### 2.3 Token Expiry

Access tokens have an expiration time, which is set when the token is generated. When an access token expires, you will need to authenticate again to get a new token. 

### 2.4* Pinging API for availability

For checking the API for availability without executing any action, partners can send OPTIONS type request to the same /auth/token endpoint without any parameter.

## 3. API Endpoints

### 3.1 Message: GetSubscriberInfo

A request from the partner to receive information about subscriber and available products for the subscriber to activate.

Type: GET

Endpoint: /api/Subscriber/{msisdn}

Example: for 37477210216, /api/Subscriber/37477210216

Input

| Parameter | Type | M |
|-----------|------|---|
| msisdn | String | M |

In response it is provided subscriber information and what products are available for him in his current state. In case of success API response will be:

Output

| Type |
|------|
| GetSubInfoResponse |

Note: Expected status codes are 200,404,429.

### 3.2 Message: AddTariffPlanProduct

We receive a request from a partner to add a product to subscriber and set the expiration date.

Type: POST

Endpoint: /api/Subscriber/AddTariffPlanProduct

Description:   add product

Input

| Parameter | Type | M |
|-----------|------|---|
| tpProduct | TpProductDTO | M |

Output

| Parameter |
|-----------|
| AddProductResponse |

Note: Expected status codes are 200,400,429.

14 | PAGE

  Software Development Unit 

 Ext-app Integrations API

### 3.3 Message: AddChargeableProduct

A request from a partner to add a chargeable product pre-configured in Viva side with fixed price.

Type: POST

Endpoint: /api/Subscriber/AddTariffPlanProduct

Description:   add product

Input

| Parameter | Type | M |
|-----------|------|---|
| tpProduct | ChargeableProductDTO | M |

Output

| Parameter |
|-----------|
| AddChargeableProductResponse |

Note: Expected status codes are 200,400,429.

### 3.4 Message: GetProductsByPhoneNum

A request from a partner to get All products which are available for subscription, and products which are already subscribed.

Type: Get

Endpoint: /api/Subscription/GetProductsByPhoneNum

Description:   Get products for mentioned number

Input

| Parameter | Type | M |
|-----------|------|---|
| phoneNum | string | M |

Output

| Parameter |
|-----------|
| ExtAppProductsRespDTO |

Note: It is only for our mobile application

Example:
Response
A successful response type 
```
{
  "resultCode": 0,
  "message": "string",
  "result": {
    "phoneNumber": "string",
    "activeSubscribedProducts": [
      {
        "productName": "string",
        "validDays": 0,
        "expDate": "2024-11-25T11:51:37.434Z"
      }
    ],
    "availableProducts": [
      {
        "productName": "string",
        "validDays": 0,
        "expDate": "2024-11-25T11:51:37.434Z"
      }
    ]
  }
}
```

### 3.5 Message: InitSubscription

A request from a partner to init subscription, then for success subscription should confirm by request mentioned in 3.6 point .

Type: POST

Endpoint: /api/Subscription/InitSubscription

Description:   Init subscription

Input

| Parameter | Type | M |
|-----------|------|---|
| phoneNum | string | M |
| productName | string | M |

Output

| Parameter |
|-----------|
| ResponseModel |

### 3.6 Message: ConfirmSubscription

A request from a partner to confirm subscription

Type: POST

Endpoint: /api/Subscription/ConfirmSubscription

Description:  Confirm subscription

Input

| Parameter | Type | M |
|-----------|------|---|
| phoneNum | string | M |
| productName | string | M |
| otp | string | M |

Output

| Parameter |
|-----------|
| ResponseModel |

### 3.7 Message: RemoveSubscription

A request from a partner to remove subscription

Type: POST

Endpoint: /api/Subscription/RemoveSubscription

Description:  Remove subscription

Input

| Parameter | Type | M |
|-----------|------|---|
| phoneNum | string | M |
| productName | string | M |

Output

| Parameter |
|-----------|
| ResponseModel |

## 4. Elements

### 4.1 Element: GetSubInfoResponse

| Name | Type | Description |
|------|------|-------------|
| msisdn | String | Phone Number |
| subNo | Int32 | Number’s SubNo |
| availableProducts | List<Int32> | Available Product for that Number |

### 4.2 Element: AddProductResponse

| Name | Type | Description |
|------|------|-------------|
| resultCode | Int32 | Result code |
| message | Int32 | Result code description |
| result | TpProductDTO | Activated product info |

### 4.3 Element: TpProductDTO

| Name | Type | Description |
|------|------|-------------|
| msisdn | String | Phone number |
| subNo | Int32 | Number’s SubNo |
| expDate | Nullable<DateTime> | Expiration Date |
| prodId | Int32 | Product Id |

### 4.4 Element: AddChargeableProductResponse

| Name | Type | Description |
|------|------|-------------|
| resultCode | Int32 | Result code |
| message | Int32 | Result code description |
| result | ChargeableProductDTO | Activated product info |

### 4.5 Element: ChargeableProductDTO

| Name | Type | Description |
|------|------|-------------|
| msisdn | String | Phone number |
| subNo | Int32 | Number’s SubNo |
| prodName | String | Product Name |

### 4.6 Element: ResponseModel

| Name | Type | Description |
|------|------|-------------|
| ResultCode | Int32 | Result code |
| Message | Int32 | Result code description |
| Result | bool | True/False |

### 4.7 Element: ExtAppProductsRespDTO

| Name | Type | Description |
|------|------|-------------|
| PhoneNumber | string | Phone Number |
| ActiveSubscribedProducts | List<ExtAppProductDTO> | List of subscribed products |
| AvailableProducts | List<ExtAppProductDTO> | Available subscriptions |

### 4.8 Element: ExtAppProductDTO 

| Name | Type | Description |
|------|------|-------------|
| PhoneNumber | string | Phone Number |
| ProductName | string | Product name |
| ValidDays | int | Product valid days per subscription |
| ExpDate | DateTime | Subscription expire date |

## 5. Result code

| Result Code | Description |
|-------------|-------------|
| 0 | Success |
| 1 | Not allowed product |
| 2 | Product not exist |
| 3 | Not allowed product for subscriber state |
| 4 | Owner changed |
| 5 | Data profile not exist |
| 6 | Partner not found |
| 7 | Already active |
| 8 | Date has passed |
| 9 | Subscriber not exist |
| 10 | Free traffic assignment failed. |
| 11 | Activation failed (other cases) |
| 12 | MSISDN Not Valid |
| 13 | Internal error – try again |
| 14 | Not Enough Funds |
| 15 | Subscriber is inactive |
| 16 | Invalid Input Params |
| 17 | No Pending Subscription |
| 18 | Verification Not Exists |
| 19 | Not Verified |
| 20 | Subscription Not Exists |
| 21 | SMS Not Send |
| 22 | Try Prolong Limit Exceeded |
| 23 | Too many request |

## 6. ExtApp Subscription workflow detailed description

### Introduction

Ext App Subscription is a platform designed to enable Viva’s partners to deliver their products to subscribers and manage product subscriptions through a standardized webhook-based integration mechanism.
This environment enables real-time communication between partner systems and Viva’s infrastructure, ensuring seamless management of the subscription lifecycle — including  activation,  prolongation, and  deactivation, depending on actions initiated by either party.

To integrate with the Ext App Subscription system, the partner must follow the steps in the sequence described below.

Partner should provide : 

Product info (Product Name, Price, Description, Has OTP verification or no, product valid period , Trial period days )

The following three callback URLs for each product: 

url for request new subscription - https://partnerUrl/ExtAppPartneerProductActivationRequest 

url for success Prolongation  - https://partnerUrl/ExtAppPartneerProductActivation 

url for remove subscription - https://partnerUrl/ExtAppPartneerProductRemove

We will provide an Ext App Subscription API, which operates according to the following endpoints.

/auth/token endpoint with their userName and password in the body. Upon successful authentication, they will receive an access token. (see 2. Authorization with client credentials)

/api/Subscription/InitSubscription endpoint with parameters phoneNumber (in format 374XXXXXXXX) and productName (means product key, which we will provide). 

Regardless of whether OTP verification is required or not, the subscription is initially created in PreActive status. PreActive indicates that the subscription was initialized but is not active yet. The subscription transitions to Active only after confirmation (see next step). If the subscriber already had a subscription for the given product, and it is not currently Active, the system initializes it again as a new PreActive subscription.

/api/Subscription/ConfirmSubscription endpoint with parameters phoneNumber (in format 374XXXXXXXX) and productName (means product key, which we will provide) and Nullable OTP. If product have OTP, and in InitSubscription subscriber received OTP via sms, we do OTP verification, then we do charge subscriber if it is not trial version and confirm subscription. Trial Period Handling - If the product includes a trial period and the subscriber has not used it before, the next activation date is set by adding the trial period days. If the trial period has already been used, the activation date is set according to the standard valid period days.

/api/Subscription/RemoveSubscription endpoint with parameters phoneNumber (in format 374XXXXXXXX) and productName (means product key, which we will provide). We check subscription existence and disable subscription.  

### Webhook Events

The system will fire webhook events in the following cases:

**New Subscription Activation**

When a subscriber attempts to enable a subscription from our side, we trigger a request to:
https://partnerUrl/ExtAppPartneerProductActivationRequest 

**Successful Auto-Prolongation**

When a subscription is successfully auto-prolonged, we trigger a request to:
https://partnerUrl/ExtAppPartneerProductActivation 

**Failed Auto-Prolongation (Subscription Removal)**

When the system cannot auto-prolong a subscription during the lifecycle process, we trigger a remove subscription event to:
https://partnerUrl/ExtAppPartneerProductRemove 

This may occur in the following cases:

The subscriber has been changed or does not exist.

TryActivationCount has exceeded its limit. This counter increases in the following situations:

Charging the subscriber fails.    

Subscriber is inactive.

For All of these requests request body is

Content-Type: application/json
```
{
  "phoneNum": "string",
  "productCode": "string"
}
```

Also for each URL we will provide secret key, which is used for signature in header “X-Signature” which is encrypted by using HMACSHA256.

Example of code in c# is here

```
var signature = Request.Headers["X-Signature"].ToString();
var computedSignature = CalculateHmacSha256(jsonPayload, " secret key");
if (signature != computedSignature)
  return Unauthorized("Invalid signature.");
```

```
private string CalculateHmacSha256(string input, string key)
{
     using (var hmac = new HMACSHA256(System.Text.Encoding.UTF8.GetBytes(key)))
     {
         var hashBytes = hmac.ComputeHash(System.Text.Encoding.UTF8.GetBytes(input));
         return BitConverter.ToString(hashBytes).Replace("-", "").ToLowerInvariant();
     }
}
```

Subscription endpoints are described in points   3.5 3.6  3.7 3.4 

You can view the workflow diagram on our Miro board at: https://miro.com/app/board/uXjVJ_Degds=/
Access password: Viva_Subscription_Management

## 7. Partner BFF (viva-api) — Landing HTTP and Postman

Сервис **viva-api** (BFF) — HTTP на `extTransport` (по умолчанию `0.0.0.0:4000`). Проксирует Viva REST (`vivaApi` в `config/viva-api.yaml`) и принимает вебхуки §6.

**Postman:** `docs/viva-api-bff-landing.postman_collection.json`  
**Переменные:** `bff_base_url` (BFF), `viva_base_url` (Viva REST), `webhook_secret` (= `extTransport.secret`).

### 7.1 Поток лендинга (init → confirm)

1. **Init** — `POST /landing/init-subscription` (или `/landing/:locale/init-subscription`).  
   Тело: `productName` (обяз.), `skipConfirm: true` (обяз.), `phoneNum` **или** заголовок `X-MSISDN` / `X-Msisdn` / `X-Phone-Number`.  
   Вызов: Viva **InitSubscription** (OTP на стороне Viva). Подтверждение без OTP через init **не выполняется**.

2. **Confirm** — `POST /landing/confirm-subscription` (или `/landing/:locale/confirm-subscription`).  
   Тело: `phoneNum` (или MSISDN в заголовке), `productName`, `otp`, **`productCode`** (обяз. для заказа в Armor).  
   Опционально: `locale`, `smsScenario` в JSON.  
   Вызов: Viva **ConfirmSubscription**. При `resultCode == 0` — `order/create` (New). При `resultCode == 7` — ответ 200 без заказа (уже активна).

**GetSubscriberInfo** — только прямой вызов Viva REST (§3.1, папка «0. Viva REST» в Postman); BFF не экспонирует `/landing/.../subscriber-info`.

Локаль SMS (`order/customData.smsLocale`): сегмент `:locale` в пути (`en`, `ru`, `hy` и алиасы) или поле `locale` в JSON. Логика: `internal/app/landing.go` (`pickLocale`, `parseLocale`).

### 7.2 Маршруты BFF

| Method | Path | Назначение |
|--------|------|------------|
| POST | `/landing/init-subscription` | InitSubscription |
| POST | `/landing/confirm-subscription` | ConfirmSubscription + order/create |
| POST | `/landing/:locale/init-subscription` | Init + локаль в пути |
| POST | `/landing/:locale/confirm-subscription` | Confirm + локаль |
| POST | `/ExtAppPartneerProductActivationRequest` | Вебхук New (§6), **X-Signature** |
| POST | `/ExtAppPartneerProductActivation` | Вебхук Renew |
| POST | `/ExtAppPartneerProductRemove` | Вебхук Cancel / remove |

На `/landing/*` — CORS, **без** подписи. На вебхуки — **X-Signature** (HMAC-SHA256 hex lower от тела JSON и `extTransport.secret`).

Тело вебхука: `phoneNum`, `productCode`; опционально `locale`, `smsScenario` (для SMS после `order/completed`).

### 7.3 Конфиг и SMS

- `vivaApi` — `baseURL`, `userName`, `password` для `internal/vivaclient`.
- `smpp` — SMPP (`endpoint`, `auth`, опционально `template` для activation SMS). Тексты сценариев (`sms2`, `sms3`, `sms4`, `sms5`, `sms14`, `sms15`, `sms_deactivated`) заданы в коде (`internal/app/sms.go`).

### 7.4 Postman — папки коллекции

| Папка | Содержимое |
|-------|------------|
| **0. Viva REST** | Прямые вызовы §2–§3.7 (токен, Init/Confirm/Remove в query) |
| **1. Landing BFF** | Init / Confirm на `bff_base_url` (с опциональным `:locale`) |
| **2. Viva → BFF webhooks** | Три POST с auto **X-Signature** (pre-request) |

Порядок проверки лендинга: Init (`skipConfirm: true`) → ввести `otp_code` → Confirm с `productCode`.

### 7.5 Viva REST в Postman (§2–§3)

| Spec | Запрос в коллекции |
|------|-------------------|
| §2.1 | POST `/auth/token` |
| §2.4 | OPTIONS `/auth/token` |
| §3.1 | GET `/api/Subscriber/{msisdn}` |
| §3.2 | POST `/api/Subscriber/AddTariffPlanProduct` |
| §3.3 | POST `/api/Subscriber/AddChargeableProduct` |
| §3.4 | GET `/api/Subscription/GetProductsByPhoneNum?phoneNum=…` |
| §3.5 | POST `/api/Subscription/InitSubscription?phoneNum=…&productName=…` |
| §3.6 | POST `/api/Subscription/ConfirmSubscription?…&otp=…` |
| §3.7 | POST `/api/Subscription/RemoveSubscription?…` |

Коды результатов: §5. Вебхуки и сценарии подписки: §6.
