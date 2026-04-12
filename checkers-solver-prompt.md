# Промпт для Claude Code — Солвер шашечных задач

## Контекст задачи

Создай полноценное веб-приложение — **солвер задач по русским шашкам**. Цель: пользователь вручную расставляет позицию на доске, нажимает "Решить", и получает полное форсированное решение с деревом вариантов. Приложение аналогично lichess.org, но для шашечных этюдов.

---

## Стек

- **Бэкенд**: Go (net/http, без фреймворков)
- **Фронтенд**: React + TypeScript, Vite, shadcn/ui, Tailwind CSS
- **Шрифт**: TT Travels Next (подключить через @font-face или CDN если доступен, иначе fallback на Geist или similar geometric sans)
- **Темы**: light / dark с переключателем, CSS variables через shadcn/ui convention

---

## Структура проекта

```
checkers-solver/
├── backend/
│   ├── main.go
│   ├── engine/
│   │   ├── board.go        # представление доски
│   │   ├── moves.go        # генерация ходов
│   │   ├── eval.go         # функция оценки позиции
│   │   ├── search.go       # alpha-beta minimax
│   │   └── solver.go       # точка входа солвера
│   └── api/
│       └── handlers.go     # HTTP handlers
└── frontend/
    ├── src/
    │   ├── components/
    │   │   ├── Board.tsx
    │   │   ├── Toolbar.tsx
    │   │   ├── SolutionPanel.tsx
    │   │   ├── ThemeToggle.tsx
    │   │   └── PieceCounter.tsx
    │   ├── lib/
    │   │   ├── api.ts       # fetch к Go бэкенду
    │   │   └── types.ts
    │   ├── App.tsx
    │   └── main.tsx
    ├── tailwind.config.ts
    └── components.json      # shadcn config
```

---

## Бэкенд — Go движок

### Представление доски (board.go)

Русские шашки — доска 8x8, только тёмные клетки (32 клетки). Использовать **битборды** (bitboard representation):

```go
type Board struct {
    White     uint32  // битмаска белых шашек
    Black     uint32  // битмаска чёрных шашек
    WhiteKing uint32  // битмаска белых дамок
    BlackKing uint32  // битмаска чёрных дамок
    Turn      Color   // чей ход
}
```

Нумерация клеток: 0-31, слева направо, сверху вниз по тёмным клеткам.

```
Визуализация нумерации (тёмные клетки):
  а  b  c  d  e  f  g  h
8 [  ][ 0][  ][ 1][  ][ 2][  ][ 3]
7 [ 4][  ][ 5][  ][ 6][  ][ 7][  ]
6 [  ][ 8][  ][ 9][  ][10][  ][11]
5 [12][  ][13][  ][14][  ][15][  ]
4 [  ][16][  ][17][  ][18][  ][19]
3 [20][  ][21][  ][22][  ][23][  ]
2 [  ][24][  ][25][  ][26][  ][27]
1 [28][  ][29][  ][30][  ][31][  ]
```

Реализовать конвертацию в/из FEN-подобной нотации для русских шашек.

### Генерация ходов (moves.go)

Правила русских шашек:
- Шашки ходят по диагонали вперёд на одну клетку
- Бьют в любом направлении (включая назад), обязательно
- При множественном бое — продолжают серию, шашка снимается после завершения серии
- Дамка ходит на любое расстояние по диагонали, бьёт через любое расстояние
- При наличии боя — ход простой шашкой запрещён (обязательный бой)
- Если есть несколько вариантов боя — можно выбрать любой (в отличие от международных шашек)
- Шашка достигшая дамочного поля в середине боя — становится дамкой и продолжает бить как дамка

```go
type Move struct {
    From      int
    To        int
    Captures  []int  // последовательность срубленных шашек
    IsCapture bool
    Promoted  bool   // стала дамкой в этом ходу
}

func GenerateMoves(board Board) []Move
func GenerateCaptures(board Board) []Move  // только бои
func ApplyMove(board Board, move Move) Board
```

Обязательно реализовать **все правила русских шашек** корректно, включая:
- Турецкий удар (шашка не может бить одну и ту же шашку дважды в серии)
- Обязательное превращение в дамку
- Обязательный бой

