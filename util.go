package migration

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/rs/zerolog/log"
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
		return body, fmt.Errorf("responded with status: %s, body: %s", res.Status, string(body))
	}
	return body, nil
}

func _close(body io.ReadCloser) {
	if err := body.Close(); err != nil {
		log.Error().Str("module", "github.com/estafette/migration").Err(err).Msg("error while closing the response body")
	}
}
