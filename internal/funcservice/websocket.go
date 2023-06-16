package funcservice

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/minhlong/go-aws-boilerplate/internal/repository"
	util "github.com/minhlong/go-aws-boilerplate/pkg"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

var sqsClient *sqs.SQS

func init() {
	sharedSession := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	sqsClient = sqs.New(sharedSession)
}

func SendInAppNotification(ctx context.Context, shopID int64, messageBody interface{}, requestID string, msg string) error {
	websocketNotificationQueueUrl := util.MustGetEnv("WEBSOCKET_NOTIFICATION_QUEUE_URL")

	now := time.Now()
	messageId := "shop:insights:" + requestID
	message, _ := json.Marshal(repository.InAppNotification{
		ShopId:    shopID,
		MessageID: messageId,
		Type:      "SA",
		Topic:     "shop:insights",
		MessageAttributes: map[string]interface{}{
			"data":      messageBody,
			"requestID": requestID,
		},
		Timestamp: now,
		Message:   msg,
		Subject:   "Get shop insights",
	})

	_, err := sqsClient.SendMessageWithContext(ctx, &sqs.SendMessageInput{
		MessageGroupId:         aws.String(messageId),
		MessageDeduplicationId: aws.String(strconv.FormatInt(now.UnixNano(), 10)),
		MessageBody:            aws.String(string(message)),
		QueueUrl:               aws.String(websocketNotificationQueueUrl),
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			"shopId": {
				DataType:    aws.String("Number"),
				StringValue: aws.String(strconv.FormatInt(shopID, 10)),
			},
		},
	})
	if err != nil {
		return errors.WithMessage(err, "can not send notification")
	}

	return nil
}
