package initializers

import (
	"stargazer/video-recording/config"
	"stargazer/video-recording/internal/db"
	"stargazer/video-recording/internal/yad"
	"stargazer/video-recording/pkg/aws/iam"
)

func Initialize_backend(env_file string) {
	config.LoadEnv(env_file)
	iam.HasPermissions()
	yad.InitializeLogger()
	db.Pg.Connect()
	db.Pg.MigrateTables()
}
