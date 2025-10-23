# Спецификация: CLI интерфейс

## Обзор

CLI интерфейс предоставляет пользователю командную строку для управления Stroidex, включая запуск/остановку мониторинга, ручную индексацию, просмотр статуса и конфигурацию системы.

## Требования

### Функциональные требования

#### 1. Основные команды
- **ID:** CLI-001
- **Описание:** Система должна поддерживать команды:
  - `monitor` - запуск мониторинга файловой системы
  - `index` - однократная индексация директории
  - `status` - отображение статуса системы
  - `config` - управление конфигурацией
  - `version` - отображение версии
  - `help` - справка

#### 2. Аргументы и флаги
- **ID:** CLI-002
- **Описание:** Система должна поддерживать:
  - `--config` - путь к файлу конфигурации
  - `--verbose` - детальный вывод
  - `--dry-run` - прогон без реальных действий
  - `--format` - формат вывода (json, table)
  - `--output` - файл для сохранения результата

#### 3. Пользовательский интерфейс
- **ID:** CLI-003
- **Описание:** Система должна поддерживать:
  - Progress bars для длительных операций
  - Понятный вывод статуса операций
  - Подтверждение опасных операций
  - Автодополнение команд в оболочке

#### 4. Обработка ошибок
- **ID:** CLI-004
- **Описание:** Система должна:
  - Показывать понятные сообщения об ошибках
  - Предлагать решения для типичных проблем
  - Логировать ошибки с детализацией
  - Корректно обрабатывать прерывания (Ctrl+C)

### Нефункциональные требования

#### Usability
- Время отклика CLI: < 100ms
- Понятные сообщения об ошибках
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
}

type CommandConfig struct {
    ConfigPath  string
    Verbose     bool
    DryRun      bool
    Format      string
    Output      string
    Timeout     time.Duration
}

// Основная структура команды
var rootCmd = &cobra.Command{
    Use:   "stroidex",
    Short: "Документационный индексатор",
    Long:  "Stroidex - система для мониторинга и индексации строительной документации",
    Version: version.Version,
}

// Команда мониторинга
var monitorCmd = &cobra.Command{
    Use:   "monitor [path]",
    Short: "Запустить мониторинг директории",
    Long:  "Запускает непрерывный мониторинг указанной директории на наличие изменений",
    Args:  cobra.MaximumNArgs(1),
    RunE:  runMonitor,
}

// Команда индексации
var indexCmd = &cobra.Command{
    Use:   "index [path]",
    Short: "Проиндексировать директорию",
    Long:  "Выполняет однократную индексацию всех документов в указанной директории",
    Args:  cobra.MaximumNArgs(1),
    RunE:  runIndex,
}
```

### Реализация команд

```go
func runMonitor(cmd *cobra.Command, args []string) error {
    config := getConfigFromContext(cmd.Context())

    var targetPath string
    if len(args) > 0 {
        targetPath = args[0]
    } else {
        // Использовать пути из конфигурации
        targetPath = config.Monitoring.Directories[0]
    }

    engine, err := NewCoreEngine(config)
    if err != nil {
        return fmt.Errorf("failed to create engine: %w", err)
    }

    // Установка обработчиков сигналов
    ctx, cancel := signal.NotifyContext(context.Background(),
        syscall.SIGINT, syscall.SIGTERM)
    defer cancel()

    // Запуск мониторинга
    if err := engine.Start(ctx); err != nil {
        return fmt.Errorf("failed to start monitoring: %w", err)
    }

    // Отображение статуса
    return showMonitorStatus(ctx, engine)
}

func runIndex(cmd *cobra.Command, args []string) error {
    config := getConfigFromContext(cmd.Context())

    targetPath := getIndexPath(args, config)

    // Проверка dry-run режима
    if config.CLI.DryRun {
        return showDryRunPlan(targetPath, config)
    }

    engine, err := NewCoreEngine(config)
    if err != nil {
        return fmt.Errorf("failed to create engine: %w", err)
    }

    progress := NewProgressBar()
    return engine.IndexDirectory(context.Background(), targetPath, progress)
}
```

### Форматирование вывода

```go
type OutputFormatter interface {
    FormatStatus(status *EngineStatus) (string, error)
    FormatStats(stats *EngineStats) (string, error)
    FormatError(err error) (string, error)
    FormatProgress(progress *ProgressInfo) (string, error)
}

type TableFormatter struct {
    writer io.Writer
}

type JSONFormatter struct {
    writer io.Writer
    pretty bool
}

func (f *TableFormatter) FormatStatus(status *EngineStatus) (string, error) {
    table := tablewriter.NewWriter(f.writer)
    table.SetHeader([]string{"Параметр", "Значение"})

    table.Append([]string{"Статус", string(status.Status)})
    table.Append([]string{"Время работы", status.Uptime.String()})
    table.Append([]string{"Обработано файлов", fmt.Sprintf("%d", status.TasksDone)})
    table.Append([]string{"Ошибок", fmt.Sprintf("%d", status.TasksFailed)})
    table.Append([]string{"Активные воркеры", fmt.Sprintf("%d", status.WorkersBusy)})
    table.Append([]string{"Размер очереди", fmt.Sprintf("%d", status.QueueSize)})

    table.Render()
    return "", nil
}

