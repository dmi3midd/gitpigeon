### Ініціалізація
- Спочатку я ініціалізував проект використовуючи [go-blueprint](https://docs.go-blueprint.dev/)

### База даних
- Буду використовувати базу даних SQLite з [go-sqlite3](https://github.com/mattn/go-sqlite3) драйвером.
- Для міграцій буду використовувати [goose](https://pressly.github.io/goose/)
- Створив два файли для міграції: створення repository та subscription сутностей.
- Під час ініціалізації бази даних додав виклик міграції структури БД.
- Створиви дві стурктури які відповідають сутностям repository та subscription.

### Конфіги
- Створив конфіги(та структури для них) для проекту використовуючи [viper](https://github.com/spf13/viper)
- Додав конфіги для бази даних, API ключа, GitHub токена, інтервалу сканування та SMTP.

### Notifier
- Створив простий email notifier використовуючи [go-mail](https://github.com/wneessen/go-mail)
- Додав можливість надсилання повідомлень на email.
- Потребує доопрацювання.

### GitHub
- Створив клієнт для взаємодії з GitHub API використовуючи [go-github](https://github.com/google/go-github)
- Потребує доопрацювання.

