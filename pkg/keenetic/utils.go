package keenetic

import (
	"bytes"
	"fmt"
	"github.com/Ponywka/go-keenetic-dns-router/pkg/errors/contextedError"
	"github.com/Ponywka/go-keenetic-dns-router/pkg/errors/parentError"
	"io"
	"net/http"
	"reflect"
)

func apiSyncRequest(method string, url string, data []byte, headers map[string]string) (resp *http.Response, body []byte, err error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		err = contextedError.NewFromFunc(&err, http.NewRequest)
		err = parentError.New("http creating error", &err)
		return
	}
	for key, val := range headers {
		req.Header.Add(key, val)
	}
	resp, err = client.Do(req)
	if err != nil {
		err = contextedError.NewFromFunc(&err, client.Do)
		err = parentError.New("client creating error", &err)
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		err = contextedError.NewFromFunc(&err, io.ReadAll)
		err = parentError.New("body reading error", &err)
		return
	}
	return
}

func parseCookies(rawCookies string) []*http.Cookie {
	header := http.Header{}
	header.Add("Cookie", rawCookies)
	req := http.Request{Header: header}
	return req.Cookies()
}

// Конвертирует map[string]interface{} в SomeStrict
func convertMapToStruct(data map[string]interface{}, obj interface{}) error {
	objValue := reflect.ValueOf(obj)
	if objValue.Kind() != reflect.Ptr || objValue.IsNil() {
		return contextedError.New("wrong struct type")
	}

	objValue = objValue.Elem()
	objType := objValue.Type()

	for i := 0; i < objValue.NumField(); i++ {
		field := objValue.Field(i)
		fieldType := objType.Field(i)

		fieldName := fieldType.Tag.Get("json")
		if fieldName == "" {
			fieldName = fieldType.Name
		}

		fieldValue, ok := data[fieldName]
		if !ok {
			continue
		}

		if !field.CanSet() {
			return contextedError.New(fmt.Sprintf("field '%s' is unavailable to write", fieldName))
		}

		if err := setFieldValue(field, fieldValue); err != nil {
			return err
		}
	}

	return nil
}

func setFieldValue(field reflect.Value, fieldValue interface{}) error {
	fieldType := field.Type()

	switch fieldType.Kind() {
	case reflect.Slice:
		sliceElemType := fieldType.Elem()

		if reflect.TypeOf(fieldValue).Kind() != reflect.Slice {
			return contextedError.New("field slice type mismatch")
		}

		slice := reflect.MakeSlice(fieldType, 0, 0)
		for _, elem := range fieldValue.([]interface{}) {
			elemValue := reflect.New(sliceElemType).Elem()

			if err := setFieldValue(elemValue, elem); err != nil {
				return err
			}

			slice = reflect.Append(slice, elemValue)
		}
		field.Set(slice)

	case reflect.Struct:
		if reflect.TypeOf(fieldValue).Kind() != reflect.Map {
			return contextedError.New("field struct type mismatch")
		}

		structValue := reflect.New(fieldType).Elem()
		if err := convertMapToStruct(fieldValue.(map[string]interface{}), structValue.Addr().Interface()); err != nil {
			return err
		}
		field.Set(structValue)

	default:
		fieldValueOfType := reflect.ValueOf(fieldValue)

		if !fieldValueOfType.Type().ConvertibleTo(fieldType) {
			return contextedError.New("field type mismatch")
		}

		field.Set(fieldValueOfType.Convert(fieldType))
	}

	return nil
}

// Конвертирует map[string]interface{} или []interface{} в []SomeStrict
func convertMapToSliceWithType(inputData interface{}, itemConverter func(map[string]interface{}) (interface{}, error)) ([]interface{}, error) {
	mapData, ok := inputData.([]interface{})
	if !ok {
		switch data := inputData.(type) {
		case map[string]interface{}:
			mapData = []interface{}{}
			for _, itemInterface := range data {
				mapData = append(mapData, itemInterface)
			}
		default:
			return nil, contextedError.New("parse error")
		}
	}

	var list []interface{}
	for _, itemInterface := range mapData {
		itemMap, ok := itemInterface.(map[string]interface{})
		if !ok {
			return nil, contextedError.New("parse error")
		}

		item, err := itemConverter(itemMap)
		if err != nil {
			return nil, contextedError.New("item conversion error")
		}

		list = append(list, item)
	}

	return list, nil
}
