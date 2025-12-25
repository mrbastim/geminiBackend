# Настройка локальной LLM

## Обзор

Проект поддерживает локальную LLM через Ollama как отдельную модель наравне с Gemini. Просто укажите локальную модель в параметре `model` при запросе. Это позволяет:
- Снизить нагрузку на внешние API (Gemini)
- Не упираться в лимиты бесплатных ключей
- Обрабатывать данные локально без отправки в облако
- Переключаться между локальной и облачной моделью в рамках одного запроса

## Технические детали

**Модель:** qwen2:1.5b (Qwen2-1.5B-Instruct)
- Размер в памяти: ~1–1.3 ГБ
- Поддержка русского языка: отличная
- Требования: 2 ГБ RAM, 2 CPU-ядра (минимум)
- Контекст: 8k-32k токенов (~12-48k символов для русского)

## Лимиты по символам

### Gemini 2.5 Flash
- Контекст: ~1M токенов (~2-3M символов)
- Рекомендация: без ограничений со стороны кода

### qwen2:1.5b (локальная)
- Контекст: 8k-32k токенов (~12-48k символов)
- **Жесткий лимит на запрос: 10,000 символов** (настраивается через `LOCAL_LLM_MAX_CHARS`)
- Для текстов больше лимита автоматически применяется чанкование

## Чанкование (для больших документов)

Если текст превышает `LOCAL_LLM_MAX_CHARS`, он автоматически разбивается на части:
1. Разбивка по абзацам (`\n\n`)
2. Если абзац больше лимита — разбивка по предложениям (`. `)
3. Каждый чанк обрабатывается отдельно
4. Результаты склеиваются через двойной перенос строки

## Конфигурация

### Переменные окружения (.env)

```env
# URL Ollama-сервера (по умолчанию: http://ollama:11434)
OLLAMA_PORT=11434
LOCAL_LLM_ENDPOINT=http://ollama:11434

# Максимум символов на запрос (по умолчанию: 10000)
LOCAL_LLM_MAX_CHARS=10000
```

### Пример .env для локальной LLM

```env
PORT=8080
JWT_SECRET=your-secret-key-change-me
DB_PATH=data.db
GEMINI_API_KEY=your-gemini-key-as-fallback
ENV=prod
LOG_LEVEL=info
RATE_LIMIT_PER_MIN=true
TRUSTED_PROXIES=

# Локальная LLM
OLLAMA_PORT=11434
LOCAL_LLM_ENDPOINT=http://ollama:11434
LOCAL_LLM_MAX_CHARS=10000
```

> Если порт 11434 на хосте занят, измените `OLLAMA_PORT` (например, 11435) и перезапустите `docker-compose up -d --force-recreate ollama`.

## Запуск

### 1. Запустить Docker Compose

```bash
docker-compose up -d
```

Это создаст два сервиса:
- `ollama` — локальный LLM-сервер
- `app` — ваш Go-бэкенд

### 2. Загрузить модель в Ollama

После первого запуска нужно загрузить модель:

```bash
docker exec gemini-ollama ollama pull qwen2:1.5b
```

**Важно:** Загрузка модели может занять 5-10 минут в зависимости от скорости интернета (~1 ГБ).

### 3. Проверить статус

```bash
# Проверить, что модель загружена
docker exec gemini-ollama ollama list

# Проверить логи приложения
docker logs gemini-backend
```

## Тестирование

### Запрос к локальной модели

Просто укажите локальную модель в параметре `model`:

```bash
curl -X POST http://localhost:8080/user/ai/text \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Исправь текст: Привт, кк дла?",
    "model": "qwen2:1.5b"
  }'
```

### Запрос к Gemini

Для облачной модели ничего не меняется:

```bash
curl -X POST http://localhost:8080/user/ai/text \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Исправь текст: Привт, кк дла?",
    "model": "gemini-2.0-flash-exp"
  }'
```

### Поддерживаемые локальные модели

Модель определяется как локальная, если её название начинается с:
- `local` (например, `local` или `local-qwen`)
- `qwen` (например, `qwen2:1.5b`, `qwen2.5:3b`)
- `phi` (например, `phi3.5:mini`, `phi-3.1-mini`)
- `llama` (например, `llama3.2:3b`)
- `mistral` (например, `mistral:7b`)
- `gemma` (например, `gemma:2b`)

