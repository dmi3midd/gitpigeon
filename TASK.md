### Ініціалізація
- Спочатку я ініціалізував проект використовуючи [go-blueprint](https://docs.go-blueprint.dev/)

### База даних
- Буду використовувати базу даних SQLite з [go-sqlite3](https://github.com/mattn/go-sqlite3) драйвером.
- Для міграцій буду використовувати [Goose](https://pressly.github.io/goose/)
- Створив два файли для міграції: створення repository та subscription сутностей.
- Під час ініціалізації бази даних додав виклик міграції структури БД.
- Створиви дві стурктури які відповідають сутностям repository та subscription.