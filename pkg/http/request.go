package http

type RequestError struct {
	status int
	msg    string
}

func (r *RequestError) Error() string {
	return r.msg
}
