package structs

type ReqContext struct {
	ReqID     string
	UserId    string
	SessionId string
	Role      string
}

type ReqCtxKey string

type RecordingCallback struct {
	RoomName      string `json:"room_name"`
	Duration      string `json:"duration"`
	ParticipantId string `json:"participant_id"`
	RoomId        string `json:"room_id"`
	Timestamp     string `json:"time_stamp"`
	Url           string `json:"url"`
}

type AppConfig struct {
	CLOUDFRONT_RESOURCE_URL       string `env:"type:url,name:CLOUDFRONT_RESOURCE_URL"`
	TWILIO_ACCOUNT_SID            string `env:"name:TWILIO_ACCOUNT_SID"`
	TWILIO_API_KEY                string `env:"name:TWILIO_API_KEY"`
	TWILIO_API_SECRET_KEY         string `env:"name:TWILIO_API_SECRET_KEY" `
	TWILIO_CALLBACK_URL           string `env:"type:url,name:TWILIO_CALLBACK_URL"`
	TWILIO_API_AUTH_TOKEN         string `env:"TWILIO_API_AUTH_TOKEN"`
	PSQL_URL                      string `env:"type:db_dsn,name:PSQL_URL"`
	S3_BUCKET_NAME                string `env:"name:S3_BUCKET_NAME"`
	S3_BUCKET_REGION              string `env:"name:S3_BUCKET_REGION"`
	S3_RECORDING_PREFIX           string `env:"name:S3_RECORDING_PREFIX"`
	S3_MEDIA_CONVERT_PREFIX       string `env:"name:S3_MEDIA_CONVERT_PREFIX"`
	MEDIA_CONVERT_ROLE            string `env:"name:MEDIA_CONVERT_ROLE"`
	AWS_REGION                    string `env:"name:AWS_REGION"`
	MEDIA_CONVERT_BASIC_AUTH_USER string `env:"name:MEDIA_CONVERT_BASIC_AUTH_USER"`
	MEDIA_CONVERT_BASIC_AUTH_PASS string `env:"name:MEDIA_CONVERT_BASIC_AUTH_PASS"`
}
