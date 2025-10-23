# Сервисы StroiMCP

## Обзор

Спецификация сервисов, обеспечивающих бизнес-логику MCP-сервера StroiMCP.

> **Терминология:** Все используемые термины определены в [`../../glossary.md`](../../glossary.md)

## SVC-1: Search Service

### Требования

#### SVC-1.1 Архитектура сервиса
- **Требование:** Централизованный сервис поиска
- **Функциональность:** Векторный и полнотекстовый поиск
- **Валидация:** Валидация поисковых запросов
- **Контракт:** Единый интерфейс поиска

#### SVC-1.2 Векторный поиск
- **Требование:** Семантический поиск через pgvector
- **Функциональность:** Поиск по эмбеддингам документов
- **Валидация:** Валидация векторных представлений
- **Контракт:** Оптимальная производительность поиска

#### SVC-1.3 Полнотекстовый поиск
- **Требование:** PostgreSQL FTS поиск
- **Функциональность:** Поиск по точному вхождению слов
- **Валидация:** Валидация FTS запросов
- **Контракт:** Высокая точность полнотекстового поиска

#### SVC-1.4 Гибридный поиск
- **Требование:** Комбинирование результатов поиска
- **Функциональность:** Алгоритм слияния и ранжирования
- **Валидация:** Валидация весов и порогов
- **Контракт:** Оптимальная релевантность результатов

### API

#### Search Service Interface
```go
type SearchService interface {
    Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error)
    HybridSearch(ctx context.Context, req *HybridSearchRequest) (*SearchResponse, error)
    VectorSearch(ctx context.Context, req *VectorSearchRequest) (*VectorSearchResponse, error)
    FullTextSearch(ctx context.Context, req *FullTextSearchRequest) (*FullTextSearchResponse, error)
}
```

#### Search Request Types
```go
type SearchRequest struct {
    Query         string
    Limit         int
    Offset        int
    Filters       map[string]interface{}
}

type HybridSearchRequest struct {
    SearchRequest
    SemanticWeight  float64
    FullTextWeight  float64
}

type VectorSearchRequest struct {
    SearchRequest
    QueryVector     []float64
    SimilarityThreshold float64
}

type FullTextSearchRequest struct {
    SearchRequest
    QueryTSVector   string
    RankingMethod   string
}
```

## SVC-2: Document Service

### Требования

#### SVC-2.1 Архитектура сервиса
- **Требование:** Сервис управления документами
- **Функциональность:** CRUD операции с документами
- **Валидация:** Валидация структуры документов
- **Контракт:** Консистентный API для документов

#### SVC-2.2 Управление метаданными
- **Требование:** Хранение и управление метаданными
- **Функциональность:** Индексация и поиск по метаданным
- **Валидация:** Валидация формата метаданных
- **Контракт:** Структурированные метаданные

#### SVC-2.3 Фрагментация документов
- **Требование:** Поддержка чанкинга контента
- **Функциональность:** Разбиение на логические фрагменты
- **Валидация:** Валидация границ фрагментов
- **Контракт:** Сохранение контекста

### API

#### Document Service Interface
```go
type DocumentService interface {
    GetDocument(ctx context.Context, id string) (*Document, error)
    GetDocumentContent(ctx context.Context, id string, opts *ContentOptions) (*DocumentContent, error)
    ListDocuments(ctx context.Context, filter *DocumentFilter) ([]*Document, error)
    GetDocumentChanges(ctx context.Context, hours int) ([]*DocumentChange, error)
}

type ContentOptions struct {
    Chunks     bool
    ChunkSize  int
    ChunkIndex int
    MaxLength  int
}

type DocumentFilter struct {
    Types      []string
    DateRange  *DateRange
    Metadata   map[string]interface{}
}
```

## SVC-3: Text-to-SQL Service

### Требования

#### SVC-3.1 Архитектура сервиса
- **Требование:** Сервис преобразования естественного языка в SQL
- **Функциональность:** LLM-based генерация запросов
- **Валидация:** Валидация и безопасность SQL
- **Контракт:** Безопасное выполнение запросов

#### SVC-3.2 LLM интеграция
- **Требование:** Интеграция с Anthropic API
- **Функциональность:** Генерация SQL из естественного языка
- **Валидация:** Валидация качества генерации
- **Контракт:** Высокая точность преобразования

