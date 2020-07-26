package main

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

func forceInt64(s string) int64 {
	s = strings.TrimSpace(s)
	s = strings.TrimLeft(s, "0")
	if s == "" {
		return 0
	}

	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}

	return val
}

func postForm(v interface{}, url string, values url.Values) ([]byte, error) {
	var client = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(values.Encode()))
	if err != nil {
		return []byte{}, errors.Wrap(err, "sms post fail")
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, errors.Wrap(err, "PostForm client fail")
	}
	defer resp.Body.Close()

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return respData, errors.Wrap(err, "read body fail")
	}

	if resp.StatusCode != 200 {
		return respData, errors.Wrapf(err, "status code: %d %s", resp.StatusCode, string(respData))
	}

	err = json.Unmarshal(respData, &v)
	if err != nil {
		return respData, errors.Wrapf(err, "unmarshal fail on: %s", string(respData))
	}

	return respData, nil
}