Все остальные модели направляются в Gemini API.

### Большой текст (автоматическое чанкование)

Если используете локальную модель и текст > 10k символов, он будет разбит на части автоматически.

Или просто закомментируйте:

```env
# USE_LOCAL_LLM=true
```

## Производительность

### Ожидаемая скорость (CPU 2 ядра, 2 ГБ RAM)

- Короткий текст (100-500 символов): 3-10 секунд
- Средний текст (1000-2000 символов): 15-30 секунд
- Большой текст (5000+ символов): 1-3 минуты (с чанкованием)

### Оптимизация

Для ускорения:
1. Увеличьте CPU-ядра (4-8 лучше)
2. Увеличьте RAM до 4 ГБ
3. Используйте SSD для хранения модели
4. Уменьшите `LOCAL_LLM_MAX_CHARS` до 5000-7000 для более быстрых ответов

## Альтернативные модели

Если нужно больше производительности или лучшее качество:

### phi-3.5-mini (легче, но слабее на русском)
```bash
# В запросе просто укажите эту модель
{
  "model": "phi3.5:mini",
  "prompt": "..."
}
```

### llama-3.2-3B (лучше качество, медленнее)
```bash
{
  "model": "llama3.2:3b",
  "prompt": "..."
}
```

**Важно:** После смены модели загрузите её:
```bash
docker exec gemini-ollama ollama pull llama3.2:3b
```

## Мониторинг

### Логи Ollama

```bash
docker logs -f gemini-ollama
```

### Использование памяти

```bash
docker stats
```

### Проверка загруженных моделей

```bash
docker exec gemini-ollama ollama list
```

## Устранение проблем

### Ollama не отвечает

```bash
# Перезапустить контейнер
docker restart gemini-ollama

# Проверить статус
docker exec gemini-ollama curl http://localhost:11434/api/tags
```

### Модель не загружена

```bash
# Загрузить заново
docker exec gemini-ollama ollama pull qwen2:1.5b
```

### Таймауты на больших текстах

Увеличьте таймаут в [local.go](internal/provider/gemini/local.go):

```go
httpClient: &http.Client{
    Timeout: 300 * time.Second, // было 120
}
```

### Out of Memory (OOM)

1. Уменьшите `LOCAL_LLM_MAX_CHARS` до 5000
2. Увеличьте RAM в docker-compose.yaml:
```yaml
deploy:
  resources:
    limits:
      memory: 3G  # было 2G
```

## Структура кода

- [config/config.go](config/config.go) — конфигурация
- [internal/provider/gemini/local.go](internal/provider/gemini/local.go) — клиент локальной LLM
- [internal/service/ai_service.go](internal/service/ai_service.go) — выбор провайдера
- [docker-compose.yaml](docker-compose.yaml) — деплой с Ollama

## Безопасность

Локальная LLM:
- ✅ Данные не покидают сервер
- ✅ Нет зависимости от внешних API
- ✅ Нет лимитов по запросам
- ⚠️ Требует собственных вычислительных ресурсов
- ⚠️ Медленнее облачных решений на GPU

## FAQ

**Q: Как система определяет, локальная это модель или облачная?**  
A: По префиксу названия. Если модель начинается с `qwen`, `phi`, `llama`, `mistral`, `gemma` или `local` — используется локальная Ollama. Иначе — Gemini API.

**Q: Можно ли использовать обе модели одновременно?**  
A: Да! Просто меняйте параметр `model` в запросе. Один запрос может идти к `qwen2:1.5b`, следующий — к `gemini-2.0-flash-exp`.

**Q: Можно ли использовать GPU?**  
A: Ollama поддерживает GPU. Для этого используйте образ с CUDA и пробросьте GPU в docker-compose.

**Q: Какие еще модели поддерживаются?**  
A: Все модели из Ollama библиотеки. Список: `docker exec gemini-ollama ollama list`

**Q: Как обновить модель?**  
A: `docker exec gemini-ollama ollama pull qwen2:1.5b` (автоматически обновит до последней версии)
