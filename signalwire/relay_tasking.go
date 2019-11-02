package signalwire

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"golang.org/x/net/context/ctxhttp"
)

// RelayTaskDeliver TODO DESCRIPTION
func (*RelaySession) RelayTaskDeliver(ctx context.Context, endpoint, project, token,
	signalwireContext string, taskMsg []byte) error {
	buf := fmt.Sprintf("{\"context\": \"%s\", \"message\": %v}", signalwireContext, string(taskMsg))
	b := []byte(buf)

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	req.SetBasicAuth(project, token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", UserAgent)

	duration, err := time.ParseDuration(fmt.Sprintf("%ds", HTTPClientTimeout))
	if err != nil {
		return err
	}

	// set timeout
	ctx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	resp, err := ctxhttp.Do(ctx, nil, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode > 300 || resp.StatusCode < 200 {
		return fmt.Errorf("%s", body)
	}

	return nil
}
