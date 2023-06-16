package repository

import "time"

type RequestInput struct {
	ShopID            int64     `json:"sid"`
	ShopCurrency      string    `json:"cur"`
	Accounts          []Account `json:"acc"`
	IAcc              int       `json:"i_acc"`
	ProgressiveImport bool      `json:"progressive"`
	AccessToken       string    `json:"access_token"`
	ConsumerID        int64     `json:"consumer_id"`
	Name              string    `json:"name"`
	ShopName          string    `json:"shop_name"`
	StartSyncTime     time.Time `json:"start_sync_time"`
	Platform          string    `json:"platform"`
}

type Account struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Currency     string `json:"cur"`
	Timezone     string `json:"tz"`
	CampaignList string `json:"cp_list"`
	IDate        int    `json:"i_date"`
}

type InAppNotification struct {
	ShopId            int64                  `json:"shop_id,omitempty"`
	MessageID         string                 `json:"message_id,omitempty"`
	Type              string                 `json:"type,omitempty"`
	Topic             string                 `json:"topic,omitempty"`
	MessageAttributes map[string]interface{} `json:"message_attributes,omitempty"`
	Timestamp         time.Time              `json:"timestamp,omitempty"`
	Message           interface{}            `json:"message,omitempty"`
	Subject           string                 `json:"subject,omitempty"`
}
