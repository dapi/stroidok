# Sequence диаграммы основных сценариев StroiDok

## Обзор

Документ содержит sequence диаграммы ключевых сценариев взаимодействия компонентов системы StroiDok.

> **Терминология:** Все используемые термины определены в [`глоссарии`](../glossary.md)

## Сценарий 1: Индексация нового документа

### 1.1 Обнаружение и индексация файла

```mermaid
sequenceDiagram
    participant FS as Файловая система
    participant SE as Stroidex Engine
    participant P as Parser
    participant LLM as Anthropic API
    participant DB as PostgreSQL stroidok
    participant User as Инженер ПТО

    Note over FS,DB: Автоматический процесс мониторинга

    FS->>SE: Новый файл обнаружен
    SE->>SE: Проверка формата файла
    SE->>P: Запрос на парсинг документа

    P->>P: Извлечение текста и метаданных
    P->>SE: Распарсенный документ

    SE->>DB: Сохранение в таблицу documents
    DB->>SE: Подтверждение сохранения

    SE->>LLM: Запрос на создание эмбеддингов
    LLM->>SE: Векторные представления
    SE->>DB: Сохранение эмбеддингов в document_embeddings
    DB->>SE: Подтверждение сохранения

    SE->>User: Уведомление об успешной индексации (опционально)
```

### 1.2 Пакетная обработка изменений

```mermaid
sequenceDiagram
    participant FS as Файловая система
    participant SE as Stroidex Engine
    participant Q as Queue Manager
    participant W as Worker Pool
    participant DB as PostgreSQL stroidok

    Note over FS,DB: Фоновая обработка

    loop Мониторинг изменений
        FS->>SE: События изменения файлов
        SE->>Q: Добавление задач в очередь
        Q->>W: Распределение задач по воркерам

        par Параллельная обработка
            W->>DB: Сохранение документа 1
        and
            W->>DB: Сохранение документа 2
        and
            W->>DB: Сохранение документа N
        end

        W->>Q: Подтверждение завершения
    end
```

## Сценарий 2: Семантический поиск через MCP

### 2.1 Поиск документов через Claude Code

```mermaid
sequenceDiagram
    participant User as Пользователь
    participant CC as Claude Code
    participant SM as StroiMCP
    participant LLM as Anthropic API
    participant DB as PostgreSQL stroidok
    participant Cache as Redis Cache

    User->>CC: "Найди требования к пожарной безопасности"
    CC->>SM: MCP запрос: search_documents(query="требования к пожарной безопасности")

    SM->>Cache: Проверка кэша поискового запроса
    alt Кэш найден
        Cache->>SM: Результаты из кэша
    else Кэш не найден
        SM->>LLM: Генерация эмбеддинга запроса
        LLM->>SM: Вектор представления запроса

        SM->>DB: Векторный поиск по document_embeddings
        DB->>SM: Релевантные документы с расстоянием

        SM->>DB: Полнотекстовый поиск по documents
        DB->>SM: FTS результаты

        SM->>SM: Слияние и ранжирование результатов
        SM->>Cache: Сохранение в кэш
    end

    SM->>CC: MCP ответ с документами
    CC->>User: Ответ с найденной информацией
```

### 2.2 Гибридный поиск (семантический + полнотекстовый)

```mermaid
sequenceDiagram
    participant User as Пользователь
    participant CC as Claude Code
    participant SM as StroiMCP
    participant DB as PostgreSQL stroidok
    participant Embeddings as document_embeddings
    participant Documents as documents

    User->>CC: Поисковый запрос
    CC->>SM: search_documents(search_mode="hybrid")

    par Параллельный поиск
        SM->>Embeddings: Векторный поиск
        Embeddings->>SM: Семантические результаты
    and
        SM->>Documents: Полнотекстовый поиск
        Documents->>SM: FTS результаты
    end

    SM->>SM: Алгоритм слияния результатов
    Note right of SM: Веса: семантика 0.7 + FTS 0.3
    SM->>SM: Переранжирование по комбинированной релевантности

    SM->>CC: Унифицированные результаты
    CC->>User: Ответ с документами
```

## Сценарий 3: Text-to-SQL запросы

### 3.1 Запрос на естественном языке

```mermaid
sequenceDiagram
    participant User as Пользователь
    participant CC as Claude Code
    participant SM as StroiMCP
    participant NL2SQL as Text-to-SQL Service
    participant LLM as Anthropic API
    participant DB as PostgreSQL stroidok

    User->>CC: "Покажи все PDF документы за последний месяц"
    CC->>SM: text_to_sql(query="все PDF документы за последний месяц")

    SM->>NL2SQL: Преобразование запроса
    NL2SQL->>LLM: Генерация SQL запроса
    Note right of LLM: Контекст: схема БД, типы полей
    LLM->>NL2SQL: SQL: SELECT * FROM documents WHERE file_type='pdf' AND created_at > NOW() - INTERVAL '1 month'

    NL2SQL->>NL2SQL: Валидация и санитизация SQL
    NL2SQL->>DB: Выполнение безопасного SQL
    DB->>NL2SQL: Результаты запроса

    NL2SQL->>SM: Форматированные результаты
    SM->>CC: MCP ответ с данными
    CC->>User: Ответ со списком документов
```

## Сценарий 4: Получение содержимого документа

### 4.1 Полный документ и фрагменты

