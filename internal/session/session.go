package session

import "time"

type Session struct {
	Name               string
	DbURI              string
	MigrationFilesPath string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
