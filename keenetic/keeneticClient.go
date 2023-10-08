package keenetic

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/Ponywka/go-keenetic-dns-router/errors/contextedError"
	"github.com/Ponywka/go-keenetic-dns-router/errors/parentError"
	"io"
	"net/http"
	"strings"
)

func apiSyncRequest(method string, url string, data []byte, headers map[string]string) (resp *http.Response, body []byte, err error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		err = contextedError.NewFromExists(&err, "http.NewRequest")
		err = parentError.New("http creating error", &err)
		return
	}
	for key, val := range headers {
		req.Header.Add(key, val)
	}
	resp, err = client.Do(req)
	if err != nil {
		err = contextedError.NewFromExists(&err, "client.Do")
		err = parentError.New("client creating error", &err)
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		err = contextedError.NewFromExists(&err, "io.ReadAll")
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
		err = contextedError.NewFromExists(&err, "json.Marshal")
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
			err = contextedError.NewFromExists(&err, "json.Unmarshal")
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

func (u *KeeneticClient) Rci(data any) (body any, err error) {
	wasAuthorisationAttempt := false
	for {
		resp, body, err := u.apiRequest("POST", "rci/", data)
		if resp.StatusCode != 401 {
			return body, err
		}
		if wasAuthorisationAttempt {
			return nil, contextedError.New("unauthorized")
		}
		wasAuthorisationAttempt = true
		ok, err := u.Auth(u.login, u.password)
		if !ok {
			return body, err
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

func NewKeeneticClient(host string) KeeneticClient {
	return KeeneticClient{
		host: host,
	}
}
