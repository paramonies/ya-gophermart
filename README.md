# go-musthave-diploma-tpl

Шаблон репозитория для индивидуального дипломного проекта курса "Самостоятельный Go-разработчик"

# Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` - адрес вашего репозитория на Github без
   префикса `https://`) для создания модуля

# Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона выполните следующую команды:

```
git remote add -m master template https://github.com/yandex-praktikum/go-musthave-diploma-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/master .github
```

затем добавьте полученые изменения в свой репозиторий.

## Команды sql-migrate

### Добавление новой миграции

`sql-migrate new -env=local migration-name`

### Применение миграций

`sql-migrate up -env=local`

### Откат миграций

Откатиться на одну миграцию: `sql-migrate down -env=local`

Откатиться на N миграций: `sql-migrate down -env=local -limit=N`