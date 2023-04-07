package util

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
)

func ShowError(err error, message string, w http.ResponseWriter, statusCode int) {
	log.Printf("%s: %s %d", message, err, statusCode)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	var tmpMsg = err.Error()
	if len(tmpMsg) > 0 && tmpMsg[0] == '"' {
		tmpMsg = tmpMsg[1:]
	}
	if len(tmpMsg) > 0 && tmpMsg[len(tmpMsg)-1] == '"' {
		tmpMsg = tmpMsg[:len(tmpMsg)-1]
	}
	msg, err := json.Marshal(fmt.Sprintf("%s: %s", message, tmpMsg))
	if err != nil {
		return
	}
	w.Write(msg)
}

func ArrayFill(startIndex int, num int, value interface{}) map[int]interface{} {
	m := make(map[int]interface{})
	var i int
	for i = 0; i < num; i++ {
		m[startIndex] = value
		startIndex++
	}
	return m
}

func Contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func InterfaceToSlice(slice interface{}) []interface{} {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		panic("InterfaceSlice() given a non-slice type")
	}

	// Keep the distinction between nil and empty slice input
	if s.IsNil() {
		return nil
	}

	ret := make([]interface{}, s.Len())

	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}

	return ret
}
