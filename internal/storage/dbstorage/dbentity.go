// Модель хранения объектов в БД
package dbstorage

// UserURLEntity связка короткой и оригинальной ссылки
type UserURLEntity struct {
	ShortURL    string
	OriginalURL string
}

// UserURL модель пользовательских ссылок
type UserURL struct {
	UserID      string
	OriginalURL string
}
