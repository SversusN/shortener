package dbstorage

type UserURLEntity struct {
	ShortURL    string
	OriginalURL string
}

type UserURL struct {
	UserID      string
	OriginalURL string
}
