# Спецификация: Парсинг документов

## Обзор

Функция парсинга документов отвечает за извлечение текстового содержимого и метаданных из файлов различных форматов (PDF, DOCX, XLSX, TXT).

## Требования

### Функциональные требования

#### 1. Поддерживаемые форматы
- **ID:** DP-001
- **Описание:** Система должна поддерживать следующие форматы:
  - **PDF:** Извлечение текстового содержимого с сохранением структуры
  - **DOCX:** Извлечение текста с форматированием
  - **XLSX:** Извлечение данных из всех листов таблицы
  - **TXT:** Прямое чтение текстового содержимого

#### 2. Извлечение метаданных
- **ID:** DP-002
- **Описание:** Система должна извлекать следующие метаданные:
  - Имя файла
  - Размер файла
  - Дата создания и изменения
  - Тип файла (MIME type)
  - Для PDF: количество страниц, автор, заголовок
  - Для DOCX: автор, тема, ключевые слова
  - Для XLSX: количество листов, название листов

#### 3. Обработка ошибок
- **ID:** DP-003
- **Описание:** Система должна корректно обрабатывать:
  - Поврежденные файлы
  - Файлы с паролями
  - Файлы неподдерживаемых версий
  - Файлы с нестандартной кодировкой

#### 4. Валидация контента
- **ID:** DP-004
- **Описание:** Система должна проверять:
  - Минимальное количество текстового контента (> 10 символов)
  - Отсутствие бинарных данных в текстовом потоке
  - Корректность кодировки UTF-8

### Нефункциональные требования

#### Производительность
- Время обработки файла: < 5 секунд для файлов до 50MB
- Потребление памяти: < 100MB для файлов до 100MB
- Параллельная обработка до 10 файлов одновременно

#### Качество
- Точность извлечения текста: > 95%
- Сохранение структуры документа (абзацы, списки)
- Корректная обработка специальных символов

## Архитектура

### Интерфейс парсера

```go
type DocumentParser interface {
    ParseDocument(ctx context.Context, path string) (*Document, error)
    SupportedExtensions() []string
    ValidateFile(path string) error
}

type Document struct {
    Path       string                 `json:"path"`
    FileName   string                 `json:"file_name"`
    FileType   string                 `json:"file_type"`
    Size       int64                  `json:"size"`
    Content    string                 `json:"content"`
    Metadata   map[string]interface{} `json:"metadata"`
    Pages      []Page                 `json:"pages,omitempty"`
    Sheets     []Sheet                `json:"sheets,omitempty"`
    CreatedAt  time.Time              `json:"created_at"`
    ModifiedAt time.Time              `json:"modified_at"`
}

type Page struct {
    Number int    `json:"number"`
    Text   string `json:"text"`
}

type Sheet struct {
    Name   string   `json:"name"`
    Rows   []Row    `json:"rows"`
}
```

### Реализация для каждого формата

#### PDF Parser
```go
type PDFParser struct {
    maxFileSize int64
}

func (p *PDFParser) ParseDocument(ctx context.Context, path string) (*Document, error) {
    // 1. Валидация файла
    // 2. Извлечение текста через github.com/ledongthuc/pdf
    // 3. Извлечение метаданных
    // 4. Разделение на страницы
    // 5. Формирование результата
}
```

#### DOCX Parser
```go
type DOCXParser struct{}

func (p *DOCXParser) ParseDocument(ctx context.Context, path string) (*Document, error) {
    // 1. Распаковка DOCX архива
    // 2. Извлечение текста из document.xml
    // 3. Извлечение метаданных из core.xml
    // 4. Сохранение структуры абзацев
}
```

#### XLSX Parser
```go
type XLSXParser struct {
    maxRows int
}

func (p *XLSXParser) ParseDocument(ctx context.Context, path string) (*Document, error) {
    // 1. Открытие файла через github.com/tealeg/xlsx/v3
    // 2. Чтение всех листов
    // 3. Извлечение данных из ячеек
    // 4. Формирование структурированного контента
}
```

### Обработка ошибок

```go
type ParseError struct {
    Path    string
    Type    string
    Message string
    Cause   error
}

func (e *ParseError) Error() string {
    return fmt.Sprintf("Parse error for %s (%s): %s", e.Path, e.Type, e.Message)
}
```

## API

### Конфигурация

```yaml
parsing:
  pdf:
    extract_images: false
    preserve_layout: true
  docx:
    extract_styles: false
    preserve_formatting: true
  xlsx:
    max_rows_per_sheet: 10000
    include_empty_cells: false
  txt:
    encoding_detection: true
    max_file_size: 100MB
```

### Методы

```go
// Создание парсера для типа файла
func NewParser(fileType string) (DocumentParser, error)

// Определение типа файла по расширению
func DetectFileType(path string) string

// Валидация файла перед парсингом
func ValidateFile(path string) error

// Парсинг документа
func ParseDocument(ctx context.Context, path string) (*Document, error)
```

## Алгоритмы обработки

### PDF обработка
1. Открытие файла
2. Проверка пароля (если требуется)
3. Извлечение текста страница за страницей
4. Обработка кодировки
5. Извлечение метаданных
6. Формирование результата

### DOCX обработка
1. Валидация ZIP архива
2. Распаковка в память
3. Парсинг document.xml
4. Извлечение стилей (если требуется)
5. Извлечение метаданных из core.xml
6. Сборка текстового контента

### XLSX обработка
1. Открытие рабочей книги
2. Определение активных листов
3. Чтение данных построчно
4. Обработка типов данных
5. Сохранение структуры таблиц
6. Извлечение метаданных

## Тестирование

### Unit тесты
- Тестирование парсинга каждого формата
- Тестирование извлечения метаданных
- Тестирование обработки ошибок
- Тестирование на пограничных значениях

### Integration тесты
- Тестирование на реальных документах
- Тестирование производительности
- Тестирование на поврежденных файлах

### Тестовые данные
```
test-docs/
├── pdf/
│   ├── simple.pdf
│   ├── complex.pdf
│   └── password-protected.pdf
├── docx/
│   ├── simple.docx
│   ├── formatted.docx
│   └── with-images.docx
├── xlsx/
│   ├── simple.xlsx
│   ├── multi-sheet.xlsx
│   └── large-data.xlsx
└── txt/
    ├── utf-8.txt
    ├── windows-1251.txt
    └── large-file.txt
```

## Метрики

### Ключевые показатели
- Время обработки по форматам
- Точность извлечения текста
- Количество ошибок парсинга
- Размер извлеченного контента

### Логирование
```go
log.Info("Document parsed successfully",
    "path", path,
    "type", fileType,
    "size", doc.Size,
    "content_length", len(doc.Content))

log.Error("Failed to parse document",
    "path", path,
    "error", err,
    "type", fileType)
```

## Зависимости

- `github.com/ledongthuc/pdf` - PDF парсинг
- `github.com/sajari/docx` - DOCX парсинг
- `github.com/tealeg/xlsx/v3` - XLSX парсинг
- `golang.org/x/text` - работа с кодировками
- `github.com/spf13/viper` - конфигурация