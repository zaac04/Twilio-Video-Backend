package initializers

import (
	"stargazer/video-recording/internal/db"
	yad "stargazer/video-recording/internal/yad"
	"stargazer/video-recording/pkg/env"
)

func Initialize_backend(env_file string) {
	env.Load_env(env_file)
	yad.InitializeLogger()
	db.Pg.Connect()
	db.Pg.MigrateTables()
}


