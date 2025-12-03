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
	Message string `json:"message"`
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

func Fail(message string, data any) Result {
	return Result{
		Status:  failure,
		Message: message,
		Data:    data,
	}
}

func ValidateError(message string) Result {
	return Result{
		Status:  validationError,
		Message: message,
	}
}

func ServerError(err error) Result {
	return Result{
		Status:  serverError,
		Message: statusMessages[serverError],
		Data:    err.Error(),
	}
}
