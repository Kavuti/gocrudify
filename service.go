package gocrudify

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	utils "github.com/Kavuti/go-service-utils"
)

type service[T Entity] struct {
	repository  repository[T]
	idFieldInfo *CrudFieldValue
	nonIdFields []CrudFieldValue
}

func (s *service[T]) Get(ctx *context.Context, id string) *T {
	tx := db.MustBegin()
	defer tx.Rollback()

	entity := s.repository.Get(ctx, tx, id)

	return entity
}

func (s *service[T]) Create(ctx *context.Context, entity *T) *T {
	err := utils.ValidateStruct(entity)
	utils.RaiseError(err, http.StatusBadRequest)

	tx := db.MustBegin()
	defer tx.Rollback()

	utils.RaiseError((*entity).ValidateCreation(tx), http.StatusBadRequest)

	newEntity := s.repository.Create(ctx, tx, entity)

	utils.CheckError(tx.Commit())

	return newEntity
}

func (s *service[T]) Update(ctx *context.Context, id string, entity *T) *T {
	utils.RaiseError(utils.ValidateStruct(entity), http.StatusBadRequest)

	tx := db.MustBegin()
	defer tx.Rollback()

	entityValue := reflect.ValueOf(*entity)

	s.repository.Get(ctx, tx, id)
	if id != fmt.Sprint(entityValue.FieldByName(s.idFieldInfo.Name).Interface()) {
		utils.RaiseError(errors.New("ids don't match"), http.StatusUnprocessableEntity)
	}

	utils.RaiseError((*entity).ValidateUpdate(tx), http.StatusBadRequest)

	s.repository.Update(ctx, tx, id, entity)

	utils.CheckError(tx.Commit())

	return entity
}

func (s *service[T]) Delete(ctx *context.Context, id string) {
	tx := db.MustBegin()
	defer tx.Rollback()

	entity := s.repository.Get(ctx, tx, id)
	utils.RaiseError((*entity).ValidateDeletion(tx), http.StatusBadRequest)
	s.repository.Delete(ctx, tx, id)

	utils.CheckError(tx.Commit())
}

func NewService[T Entity](tableName string, idFieldInfo *CrudFieldValue, nonIdFields []CrudFieldValue) *service[T] {
	service := &service[T]{}

	service.idFieldInfo = idFieldInfo
	service.nonIdFields = nonIdFields

	service.repository = repository[T](*NewRepository[T](tableName, idFieldInfo, nonIdFields))
	return service
}