#### SVC-3.3 Безопасность SQL
- **Требование:** Защита от SQL инъекций
- **Функциональность:** Санитизация и валидация запросов
- **Валидация:** Проверка прав доступа
- **Контракт:** Только read-only операции

### API

#### Text-to-SQL Service Interface
```go
type TextToSQLService interface {
    GenerateSQL(ctx context.Context, req *GenerateSQLRequest) (*GenerateSQLResponse, error)
    ValidateSQL(sql string) error
    ExecuteQuery(ctx context.Context, sql string, params []interface{}) (*QueryResult, error)
}

type GenerateSQLRequest struct {
    Query     string
    Context   map[string]interface{}
    Schema    *DatabaseSchema
}

type GenerateSQLResponse struct {
    SQL        string
    Confidence float64
    Explanation string
    Warnings   []string
}
```

## SVC-4: Caching Service

### Требования

#### SVC-4.1 Архитектура кэширования
- **Требование:** Многоуровневое кэширование
- **Функциональность:** L1/L2/L3 уровни кэша
- **Валидация:** Валидация целостности кэша
- **Контракт:** Консистентность данных

#### SVC-4.2 Redis кэш
- **Требование:** Распределенный кэш L2
- **Функциональность:** Хранение поисковых результатов
- **Валидация:** TTL и инвалидация кэша
- **Контракт:** Высокая доступность

#### SVC-4.3 In-memory кэш
- **Требование:** Локальный кэш L1
- **Функциональность:** Быстрый доступ к горячим данным
- **Валидация:** Размер и TTL лимиты
- **Контракт:** Минимальная задержка

### API

#### Caching Service Interface
```go
type CachingService interface {
    Get(ctx context.Context, key string) (interface{}, error)
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    InvalidatePattern(ctx context.Context, pattern string) error
}

type CacheLevel int

const (
    CacheLevelL1 CacheLevel = iota // In-memory
    CacheLevelL2                    // Redis
    CacheLevelL3                    // Database (query cache)
)
```

## Алгоритмы

### Hybrid Search Algorithm
1. **Query Analysis** - Анализ поискового запроса
2. **Vector Generation** - Генерация эмбеддинга запроса
3. **Parallel Search** - Параллельный векторный и полнотекстовый поиск
4. **Score Normalization** - Нормализация оценок релевантности
5. **Result Fusion** - Слияние результатов с весами
6. **Re-ranking** - Переранжирование по комбинированной релевантности
7. **Cache Storage** - Сохранение в кэш

### Content Retrieval Algorithm
1. **Document Lookup** - Поиск документа в БД
2. **Access Check** - Проверка прав доступа
3. **Content Loading** - Загрузка содержимого
4. **Chunking** - Разбиение на фрагменты (если требуется)
5. **Metadata Enrichment** - Обогащение метаданными
6. **Cache Update** - Обновление кэша
7. **Response Formatting** - Форматирование ответа

## Тестирование

### Unit Tests
- Тестирование каждого сервиса
- Валидация параметров
- Обработка ошибок
- Кэширование

### Integration Tests
- End-to-end сценарии
- Взаимодействие сервисов
- Тестирование кэша
- Производительность

### Performance Tests
- Нагрузочное тестирование
- Latency измерения
- Throughput тесты
- Stress тестирование

## Метрики

### Performance Metrics
- **Search Latency:** < 500ms (95th percentile)
- **Content Retrieval:** < 200ms
- **Cache Hit Rate:** > 80%
- **SQL Generation:** < 2s

### Quality Metrics
- **Search Relevance:** > 0.8
- **Content Accuracy:** 100%
- **SQL Success Rate:** > 95%
- **Cache Consistency:** 100%

## Конфигурация

```yaml
services:
  search:
    semantic_weight: 0.7
    fulltext_weight: 0.3
    max_results: 100
    cache_ttl: 5m

  document:
    max_content_length: 1MB
    default_chunk_size: 1000
    cache_ttl: 10m

  text_to_sql:
    timeout: 30s
    confidence_threshold: 0.7
    allowed_operations: ["SELECT"]
    cache_ttl: 15m

  cache:
    l1_max_size: 100MB
    l2_ttl: 1h
    l3_ttl: 24h
```