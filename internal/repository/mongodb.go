package repository

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
	"os"
)

var mongoClient *mongo.Client

type MongodbRepository struct {
	Database       *mongo.Database
	CollectionName *mongo.Collection
}

func getMongoClient(ctx context.Context, connectionUri string) (_ *mongo.Client, err error) {
	if mongoClient != nil {
		if err := mongoClient.Ping(ctx, readpref.Primary()); err == nil {
			return mongoClient, nil
		}
		mongoClient.Disconnect(ctx)
	}

	appName := os.Getenv("AWS_REGION") + ":" + os.Getenv("AWS_LAMBDA_FUNCTION_NAME")
	clientOptions := options.Client().ApplyURI(connectionUri).SetAppName(appName)

	mongoClient, err = mongo.NewClient(clientOptions)
	if err != nil {
		return nil, err
	}

	if err := mongoClient.Connect(ctx); err != nil {
		return nil, err
	}

	if err := mongoClient.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	return mongoClient, nil
}

func NewMongoDb(ctx context.Context, shopID int64) (*MongodbRepository, error) {
	databaseName, ok := os.LookupEnv("DB_NAME")
	if !ok {
		err := errors.New("DB_NAME is missing")

		return nil, err
	}
	connectionURI, ok := os.LookupEnv("DB_URI")
	if !ok {
		err := errors.New("DB_URI is missing")

		return nil, err
	}
	mongoClient, err := getMongoClient(ctx, connectionURI)
	if err != nil {
		return nil, err
	}
	database := mongoClient.Database(databaseName)
	tmpName := database.Collection(fmt.Sprintf("acction_%d", shopID))

	repo := &MongodbRepository{
		Database:       database,
		CollectionName: tmpName,
	}

	return repo, nil
}

func (m *MongodbRepository) Disconnect(ctx context.Context) {
	if mongoClient != nil {
		mongoClient.Disconnect(ctx)
	}
}

