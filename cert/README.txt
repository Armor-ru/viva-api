Сюда положите файлы, выданные для доступа к NATS Armor (как в examples.js):

  tls.crt   — клиентский сертификат
  tls.key   — приватный ключ
  ca.crt    — CA для проверки сервера

Пути заданы в config/yandex-plus-api.yaml (intTransport.tls).
Не коммитьте секретные ключи в git.
