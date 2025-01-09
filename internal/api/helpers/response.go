package helpers

import (
	"errors"
	"fmt"
	"net/http"

	"gorm.io/gorm"
)

func SendResponse(w http.ResponseWriter, res []byte) {
	w.Write(res)
}

func GetDBError(res *gorm.DB) error {
	return errors.Join(fmt.Errorf("error:%s", res.Error), fmt.Errorf("query: %s", res.Statement.SQL.String()))
}
