package gocrudify

import (
	"encoding/json"
	"fmt"
	"net/http"

	utils "github.com/Kavuti/go-service-utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/jmoiron/sqlx"
)

type Handler[T Entity] struct {
	service     service[T]
	idFieldInfo *CrudFieldValue
	nonIdFields []CrudFieldValue
}

func (h *Handler[T]) Search(w http.ResponseWriter, r *http.Request) ([]T, error) {
	return []T{}, nil
}

func (h *Handler[T]) Get(w http.ResponseWriter, r *http.Request) {
	defer utils.RecoverIfError(w, r)
	id := chi.URLParam(r, h.idFieldInfo.JsonName)
	ctx := r.Context()
	render.Render(w, r, *h.service.Get(&ctx, id))
	w.WriteHeader(http.StatusOK)
}

func (h *Handler[T]) Create(w http.ResponseWriter, r *http.Request) {
	defer utils.RecoverIfError(w, r)
	var request T
	err := json.NewDecoder(r.Body).Decode(&request)
	utils.CheckError(err)
	ctx := r.Context()
	render.Render(w, r, *h.service.Create(&ctx, &request))
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler[T]) Update(w http.ResponseWriter, r *http.Request) {
	defer utils.RecoverIfError(w, r)
	id := chi.URLParam(r, h.idFieldInfo.JsonName)
	var request T
	err := json.NewDecoder(r.Body).Decode(&request)
	utils.CheckError(err)
	ctx := r.Context()
	render.Render(w, r, *h.service.Update(&ctx, id, &request))
	w.WriteHeader(http.StatusOK)
}

func (h *Handler[T]) Delete(w http.ResponseWriter, r *http.Request) {
	defer utils.RecoverIfError(w, r)
	id := chi.URLParam(r, h.idFieldInfo.JsonName)
	ctx := r.Context()
	h.service.Delete(&ctx, id)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler[T]) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.Create)

	r.Route(fmt.Sprintf("/{%s}", h.idFieldInfo.JsonName), func(r chi.Router) {
		r.Get("/", h.Get)
		r.Put("/", h.Update)
		r.Delete("/", h.Delete)
	})

	return r
}

func Expose[T Entity](tableName string, db *sqlx.DB) *Handler[T] {
	initializeDB(db)
	handler := &Handler[T]{}

	handler.idFieldInfo = GetIdField[T]()
	if handler.idFieldInfo == nil {
		panic("Some entity is missing the `crud` tag")
	}

	handler.nonIdFields = GetNonIdFields[T]()
	if handler.idFieldInfo.Type.String() != "int64" && handler.idFieldInfo.Type.String() != "string" {
		panic(fmt.Sprintf("Some entity has the wrong id type: %s instead of int/string", handler.idFieldInfo.Type.String()))
	}

	handler.service = *NewService[T](tableName, handler.idFieldInfo, handler.nonIdFields)

	return handler
}
