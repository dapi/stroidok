# MCP Протокол и Транспортный уровень

## Обзор

Спецификация MCP (Model Context Protocol) протокола для интеграции StroiMCP с Claude Code и другими AI-ассистентами.

> **Терминология:** Все используемые термины определены в [`../../glossary.md`](../../glossary.md)

## MCP-1: MCP Protocol Specification

### Требования

#### MCP-1.1 Формат сообщений
- **Требование:** Поддержка JSON-RPC 2.0 формата
- **Функциональность:** Обработка запросов/ответов в формате JSON-RPC
- **Валидация:** Проверка структуры входящих сообщений
- **Контракт:** Совместимость с MCP спецификацией 1.0+

#### MCP-1.2 Обработка ошибок
- **Требование:** Graceful handling ошибок протокола
- **Функциональность:** Информативные сообщения об ошибках
- **Валидация:** Корректные коды ошибок MCP
- **Контракт:** Соответствие JSON-RPC error format

#### MCP-1.3 Валидация запросов
- **Требование:** Валидация параметров MCP запросов
- **Функциональность:** Проверка типов и значений
- **Валидация:** Early detection malformed запросов
- **Контракт:** Detailed error messages

### API

#### Request/Response Structure
```go
type MCPRequest struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      interface{} `json:"id"`
    Method  string      `json:"method"`
    Params  interface{} `json:"params,omitempty"`
}

type MCPResponse struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      interface{} `json:"id"`
    Result  interface{} `json:"result,omitempty"`
    Error   *MCPError   `json:"error,omitempty"`
}
```

#### Error Handling
```go
type MCPError struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
```

## MCP-2: Transport Layer Specification

### Требования

#### MCP-2.1 Обработка протокола
- **Требование:** Маршрутизация MCP запросов
- **Функциональность:** Диспетчеризация методов
- **Валидация:** Проверка доступности инструментов
- **Контракт:** Поддержка всех MCP методов

#### MCP-2.2 Управление соединениями
- **Требование:** Управление lifecycle соединений
- **Функциональность:** Graceful shutdown
- **Валидация:** Проверка состояния соединений
- **Контракт:** Reconnection логика

#### MCP-2.3 Форматирование ответов
- **Требование:** Стандартизированный формат ответов
- **Функциональность:** Консистентная структура
- **Валидация:** Валидация JSON формата
- **Контракт:** Соответствие MCP spec

### API

#### Transport Interface
```go
type MCPTransport interface {
    HandleRequest(ctx context.Context, req *MCPRequest) (*MCPResponse, error)
    RegisterTool(name string, tool MCPTool) error
    ListTools() []MCPTool
    Close() error
}
```

#### Tool Registry
```go
type MCPTool interface {
    Name() string
    Description() string
    Parameters() jsonschema.Schema
    Execute(ctx context.Context, params map[string]interface{}) (interface{}, error)
}
```

## Алгоритмы

### Request Processing Flow
1. **Parse Request** - Десериализация JSON-RPC
2. **Validate Request** - Проверка формата и параметров
3. **Route to Tool** - Диспетчеризация в соответствующий инструмент
4. **Execute Tool** - Выполнение бизнес-логики
5. **Format Response** - Сериализация результата
6. **Send Response** - Отправка ответа клиенту

### Error Handling Strategy
1. **Capture Error** - Перехват ошибок выполнения
2. **Classify Error** - Определение типа ошибки
3. **Format Error** - Создание MCP error response
4. **Log Error** - Логирование для отладки
5. **Return Response** - Отправка ошибки клиенту

## Тестирование

### Unit Tests
- Тестирование парсинга JSON-RPC сообщений
- Валидация параметров запросов
- Обработка ошибок протокола
- Форматирование ответов

### Integration Tests
- End-to-end MCP communication
- Тестирование всех инструментов
- Graceful shutdown сценарии
- Performance под нагрузкой

### Test Scenarios
- Valid request/response flow
- Invalid JSON handling
- Unknown method handling
- Parameter validation errors
- Network interruption handling
- Concurrent request processing

## Метрики

### Performance Metrics
- **Request Latency:** < 100ms (95th percentile)
- **Throughput:** > 100 requests/second
- **Error Rate:** < 0.1%
- **Connection Setup Time:** < 50ms

### Monitoring Metrics
- Active connections count
- Request rate per tool
- Error rate by type
- Response time distribution
- Memory usage per connection

## Конфигурация

```yaml
mcp:
  server:
    host: "0.0.0.0"
    port: 8080
    read_timeout: 30s
    write_timeout: 30s
    max_connections: 100

  protocol:
    max_request_size: 10MB
    enable_compression: true
    log_requests: true
    log_responses: false

  tools:
    auto_discover: true
    timeout: 30s
    max_concurrent: 10
```

## Зависимости

### External Dependencies
- Claude Code MCP client
- Anthropic API (для LLM операций)
- PostgreSQL (для данных)

### Internal Dependencies
- Search Service
- Document Service
- LLM Integration Service
- Configuration Service