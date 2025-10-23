# StroiDok - Интеллектуальная система анализа документов

## Обзор

**StroiDok** - это система интеллектуального анализа строительной документации на базе LLM, состоящая из двух основных компонентов:

- **Stroidex** - CLI приложение для мониторинга и индексации документов (смотри
  docs/stroidex)
- **StroiMCP** - MCP-сервер для поиска и взаимодействия с документами (смотри
  docs/StroiMCP)

Система обеспечивает естественный интерфейс для взаимодействия с документальной базой знаний через CLI приложение и MCP-протокол.

## Компоненты системы

### Stroidex (CLI приложение)
**Основные функции:**
- Поддержка форматов документов: PDF, DOC, DOCX, XLSX, TXT
- Отслеживание изменений файловой системы в реальном времени
- Автоматическое извлечение текстового содержимого
- Индексация и хранение обработанных данных в PostgreSQL
- Подготовка данных для поиска и анализа

**Stroidex НЕ делает:**
- Поиск документов
- Ответы на вопросы
- RAG функциональность
- Генерацию эмбеддингов

### StroiMCP (MCP-сервер)
**Основные функции:**
- Поиск документов (семантический и полнотекстовый)
- RAG (Retrieval-Augmented Generation)
- Text-to-SQL функциональность
- Интеграция с Claude Code через MCP-протокол
- Предоставление инструментов для AI-ассистентов

## Архитектура

```
Файловая система → Stroidex → PostgreSQL → StroiMCP → Claude Code
```

## Установка

### Предварительные требования
- Go 1.21 или выше
- PostgreSQL с pgvector
- Поддерживаемые платформы: macOS, Windows 10/11, Linux

### Сборка из исходников
```bash
# Клонирование репозитория
git clone https://github.com/your-org/stroidok.git
cd stroidok

# Сборка CLI приложения (Stroidex)
go build -o stroidex main.go

# Или через Makefile
make build
```

### Установка бинарных файлов
```bash
# Перемещение в PATH
sudo mv stroidex /usr/local/bin/

# Или добавление в текущую директорию
export PATH=$PATH:$(pwd)
```

## Использование

### Stroidex команды

#### Помощь и версия
```bash
# Показать справку
stroidex --help

# Показать версию
stroidex --version
```

#### Индексация документов
```bash
# Индексация текущей директории
stroidex index .

# Индексация указанных путей
stroidex index ./docs ./specifications

# Индексация с настройками
stroidex index . --workers 8 --batch-size 100

# Прогон (показать что будет проиндексировано)
stroidex index . --dry-run
```

#### Мониторинг изменений файловой системы
```bash
# Мониторинг текущей директории
stroidex monitor .

# Мониторинг с кастомным интервалом
stroidex monitor . --interval 30s

# Запуск как демон
stroidex monitor . --daemon

# Мониторинг с фильтрацией файлов
stroidex monitor . --pattern "*.pdf,*.docx,*.xlsx,*.txt"
```

#### Проверка статуса
```bash
# Комплексный статус
stroidex status

# Статус индексации
stroidex status --index

# Статус мониторинга
stroidex status --monitor

# Вывод в разных форматах
stroidex status --output json
```

### StroiMCP команды (планируется)

#### Запуск MCP-сервера
```bash
# Запуск с настройками по умолчанию
stroimcp

# Запуск с конфигурацией
stroimcp --config config.yaml

# Указание порта
stroimcp --port 8080
```

*Примечание: StroiMCP находится в разработке и пока не доступен.*

## Конфигурация

### Файл конфигурации Stroidex
```yaml
# config/stroidex.yaml
database:
  host: localhost
  port: 5432
  name: stroidok
  user: stroidok_writer
  password: ${DB_PASSWORD}

monitoring:
  directories:
    - /path/to/documents
    - /path/to/specs
  file_patterns:
    - "*.pdf"
    - "*.docx"
    - "*.xlsx"
    - "*.txt"
  interval: 30s
  max_file_size: 100MB

indexing:
  batch_size: 10
  workers: 4
  watch_interval: 30s
```

### Файл конфигурации StroiMCP
```yaml
# config/stroimcp.yaml
server:
  host: 0.0.0.0
  port: 8080
  read_timeout: 30s
  write_timeout: 30s

database:
  host: localhost
  port: 5432
  name: stroidok
  user: stroidok_reader
  password: ${DB_PASSWORD}

llm:
  provider: openai  # openai или anthropic
  api_key: ${LLM_API_KEY}
  model: gpt-4
  base_url: https://api.openai.com/v1

cache:
  redis:
    host: localhost
    port: 6379
    db: 0
    ttl: 5m
```

## Архитектура

### Структура проекта
```
stroidok/
├── main.go               # Основная точка входа
├── go.mod                # Go модуль
├── go.sum                # Зависимости
├── Makefile              # Сборочные скрипты
├── internal/
│   ├── cli/              # CLI интерфейс (Stroidex)
│   │   ├── root.go
│   │   ├── monitor.go
│   │   ├── index.go
│   │   ├── status.go
│   │   └── progress.go
│   ├── core/             # Основная логика
│   │   ├── engine.go
│   │   └── config.go
│   ├── parser/           # Парсеры документов
│   ├── indexer/          # Индексация
│   └── monitor/          # Мониторинг ФС
├── pkg/                  # Общие пакеты
├── config/               # Конфигурации
├── docs/                 # Документация
└── example/              # Примеры документов
```

## Интеграция с Claude Code

StroiMCP предоставляет MCP-инструменты для Claude Code:

- `search_documents` - семантический поиск по документам
- `get_document_content` - получение содержимого документа
- `list_recent_changes` - последние изменения в документах
- `text_to_sql` - запросы на естественном языке

## Разработка

### Сборка
```bash
# Сборка всех компонентов
make build

# Кросс-компиляция
make build-all

# Очистка
make clean
```

### Тестирование
```bash
# Запуск всех тестов
make test

# Запуск с покрытием
make test-coverage

# Интеграционные тесты
make test-integration
```

## Лицензия

[Укажите вашу лицензию]

---

Больше информации и отчеты об ошибках:
https://github.com/your-org/stroidok
