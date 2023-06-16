package handler

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/minhlong/go-aws-boilerplate/internal/funcservice"
	"github.com/minhlong/go-aws-boilerplate/internal/repository"
	"github.com/minhlong/go-aws-boilerplate/internal/service"
	"go.uber.org/zap"
)

func HandleLambdaEvent(ctx context.Context, event events.SQSEvent) (errC error) {
	// Parse request
	request, errP := ParseRequest(event)
	if errP != nil {
		return errP
	}

	// Init mongo connection
	repo, errC := repository.NewMongoDb(ctx, request.ShopID)
	if errC != nil {
		zap.L().Error("can not init mongo connection", zap.Error(errC))
		return errC
	}

	// Get data
	result, errD := service.GetInsights(ctx, *request, repo)
	if errD != nil {
		zap.L().Error("can not aggregate data", zap.Error(errD))
		return errD
	}

	// Trigger notification
	errW := funcservice.SendInAppNotification(ctx, request.ShopID, result, "insight-request", "Success")
	if errW != nil {
		return errD
	}

	return nil
}
