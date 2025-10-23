# MCP Инструменты

## Обзор

Спецификация MCP инструментов, предоставляемых StroiMCP для интеграции с Claude Code и другими AI-ассистентами.

> **Терминология:** Все используемые термины определены в [`../../glossary.md`](../../glossary.md)

## MT-1: Search Documents Tool

### Требования

#### MT-1.1 API Спецификация
- **Требование:** Семантический и полнотекстовый поиск документов
- **Функциональность:** Гибридный поиск с различными режимами
- **Валидация:** Валидация поисковых запросов
- **Контракт:** Консистентный формат результатов

#### MT-1.2 Параметры поиска
- **Требование:** Гибкие параметры поиска
- **Функциональность:** Фильтрация по типам, датам, источникам
- **Валидация:** Проверка корректности параметров
- **Контракт:** Предиктивные значения по умолчанию

#### MT-1.3 Режимы поиска
- **Требование:** Поддержка различных стратегий поиска
- **Функциональность:** Semantic, Full-text, Hybrid режимы
- **Валидация:** Валидация доступности режимов
- **Контракт:** Оптимальный выбор стратегии

### API

#### Search Request
```go
type SearchDocumentsRequest struct {
    Query         string            `json:"query"`
    Limit         int              `json:"limit,omitempty"`
    Offset        int              `json:"offset,omitempty"`
    SearchMode    SearchMode       `json:"search_mode,omitempty"`
    DocumentTypes []string         `json:"document_types,omitempty"`
    DateRange     *DateRange       `json:"date_range,omitempty"`
    Filters       map[string]interface{} `json:"filters,omitempty"`
}

type SearchMode string

const (
    SearchModeSemantic  SearchMode = "semantic"
    SearchModeFulltext  SearchMode = "fulltext"
    SearchModeHybrid    SearchMode = "hybrid"
)
```

#### Search Response
```go
type SearchDocumentsResponse struct {
    Results    []DocumentResult `json:"results"`
    Total      int             `json:"total"`
    QueryTime  time.Duration   `json:"query_time"`
    SearchInfo SearchInfo      `json:"search_info"`
}

type DocumentResult struct {
    ID           string            `json:"id"`
    Title        string            `json:"title"`
    Content      string            `json:"content,omitempty"`
    DocumentType string            `json:"document_type"`
    Score        float64           `json:"score"`
    Metadata     map[string]interface{} `json:"metadata"`
    Highlights   []string          `json:"highlights,omitempty"`
}
```

## MT-2: Get Document Content Tool

### Требования

#### MT-2.1 API Спецификация
- **Требование:** Получение полного содержимого документа
- **Функциональность:** Поддержка фрагментации контента
- **Валидация:** Проверка доступа к документу
- **Контракт:** Структурированный формат контента

#### MT-2.2 Параметры доступа
- **Требование:** Гибкие параметры извлечения
- **Функциональность:** Полный документ или фрагменты
- **Валидация:** Валидация ID документа
- **Контракт:** Оптимизация размера ответа

#### MT-2.3 Фрагментация
- **Требование:** Поддержка чанкинга контента
- **Функциональность:** Разбиение на страницы/секции
- **Валидация:** Валидация границ фрагментов
- **Контракт:** Сохранение контекста

### API

#### Document Request
```go
type GetDocumentContentRequest struct {
    DocumentID string `json:"document_id"`
    Chunks     bool   `json:"chunks,omitempty"`
    ChunkSize  int    `json:"chunk_size,omitempty"`
    ChunkIndex int    `json:"chunk_index,omitempty"`
    MaxLength  int    `json:"max_length,omitempty"`
}
```

#### Document Response
```go
type GetDocumentContentResponse struct {
    Document   DocumentMetadata `json:"document"`
    Content    string           `json:"content"`
    Chunks     []DocumentChunk  `json:"chunks,omitempty"`
    ChunkIndex int              `json:"chunk_index,omitempty"`
    TotalChunks int             `json:"total_chunks,omitempty"`
}

type DocumentChunk struct {
    Index   int    `json:"index"`
    Content string `json:"content"`
    StartPos int  `json:"start_pos"`
    EndPos   int  `json:"end_pos"`
}
```

## MT-3: List Recent Changes Tool

### Требования

#### MT-3.1 API Спецификация
- **Требование:** Отслеживание изменений в документах
- **Функциональность:** Список недавних добавлений/изменений
- **Валидация:** Валидация временных диапазонов
- **Контракт:** Сортировка по времени изменения

#### MT-3.2 Фильтрация изменений
- **Требование:** Гибкая фильтрация результатов
- **Функциональность:** По типам документов, источникам
- **Валидация:** Проверка корректности фильтров
- **Контракт:** Релевантные результаты

