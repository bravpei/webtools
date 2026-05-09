package response

type Status int

const (
	success Status = iota
	failure
	validationError
	serverError
)

type Result struct {
	Status  Status `json:"status"`
	ErrCode string `json:"err_code,omitempty"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	TraceId string `json:"trace_id,omitempty"`
}

var statusMessages = map[Status]string{
	success:         "success",
	failure:         "failure",
	validationError: "validationError",
	serverError:     "serverError",
}

func Succeed(data any) Result {
	return Result{
		Status:  success,
		Message: statusMessages[success],
		Data:    data,
	}
}

func Fail(errCode string, message ...string) Result {
	return buildResult(failure, errCode, message...)
}

func ValidateError(errCode string, message ...string) Result {
	return buildResult(validationError, errCode, message...)
}

func ServerError(errCode string, message ...string) Result {
	return buildResult(serverError, errCode, message...)
}

func buildResult(status Status, errCode string, message ...string) Result {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	return Result{
		Status:  status,
		ErrCode: errCode,
		Message: msg,
	}
}
