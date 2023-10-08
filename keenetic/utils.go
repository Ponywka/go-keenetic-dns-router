package keenetic

import (
	"bytes"
	"github.com/Ponywka/go-keenetic-dns-router/errors/contextedError"
	"github.com/Ponywka/go-keenetic-dns-router/errors/parentError"
	"io"
	"net/http"
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
