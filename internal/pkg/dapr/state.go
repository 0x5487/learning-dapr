package dapr

import (
	"errors"
	"fmt"

	"github.com/jasonsoft/log/v2"
	"github.com/jasonsoft/request"
)

var ErrStateKeyNotFound = errors.New("actor state key not found")

func ActorState(actorType, actorID, key string) (string, error) {

	// invoke get sate
	url := fmt.Sprintf("http://localhost:3500/v1.0/actors/%s/%s/state/%s", actorType, actorID, key)
	resp, err := request.GET(url).
		Set("Content-Type", "application/json").
		End()

	if err != nil {
		log.Err(err).Error("can't send to dapr: can't get actor state")
		return "", err
	}

	if resp.StatusCode == 404 {
		return "", ErrStateKeyNotFound
	}

	if resp.OK == false {
		log.Debugf("can't get actor state failed %s", resp.String())
		return "", errors.New("can't get actor state")
	}

	return resp.String(), nil

}

type StateRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type State struct {
	Operation string       `json:"operation"`
	Request   StateRequest `json:"request"`
}

func SaveActorState(actorType, actorID string, key, value string) error {

	url := fmt.Sprintf("http://localhost:3500/v1.0/actors/%s/%s/state", actorType, actorID)

	payload := []State{}
	req := State{
		Operation: "upsert",
		Request: StateRequest{
			Key:   key,
			Value: value,
		},
	}
	payload = append(payload, req)

	resp, err := request.PUT(url).
		SendJSON(payload).
		End()

	if err != nil {
		log.Err(err).Error("can't send to dapr: can't save actor state")
		return err
	}

	if resp.OK == false {
		log.Debugf("save actor state failed %s", resp.String())
		return fmt.Errorf("save actor state failed")
	}

	return nil
}
