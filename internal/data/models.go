package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Links  LinksModel
	Admin  AdminModel
	Tokens TokenModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Links:  LinksModel{DB: db},
		Admin:  AdminModel{DB: db},
		Tokens: TokenModel{DB: db},
	}
}
