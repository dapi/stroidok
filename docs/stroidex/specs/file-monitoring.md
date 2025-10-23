# Спецификация: Мониторинг файловой системы

## Обзор

Функция мониторинга файловой системы отвечает за отслеживание изменений в указанных директориях и обнаружение документов, требующих обработки.

## Требования

### Функциональные требования

#### 1. Отслеживание изменений
- **ID:** FM-001
- **Описание:** Система должна отслеживать следующие события:
  - Создание новых файлов
  - Изменение существующих файлов
  - Удаление файлов
  - Переименование файлов
  - Создание/удаление поддиректорий

#### 2. Фильтрация файлов
- **ID:** FM-002
- **Описание:** Система должна обрабатывать только файлы с расширениями:
  - `.pdf` - документы PDF
  - `.docx` - документы Microsoft Word
  - `.xlsx` - таблицы Microsoft Excel
  - `.txt` - текстовые файлы

#### 3. Дебаунсинг событий
- **ID:** FM-003
- **Описание:** Система должна группировать множественные события одного файла в временном окне (30 секунд) для избежания дублирования обработки

#### 4. Рекурсивный обход
- **ID:** FM-004
- **Описание:** Мониторинг должен распространяться на все вложенные поддиректории

### Нефункциональные требования

#### Производительность
- Время реакции на изменение файла: < 5 секунд
- Потребление памяти: < 50MB
- Поддержка до 10 000 файлов в мониторинге

#### Надежность
- Автоматическое восстановление после ошибок файловой системы
- Логирование всех событий с ошибками
- Graceful shutdown с сохранением состояния

## Архитектура

### Компоненты

```go
type FileMonitor struct {
    watcher     *fsnotify.Watcher
    processor   DocumentProcessor
    debouncer   *EventDebouncer
    config      *MonitorConfig
    eventChan   chan FileEvent
    stopChan    chan struct{}
}

type FileEvent struct {
    Path        string
    Type        EventType // CREATE, MODIFY, DELETE, RENAME
    Timestamp   time.Time
    Size        int64
    IsDirectory bool
}

type MonitorConfig struct {
    Directories   []string
    FilePatterns  []string
    WatchInterval time.Duration
    MaxFileSize   int64
}
```

### Алгоритм работы

1. **Инициализация**
   - Создание fsnotify watcher
   - Добавление директорий для мониторинга
   - Запуск дебаунсера

2. **Обработка событий**
   ```mermaid
   graph TD
       A[Файловое событие] --> B{Проверка типа файла}
       B -->|Поддерживаемый| C[Дебаунсинг]
       B -->|Не поддерживаемый| D[Игнорировать]
       C --> E{Событие актуально?}
       E -->|Да| F[Отправка в обработку]
       E -->|Нет| G[Игнорировать]
       F --> H[Обработка документа]
   ```

3. **Обработка ошибок**
   - Потеря подключения к файловой системе
   - Отсутствие прав доступа
   - Превышение лимита файловых дескрипторов

## API

### Конфигурация

```yaml
monitoring:
  directories:
    - "/path/to/documents"
    - "/path/to/specs"
  file_patterns:
    - "*.pdf"
    - "*.docx"
    - "*.xlsx"
    - "*.txt"
  watch_interval: 30s
  max_file_size: 100MB
  exclude_patterns:
    - "*/temp/*"
    - "*/.git/*"
    - "*.tmp"
```

### Методы

```go
// Запуск мониторинга
func (fm *FileMonitor) Start(ctx context.Context) error

// Остановка мониторинга
func (fm *FileMonitor) Stop() error

// Добавление директории для мониторинга
func (fm *FileMonitor) AddDirectory(path string) error

// Удаление директории из мониторинга
func (fm *FileMonitor) RemoveDirectory(path string) error

// Получение статуса
func (fm *FileMonitor) GetStatus() *MonitorStatus
```

## Тестирование

### Unit тесты
- Тестирование фильтрации файлов
- Тестирование дебаунсинга
- Тестирование обработки ошибок

### Integration тесты
- Тестирование на реальной файловой системе
- Тестирование с большим количеством файлов
- Тестирование восстановления после сбоев

## Метрики

### Ключевые показатели
- Количество обработанных файлов
- Среднее время реакции на событие
- Количество ошибок мониторинга
- Размер очереди событий

### Логирование
```go
log.Info("File monitor started", "directories", config.Directories)
log.Debug("File event received", "path", event.Path, "type", event.Type)
log.Error("Failed to watch directory", "path", dirPath, "error", err)
```

## Зависимости

- `github.com/fsnotify/fsnotify` - кроссплатформенный файловый вотчер
- `github.com/spf13/viper` - управление конфигурацией
- `go.uber.org/zap` - структурированное логирование