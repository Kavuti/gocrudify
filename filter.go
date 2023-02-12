package gocrudify

import (
	"errors"
	"net/http"
	"reflect"
	"strings"

	utils "github.com/Kavuti/go-service-utils"
	sq "github.com/Masterminds/squirrel"
)

func ParseFilter(payload map[string]interface{}, nonIdFields []CrudFieldValue) sq.Sqlizer {

	expressions := sq.And{}

	for field, value := range payload {
		realField, err := findInSlice(field, nonIdFields)

		if err == nil {
			operator := strings.ReplaceAll(field, realField.JsonName, "")
			switch operator {
			case "GreaterThan":
				expressions = sq.And{expressions, sq.Gt{realField.DbName: value}}
			case "LessThan":
				expressions = sq.And{expressions, sq.Lt{realField.DbName: value}}
			case "GreaterThanOrEqualTo":
				expressions = sq.And{expressions, sq.GtOrEq{realField.DbName: value}}
			case "LessThanOrEqualTo":
				expressions = sq.And{expressions, sq.LtOrEq{realField.DbName: value}}
			case "In":
				expressions = sq.And{expressions, sq.Eq{realField.DbName: value}}
			case "Like":
				if reflect.TypeOf(value).Name() != "string" {
					utils.RaiseError(errors.New("wrong usage of Like filter: value must be string"), http.StatusBadRequest)
				}
				expressions = sq.And{expressions, sq.Like{realField.DbName: "%" + value.(string) + "%"}}
			}
		}
	}

	return expressions

}

func findInSlice(elem string, slice []CrudFieldValue) (CrudFieldValue, error) {
	for _, sliceElem := range slice {
		if strings.HasPrefix(elem, sliceElem.JsonName) {
			return sliceElem, nil
		}
	}

	return CrudFieldValue{}, errors.New("not found in slice")
}
