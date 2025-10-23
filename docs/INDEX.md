# Индекс документации StroiDok

## 📋 Карта документации

Этот документ предоставляет навигационную карту всей документации проекта StroiDok.

## 🏗️ Структура документации

```
docs/
├── INDEX.md                           # Этот файл - навигация по документации
├── README.md                          # Обзор документации (если есть)
├── architecture/                     # Архитектурная документация
│   ├── overview.md                    # C4 модель, контекст системы
│   ├── sequence-diagrams.md          # Сценарии взаимодействия компонентов
│   ├── stroidex.md                   # Архитектура CLI приложения
│   └── stroimcp.md                   # Архитектура MCP сервера
├── specs/                            # Системные спецификации
│   ├── README.md                     # Обзор системных спецификаций
│   ├── indexing-storage-implementation-plan.md    # Индексация и хранение
│   └── process-orchestration-implementation-plan.md  # Оркестрация процессов
├── stroidex/                         # Документация Stroidex CLI
│   ├── README.md                     # Обзор Stroidex
│   └── specs/                        # Спецификации Stroidex
│       ├── README.md                 # Обзор спецификаций
│       ├── cli-implementation-plan.md
│       ├── cli-interface.md
│       ├── document-parsing.md
│       ├── document-parsing-implementation-plan.md
│       ├── file-monitoring.md
│       ├── file-monitoring-implementation-plan.md
│       ├── indexing-storage.md
│       └── process-orchestration.md
├── StroiMCP/                         # Документация StroiMCP сервера
│   ├── README.md                     # Обзор StroiMCP
│   ├── Features.md                   # Дополнительные возможности
│   └── specs/                        # Спецификации StroiMCP
│       ├── README.md                 # Обзор спецификаций
│       ├── command-search-specification.md
│       ├── mcp-protocol.md
│       ├── mcp-tools.md
│       └── services.md
├── component-responsibility-matrix.md # Матрица ответственности
├── database-schema.md                # Схема базы данных
├── glossary.md                       # Глоссарий терминов
├── mcp-claude-code-integration.md    # Интеграция с Claude Code
└── phase1-summary.md                 # Итоги фазы 1
```

## 🎯 Быстрый старт

### Для понимания проекта
1. **[Основной README](../README.md)** - обзор проекта
2. **[Архитектура системы](architecture/overview.md)** - контекст и компоненты
3. **[Глоссарий](glossary.md)** - терминология

### Для разработки
1. **[Матрица ответственности](component-responsibility-matrix.md)** - разделение функций
2. **[Системные спецификации](specs/)** - сквозная функциональность
3. **[Спецификации компонентов](stroidex/specs/, StroiMCP/specs/)** - детальная реализация

### Для интеграции
1. **[Интеграция с Claude Code](mcp-claude-code-integration.md)** - MCP протокол
2. **[Sequence диаграммы](architecture/sequence-diagrams.md)** - сценарии взаимодействия

## 📚 Детальная навигация

### 🏗️ Архитектура

#### [Архитектура системы](architecture/overview.md)
- C4 Level 1 контекст
- Внешние зависимости
- Основные компоненты
- Технологический стек
- Потоки данных

#### [Sequence диаграммы](architecture/sequence-diagrams.md)
- Сценарии индексации
- Процессы поиска
- MCP взаимодействия
- Обработка ошибок

#### [Архитектура компонентов]
- [Stroidex](architecture/stroidex.md) - CLI приложение
- [StroiMCP](architecture/stroimcp.md) - MCP сервер

### 📋 Спецификации

#### Системные спецификации ([docs/specs/](specs/))
Определяют взаимодействие между компонентами:
- [Индексация и хранение](specs/indexing-storage-implementation-plan.md) - архитектура данных
- [Оркестрация процессов](specs/process-orchestration-implementation-plan.md) - управление жизненным циклом

#### Спецификации Stroidex ([docs/stroidex/specs/](stroidex/specs/))
- [CLI интерфейс](stroidex/specs/cli-interface.md) - команды и опции
- [Мониторинг ФС](stroidex/specs/file-monitoring.md) - отслеживание изменений
- [Парсинг документов](stroidex/specs/document-parsing.md) - извлечение текста
- [Индексация](stroidex/specs/indexing-storage.md) - хранение в БД
- [Оркестрация](stroidex/specs/process-orchestration.md) - управление процессами

#### Спецификации StroiMCP ([docs/StroiMCP/specs/](StroiMCP/specs/))
- [MCP протокол](StroiMCP/specs/mcp-protocol.md) - интеграция
- [MCP инструменты](StroiMCP/specs/mcp-tools.md) - доступные инструменты
- [Поиск команд](StroiMCP/specs/command-search-specification.md) - функциональность поиска
- [Сервисы](StroiMCP/specs/services.md) - архитектура сервисов

### 🔧 Техническая документация

#### [База данных](database-schema.md)
- Полная схема PostgreSQL
- pgvector расширение
- Индексы и оптимизации
- Миграции

#### [Матрица ответственности](component-responsibility-matrix.md)
- Разделение функций между Stroidex и StroiMCP
- Потоки данных
- Права доступа
- Границы ответственности

#### [Глоссарий](glossary.md)
- Термины и определения
- Стандарты написания
- Акронимы и сокращения
- Правила использования

### 🔌 Интеграция

#### [Интеграция с Claude Code](mcp-claude-code-integration.md)
- MCP протокол
- Доступные инструменты
- Сценарии использования
- Примеры интеграции

## 🏷️ Теги документации

- **📋** - Спецификации и требования
- **🏗️** - Архитектура и дизайн
- **🔧** - Техническая реализация
- **🔌** - Интеграция и API
- **📚** - Документация для пользователей
- **🎯** - Быстрый старт и обзоры

## 📝 Поиск информации

### По типу задачи
- **Разработка CLI**: [stroidex/specs/](stroidex/specs/)
- **Разработка MCP**: [StroiMCP/specs/](StroiMCP/specs/)
- **Архитектурные решения**: [architecture/](architecture/)
- **Системная интеграция**: [specs/](specs/)

### По компоненту
- **Stroidex**: [docs/stroidex/](stroidex/)
- **StroiMCP**: [docs/StroiMCP/](StroiMCP/)
- **Общая архитектура**: [docs/architecture/](architecture/)
- **База данных**: [docs/database-schema.md](database-schema.md)

### По тематике
- **Поиск**: [StroiMCP/specs/command-search-specification.md](StroiMCP/specs/command-search-specification.md)
- **Индексация**: [specs/indexing-storage-implementation-plan.md](specs/indexing-storage-implementation-plan.md)
- **Мониторинг**: [stroidex/specs/file-monitoring.md](stroidex/specs/file-monitoring.md)
- **MCP**: [StroiMCP/specs/mcp-protocol.md](StroiMCP/specs/mcp-protocol.md)

---

## 📞 Поддержка

При возникновении вопросов или предложений по улучшению документации:
1. Проверьте [глоссарий](glossary.md) для уточнения терминов
2. Обратитесь к [матрице ответственности](component-responsibility-matrix.md) для понимания границ компонентов
3. Используйте [sequence диаграммы](architecture/sequence-diagrams.md) для понимания сценариев взаимодействия

**Версия индекса:** 1.0
**Последнее обновление:** 2024-10-23
**Ответственный:** Архитектор системы