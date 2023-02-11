package gocrudify

import (
	"github.com/go-chi/render"
	"github.com/jmoiron/sqlx"
)

type Entity interface {
	render.Renderer
	ValidateCreation(db *sqlx.Tx) error
	ValidateUpdate(db *sqlx.Tx) error
	ValidateDeletion(db *sqlx.Tx) error
}
