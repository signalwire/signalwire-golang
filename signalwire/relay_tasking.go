package signalwire

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

// RelayTaskDeliver TODO DESCRIPTION
func (*RelaySession) RelayTaskDeliver(_ context.Context, endpoint, project, token,
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

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode > 300 || resp.StatusCode < 200 {
		err := fmt.Sprintf("%s", body)
		return errors.New(err)
	}

	return nil
}
