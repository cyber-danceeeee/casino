package content

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// ConfigurationValues builds the single live representation after validation.
// Reflection preserves integer identifiers and time pointers without JSON
// round-tripping them through floating point numbers.
func ConfigurationValues(configuration any, columns map[string]string) (map[string]any, error) {
	value := reflect.Indirect(reflect.ValueOf(configuration))
	if value.Kind() != reflect.Struct {
		return nil, fmt.Errorf("configuration must be a struct")
	}
	result := map[string]any{}
	typ := value.Type()
	for index := 0; index < value.NumField(); index++ {
		field := typ.Field(index)
		column, ok := columns[field.Name]
		if !ok {
			continue
		}
		item := value.Field(index).Interface()
		if value.Field(index).Kind() == reflect.Map || value.Field(index).Kind() == reflect.Slice {
			encoded, err := json.Marshal(item)
			if err != nil {
				return nil, err
			}
			item = string(encoded)
		}
		result[column] = item
	}
	encoded, err := json.Marshal(configuration)
	if err != nil {
		return nil, err
	}
	result["configuration_json"] = string(encoded)
	return result, nil
}
