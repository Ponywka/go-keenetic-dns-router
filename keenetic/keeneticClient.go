package keenetic

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func apiSyncRequest(method string, url string, data []byte, headers map[string]string) (resp *http.Response, body []byte, err error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return
	}
	for key, val := range headers {
		req.Header.Add(key, val)
	}
	resp, err = client.Do(req)
	if err != nil {
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	body, err = io.ReadAll(resp.Body)
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

func (u *KeeneticClient) apiRequest(method string, path string, data any) (resp *http.Response, body map[string]interface{}, err error) {
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
		return
	}

	resp, outBody, err := apiSyncRequest(method, fmt.Sprintf("%s/%s", u.host, path), jsonData, headers)
	if err != nil {
		return
	}

	for _, cookie := range parseCookies(resp.Header.Get("Set-Cookie")) {
		u.cookies[cookie.Name] = cookie.Value
	}

	// TODO: Catch JSON parse error
	_ = json.Unmarshal(outBody, &body)
	return
}

func (u *KeeneticClient) resetAuth() (data AuthData, err error) {
	u.cookies = make(map[string]string)
	resp, _, err := u.apiRequest("GET", "auth", nil)
	if err != nil {
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
		return
	}
	res = resp.StatusCode == 200
	return
}

func (u *KeeneticClient) Auth(login string, password string) (res bool, err error) {
	log.Printf("Reset auth")
	authData, err := u.resetAuth()
	if err != nil {
		return
	}

	log.Printf("Generating hash")
	var passHash string
	passHash = fmt.Sprintf("%s:%s:%s", login, authData.Realm, password)
	passHash = fmt.Sprintf("%x", md5.Sum([]byte(passHash)))
	passHash = fmt.Sprintf("%s%s", authData.Challenge, passHash)
	passHash = fmt.Sprintf("%x", sha256.Sum256([]byte(passHash)))

	log.Printf("Sending request")
	resp, _, err := u.apiRequest("POST", "auth", map[string]string{
		"login":    login,
		"password": passHash,
	})
	if err != nil {
		return
	}

	log.Printf("Result")
	res = resp.StatusCode == 200
	u.login = login
	u.password = password
	return
}

func NewKeeneticClient(host string) KeeneticClient {
	return KeeneticClient{
		host: host,
	}
}
