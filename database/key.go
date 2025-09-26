package database

import (
	"fmt"
	"reflect"
	"sort"
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

func extractTimestamp(key string) uint64 {
	_, uuidStr, found := strings.Cut(key, ":")
	if found == false {
		return uint64(key[0])
	}

	u, err := uuid.Parse(uuidStr)
	if err != nil {
		return 0
	}

	timestamp := u[0:8]
	var ts uint64
	for i := 0; i < 8; i++ {
		ts = ts<<8 | uint64(timestamp[i])
	}
	return ts
}

func SortKeys(keys []string) {
	sort.Slice(keys, func(i, j int) bool {
		tsI := extractTimestamp(keys[i])
		tsJ := extractTimestamp(keys[j])
		return tsI < tsJ
	})
}
