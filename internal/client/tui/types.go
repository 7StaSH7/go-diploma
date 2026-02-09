package tui

const defaultSecretType = "note"
const actionUpdateSelected = "update_selected_secret"
const fieldFindTitle = "find_title"
const fieldFindTags = "find_tags"
const fieldFindDate = "find_date"

type tuiMode int

const (
	tuiModeMenu tuiMode = iota
	tuiModeForm
	tuiModeSelect
	tuiModeConfirmDelete
)

type tuiField struct {
	Key      string
	Label    string
	Hint     string
	Required bool
	Secret   bool
}

type tuiAction struct {
	ID          string
	Title       string
	Description string
	Fields      []tuiField
}

var tuiActions = []tuiAction{
	{
		ID:          "signup",
		Title:       "Регистрация",
		Description: "Создать пользователя и сохранить сессию",
		Fields: []tuiField{
			{Key: "login", Label: "Логин", Hint: "Например: alice", Required: true},
			{Key: "password", Label: "Пароль", Hint: "Минимум 8 символов", Required: true, Secret: true},
		},
	},
	{
		ID:          "signin",
		Title:       "Вход",
		Description: "Аутентифицировать пользователя и сохранить сессию",
		Fields: []tuiField{
			{Key: "login", Label: "Логин", Hint: "Ваш логин", Required: true},
			{Key: "password", Label: "Пароль", Required: true, Secret: true},
		},
	},
	{
		ID:          "search",
		Title:       "Поиск Секретов",
		Description: "Показать последние секреты или найти по названию, тегам и дате",
		Fields: []tuiField{
			{Key: fieldFindTitle, Label: "Название содержит", Hint: "Пусто = искать без фильтра по названию"},
			{Key: fieldFindTags, Label: "Теги через запятую", Hint: "Например: работа,почта"},
			{Key: fieldFindDate, Label: "С даты (ГГГГ-ММ-ДД)", Hint: "Например: 2026-02-09"},
		},
	},
	{
		ID:          "create",
		Title:       "Создать Секрет",
		Description: "Сохранить данные (шифрование выполнится автоматически)",
		Fields: []tuiField{
			{Key: "data", Label: "Данные", Hint: "Любой текст, который хотите сохранить", Required: true},
			{Key: "title", Label: "Заголовок", Hint: "Короткое имя секрета"},
			{Key: "tags", Label: "Теги через запятую", Hint: "Например: работа,почта"},
			{Key: "site", Label: "Сайт", Hint: "Например: https://example.com"},
		},
	},
	{
		ID:          "update",
		Title:       "Изменить Секрет",
		Description: "Сначала выбрать секрет, затем обновить его данные",
		Fields: []tuiField{
			{Key: fieldFindTitle, Label: "Название содержит", Hint: "Можно оставить пустым"},
			{Key: fieldFindTags, Label: "Теги через запятую", Hint: "Например: работа,почта"},
			{Key: fieldFindDate, Label: "С даты (ГГГГ-ММ-ДД)", Hint: "Например: 2026-02-09"},
		},
	},
	{
		ID:          "delete",
		Title:       "Удалить Секрет",
		Description: "Найти секрет и удалить его без ввода ID",
		Fields: []tuiField{
			{Key: fieldFindTitle, Label: "Название содержит", Hint: "Можно оставить пустым"},
			{Key: fieldFindTags, Label: "Теги через запятую", Hint: "Например: работа,почта"},
			{Key: fieldFindDate, Label: "С даты (ГГГГ-ММ-ДД)", Hint: "Например: 2026-02-09"},
		},
	},
	{
		ID:          "version",
		Title:       "Версия",
		Description: "Показать метаданные сборки",
	},
}

type operationResultMsg struct {
	ActionID string
	Output   string
	Err      error
}

type secretSelectionLoadedMsg struct {
	Items []secretOutputItem
	Err   error
}

type syncTickMsg struct{}

var updateSelectedAction = tuiAction{
	ID:          actionUpdateSelected,
	Title:       "Обновить выбранный секрет",
	Description: "Измените только нужные поля. Пусто = оставить как есть",
	Fields: []tuiField{
		{Key: "data", Label: "Новые данные", Hint: "Пусто = оставить текущие данные"},
		{Key: "title", Label: "Заголовок", Hint: "Пусто = оставить, '-' = очистить"},
		{Key: "tags", Label: "Теги через запятую", Hint: "Пусто = оставить, '-' = очистить"},
		{Key: "site", Label: "Сайт", Hint: "Пусто = оставить, '-' = очистить"},
	},
}

type tuiModel struct {
	mode             tuiMode
	cursor           int
	authorized       bool
	currentAction    tuiAction
	selectedSecret   secretOutputItem
	fieldIndex       int
	fieldValues      map[string]string
	selectionAction  string
	selectionFilters map[string]string
	selectionItems   []secretOutputItem
	selectionCursor  int
	input            string
	status           string
	output           string
	autoSync         bool
	syncInFlight     bool
}
