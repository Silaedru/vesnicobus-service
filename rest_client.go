package main

import (
	"fmt"
	"net/http"
	"strings"
)

func httpCall(method string, url string, headers []string) (*http.Response, int) {
	req, err := http.NewRequest(method, url, nil)
	processFatalError(err)

	for _, header := range headers {
		h := strings.Split(header, ":")
		req.Header.Add(h[0], h[1])
	}

	client := http.Client{}
	resp, err := client.Do(req)
	processFatalError(err)

	return resp, resp.StatusCode
}

func golemioHttpCall(url string, limit int, offset int) (*http.Response, int) {
	url = fmt.Sprintf("%s&limit=%d&offset=%d", url, limit, offset)
	return httpCall(http.MethodGet, url, []string{"x-access-token:" + golemioApiKey})
}