```mermaid
sequenceDiagram
    participant User as Пользователь
    participant CC as Claude Code
    participant SM as StroiMCP
    participant DB as PostgreSQL stroidok
    participant FS as Файловая система

    User->>CC: "Покажи полный текст СП 4.13130.2013"
    CC->>SM: get_document_content(document_id="doc_12345", chunks=true)

    SM->>DB: Запрос метаданных документа
    DB->>SM: Информация о документе

    SM->>DB: Запрос содержимого из documents
    DB->>SM: Полный текст документа

    alt Документ разбит на страницы/листы
        SM->>DB: Запрос из document_pages/document_sheets
        DB->>SM: Структурированные фрагменты
    end

    SM->>SM: Форматирование фрагментов
    SM->>CC: MCP ответ с контентом
    CC->>User: Ответ с содержимым документа
```

## Сценарий 5: Оповещение об изменениях

### 5.1 Мониторинг последних изменений

```mermaid
sequenceDiagram
    participant User as Пользователь
    participant CC as Claude Code
    participant SM as StroiMCP
    participant DB as PostgreSQL stroidok
    participant SE as Stroidex

    User->>CC: "Что нового в документации за сегодня?"
    CC->>SM: list_recent_changes(hours=24)

    SM->>DB: Запрос недавних изменений
    DB->>SM: Список измененных документов

    alt Есть свежие индексированные документы
        SM->>CC: Список новых/измененных документов
        CC->>User: "Добавлены следующие документы: ..."
    else Нет изменений
        SM->>CC: Пустой результат
        CC->>User: "За сегодня изменений не было"
    end
```

## Сценарий 6: Обработка ошибок

### 6.1 Ошибка парсинга документа

```mermaid
sequenceDiagram
    participant FS as Файловая система
    participant SE as Stroidex Engine
    participant P as Parser
    participant DB as PostgreSQL stroidok
    participant LOG as Логирование

    FS->>SE: Новый файл (поврежденный PDF)
    SE->>P: Запрос на парсинг

    P->>P: Попытка чтения файла
    P->>P: Обнаружение ошибки парсинга

    P->>SE: Ошибка парсинга с деталями
    SE->>DB: Запись в processing_errors
    SE->>LOG: Логирование ошибки с stack trace

    SE->>SE: Продолжение обработки других файлов
    Note right of SE: Система не останавливается при ошибке одного файла
```

### 6.2 Недоступность базы данных

```mermaid
sequenceDiagram
    participant SE as Stroidex Engine
    participant DB as PostgreSQL stroidok
    participant Retry as Retry Mechanism
    participant User as Инженер ПТО

    SE->>DB: Попытка подключения
    DB-->>SE: Connection refused

    SE->>Retry: Запуск retry механизма
    Retry->>Retry: Экспоненциальный backoff: 1с, 2с, 4с, 8с...

    loop Повторные попытки
        Retry->>DB: Попытка подключения
        alt Успешное подключение
            DB->>Retry: Connection established
            Retry->>SE: Возобновление работы
            SE->>User: Уведомление о восстановлении
        else Провал подключения
            Retry->>Retry: Следующая попытка
        end
    end

    alt Все попытки исчерпаны
        Retry->>SE: Ошибка подключения
        SE->>User: Критическая ошибка: БД недоступна
    end
```

## Сценарий 7: Кэширование результатов

### 7.1 Многоуровневое кэширование

```mermaid
sequenceDiagram
    participant User as Пользователь
    participant CC as Claude Code
    participant SM as StroiMCP
    participant L1 as In-Memory Cache
    participant L2 as Redis Cache
    participant DB as PostgreSQL stroidok

    User->>CC: Повторный поисковый запрос
    CC->>SM: search_documents(...)

    SM->>L1: Проверка L1 кэша
    alt L1 cache hit
        L1->>SM: Результаты из памяти
    else L1 cache miss
        SM->>L2: Проверка L2 кэша (Redis)
        alt L2 cache hit
            L2->>SM: Результаты из Redis
            SM->>L1: Сохранение в L1
        else L2 cache miss
            SM->>DB: Запрос к базе данных
            DB->>SM: Результаты из БД
            SM->>L2: Сохранение в Redis (TTL: 5 минут)
            SM->>L1: Сохранение в памяти (TTL: 1 минута)
        end
    end

    SM->>CC: Результаты поиска
    CC->>User: Ответ с документами
```

## Метрики производительности сценариев

| Сценарий | Целевое время | Максимальное время | Частота выполнения |
|----------|---------------|-------------------|-------------------|
| Индексация документа | < 5 секунд | 30 секунд | По мере поступления |
| Семантический поиск | < 1 секунда | 3 секунды | Высокая |
| Text-to-SQL запрос | < 2 секунды | 10 секунд | Средняя |
| Получение документа | < 500мс | 2 секунды | Средняя |
| Проверка изменений | < 200мс | 1 секунда | Периодическая |

## Примечания к диаграммам

1. **Асинхронность:** Многие операции выполняются асинхронно для повышения производительности
2. **Отказоустойчивость:** Все критические операции имеют механизмы retry и graceful degradation
3. **Кэширование:** Многоуровневое кэширование снижает нагрузку на базу данных
4. **Масштабируемость:** Использование worker pool для параллельной обработки
5. **Безопасность:** Валидация всех входных данных и SQL запросов