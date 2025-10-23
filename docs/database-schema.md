# Единая схема базы данных StroiDok

## Обзор

Документ описывает полную схему базы данных PostgreSQL, используемую компонентами StroiDok. База данных служит единым хранилищем для Stroidex (который пишет данные) и StroiMCP (который читает данные).

**Название базы данных:** `stroidok`

## Расширения

```sql
-- Включаем необходимые расширения
CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS pg_trgm;
```

## Основные таблицы

### 1. documents - Основная таблица документов

```sql
CREATE TABLE documents (
    id SERIAL PRIMARY KEY,
    file_path TEXT UNIQUE NOT NULL,
    file_name TEXT NOT NULL,
    file_type TEXT NOT NULL,
    file_size BIGINT NOT NULL,
    content TEXT,
    metadata JSONB,
    processed_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL,
    modified_at TIMESTAMP NOT NULL,
    file_hash TEXT UNIQUE,
    status TEXT DEFAULT 'processed' CHECK (status IN ('processed', 'error', 'pending', 'deleted'))
);
```

**Назначение:** Хранение основной информации о документах
**Кто пишет:** Stroidex
**Кто читает:** Stroidex, StroiMCP

### 2. document_embeddings - Векторные представления

```sql
CREATE TABLE document_embeddings (
    id SERIAL PRIMARY KEY,
    document_id INTEGER REFERENCES documents(id) ON DELETE CASCADE,
    embedding vector(1536) NOT NULL,
    model_name TEXT NOT NULL,
    chunk_index INTEGER DEFAULT 0,
    chunk_text TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(document_id, model_name, chunk_index)
);
```

**Назначение:** Хранение эмбеддингов для семантического поиска
**Кто пишет:** Stroidex
**Кто читает:** Stroidex, StroiMCP

### 3. document_pages - Страницы документов (для PDF)

```sql
CREATE TABLE document_pages (
    id SERIAL PRIMARY KEY,
    document_id INTEGER REFERENCES documents(id) ON DELETE CASCADE,
    page_number INTEGER NOT NULL,
    page_content TEXT,
    page_metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);
```

**Назначение:** Хранение построчного содержимого PDF документов
**Кто пишет:** Stroidex
**Кто читает:** Stroidex, StroiMCP

### 4. document_sheets - Листы документов (для XLSX)

```sql
CREATE TABLE document_sheets (
    id SERIAL PRIMARY KEY,
    document_id INTEGER REFERENCES documents(id) ON DELETE CASCADE,
    sheet_name TEXT NOT NULL,
    sheet_content JSONB,
    row_count INTEGER,
    column_count INTEGER,
    created_at TIMESTAMP DEFAULT NOW()
);
```

**Назначение:** Хранение содержимого таблиц Excel
**Кто пишет:** Stroidex
**Кто читает:** Stroidex, StroiMCP

### 5. processing_errors - Ошибки обработки

```sql
CREATE TABLE processing_errors (
    id SERIAL PRIMARY KEY,
    file_path TEXT NOT NULL,
    error_type TEXT NOT NULL,
    error_message TEXT,
    stack_trace TEXT,
    occurred_at TIMESTAMP DEFAULT NOW(),
    resolved_at TIMESTAMP,
    resolution TEXT
);
```

**Назначение:** Логирование ошибок обработки документов
**Кто пишет:** Stroidex
**Кто читает:** Stroidex

## Индексы

### Индексы для таблицы documents

```sql
-- Основные индексы
CREATE INDEX documents_file_path_idx ON documents(file_path);
CREATE INDEX documents_file_type_idx ON documents(file_type);
CREATE INDEX documents_modified_at_idx ON documents(modified_at);
CREATE INDEX documents_status_idx ON documents(status);
CREATE INDEX documents_file_hash_idx ON documents(file_hash);
CREATE INDEX documents_processed_at_idx ON documents(processed_at);

-- Полнотекстовый поиск
CREATE INDEX documents_content_fts ON documents
USING gin(to_tsvector('russian', content));

-- Триграммный поиск для нечеткого поиска
CREATE INDEX documents_content_trgm_idx ON documents
USING gin(content gin_trgm_ops);

-- JSONB индексы
CREATE INDEX documents_metadata_idx ON documents USING gin(metadata);
```

