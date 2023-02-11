package gocrudify

import (
	"reflect"
	"strings"
)

type CrudFieldValue struct {
	Name     string
	DbName   string
	JsonName string
	Type     reflect.Type
}

func GetIdField[T Entity]() *CrudFieldValue {
	entityType := reflect.TypeOf((*T)(nil)).Elem()
	for i := 0; i < entityType.NumField(); i++ {
		field := entityType.Field(i)
		if field.Tag.Get("crud") == "id" {
			return &CrudFieldValue{Name: field.Name, DbName: field.Tag.Get("db"), JsonName: field.Tag.Get("json"), Type: field.Type}
		}
	}
	return nil
}

func GetNonIdFields[T Entity]() []CrudFieldValue {
	entityType := reflect.TypeOf((*T)(nil)).Elem()
	result := []CrudFieldValue{}
	for i := 0; i < entityType.NumField(); i++ {
		field := entityType.Field(i)
		if field.Tag.Get("crud") != "id" {
			kind := field.Type.Kind().String()
			// For now slice and array are not considered, next a solution to manage relations will come out.
			if kind != "ptr" && kind != "struct" && kind != "slice" && kind != "array" {
				missingName := strings.ToLower(field.Name)
				dbName := field.Tag.Get("db")
				if dbName == "" {
					dbName = missingName
				}

				jsonName := field.Tag.Get("json")
				if jsonName == "" {
					jsonName = missingName
				}
				result = append(result, CrudFieldValue{
					Name:     field.Name,
					DbName:   dbName,
					JsonName: jsonName,
					Type:     field.Type,
				})
			}
		}
	}
	return result
}
