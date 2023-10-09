package keenetic

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/Ponywka/go-keenetic-dns-router/errors/contextedError"
	"github.com/Ponywka/go-keenetic-dns-router/errors/parentError"
	"net/http"
	"strings"
)

type AuthData struct {
	Realm     string
	Challenge string
}

type Client struct {
	cookies  map[string]string
	login    string
	password string
	host     string
}

func (u *Client) apiRequest(method, path string, data interface{}) (resp *http.Response, body interface{}, err error) {
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
		err = json.Unmarshal(outBody, &body)
		if err != nil {
			err = contextedError.NewFromFunc(&err, json.Unmarshal)
			err = parentError.New("json decoding error", &err)
			return
		}
	}
	return
}

func (u *Client) resetAuth() (data AuthData, err error) {
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

func (u *Client) checkAuth() (res bool, err error) {
	resp, _, err := u.apiRequest("GET", "auth", nil)
	if err != nil {
		err = parentError.New("api requesting error", &err)
		return
	}
	res = resp.StatusCode == 200
	return
}

func (u *Client) Auth(login, password string) (res bool, err error) {
	authData, err := u.resetAuth()
	if err != nil {
		err = parentError.New("auth reset error", &err)
		return
	}

	passHash := fmt.Sprintf("%s:%s:%s", login, authData.Realm, password)
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

func (u *Client) Rci(data interface{}) (res []interface{}, err error) {
	wasAuthorizationAttempt := false
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
		if wasAuthorizationAttempt {
			return nil, contextedError.New("unauthorized")
		}
		wasAuthorizationAttempt = true
		ok, err := u.Auth(u.login, u.password)
		if err != nil {
			return nil, parentError.New("reauth error", &err)
		}
		if !ok {
			return nil, contextedError.New("reauth error")
		}
	}
}

func (u *Client) ToRciQueryList(list *[]map[string]interface{}, path string, data interface{}) error {
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

func (u *Client) getByRciQuery(path string, data interface{}) (res interface{}, err error) {
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

func (u *Client) getListByRciQuery(query string, data any, itemConverter func(map[string]interface{}) (interface{}, error)) ([]interface{}, error) {
	body, err := u.getByRciQuery(query, data)
	if err != nil {
		return nil, parentError.New("rci request error", &err)
	}

	list, err := convertMapToSliceWithType(body, itemConverter)
	if err != nil {
		return nil, parentError.New("conversation error", &err)
	}

	return list, nil
}

func (u *Client) GetInterfaceList() ([]InterfaceBase, error) {
	list, err := u.getListByRciQuery("show.interface", nil, func(mapItem map[string]interface{}) (interface{}, error) {
		item := *new(InterfaceBase)
		err := convertMapToStruct(mapItem, &item)
		return item, err
	})
	if err != nil {
		return nil, err
	}

	var listMap []InterfaceBase
	for _, val := range list {
		v, ok := val.(InterfaceBase)
		if !ok {
			return nil, contextedError.New("parse error")
		}
		listMap = append(listMap, v)
	}

	return listMap, nil
}

func (u *Client) GetRouteList() ([]Route, error) {
	list, err := u.getListByRciQuery("show.ip.route", nil, func(mapItem map[string]interface{}) (interface{}, error) {
		item := *new(Route)
		err := convertMapToStruct(mapItem, &item)
		return item, err
	})
	if err != nil {
		return nil, err
	}

	var listMap []Route
	for _, val := range list {
		v, ok := val.(Route)
		if !ok {
			return nil, contextedError.New("parse error")
		}
		listMap = append(listMap, v)
	}

	return listMap, nil
}

func New(host string) Client {
	return Client{
		host: host,
	}
}