### Индексы для таблицы document_embeddings

```sql
-- Векторный индекс для семантического поиска
CREATE INDEX document_embeddings_vector_idx ON document_embeddings
USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- Индексы для JOIN операций
CREATE INDEX document_embeddings_document_id_idx ON document_embeddings(document_id);
CREATE INDEX document_embeddings_model_name_idx ON document_embeddings(model_name);
```

### Индексы для других таблиц

```sql
-- Индексы для document_pages
CREATE INDEX document_pages_document_id_idx ON document_pages(document_id);
CREATE INDEX document_pages_page_number_idx ON document_pages(page_number);

-- Индексы для document_sheets
CREATE INDEX document_sheets_document_id_idx ON document_sheets(document_id);
CREATE INDEX document_sheets_sheet_name_idx ON document_sheets(sheet_name);

-- Индексы для processing_errors
CREATE INDEX processing_errors_file_path_idx ON processing_errors(file_path);
CREATE INDEX processing_errors_occurred_at_idx ON processing_errors(occurred_at);
CREATE INDEX processing_errors_error_type_idx ON processing_errors(error_type);
```

## Триггеры и функции

### Функция обновления modified_at

```sql
CREATE OR REPLACE FUNCTION update_modified_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.modified_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_documents_modified_at
    BEFORE UPDATE ON documents
    FOR EACH ROW EXECUTE FUNCTION update_modified_at_column();
```

### Функция обновления updated_at в document_embeddings

```sql
CREATE OR REPLACE FUNCTION update_embeddings_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_document_embeddings_updated_at
    BEFORE UPDATE ON document_embeddings
    FOR EACH ROW EXECUTE FUNCTION update_embeddings_updated_at();
```

## Типы данных и Enum

```sql
-- Типы документов
CREATE TYPE document_type AS ENUM (
    'pdf', 'docx', 'doc', 'xlsx', 'xls', 'txt', 'rtf', 'odt', 'ods'
);

-- Статусы обработки
CREATE TYPE processing_status AS ENUM (
    'pending', 'processing', 'processed', 'error', 'deleted'
);

-- Типы ошибок
CREATE TYPE error_type AS ENUM (
    'file_not_found', 'parse_error', 'size_limit', 'format_error', 'database_error'
);
```

## Вью для удобного доступа

### Вью для поиска документов

```sql
CREATE VIEW searchable_documents AS
SELECT
    d.id,
    d.file_name,
    d.file_path,
    d.file_type,
    d.content,
    d.metadata,
    d.processed_at,
    d.modified_at,
    ts_rank_cd(
        to_tsvector('russian', d.content),
        plainto_tsquery('russian', '')
    ) as default_rank
FROM documents d
WHERE d.status = 'processed';
```

### Вью для статистики

```sql
CREATE VIEW document_stats AS
SELECT
    file_type,
    COUNT(*) as total_count,
    SUM(file_size) as total_size,
    AVG(file_size) as avg_size,
    MAX(processed_at) as last_processed
FROM documents
WHERE status = 'processed'
GROUP BY file_type;
```

## Роли и права доступа

### Роли

```sql
-- Роль для Stroidex (запись и чтение данных)
CREATE ROLE stroidex WITH LOGIN PASSWORD 'secure_password';

-- Роль для StroiMCP (только чтение данных)
CREATE ROLE stroimcp WITH LOGIN PASSWORD 'secure_password';

-- Администратор базы данных
CREATE ROLE stroidok_admin WITH LOGIN PASSWORD 'admin_password';
```

### Права доступа