### API

#### Recent Changes Request
```go
type ListRecentChangesRequest struct {
    Hours      int               `json:"hours,omitempty"`
    Limit      int               `json:"limit,omitempty"`
    Offset     int               `json:"offset,omitempty"`
    DocumentTypes []string       `json:"document_types,omitempty"`
    ChangeType []ChangeType      `json:"change_type,omitempty"`
}

type ChangeType string

const (
    ChangeTypeCreated  ChangeType = "created"
    ChangeTypeUpdated  ChangeType = "updated"
    ChangeTypeDeleted  ChangeType = "deleted"
)
```

#### Recent Changes Response
```go
type ListRecentChangesResponse struct {
    Changes []DocumentChange `json:"changes"`
    Total   int              `json:"total"`
}

type DocumentChange struct {
    DocumentID   string    `json:"document_id"`
    Title        string    `json:"title"`
    ChangeType   ChangeType `json:"change_type"`
    ChangedAt    time.Time `json:"changed_at"`
    DocumentType string    `json:"document_type"`
    Metadata     map[string]interface{} `json:"metadata"`
}
```

## MT-4: Text-to-SQL Tool

### Требования

#### MT-4.1 API Спецификация
- **Требование:** Преобразование естественного языка в SQL
- **Функциональность:** Безопасное выполнение запросов
- **Валидация:** Валидация и санитизация SQL
- **Контракт:** Структурированные результаты

#### MT-4.2 Обработка запросов
- **Требование:** Интеграция с LLM для генерации SQL
- **Функциональность:** Контекст-aware преобразование
- **Валидация:** Проверка синтаксиса SQL
- **Контракт:** Безопасное выполнение

### API

#### Text-to-SQL Request
```go
type TextToSQLRequest struct {
    Query     string                 `json:"query"`
    Context   map[string]interface{} `json:"context,omitempty"`
    Limit     int                    `json:"limit,omitempty"`
    DryRun    bool                   `json:"dry_run,omitempty"`
}
```

#### Text-to-SQL Response
```go
type TextToSQLResponse struct {
    SQL        string        `json:"sql"`
    Results    []map[string]interface{} `json:"results,omitempty"`
    Columns    []ColumnInfo  `json:"columns,omitempty"`
    RowCount   int           `json:"row_count"`
    ExecTime   time.Duration `json:"exec_time"`
    Confidence float64       `json:"confidence"`
}

type ColumnInfo struct {
    Name string `json:"name"`
    Type string `json:"type"`
}
```

## Алгоритмы

### Search Algorithm
1. **Query Analysis** - Анализ поискового запроса
2. **Mode Selection** - Выбор оптимального режима поиска
3. **Vector Search** - Семантический поиск по эмбеддингам
4. **Full-text Search** - Полнотекстовый поиск
5. **Result Fusion** - Слияние и ранжирование результатов
6. **Relevance Scoring** - Расчет релевантности
7. **Result Formatting** - Форматирование ответа

### Content Retrieval Algorithm
1. **Document Lookup** - Поиск документа в БД
2. **Access Validation** - Проверка прав доступа
3. **Content Loading** - Загрузка содержимого
4. **Chunking** - Разбиение на фрагменты (если требуется)
5. **Metadata Enrichment** - Обогащение метаданными
6. **Response Formatting** - Форматирование ответа

## Тестирование

### Unit Tests
- Тестирование каждого инструмента
- Валидация параметров
- Обработка ошибок
- Форматирование ответов

### Integration Tests
- End-to-end поиск документов
- Извлечение содержимого
- Отслеживание изменений
- Text-to-SQL преобразование

### Performance Tests
- Поиск в больших коллекциях
- Извлечение больших документов
- Обработка множественных запросов
- Stress тестирование

## Метрики

### Performance Metrics
- **Search Latency:** < 1s (95th percentile)
- **Content Retrieval:** < 500ms
- **Changes Lookup:** < 200ms
- **Text-to-SQL:** < 2s

### Quality Metrics
- **Search Relevance:** > 0.8 average score
- **Content Accuracy:** 100%
- **SQL Generation Success:** > 95%
- **Error Rate:** < 0.1%

## Конфигурация

```yaml
mcp_tools:
  search:
    default_limit: 10
    max_limit: 100
    default_mode: "hybrid"
    cache_ttl: 5m

  content:
    max_chunk_size: 2000
    default_chunk_size: 1000
    max_content_length: 100000

  changes:
    default_hours: 24
    max_hours: 168
    cache_ttl: 1m

  text_to_sql:
    timeout: 30s
    max_results: 1000
    confidence_threshold: 0.7
```