func (m *MongodbRepository) Insights(ctx context.Context, input RequestInput) ([]Account, error) {
	match := bson.M{
		"date": bson.M{
			"$gte": input.StartSyncTime,
			"$lte": input.StartSyncTime,
		},
	}

	pipeLine := []bson.M{
		{
			"$match": match,
		},
		{
			"$addFields": bson.M{
				"purchases": bson.M{
					"$ifNull": bson.A{
						bson.M{
							"$toDouble": "$purchases",
						}, 0,
					},
				},
				"purchases_value": bson.M{
					"$ifNull": bson.A{
						"$purchases_value",
						bson.M{
							"$toDecimal": 0,
						},
					},
				},
				"add_to_cart": bson.M{
					"$ifNull": bson.A{
						bson.M{
							"$toDouble": "$add_to_cart",
						}, 0,
					},
				},
				"assisted_purchase": bson.M{
					"$ifNull": bson.A{
						"$assisted_purchase", 0,
					},
				},
				"direct_purchase": bson.M{
					"$ifNull": bson.A{
						"$direct_purchase", 0,
					},
				},
				"ad_status": bson.M{
					"$ifNull": bson.A{
						"$ad_status", "INACTIVE",
					},
				},
				"adset_status": bson.M{
					"$ifNull": bson.A{
						"$adset_status", "INACTIVE",
					},
				},
				"campaign_status": bson.M{
					"$ifNull": bson.A{
						"$campaign_status", "INACTIVE",
					},
				},
				"valid_parameters": bson.M{
					"$ifNull": bson.A{
						"$valid_parameters", false,
					},
				},
			},
		},
		{
			"$group": bson.M{
				"_id":               "$ad_id",
				"ad_id":             bson.M{"$first": "$ad_id"},
				"ad_name":           bson.M{"$first": "$ad_name"},
				"clicks":            bson.M{"$sum": "$clicks"},
				"spend":             bson.M{"$sum": "$spend"},
				"impressions":       bson.M{"$sum": "$impressions"},
				"add_to_cart":       bson.M{"$sum": "$add_to_cart"},
				"purchases":         bson.M{"$sum": "$purchases"},
				"purchases_value":   bson.M{"$sum": "$purchases_value"},
				"assisted_purchase": bson.M{"$sum": "$assisted_purchase"},
				"direct_purchase":   bson.M{"$sum": "$direct_purchase"},
				"ad_group_id":       bson.M{"$first": "$adset_id"},
				"ad_group_name":     bson.M{"$first": "$adset_name"},
				"campaign_id":       bson.M{"$first": "$campaign_id"},
				"campaign_name":     bson.M{"$first": "$campaign_name"},
				"account_id":        bson.M{"$first": "$ad_account_id"},
				"account_name":      bson.M{"$first": "$ad_account_name"},
				"ad_group_status":   bson.M{"$first": "$adset_status"},
				"campaign_status":   bson.M{"$first": "$campaign_status"},
				"ad_status":         bson.M{"$first": "$ad_status"},
				"valid_parameters":  bson.M{"$first": "$valid_parameters"},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.D{
					{"ad_group_id", "$ad_group_id"},
					{"campaign_id", "$campaign_id"},
				},
				"ads": bson.M{"$push": bson.M{
					"ad_id":           "$ad_id",
					"ad_name":         "$ad_name",
					"clicks":          "$clicks",
					"spend":           "$spend",
					"impressions":     "$impressions",
					"add_to_cart":     "$add_to_cart",
					"purchases":       "$purchases",
					"purchases_value": "$purchases_value",
					"ad_status":       "$ad_status",
					"ctr": bson.M{
						"$multiply": bson.A{
							bson.M{
								"$cond": bson.M{
									"if":   bson.M{"$gt": bson.A{"$impressions", 0}},
									"then": bson.M{"$toDouble": bson.M{"$divide": bson.A{"$clicks", "$impressions"}}},
									"else": bson.M{"$toDouble": 0},
								},
							}, 100,
						},
					},
					"cost_per_atc": bson.M{
						"$cond": bson.M{
							"if":   bson.M{"$gt": bson.A{"$add_to_cart", 0}},
							"then": bson.M{"$toDouble": bson.M{"$divide": bson.A{"$spend", "$add_to_cart"}}},
							"else": bson.M{"$toDouble": 0},
						},
					},
					"cost_per_purchase": bson.M{
						"$cond": bson.M{
							"if":   bson.M{"$gt": bson.A{"$purchases", 0}},
							"then": bson.M{"$toDouble": bson.M{"$divide": bson.A{"$spend", "$purchases"}}},
							"else": bson.M{"$toDouble": 0},
						},
					},
					"conversion_rate": bson.M{"$multiply": bson.A{
						bson.M{
							"$cond": bson.M{
								"if":   bson.M{"$gt": bson.A{"$purchases", 0}},
								"then": bson.M{"$toDouble": bson.M{"$divide": bson.A{"$purchases", "$clicks"}}},
								"else": bson.M{"$toDouble": 0},
							}}, 100}},
					"roas": bson.M{
						"$cond": bson.M{
							"if":   bson.M{"$gt": bson.A{"$spend", 0}},
							"then": bson.M{"$toDouble": bson.M{"$divide": bson.A{"$purchases_value", "$spend"}}},
							"else": bson.M{"$toDouble": 0},
						},
					},
					"assisted_purchase": "$assisted_purchase",
					"direct_purchase":   "$direct_purchase",
					"valid_parameters":  "$valid_parameters",
				}},
				"clicks":            bson.M{"$sum": "$clicks"},
				"spend":             bson.M{"$sum": "$spend"},
				"impressions":       bson.M{"$sum": "$impressions"},
				"add_to_cart":       bson.M{"$sum": "$add_to_cart"},
				"purchases":         bson.M{"$sum": "$purchases"},
				"purchases_value":   bson.M{"$sum": "$purchases_value"},
				"assisted_purchase": bson.M{"$sum": "$assisted_purchase"},
				"direct_purchase":   bson.M{"$sum": "$direct_purchase"},
				"ad_group_id":       bson.M{"$first": "$ad_group_id"},
				"ad_group_name":     bson.M{"$first": "$ad_group_name"},
				"campaign_id":       bson.M{"$first": "$campaign_id"},
				"campaign_name":     bson.M{"$first": "$campaign_name"},
				"account_id":        bson.M{"$first": "$account_id"},
				"account_name":      bson.M{"$first": "$account_name"},
				"ad_group_status":   bson.M{"$first": "$ad_group_status"},
				"campaign_status":   bson.M{"$first": "$campaign_status"},
			},
		},
		{
			"$group": bson.M{
				"_id": "$_id.campaign_id",
				"ad_group": bson.M{
					"$push": bson.M{
						"ads":               "$ads",
						"clicks":            "$clicks",
						"spend":             "$spend",
						"impressions":       "$impressions",
						"add_to_cart":       "$add_to_cart",
						"purchases":         "$purchases",
						"purchases_value":   "$purchases_value",
						"assisted_purchase": "$assisted_purchase",
						"direct_purchase":   "$direct_purchase",
						"ad_group_id":       "$ad_group_id",
						"ad_group_name":     "$ad_group_name",
						"ad_group_status":   "$ad_group_status",
						"ctr": bson.M{
							"$multiply": bson.A{
								bson.M{
									"$cond": bson.M{
										"if":   bson.M{"$gt": bson.A{"$impressions", 0}},
										"then": bson.M{"$toDouble": bson.M{"$divide": bson.A{"$clicks", "$impressions"}}},
										"else": bson.M{"$toDouble": 0},
									},
								}, 100,
							},
						},
						"cost_per_atc": bson.M{
							"$cond": bson.M{
								"if":   bson.M{"$gt": bson.A{"$add_to_cart", 0}},
								"then": bson.M{"$toDouble": bson.M{"$divide": bson.A{"$spend", "$add_to_cart"}}},
								"else": bson.M{"$toDouble": 0},
							},
						},
						"cost_per_purchase": bson.M{
							"$cond": bson.M{
								"if":   bson.M{"$gt": bson.A{"$purchases", 0}},
								"then": bson.M{"$toDouble": bson.M{"$divide": bson.A{"$spend", "$purchases"}}},
								"else": bson.M{"$toDouble": 0},
							},
						},
						"conversion_rate": bson.M{"$multiply": bson.A{
							bson.M{
								"$cond": bson.M{
									"if":   bson.M{"$gt": bson.A{"$purchases", 0}},
									"then": bson.M{"$toDouble": bson.M{"$divide": bson.A{"$purchases", "$clicks"}}},
									"else": bson.M{"$toDouble": 0},
								}}, 100}},
						"roas": bson.M{
							"$cond": bson.M{
								"if":   bson.M{"$gt": bson.A{"$spend", 0}},
								"then": bson.M{"$toDouble": bson.M{"$divide": bson.A{"$purchases", "$spend"}}},
								"else": bson.M{"$toDouble": 0},
							},
						},
					},
				},
				"clicks":            bson.M{"$sum": "$clicks"},
				"spend":             bson.M{"$sum": "$spend"},
				"impressions":       bson.M{"$sum": "$impressions"},
				"add_to_cart":       bson.M{"$sum": "$add_to_cart"},
				"purchases":         bson.M{"$sum": "$purchases"},
				"purchases_value":   bson.M{"$sum": "$purchases_value"},
				"assisted_purchase": bson.M{"$sum": "$assisted_purchase"},
				"direct_purchase":   bson.M{"$sum": "$direct_purchase"},
				"campaign_id":       bson.M{"$first": "$campaign_id"},
				"campaign_name":     bson.M{"$first": "$campaign_name"},
				"account_id":        bson.M{"$first": "$account_id"},
				"account_name":      bson.M{"$first": "$account_name"},
				"campaign_status":   bson.M{"$first": "$campaign_status"},
			},
		},
		{
			"$group": bson.M{
				"_id": "$account_id",
				"campaigns": bson.M{
					"$push": bson.M{
						"ad_groups":         "$ad_group",
						"clicks":            "$clicks",
						"spend":             "$spend",
						"impressions":       "$impressions",
						"add_to_cart":       "$add_to_cart",
						"purchases":         "$purchases",
						"purchases_value":   "$purchases_value",
						"assisted_purchase": "$assisted_purchase",
						"direct_purchase":   "$direct_purchase",
						"campaign_id":       "$campaign_id",
						"campaign_name":     "$campaign_name",
						"campaign_status":   "$campaign_status",
						"ctr": bson.M{
							"$multiply": bson.A{
								bson.M{
									"$cond": bson.M{
										"if":   bson.M{"$gt": bson.A{"$impressions", 0}},
										"then": bson.M{"$toDouble": bson.M{"$divide": bson.A{"$clicks", "$impressions"}}},
										"else": bson.M{"$toDouble": 0},
									},
								}, 100,
							},
						},
						"cost_per_atc": bson.M{
							"$cond": bson.M{
								"if":   bson.M{"$gt": bson.A{"$add_to_cart", 0}},
								"then": bson.M{"$toDouble": bson.M{"$divide": bson.A{"$spend", "$add_to_cart"}}},
								"else": bson.M{"$toDouble": 0},
							},
						},
						"cost_per_purchase": bson.M{
							"$cond": bson.M{
								"if":   bson.M{"$gt": bson.A{"$purchases", 0}},
								"then": bson.M{"$toDouble": bson.M{"$divide": bson.A{"$spend", "$purchases"}}},
								"else": bson.M{"$toDouble": 0},
							},
						},
						"conversion_rate": bson.M{"$multiply": bson.A{
							bson.M{
								"$cond": bson.M{
									"if":   bson.M{"$gt": bson.A{"$purchases", 0}},
									"then": bson.M{"$toDouble": bson.M{"$divide": bson.A{"$purchases", "$clicks"}}},
									"else": bson.M{"$toDouble": 0},
								}}, 100}},
						"roas": bson.M{
							"$cond": bson.M{
								"if":   bson.M{"$gt": bson.A{"$spend", 0}},
								"then": bson.M{"$toDouble": bson.M{"$divide": bson.A{"$purchases_value", "$spend"}}},
								"else": bson.M{"$toDouble": 0},
							},
						},
					},
				},
				"clicks":            bson.M{"$sum": "$clicks"},
				"spend":             bson.M{"$sum": "$spend"},
				"impressions":       bson.M{"$sum": "$impressions"},
				"add_to_cart":       bson.M{"$sum": "$add_to_cart"},
				"purchases":         bson.M{"$sum": "$purchases"},
				"purchases_value":   bson.M{"$sum": "$purchases_value"},
				"assisted_purchase": bson.M{"$sum": "$assisted_purchase"},
				"direct_purchase":   bson.M{"$sum": "$direct_purchase"},
				"account_id":        bson.M{"$first": "$account_id"},
				"account_name":      bson.M{"$first": "$account_name"},
			},
		},
		{
			"$addFields": bson.M{
				"ctr": bson.M{
					"$multiply": bson.A{
						bson.M{
							"$cond": bson.M{
								"if":   bson.M{"$gt": bson.A{"$impressions", 0}},
								"then": bson.M{"$toDouble": bson.M{"$divide": bson.A{"$clicks", "$impressions"}}},
								"else": bson.M{"$toDouble": 0},
							},
						}, 100,
					},
				},
				"cost_per_atc": bson.M{
					"$cond": bson.M{
						"if":   bson.M{"$gt": bson.A{"$add_to_cart", 0}},
						"then": bson.M{"$toDouble": bson.M{"$divide": bson.A{"$spend", "$add_to_cart"}}},
						"else": bson.M{"$toDouble": 0},
					},
				},
				"cost_per_purchase": bson.M{
					"$cond": bson.M{
						"if":   bson.M{"$gt": bson.A{"$purchases", 0}},
						"then": bson.M{"$toDouble": bson.M{"$divide": bson.A{"$spend", "$purchases"}}},
						"else": bson.M{"$toDouble": 0},
					},
				},
				"conversion_rate": bson.M{"$multiply": bson.A{
					bson.M{
						"$cond": bson.M{
							"if":   bson.M{"$gt": bson.A{"$purchases", 0}},
							"then": bson.M{"$toDouble": bson.M{"$divide": bson.A{"$purchases", "$clicks"}}},
							"else": bson.M{"$toDouble": 0},
						}}, 100}},
				"roas": bson.M{
					"$cond": bson.M{
						"if":   bson.M{"$gt": bson.A{"$spend", 0}},
						"then": bson.M{"$toDouble": bson.M{"$divide": bson.A{"$purchases_value", "$spend"}}},
						"else": bson.M{"$toDouble": 0},
					},
				},
				"platform": "pinterest",
			},
		},
	}

	zap.L().Info("pipeLine", zap.Any("pipeLine", pipeLine))

	result, err := m.CollectionName.Aggregate(ctx, pipeLine)
	if err != nil {
		return nil, err
	}

	var AccountInsights []Account
	if err = result.All(ctx, &AccountInsights); err != nil {
		return nil, err
	}

	return AccountInsights, nil
}
