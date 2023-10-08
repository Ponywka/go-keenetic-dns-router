package keenetic

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/Ponywka/go-keenetic-dns-router/errors/contextedError"
	"github.com/Ponywka/go-keenetic-dns-router/errors/parentError"
	"net/http"
	"reflect"
	"strings"
)

type AuthData struct {
	Realm     string
	Challenge string
}

type KeeneticClient struct {
	cookies  map[string]string
	login    string
	password string
	host     string
}

func (u *KeeneticClient) apiRequest(method string, path string, data any) (resp *http.Response, body any, err error) {
	var cookieStr string
	for key, val := range u.cookies {
		// TODO: Escape symbols
		cookieStr += fmt.Sprintf("%s=%s;", key, val)
	}

	headers := map[string]string{
		"Content-Type": "application/json",
		"Cookie":       cookieStr,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		err = contextedError.NewFromFunc(&err, json.Marshal)
		err = parentError.New("json encoding error", &err)
		return
	}

	resp, outBody, err := apiSyncRequest(method, fmt.Sprintf("%s/%s", u.host, path), jsonData, headers)
	if err != nil {
		err = parentError.New("api requesting error", &err)
		return
	}

	for _, cookie := range parseCookies(resp.Header.Get("Set-Cookie")) {
		u.cookies[cookie.Name] = cookie.Value
	}

	if len(outBody) > 0 {
		_ = json.Unmarshal(outBody, &body)
		if err != nil {
			err = contextedError.NewFromFunc(&err, json.Unmarshal)
			err = parentError.New("json decoding error", &err)
			return
		}
	}
	return
}

func (u *KeeneticClient) resetAuth() (data AuthData, err error) {
	u.cookies = make(map[string]string)
	resp, _, err := u.apiRequest("GET", "auth", nil)
	if err != nil {
		err = parentError.New("api requesting error", &err)
		return
	}
	data = AuthData{
		Realm:     resp.Header.Get("X-Ndm-Realm"),
		Challenge: resp.Header.Get("X-Ndm-Challenge"),
	}
	return
}

func (u *KeeneticClient) checkAuth() (res bool, err error) {
	resp, _, err := u.apiRequest("GET", "auth", nil)
	if err != nil {
		err = parentError.New("api requesting error", &err)
		return
	}
	res = resp.StatusCode == 200
	return
}

func (u *KeeneticClient) Auth(login string, password string) (res bool, err error) {
	authData, err := u.resetAuth()
	if err != nil {
		err = parentError.New("auth reset error", &err)
		return
	}

	var passHash string
	passHash = fmt.Sprintf("%s:%s:%s", login, authData.Realm, password)
	passHash = fmt.Sprintf("%x", md5.Sum([]byte(passHash)))
	passHash = fmt.Sprintf("%s%s", authData.Challenge, passHash)
	passHash = fmt.Sprintf("%x", sha256.Sum256([]byte(passHash)))

	resp, _, err := u.apiRequest("POST", "auth", map[string]string{
		"login":    login,
		"password": passHash,
	})
	if err != nil {
		err = parentError.New("auth error", &err)
		return
	}

	res = resp.StatusCode == 200
	u.login = login
	u.password = password
	return
}

func (u *KeeneticClient) Rci(data any) (res []interface{}, err error) {
	wasAuthorisationAttempt := false
	for {
		resp, body, err := u.apiRequest("POST", "rci/", data)
		if err != nil {
			return nil, parentError.New("api requesting error", &err)
		}
		if resp.StatusCode != 401 {
			res, ok := body.([]interface{})
			if !ok {
				return nil, contextedError.New("parse error")
			}
			return res, nil
		}
		if wasAuthorisationAttempt {
			return nil, contextedError.New("unauthorized")
		}
		wasAuthorisationAttempt = true
		ok, err := u.Auth(u.login, u.password)
		if err != nil {
			return nil, parentError.New("reauth error", &err)
		}
		if !ok {
			return nil, contextedError.New("reauth error")
		}
	}
}

func (u *KeeneticClient) ToRciQueryList(list *[]map[string]interface{}, path string, data any) error {
	pathSplitted := strings.Split(path, ".")
	if data == nil {
		data = map[string]interface{}{}
	}
	outData := map[string]interface{}{}
	for i := len(pathSplitted) - 1; i >= 0; i-- {
		if len(pathSplitted[i]) == 0 {
			return contextedError.New("empty path name was detected")
		}
		if i == len(pathSplitted)-1 {
			outData = map[string]interface{}{pathSplitted[i]: data}
		} else {
			outData = map[string]interface{}{pathSplitted[i]: outData}
		}
	}
	*list = append(*list, outData)
	return nil
}

func (u *KeeneticClient) getByRciQuery(path string, data any) (res any, err error) {
	var list []map[string]interface{}
	err = u.ToRciQueryList(&list, path, data)
	if err != nil {
		return nil, parentError.New("list generating error", &err)
	}

	body, err := u.Rci(list)
	if err != nil {
		return nil, parentError.New("rci request error", &err)
	}

	if len(body) == 0 {
		return nil, contextedError.New("parse error")
	}
	res = body[0]
	for _, key := range strings.Split(path, ".") {
		v, ok := res.(map[string]interface{})
		if !ok {
			return nil, contextedError.New("parse error")
		}
		res = v[key]
	}
	return
}

func convertMapToStruct(d any, s interface{}) error {
	m, ok := d.(map[string]interface{})
	if !ok {
		return contextedError.New("parse error")
	}
	stValue := reflect.ValueOf(s).Elem()
	sType := stValue.Type()
	for i := 0; i < sType.NumField(); i++ {
		field := sType.Field(i)
		tagName := string(field.Tag.Get("json"))
		if len(tagName) == 0 {
			tagName = field.Name
		}
		if value, ok := m[tagName]; ok {
			stValue.Field(i).Set(reflect.ValueOf(value))
		}
	}
	return nil
}

func (u *KeeneticClient) GetInterfaceList() (res map[string]InterfaceBase, err error) {
	body, err := u.getByRciQuery("show.interface", nil)
	v, ok := body.(map[string]interface{})
	if !ok {
		return nil, contextedError.New("parse error")
	}
	res = map[string]InterfaceBase{}
	for key, val := range v {
		interfaceBase := *new(InterfaceBase)
		err := convertMapToStruct(val, &interfaceBase)
		if err != nil {
			return nil, err
		}
		res[key] = interfaceBase
	}
	return
}

func NewKeeneticClient(host string) KeeneticClient {
	return KeeneticClient{
		host: host,
	}
}
