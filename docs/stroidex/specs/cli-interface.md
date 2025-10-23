# Спецификация: Unified CLI интерфейс

## Обзор

CLI интерфейс предоставляет пользователю единую команду для управления Stroidex в foreground режиме. Приложение выполняет непрерывную индексацию и мониторинг документов пока активно.

## Требования

### Функциональные требования

#### 1. Основная команда
- **ID:** CLI-001
- **Описание:** Система должна поддерживать единую команду:
  - `stroidex [path]` - запуск непрерывной индексации и мониторинга

#### 2. Опции команды
- **ID:** CLI-002
- **Описание:** Система должна поддерживать опции:
  - `--config` - путь к файлу конфигурации
  - `--once` - однократная индексация и выход
  - `--watch-interval` - интервал проверки изменений
  - `--patterns` - паттерны файлов для обработки
  - `--workers` - количество воркеров
  - `--batch-size` - размер пакета обработки
  - `--verbose` - детальный вывод
  - `--dry-run` - показ плана обработки без реальных действий

#### 3. Пользовательский интерфейс
- **ID:** CLI-003
- **Описание:** Система должна поддерживать:
  - Progress bars для обработки файлов
  - Real-time статус индексации
  - Информацию о обработанных файлах
  - Корректную обработку сигналов (SIGINT, SIGTERM)

#### 4. Обработка ошибок
- **ID:** CLI-004
- **Описание:** Система должна:
  - Показывать понятные сообщения об ошибках
  - Предлагать решения для типичных проблем
  - Логировать ошибки с детализацией при verbose режиме
  - Корректно останавливаться по сигналам

### Нефункциональные требования

#### Usability
- Время отклика CLI: < 100ms
- Понятные сообщения о статусе
- Поддержка Unicode в выводе
- Адаптация под размер терминала

#### Совместимость
- Поддержка Linux, macOS, Windows
- Минимальные зависимости
- Работа без установки (single binary)

## Архитектура

### Структура CLI

```go
type CLI struct {
    rootCmd *cobra.Command
    config  *Config
    engine  *CoreEngine
    logger  *zap.Logger
    signals chan os.Signal
}

type CommandConfig struct {
    ConfigPath     string
    Once           bool
    Verbose        bool
    DryRun         bool
    WatchInterval  time.Duration
    Patterns       []string
    Workers        int
    BatchSize      int
}

// Основная структура команды
var rootCmd = &cobra.Command{
    Use:   "stroidex [path]",
    Short: "Документационный индексатор",
    Long:  "Stroidex - система для непрерывной индексации и мониторинга строительной документации в foreground режиме",
    Args:  cobra.MaximumNArgs(1),
    RunE:  runStroidex,
    Version: version.Version,
}
```

### Реализация основной команды

```go
func runStroidex(cmd *cobra.Command, args []string) error {
    config := getConfigFromFlags(cmd)

    var targetPath string
    if len(args) > 0 {
        targetPath = args[0]
    } else {
        // Использовать пути из конфигурации
        targetPath = config.Processing.Directories[0]
    }

    // Проверка dry-run режима
    if config.Runtime.DryRun {
        return showDryRunPlan(targetPath, config)
    }

    engine, err := NewCoreEngine(config)
    if err != nil {
        return fmt.Errorf("failed to create engine: %w", err)
    }

    // Установка обработчиков сигналов для graceful shutdown
    ctx, cancel := signal.NotifyContext(context.Background(),
        syscall.SIGINT, syscall.SIGTERM)
    defer cancel()

    // Запуск в foreground режиме
    return engine.StartForeground(ctx, targetPath)
}
```

### Обработка сигналов

```go
func (e *CoreEngine) StartForeground(ctx context.Context, path string) error {
    // Initial indexing
    if err := e.indexDirectory(ctx, path); err != nil {
        return fmt.Errorf("initial indexing failed: %w", err)
    }

    // Start monitoring
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return fmt.Errorf("failed to create watcher: %w", err)
    }
    defer watcher.Close()

    // Start monitoring goroutine
    go e.monitorLoop(ctx, watcher, path)

    // Main loop - вывод статуса
    ticker := time.NewTicker(e.config.Processing.WatchInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            e.logger.Info("Received shutdown signal, stopping gracefully...")
            return e.shutdown()
        case <-ticker.C:
            e.showRealTimeStatus()
        }
    }
}
```

### Форматирование вывода

```go
type OutputFormatter interface {
    FormatStatus(status *EngineStatus) (string, error)
    FormatProgress(progress *ProgressInfo) (string, error)
    FormatFileProcessed(file *ProcessedFile) (string, error)
}

type RealTimeFormatter struct {
    writer io.Writer
    colors bool
}

func (f *RealTimeFormatter) FormatFileProcessed(file *ProcessedFile) (string, error) {
    status := "✓"
    if file.Error != "" {
        status = "✗"
    }

    return fmt.Sprintf("%s %s (%s) - %s",
        status, file.Path,
        humanize.Bytes(uint64(file.Size)),
        file.Status), nil
}

func (f *RealTimeFormatter) FormatStatus(status *EngineStatus) (string, error) {
    return fmt.Sprintf("\r📁 Indexed: %d | 🔄 Processing: %d | ⚠️  Errors: %d | 📊 Queue: %d",
        status.FilesIndexed,
        status.FilesProcessing,
        status.FilesError,
        status.QueueSize), nil
}
```

### Progress bars

