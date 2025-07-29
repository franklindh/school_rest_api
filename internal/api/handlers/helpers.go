package handlers

import (
	"errors"
	"reflect"
	"restapi/pkg/utils"
	"strings"
)

func CheckBlankFields(value any) error {
	val := reflect.ValueOf(value)
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if field.Kind() == reflect.String && field.String() == "" {
			// http.Error(w, "All field are required", http.StatusBadRequest)
			return utils.ErrorHandler(errors.New("all fields are required"), "All fields are required")
		}
	}
	return nil
}

func GetFieldNames(model any) []string {
	val := reflect.TypeOf(model)
	fields := []string{}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldToAdd := strings.TrimSuffix(field.Tag.Get("json"), ",omitempty")
		fields = append(fields, fieldToAdd)
	}
	return fields
}
