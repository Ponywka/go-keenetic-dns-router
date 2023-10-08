package keenetic

import (
	"bytes"
	"fmt"
	"github.com/Ponywka/go-keenetic-dns-router/errors/contextedError"
	"github.com/Ponywka/go-keenetic-dns-router/errors/parentError"
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

		if fieldType.Type.Kind() == reflect.Slice {
			sliceType := fieldType.Type
			sliceElemType := sliceType.Elem()

			if reflect.TypeOf(fieldValue).Kind() == reflect.Slice {
				slice := reflect.MakeSlice(sliceType, 0, 0)
				for _, elem := range fieldValue.([]interface{}) {
					if !reflect.ValueOf(elem).Type().ConvertibleTo(sliceElemType) {
						return contextedError.New(fmt.Sprintf("field '%s' is unavailable to convert", fieldName))
					}

					slice = reflect.Append(slice, reflect.ValueOf(elem).Convert(sliceElemType))
				}
				field.Set(slice)
			} else {
				return contextedError.New(fmt.Sprintf("field '%s' is unavailable to convert", fieldName))
			}
		} else {
			if !reflect.ValueOf(fieldValue).Type().ConvertibleTo(field.Type()) {
				return contextedError.New(fmt.Sprintf("field '%s' is unavailable to convert", fieldName))
			}

			field.Set(reflect.ValueOf(fieldValue).Convert(field.Type()))
		}
	}

	return nil
}