### Функция оценки (eval.go)

```go
func Evaluate(board Board, perspective Color) int
```

Компоненты оценки (настраиваемые веса):

```go
const (
    PawnValue    = 100  // простая шашка
    KingValue    = 350  // дамка (в 3.5 раза ценнее)
    
    // Позиционные бонусы
    CenterBonus       = 10  // клетки d4,e4,d5,e5
    BackRankBonus     = 15  // защита своей последней линии
    AdvancedPawnBonus = 5   // за каждую линию продвижения
    EdgePenalty       = -8  // штраф за крайние вертикали
    
    // Структурные бонусы
    TempoBonus    = 3   // количество возможных ходов
    KingMobility  = 12  // подвижность дамок
)

// Эндшпильные корректировки
// Если осталось мало шашек — менять веса (дамки важнее)
```

Оценка должна быть **симметричной**: evaluate(board, White) == -evaluate(board, Black).

### Alpha-Beta поиск (search.go)

Реализовать **negamax с alpha-beta отсечением**:

```go
type SearchResult struct {
    Score    int
    BestMove Move
    PV       []Move  // principal variation — главный вариант
    Nodes    int64   // количество просмотренных узлов
    Depth    int
}

func Search(board Board, maxDepth int, timeLimit time.Duration) SearchResult
```

Обязательные оптимизации:

1. **Iterative Deepening** — искать с глубины 1, увеличивая до maxDepth. Позволяет прервать по времени с лучшим найденным результатом.

2. **Transposition Table** — хэш-таблица посещённых позиций:
```go
type TTEntry struct {
    Hash  uint64
    Depth int
    Score int
    Flag  TTFlag  // EXACT, LOWERBOUND, UPPERBOUND
    Move  Move
}
```
Использовать Zobrist hashing для быстрого обновления хэша позиции.

3. **Move Ordering** — сортировать ходы перед перебором:
   - Сначала бои (captures) — всегда первыми
   - Ходы из transposition table
   - Killer moves (ходы вызвавшие отсечение на том же уровне)
   - History heuristic
   
4. **Quiescence Search** — на листьях дерева продолжать поиск только боёв (чтобы избежать horizon effect)

5. **Null Move Pruning** — пропустить ход и проверить не слишком ли хороша позиция

### Солвер задач (solver.go)

Для шашечных **задач** (в отличие от игры) нужен особый режим:

```go
type SolveRequest struct {
    Board     Board
    GoalType  GoalType  // WIN, DRAW, MATE_IN_N
    MaxDepth  int       // максимальная глубина поиска
    TimeLimit int       // секунды
}

type SolveResult struct {
    Found      bool
    Score      int
    Depth      int      // на какой глубине найдено
    Moves      []Move   // главный вариант (форсированное решение)
    Tree       *Node    // дерево вариантов (для отображения)
    NodesCount int64
    TimeMs     int64
}

type Node struct {
    Move     Move
    Score    int
    Children []*Node
    IsBest   bool
}
```

Задачи решаются **на WIN** — найти форсированный выигрыш при любом ответе противника. Глубина поиска для задач: 5-20 полуходов (настраивается).

### HTTP API (handlers.go)

```
POST /api/solve
Content-Type: application/json

Request:
{
    "fen": "W:W21,22,23,24,25,26,27,28,29,30,31,32:B1,2,3,4,5,6,7,8,9,10,11,12",
    "turn": "white",
    "goal": "win",
    "max_depth": 15,
    "time_limit": 5
}

Response:
{
    "found": true,
    "score": 9999,
    "depth": 7,
    "moves": [
        {"from": 21, "to": 17, "captures": [], "notation": "e3-d4"},
        {"from": 8, "to": 12, "captures": [17], "notation": "d6:e5"}
    ],
    "tree": { ... },
    "nodes": 125847,
    "time_ms": 43
}

GET /api/health
```

CORS настроить для dev (localhost:5173).

---

## Фронтенд

### Тема и дизайн

Использовать shadcn/ui компоненты. Цветовая схема — нейтральная (zinc/slate), не синяя. Переключатель темы через next-themes или ручной CSS variables toggle.

