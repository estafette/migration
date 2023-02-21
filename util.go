package migration

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"net/url"
)

func _urlJoin(URL string, path ...string) string {
	if u, err := url.JoinPath(URL, path...); err != nil {
		panic(err)
	} else {
		return u
	}
}

func _successful(res *http.Response) ([]byte, error) {
	defer _close(res.Body)
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading resposne body: %w", err)
	}
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return body, fmt.Errorf("response with status: %s, body: %s", res.Status, string(body))
	}
	return body, nil
}

func _close(body io.ReadCloser) {
	if err := body.Close(); err != nil {
		log.Error().Err(err).Msg("error while closing the response body")
	}
}
