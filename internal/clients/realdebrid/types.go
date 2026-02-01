package realdebrid

type UnrestrictResponse struct {
	ID         string `json:"id"`
	Filename   string `json:"filename"`
	MimeType   string `json:"mimeType"`
	Filesize   int64  `json:"filesize"`
	Link       string `json:"link"`
	Host       string `json:"host"`
	Download   string `json:"download"`
	Streamable int    `json:"streamable"`
}

type ErrorResponse struct {
	Error     string `json:"error"`
	ErrorCode *int   `json:"error_code"`
}

func (e ErrorResponse) Error() string {
	if e.ErrorCode != nil {
		return e.Error + " (code=" + string(rune(*e.ErrorCode)) + ")"
	}
	return e.Error
}
