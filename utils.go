package transit_go

import (
	"fmt"
	"reflect"
)

func mapToMapEntries(m interface{}) (mapEntries, error) {
	entries := NewSet()

	mapAsValue := reflect.ValueOf(m)

	if mapAsValue.Kind() != reflect.Map {
		return entries, fmt.Errorf("Can only process entries of a map, was a %+v", mapAsValue.Kind())
	}

	keys := mapAsValue.MapKeys()
	for _, key := range keys {
		value := mapAsValue.MapIndex(key).Interface()
		entries.Add(mapEntry{key: key.Interface(), value: value})
	}

	return entries, nil
}
