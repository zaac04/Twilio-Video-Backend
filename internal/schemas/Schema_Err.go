package schemas

type Login_Response_Error struct {
	Data string `json:"data"`
}

type ErrResponse struct {
	Error string `json:"error" `
	Hint  string `json:"hint,omitempty"`
}
