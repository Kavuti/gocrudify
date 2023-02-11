package gocrudify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	utils "github.com/Kavuti/go-service-utils"
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type repository[T Entity] struct {
	tableName   string
	idFieldInfo *CrudFieldValue
	nonIdFields []CrudFieldValue
}

func (r *repository[T]) Get(ctx *context.Context, tx *sqlx.Tx, id string) *T {
	var entity []T
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	idValue := r.getId(id)

	sql, args, err := psql.Select("*").From(r.tableName).Where(sq.Eq{r.idFieldInfo.Name: idValue}).ToSql()
	utils.CheckError(err)

	err = tx.Select(&entity, sql, args...)
	utils.CheckError(err)

	if len(entity) == 0 {
		utils.RaiseError(errors.New("not found"), http.StatusNotFound)
	}

	return &entity[0]
}

func (r *repository[T]) Create(ctx *context.Context, tx *sqlx.Tx, entity *T) *T {

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	subSql, _, err := psql.Select(fmt.Sprintf("nextval('seq_%s')", r.tableName)).ToSql()
	utils.CheckError(err)

	columns := []string{r.idFieldInfo.DbName}
	values := []interface{}{sq.Expr("(" + subSql + ")")}

	entityValue := reflect.ValueOf(*entity)

	for _, field := range r.nonIdFields {
		columns = append(columns, strings.ToLower(field.DbName))
		values = append(values, entityValue.FieldByName(field.Name).Interface())
	}

	sql, args, err := psql.Insert(r.tableName).Columns(columns...).Values(values...).Suffix(fmt.Sprintf("RETURNING \"%s\"", r.idFieldInfo.DbName)).ToSql()
	utils.CheckError(err)

	var idValue interface{}
	utils.CheckError(sqlx.GetContext(*ctx, tx, &idValue, sql, args...))

	var entityMap map[string]interface{}
	entityBytes, err := json.Marshal(entity)
	utils.CheckError(err)
	utils.CheckError(json.Unmarshal(entityBytes, &entityMap))

	entityMap[r.idFieldInfo.JsonName] = idValue

	var newEntity T
	entityBytes, err = json.Marshal(entityMap)
	utils.CheckError(err)
	utils.CheckError(json.Unmarshal(entityBytes, &newEntity))

	return &newEntity
}

func (r *repository[T]) Update(ctx *context.Context, tx *sqlx.Tx, id string, entity *T) *T {
	idValue := r.getId(id)

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	updateClause := psql.Update(r.tableName).Set(r.idFieldInfo.DbName, idValue)

	var entityMap map[string]interface{}
	entityBytes, err := json.Marshal(entity)
	utils.CheckError(err)
	err = json.Unmarshal(entityBytes, &entityMap)
	utils.CheckError(err)

	for _, field := range r.nonIdFields {
		value, ok := entityMap[field.JsonName]
		if ok {
			updateClause = updateClause.Set(strings.ToLower(field.DbName), value)
		}
	}

	updateClause = updateClause.Where(fmt.Sprintf("%s = ?", r.idFieldInfo.DbName), idValue)
	sql, args, err := updateClause.ToSql()
	utils.CheckError(err)
	_, err = tx.Exec(sql, args...)
	utils.CheckError(err)

	return entity
}

func (r *repository[T]) Delete(ctx *context.Context, tx *sqlx.Tx, id string) {
	idValue := r.getId(id)
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	sql, args, err := psql.Delete(r.tableName).Where(fmt.Sprintf("%s = ?", r.idFieldInfo.DbName), idValue).ToSql()
	utils.CheckError(err)

	_, err = tx.Exec(sql, args...)
	utils.CheckError(err)
}

func (r *repository[T]) getId(id string) interface{} {
	var idValue interface{}
	if r.idFieldInfo.Type.Name() == "int64" {
		intId, err := strconv.Atoi(id)
		utils.CheckError(err)
		idValue = intId
	} else if r.idFieldInfo.Type.Name() == "string" {
		idValue = id
	} else {
		utils.RaiseError(errors.New("invalid id type"), http.StatusInternalServerError)
	}
	return idValue
}

func NewRepository[T Entity](tableName string, idFieldInfo *CrudFieldValue, nonIdFields []CrudFieldValue) *repository[T] {
	repo := &repository[T]{tableName: tableName}

	repo.idFieldInfo = idFieldInfo
	repo.nonIdFields = nonIdFields

	return repo
}
