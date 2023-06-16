package service

import (
	"context"
	"github.com/minhlong/go-aws-boilerplate/internal/repository"
)

func GetInsights(ctx context.Context, request repository.RequestInput, repo *repository.MongodbRepository) ([]repository.Account, error) {
	result, err := repo.Insights(ctx, request)
	if err != nil {
		return nil, err
	}

	return result, nil
}