```go
type ProgressBar struct {
    bar     *progressbar.ProgressBar
    logger  *zap.Logger
    stats   *RealTimeStats
}

func NewProgressBar() *ProgressBar {
    return &ProgressBar{
        bar: progressbar.NewOptions64(
            -1, // Unknown total for continuous mode
            progressbar.OptionSetDescription("Индексация документов"),
            progressbar.OptionSetWriter(os.Stderr),
            progressbar.OptionShowCount(),
            progressbar.OptionShowIts(),
            progressbar.OptionThrottle(100*time.Millisecond),
        ),
        stats: &RealTimeStats{},
    }
}

func (p *ProgressBar) UpdateFile(file *ProcessedFile) {
    p.stats.IncrementProcessed()
    p.bar.Describe(fmt.Sprintf("Обработано: %d | Текущий: %s",
        p.stats.TotalProcessed, filepath.Base(file.Path)))
}
```

## API

### Основная команда

```bash
# Основной режим - непрерывный мониторинг
stroidex [path] [flags]

# Примеры:
stroidex ./documents
stroidex . --verbose
stroidex ./docs --config ./prod.yaml
```

### Опции

```bash
--config string        Путь к файлу конфигурации (default "config.yaml")
--once                 Выполнить однократную индексацию и выйти
--watch-interval       Интервал проверки изменений (default 30s)
--patterns strings     Паттерны файлов (default "*.pdf,*.docx,*.xlsx,*.txt")
--workers int          Количество воркеров (default 4)
--batch-size int       Размер пакета обработки (default 10)
--verbose              Детальный вывод логов
--dry-run              Показать план обработки без действий
```

### Примеры использования

```bash
# Основной режим - непрерывный мониторинг
stroidex ./documents

# Однократная индексация
stroidex ./documents --once

# Детальный вывод с кастомными настройками
stroidex ./documents --verbose --watch-interval 10s --workers 8

# Проверка конфигурации
stroidex ./documents --dry-run

# Использование кастомного конфига
stroidex ./documents --config ./prod.yaml

# Индексация с фильтрацией файлов
stroidex ./documents --patterns "*.pdf,*.docx"
```

### Поведение в foreground режиме

1. **Запуск:** Приложение стартует и сразу начинает индексацию
2. **Мониторинг:** После индексации продолжает отслеживать изменения
3. **Статус:** В реальном времени показывает обработанные файлы
4. **Остановка:** Корректно останавливается по SIGINT/SIGTERM

### Пример вывода

```
🚀 Starting Stroidex v1.0.0
📂 Target directory: ./documents
🔧 Configuration: config.yaml
⚡ Workers: 4 | Batch size: 10 | Interval: 30s

📊 Scanning directory structure...
📁 Found 156 files to process

✓ ./documents/spec.pdf (2.3 MB) - Indexed
✓ ./docs/contract.docx (1.1 MB) - Indexed
✗ ./docs/corrupted.xlsx (0 bytes) - Error: invalid format
✓ ./images/plan.png (5.2 MB) - Skipped: unsupported format

📁 Indexed: 124 | 🔄 Processing: 2 | ⚠️  Errors: 1 | 📊 Queue: 0

🔄 Monitoring for changes... (Press Ctrl+C to stop)
✓ New file detected: ./docs/new_spec.pdf - Indexed
```

## Конфигурация

### Структура конфигурации

```yaml
# config/stroidex.yaml
database:
  host: localhost
  port: 5432
  name: stroidok
  user: stroidex
  password: ${DB_PASSWORD}

llm:
  provider: anthropic
  api_key: ${ANTHROPIC_API_KEY}
  model: claude-3-5-sonnet
  embedding_model: claude-3-5-sonnet

# Общие настройки обработки документов
processing:
  directories:
    - /path/to/documents
  file_patterns:
    - "*.pdf"
    - "*.docx"
    - "*.xlsx"
    - "*.txt"
  watch_interval: 30s
  max_file_size: 100MB
  batch_size: 10
  workers: 4

# Настройки foreground режима
runtime:
  verbose: false
  dry_run: false
  once: false

# CLI настройки
cli:
  colors_enabled: true
  progress_bars: true
  real_time_status: true
```

## Тестирование

### Unit тесты
- Тестирование парсинга аргументов
- Тестирование обработки опций
- Тестирование форматирования вывода
- Тестирование обработки сигналов

### Integration тесты
- Тестирование выполнения команды
- Тестирование с реальной файловой системой
- Тестирование прерывания по SIGINT/SIGTERM

### E2E тесты
```bash
# Тестовый сценарий
stroidex ./test-docs --dry-run
stroidex ./test-docs --once
stroidex ./test-docs --verbose --once
```

## Метрики

### Ключевые показатели
- Количество обработанных файлов
- Скорость обработки (файлы/сек)
- Количество ошибок
- Использование памяти
- Размер очереди задач

### Логирование
```go
// Запуск приложения
log.Info("Starting Stroidex",
    "version", version.Version,
    "target", targetPath,
    "config", configPath)

// Обработка файла
log.Debug("Processing file",
    "path", filePath,
    "size", fileSize,
    "duration", processingTime)

// Ошибки
log.Error("Failed to process file",
    "path", filePath,
    "error", err)
```

## Зависимости

- `github.com/spf13/cobra` - CLI фреймворк
- `github.com/spf13/viper` - управление конфигурацией
- `github.com/spf13/pflag` - парсинг флагов
- `github.com/schollz/progressbar/v3` - progress bars
- `github.com/fatih/color` - цветной вывод
- `go.uber.org/zap` - логирование
- `github.com/fsnotify/fsnotify` - мониторинг файловой системы