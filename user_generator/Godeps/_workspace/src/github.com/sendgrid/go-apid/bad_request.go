package apid

type BadRequest struct {
	ErrorText string
}

func (e BadRequest) Error() string {
	return e.ErrorText
}
