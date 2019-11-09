package main

import (
	"fmt"
	"net/http"
	"strings"
)

func httpCall(method string, url string, headers []string) (*http.Response, int, error) {
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return nil, 0, err
	}

	for _, header := range headers {
		h := strings.Split(header, ":")
		req.Header.Add(h[0], h[1])
	}

	client := http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return nil, 0, err
	}

	return resp, resp.StatusCode, nil
}

func golemioHttpCall(url string, limit int, offset int) (*http.Response, int, error) {
	url = fmt.Sprintf("%s&limit=%d&offset=%d", url, limit, offset)
	return httpCall(http.MethodGet, url, []string{"x-access-token:" + golemioApiKey})
}
