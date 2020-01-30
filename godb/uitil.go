package db

import (
	"fmt"
	"time"
)

// key and value extract
func whereExactKeyValues(where map[string]interface{}) (keys string, values []interface{}) {
	var i = 1
	if where != nil && len(where) > 0 {
		for k, v := range where {
			if keys != "" {
				keys = keys + " and "
			}
			if vv, ok := v.(string); ok {
				values = append(values, vv)                   //  value
				keys = fmt.Sprintf("%s %s = $%d", keys, k, i) // key like $#
			} else {
				values = append(values, v)                    // value
				keys = fmt.Sprintf("%s %s = $%d", keys, k, i) // key = $#
			}
			i = i + 1
		}
	}
	return
}

//convert milli time to date format
func TimestampDate(m int64) time.Time {
	return time.Unix(m/1e3, (m%1e3)*int64(time.Millisecond)/int64(time.Nanosecond))
}

// make time to milli format
func TimestampMilli(t time.Time) int64 {
	return unixMilli(t)
}

// make milli time
func unixMilli(t time.Time) int64 {
	return t.Round(time.Millisecond).UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

// return true if float64 type , else false
func IsFloat64(value interface{}) bool {
	if _, ok := value.(float64); ok {
		return true
	}
	return false
}

// return true if float32 type , else false
func IsFloat32(value interface{}) bool {
	if _, ok := value.(float32); ok {
		return true
	}
	return false
}

// return true if bool type , else false
func IsBoolean(value interface{}) bool {
	if _, ok := value.(bool); ok {
		return true
	}
	return false
}

// return true if int type , else false
func IsInt(value interface{}) bool {
	if _, ok := value.(int); ok {
		return true
	}
	return false
}

// return true if int64 type , else false
func IsInt64(value interface{}) bool {
	if _, ok := value.(int64); ok {
		return true
	}
	return false
}

// return true if map[string]interface{} type , else false
func IsMap(value interface{}) bool {
	if _, ok := value.(map[string]interface{}); ok {
		return true
	}
	return false
}
