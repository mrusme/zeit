package database

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/google/uuid"
)

func NewKey(v Model) string {
	var modelName string
	if t := reflect.TypeOf(v); t.Kind() == reflect.Ptr {
		modelName = t.Elem().Name()
	} else {
		modelName = t.Name()
	}

	uu, err := uuid.NewV7()
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("%s:%s", strings.ToLower(modelName), uu.String())
}