```sql
-- Права для stroidex (индексатор + создание эмбеддингов)
GRANT SELECT, INSERT, UPDATE ON documents TO stroidex;
GRANT SELECT, INSERT, UPDATE ON document_pages TO stroidex;
GRANT SELECT, INSERT, UPDATE ON document_sheets TO stroidex;
GRANT SELECT, INSERT, UPDATE ON processing_errors TO stroidex;
GRANT SELECT, INSERT, UPDATE ON document_embeddings TO stroidex; -- Полные права на эмбеддинги
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO stroidex;

-- Права для stroimcp (поисковый сервер - только чтение)
GRANT SELECT ON documents TO stroimcp;
GRANT SELECT ON document_pages TO stroimcp;
GRANT SELECT ON document_sheets TO stroimcp;
GRANT SELECT ON document_embeddings TO stroimcp; -- Только чтение эмбеддингов
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO stroimcp;

-- Права для администратора
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO stroidok_admin;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO stroidok_admin;
```

## Оптимизации производительности

### Partitioning (для больших объемов данных)

```sql
-- Партиционирование таблицы documents по дате обработки
CREATE TABLE documents_partitioned (
    LIKE documents INCLUDING ALL
) PARTITION BY RANGE (processed_at);

-- Создание партиций по месяцам
CREATE TABLE documents_2024_01 PARTITION OF documents_partitioned
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

CREATE TABLE documents_2024_02 PARTITION OF documents_partitioned
    FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');
```

### Материализованные вью для агрегации

```sql
CREATE MATERIALIZED VIEW daily_stats AS
SELECT
    DATE(processed_at) as date,
    COUNT(*) as documents_processed,
    COUNT(DISTINCT file_type) as unique_types,
    SUM(file_size) as total_size,
    AVG(file_size) as avg_size
FROM documents
WHERE status = 'processed'
GROUP BY DATE(processed_at)
ORDER BY date DESC;

-- Создание индекса для материализованной вью
CREATE UNIQUE INDEX daily_stats_date_idx ON daily_stats(date);

-- Функция обновления статистики
CREATE OR REPLACE FUNCTION refresh_daily_stats()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY daily_stats;
END;
$$ LANGUAGE plpgsql;
```

## Резервное копирование и восстановление

### Полное резервирование

```bash
pg_dump -h localhost -U stroidok_admin -d stroidok > stroidok_backup_$(date +%Y%m%d).sql
```

### Восстановление

```bash
psql -h localhost -U stroidok_admin -d stroidok < stroidok_backup_$(date +%Y%m%d).sql
```

## Мониторинг и обслуживание

### Запросы для мониторинга

```sql
-- Статистика по размеру таблиц
SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size,
    pg_total_relation_size(schemaname||'.'||tablename) as size_bytes
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY size_bytes DESC;

-- Анализ медленных запросов
SELECT
    query,
    calls,
    total_time,
    mean_time,
    rows
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;

-- Мониторинг индексов
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC;
```

### Обслуживание таблиц

```sql
-- Автовакуум и автоанализ
ALTER TABLE documents SET (autovacuum_vacuum_scale_factor = 0.1);
ALTER TABLE documents SET (autovacuum_analyze_scale_factor = 0.05);

-- Ручное обновление статистики
ANALYZE documents;
ANALYZE document_embeddings;
```

## Примеры запросов

### Поисковые запросы

```sql
-- Полнотекстовый поиск
SELECT id, file_name, content
FROM searchable_documents
WHERE to_tsvector('russian', content) @@ to_tsquery('russian', 'строительство & фундамент');

-- Векторный поиск
SELECT d.id, d.file_name,
       e.embedding <=> '[0.1,0.2,0.3...]'::vector as similarity
FROM documents d
JOIN document_embeddings e ON d.id = e.document_id
WHERE e.model_name = 'text-embedding-ada-002'
ORDER BY similarity
LIMIT 10;

-- Поиск по метаданным
SELECT id, file_name, metadata
FROM documents
WHERE metadata->>'project' = 'ЖК Новатор'
  AND metadata->>'document_type' = 'чертеж';
```

Эта схема обеспечивает единую основу для работы обоих компонентов системы StroiDok с четким разделением прав доступа и ответственности.