package env

import (
	"stargazer/video-recording/internal/enums"
	"stargazer/video-recording/internal/utils"

	"github.com/joho/godotenv"
)

func Load_env(fl string) {
	err := godotenv.Load(fl)
	utils.CheckError(err, enums.EnvLoadFailed)
}

func Verify_Env(map[string]string) {

}
