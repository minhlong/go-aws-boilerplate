package handler

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/minhlong/go-aws-boilerplate/internal/repository"
	"github.com/pkg/errors"
)

func ParseRequest(event events.SQSEvent) (*repository.RequestInput, error) {
	var payload repository.RequestInput

	if len(event.Records) != 1 {
		err := errors.New("this is not sqs event")
		return nil, err
	}

	if err := json.Unmarshal([]byte(event.Records[0].Body), &payload); err != nil {
		return nil, err
	}

	return &payload, nil
}
