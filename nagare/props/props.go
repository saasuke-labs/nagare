package props

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Props is the interface that all component props must implement
type Props interface {
	Parse(input string) error
}

// ParseProps is a helper function to parse props using struct tags
func ParseProps(input string, target interface{}) error {
	// Remove parentheses
	input = strings.Trim(input, "()")

	// Split by comma but respect quoted strings
	var pairs []string
	var current strings.Builder
	inQuotes := false

	// Always read input into the current buffer
	current.WriteString(strings.TrimSpace(input))

	// Split into pairs by commas outside quotes
	inQuotes = false
	var result strings.Builder
	for i := 0; i < current.Len(); i++ {
		char := current.String()[i]

		if char == '"' {
			inQuotes = !inQuotes
		}

		if char == ',' && !inQuotes {
			// Found a valid separator - add pair if non-empty
			if str := strings.TrimSpace(result.String()); str != "" {
				pairs = append(pairs, str)
			}
			result.Reset()
		} else {
			result.WriteByte(char)
		}
	}

	// Don't forget the last pair
	if str := strings.TrimSpace(result.String()); str != "" {
		pairs = append(pairs, str)
	}

	// Now parse each key:value pair
	for _, pair := range pairs {
		// Split at first colon only
		pair = strings.TrimSpace(pair)
		colonIndex := -1
		inQuotes = false
		for i := 0; i < len(pair); i++ {
			char := pair[i]
			if char == '"' {
				inQuotes = !inQuotes
			} else if char == ':' && !inQuotes && colonIndex == -1 {
				colonIndex = i
			}
		}

		if colonIndex == -1 {
			continue // No valid key:value separator found
		}

		key := strings.TrimSpace(pair[:colonIndex])
		value := strings.TrimSpace(pair[colonIndex+1:])

		// Debug output
		fmt.Printf("Parsing prop key=%q value=%q\n", key, value)

		// Clean up any unwanted spaces in the value
		value = strings.TrimPrefix(value, " ")
		value = strings.TrimSuffix(value, " ")

		// Handle quoted values
		if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
			value = strings.Trim(value, `"`)
		}

		// Debug after cleanup
		fmt.Printf("After cleanup: key=%q value=%q\n", key, value)

		// Use reflection to find matching field
		v := reflect.ValueOf(target).Elem()
		t := v.Type()

		var found bool
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if prop := field.Tag.Get("prop"); prop == key {
				fmt.Printf("Setting field %s with tag %q to %q\n", field.Name, prop, value)
				fieldValue := v.Field(i)

				// Handle different field types
				switch fieldValue.Kind() {
				case reflect.String:
					fieldValue.SetString(value)
				case reflect.Int:
					// Parse string as integer
					if intValue, err := strconv.Atoi(value); err == nil {
						fieldValue.SetInt(int64(intValue))
					} else {
						return fmt.Errorf("failed to parse %q as integer for field %s: %v", value, field.Name, err)
					}
				case reflect.Ptr:
					elemType := fieldValue.Type().Elem()
					switch elemType.Kind() {
					case reflect.String:
						strValue := value
						fieldValue.Set(reflect.ValueOf(&strValue))
					case reflect.Int:
						if intValue, err := strconv.Atoi(value); err == nil {
							fieldValue.Set(reflect.ValueOf(&intValue))
						} else {
							return fmt.Errorf("failed to parse %q as integer for field %s: %v", value, field.Name, err)
						}
					default:
						return fmt.Errorf("unsupported pointer type for field %s", field.Name)
					}
				}

				found = true
				break
			}
		}
		if !found {
			fmt.Printf("Warning: no field found with prop tag %q\n", key)
		}
	}

	return nil
}