func (f *JSONFormatter) FormatStatus(status *EngineStatus) (string, error) {
    data, err := json.Marshal(status)
    if err != nil {
        return "", err
    }

    if f.pretty {
        var pretty bytes.Buffer
        json.Indent(&pretty, data, "", "  ")
        return pretty.String(), nil
    }

    return string(data), nil
}
```

### Progress bars

```go
type ProgressBar struct {
    bar    *progressbar.ProgressBar
    logger *zap.Logger
}

func NewProgressBar() *ProgressBar {
    return &ProgressBar{
        bar: progressbar.NewOptions64(
            100,
            progressbar.OptionSetDescription("Индексация"),
            progressbar.OptionSetWriter(os.Stderr),
            progressbar.OptionShowCount(),
            progressbar.OptionShowIts(),
            progressbar.OptionOnCompletion(func() {
                fmt.Fprintln(os.Stderr, "Завершено!")
            }),
        ),
    }
}

func (p *ProgressBar) Update(current, total int64, filename string) {
    if total != p.bar.GetMax64() {
        p.bar.ChangeMax64(total)
    }

    p.bar.Set64(current)
    p.bar.Describe(fmt.Sprintf("Обработка: %s", filename))
}
```

## API

### Команды и опции

```bash
# Основные команды
stroidex monitor [path]          # Запуск мониторинга
stroidex index [path]            # Индексация директории
stroidex status                  # Статус системы
stroidex config [action]         # Управление конфигурацией
stroidex version                 # Версия приложения

# Глобальные опции
--config string     Путь к файлу конфигурации (default "config.yaml")
--verbose           Детальный вывод
--dry-run           Прогон без реальных действий
--format string     Формат вывода: json|table (default "table")
--output string     Файл для сохранения результата
--timeout duration  Таймаут операций (default 30s)
```

### Команда monitor

```bash
stroidex monitor [path] [flags]

# Примеры:
stroidex monitor ./documents
stroidex monitor --config ./prod.yaml --verbose
stroidex monitor --dry-run  # Проверка конфигурации
```

**Флаги:**
- `--watch-interval duration` Интервал проверки (default 30s)
- `--exclude strings` Паттерны исключения файлов
- `--max-workers int` Максимальное количество воркеров

### Команда index

```bash
stroidex index [path] [flags]

# Примеры:
stroidex index ./documents
stroidex index --recursive --format json
stroidex index --dry-run --output plan.json
```

**Флаги:**
- `--recursive` Рекурсивная обработка поддиректорий
- `--patterns strings` Паттерны файлов для обработки
- `--batch-size int` Размер пакета обработки

### Команда status

```bash
stroidex status [flags]

# Примеры:
stroidex status
stroidex status --format json --output status.json
stroidex status --verbose
```

**Вывод включает:**
- Статус системы
- Количество обработанных файлов
- Текущую нагрузку
- Последние ошибки

### Команда config

```bash
stroidex config [action] [flags]

# Действия:
stroidex config show           # Показать текущую конфигурацию
stroidex config validate       # Валидировать конфигурацию
stroidex config init           # Создать шаблон конфигурации
stroidex config set key value  # Установить параметр
```

## Конфигурация

### Структура конфигурации CLI

```yaml
cli:
  default_format: "table"          # table|json
  colors_enabled: true             # цветной вывод
  progress_enabled: true           # progress bars
  auto_confirm_dangerous: false    # автоподтверждение опасных операций
  history_file: "~/.stroidex_history"
  completion_script: "~/.stroidex_completion"

logging:
  level: "info"                    # debug|info|warn|error
  format: "console"                # console|json
  file: ""                         # путь к файлу логов
  max_size: "100MB"
  max_backups: 5
  max_age: 30                      # дней
```

### Валидация конфигурации

```go
func ValidateCLIConfig(config *CLIConfig) error {
    var errors []string

    if config.DefaultFormat != "table" && config.DefaultFormat != "json" {
        errors = append(errors, "default_format must be 'table' or 'json'")
    }

    if config.Timeout < 0 {
        errors = append(errors, "timeout must be positive")
    }

    if len(errors) > 0 {
        return fmt.Errorf("validation errors: %s", strings.Join(errors, ", "))
    }

    return nil
}
```

## Тестирование

### Unit тесты
- Тестирование парсинга аргументов
- Тестирование форматирования вывода
- Тестирование валидации конфигурации
- Тестирование обработки ошибок

### Integration тесты
- Тестирование выполнения команд
- Тестирование с реальной файловой системой
- Тестирование прерывания команд

### E2E тесты
```bash
# Тестовый сценарий
stroidex config init
stroidex config validate
stroidex index ./test-docs --dry-run
stroidex status --format json
stroidex version
```

## Метрики

### Ключевые показатели
- Время выполнения команд
- Количество ошибок CLI
- Успешность операций
- Использование памяти

### Логирование
```go
// Успешное выполнение команды
log.Info("Command completed successfully",
    "command", cmd.Use(),
    "duration", time.Since(start),
    "args", args)

// Ошибка выполнения команды
log.Error("Command failed",
    "command", cmd.Use(),
    "error", err,
    "args", args)
```

## Зависимости

- `github.com/spf13/cobra` - CLI фреймворк
- `github.com/spf13/viper` - управление конфигурацией
- `github.com/spf13/pflag` - парсинг флагов
- `github.com/schollz/progressbar/v3` - progress bars
- `github.com/olekukonko/tablewriter` - таблицы
- `go.uber.org/zap` - логирование
- `github.com/fatih/color` - цветной вывод