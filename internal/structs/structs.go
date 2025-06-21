package structs

type RecordingCallback struct {
	RoomName      string `json:"room_name"`
	Duration      string `json:"duration"`
	ParticipantId string `json:"participant_id"`
	RoomId        string `json:"room_id"`
	Timestamp     string `json:"time_stamp"`
	Url           string `json:"url"`
}

type AppConfig struct {
	CLOUDFRONT_RESOURCE_URL string `env:"type:url,name:CLOUDFRONT_RESOURCE_URL"`
	TWILIO_ACCOUNT_SID      string `env:"name:TWILIO_ACCOUNT_SID"`
	TWILIO_API_KEY          string `env:"name:TWILIO_API_KEY"`
	TWILIO_API_SECRET_KEY   string `env:"name:TWILIO_API_SECRET_KEY" `
	TWILIO_CALLBACK_URL     string `env:"type:url,name:TWILIO_CALLBACK_URL"`
	TWILIO_API_AUTH_TOKEN   string `env:"name:TWILIO_API_AUTH_TOKEN"`
	S3_BUCKET_NAME          string `env:"name:S3_BUCKET_NAME"`
	S3_BUCKET_REGION        string `env:"name:S3_BUCKET_REGION"`
	S3_RECORDING_PREFIX     string `env:"name:S3_RECORDING_PREFIX"`

	S3_MEDIA_CONVERT_PREFIX       string `env:"name:S3_MEDIA_CONVERT_PREFIX"`
	MEDIA_CONVERT_ROLE            string `env:"name:MEDIA_CONVERT_ROLE"`
	AWS_REGION                    string `env:"name:AWS_REGION"`
	MEDIA_CONVERT_BASIC_AUTH_USER string `env:"name:MEDIA_CONVERT_BASIC_AUTH_USER"`
	MEDIA_CONVERT_BASIC_AUTH_PASS string `env:"name:MEDIA_CONVERT_BASIC_AUTH_PASS"`
	MEDIA_CONVERT_SAVE_FILE_NAME  string `env:"name:MEDIA_CONVERT_SAVE_FILE_NAME"`

	APP_PORT int    `env:"name:APP_PORT"`
	DB_HOST  string `env:"name:DB_HOST"`
	DB_PORT  string `env:"name:DB_PORT"`
	DB_NAME  string `env:"name:DB_NAME"`
	DB_USER  string `env:"name:DB_USER"`
	DB_PASS  string `env:"name:DB_PASS"`

	SERVICE_KEY  string `env:"name:SERVICE_KEY"`
	AUTH_KEY     string `env:"name:AUTH_KEY"`
	AUTH_SVC_URL string `env:"name:AUTH_SVC_URL"`
	SERVICE_NAME string `env:"name:SERVICE_NAME"`

	CORS_ALLOWED_ORIGINS string `env:"name:CORS_ALLOWED_ORIGINS"`
	ENV                  string `env:"name:ENV"`
	INTERVIEW_SVC_URL    string `env:"INTERVIEW_SVC_URL"`
	INTERVIEW_SVC_NAME   string `env:"INTERVIEW_SVC_NAME"`

	KAFKA_ENDPOINT              string `env:"KAFKA_ENDPOINT"`
	KAFKA_TRIGGER_ANALYZE_TOPIC string `env:"KAFKA_TRIGGER_ANALYZE_TOPIC"`
	KAFKA_RESULTS_TOPIC         string `env:"KAFKA_RESULTS_TOPIC"`
}

type AnalyzeVideoPresignedUrls struct {
	Source_Url      string
	Destination_url string
}