**TT Travels Next** — подключить шрифт:
```css
@font-face {
    font-family: 'TT Travels Next';
    /* если недоступен через CDN — использовать Geist от Vercel как fallback */
}
```

Если TT Travels Next недоступен через open CDN, использовать **Geist** (vercel.com/font) — близкий по духу geometric sans.

### Компонент доски (Board.tsx)

```tsx
interface BoardProps {
    position: Position       // текущая расстановка
    onSquareClick: (idx: number) => void
    selectedTool: Tool
    highlightedSquares?: number[]  // подсветка хода решения
    lastMove?: Move
}
```

- SVG или CSS Grid — на выбор, но SVG предпочтительнее для точного рендера
- Тёмные клетки кликабельны
- Hover state на тёмных клетках
- Анимация появления/исчезновения шашек
- Подсветка клеток при отображении решения (зелёный for best move)
- Номера строк (1-8) и букв колонок (a-h) вокруг доски
- Шашки: CSS-styled divы с градиентом и тенью, дамка имеет корону/кольцо внутри

### Toolbar (Toolbar.tsx)

Инструменты расстановки:
- Белая шашка
- Чёрная шашка  
- Белая дамка
- Чёрная дамка
- Ластик
- (Активный инструмент подсвечен)

Переключатель хода: "Ход белых / Ход чёрных"

### SolutionPanel (SolutionPanel.tsx)

После решения показывать:
- Оценка позиции (выигрыш/ничья/проигрыш)
- Глубина найденного решения
- Главный вариант в нотации (ходы кликабельны — при клике подсвечиваются на доске)
- Время поиска и количество просмотренных узлов
- Кнопки: ← → для навигации по вариантам

### Нотация ходов

Русская шашечная нотация:
- Простой ход: `e3-d4`
- Бой: `e3:d4` или `e3:c5:a3` для серии

Конвертировать индекс клетки (0-31) в буквенно-цифровую нотацию.

---

## Детали реализации

### FEN для русских шашек

Стандарт PDN (Portable Draughts Notation):
```
W:W21,22,23,24,25,26,27,28,29,30,31,32:B1,2,3,4,5,6,7,8,9,10,11,12
```
- W в начале = ход белых (B = ход чёрных)
- W... = позиции белых шашек
- B... = позиции чёрных шашек
- K перед числом = дамка (K5 = дамка на клетке 5)

### Zobrist Hashing

Предгенерировать случайные 64-битные числа для каждой комбинации (клетка × тип фигуры). XOR при добавлении/удалении фигуры. Хэш цвета хода — отдельное число.

### Конфигурация движка

```go
type EngineConfig struct {
    MaxDepth      int           // default: 20
    TimeLimit     time.Duration // default: 5s
    TTSize        int           // записей в таблице транспозиций, default: 1<<22
    UseNullMove   bool          // default: true
    UseKillers    bool          // default: true
    UseHistory    bool          // default: true
    QuiesceDepth  int           // default: 10
}
```

---

## Что нужно реализовать в первую очередь

1. Корректная генерация всех ходов по правилам русских шашек (это самое важное — баги здесь убивают всё)
2. Базовый negamax с alpha-beta
3. HTTP endpoint `/api/solve`
4. Фронтенд с доской и вызовом API
5. Затем — оптимизации (TT, move ordering, iterative deepening)

---

## Тесты

Обязательно написать тесты для генерации ходов:

```go
func TestMandatoryCapture(t *testing.T)      // обязательный бой
func TestMultipleCapture(t *testing.T)       // серия боёв
func TestKingPromotion(t *testing.T)         // превращение в дамку
func TestKingMoves(t *testing.T)             // ходы дамки
func TestTurkishStrike(t *testing.T)         // турецкий удар
```

Взять несколько позиций из реальных задач и проверить что солвер находит правильное решение.

---

## Примечания

- Движок должен работать в отдельной горутине, HTTP handler — неблокирующий с context timeout
- Для задач глубина 5-15 полуходов решается за миллисекунды на хорошем alpha-beta
- Если решение не найдено за timeLimit — вернуть лучший найденный ход
- Логировать количество nodes/second для бенчмаркинга
- Фронт запускается на :5173, бэк на :8080
