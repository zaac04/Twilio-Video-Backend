package initializers

import (
	"stargazer/video-recording/config"
	"stargazer/video-recording/internal/comms"
	"stargazer/video-recording/internal/db"
	"stargazer/video-recording/internal/utils"
	"stargazer/video-recording/pkg/yad"
)

func Initialize_backend(env_file string) {
	config.LoadEnv(env_file)
	yad.InitializeLogger()

	// utils.ExitOnError(iam.HasPermissions())
	utils.ExitOnError(db.Pg.Connect())
	utils.ExitOnError(db.Pg.MigrateTables())
	utils.ExitOnError(comms.AuthenticateService())
}
