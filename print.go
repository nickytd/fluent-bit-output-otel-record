package main

import "fmt"

func printValue(v interface{}) {
	switch val := v.(type) {
	case string:
		fmt.Printf("%s", val)
	case []byte:
		fmt.Printf("\"%s\"", string(val))
	case map[interface{}]interface{}:
		fmt.Printf("{")
		printMap(val)
		fmt.Printf("}")
	case []map[interface{}]interface{}:
		fmt.Printf("[")
		printMapSlice(val)
		fmt.Printf("]")
	case []interface{}:
		// Check if it contains maps
		if containsMaps(val) {
			fmt.Printf("[")
			for i, elem := range val {
				if i > 0 {
					fmt.Printf(", ")
				}
				printValue(elem)
			}
			fmt.Printf("]")
		} else if isBytes(val) {
			// Check if it's a byte array (all values are integers 0-255)
			fmt.Printf("\"%s\"", interfaceSliceToString(val))
		} else {
			fmt.Printf("[")
			for i, elem := range val {
				if i > 0 {
					fmt.Printf(", ")
				}
				printValue(elem)
			}
			fmt.Printf("]")
		}
	default:
		fmt.Printf("\"%v\"", val)
	}
}

func printMapSlice(slice []map[interface{}]interface{}) {
	for i, m := range slice {
		if i > 0 {
			fmt.Printf(", ")
		}
		fmt.Printf("{")
		printMap(m)
		fmt.Printf("}")
	}
}

func isBytes(slice []interface{}) bool {
	for _, v := range slice {
		switch num := v.(type) {
		case int:
			if num < 0 || num > 255 {
				return false
			}
		case int8:
			if num < 0 {
				return false
			}
		case int64:
			if num < 0 || num > 255 {
				return false
			}
		case uint8:
			// valid byte
		case uint64:
			if num > 255 {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func containsMaps(slice []interface{}) bool {
	for _, v := range slice {
		switch v.(type) {
		case map[interface{}]interface{}, map[string]interface{}:
			return true
		}
	}
	return false
}

func interfaceSliceToString(slice []interface{}) string {
	bytes := make([]byte, len(slice))
	for i, v := range slice {
		switch num := v.(type) {
		case int:
			bytes[i] = byte(num)
		case int8:
			bytes[i] = byte(num)
		case int64:
			bytes[i] = byte(num)
		case uint8:
			bytes[i] = num
		case uint64:
			bytes[i] = byte(num)
		}
	}
	return string(bytes)
}

func printMap(m map[interface{}]interface{}) {
	first := true
	for k, v := range m {
		if !first {
			fmt.Printf(", ")
		}
		first = false
		fmt.Printf("\"%v\": ", k)
		printValue(v)
	}
}
