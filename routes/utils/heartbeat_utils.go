package utils

import (
	"bytes"
	"encoding/json"
	"github.com/muety/wakapi/models"
	"io"
	"net/http"
)

func ParseHeartbeats(r *http.Request) ([]*models.Heartbeat, error) {
	heartbeats, err := tryParseBulk(r)
	if err == nil {
		return heartbeats, err
	}

	heartbeats, err = tryParseSingle(r)
	if err == nil {
		return heartbeats, err
	}

	return []*models.Heartbeat{}, err
}

func tryParseBulk(r *http.Request) ([]*models.Heartbeat, error) {
	var heartbeats []*models.Heartbeat

	body, _ := io.ReadAll(r.Body)
	r.Body.Close()
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	dec := json.NewDecoder(io.NopCloser(bytes.NewBuffer(body)))
	if err := dec.Decode(&heartbeats); err != nil {
		return nil, err
	}

	return heartbeats, nil
}

func tryParseSingle(r *http.Request) ([]*models.Heartbeat, error) {
	var heartbeat models.Heartbeat

	body, _ := io.ReadAll(r.Body)
	r.Body.Close()
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	dec := json.NewDecoder(io.NopCloser(bytes.NewBuffer(body)))
	if err := dec.Decode(&heartbeat); err != nil {
		return nil, err
	}

	return []*models.Heartbeat{&heartbeat}, nil
}
