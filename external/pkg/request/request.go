package request

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/bravpei/webtools/external/pkg/response"
	"github.com/gookit/validate"
	"github.com/kataras/iris/v12"
)

// BusinessResult defines the unified return structure for business logic
type BusinessResult struct {
	Data    any
	ErrCode string
	Error   error
}

// ControllerTemplate is a template function for handling requests with JSON parameters.
func ControllerTemplate[Params any](ctx iris.Context, f func(p Params) BusinessResult) {
	var params Params

	// Parameter parsing
	if err := ctx.ReadBody(&params); err != nil {
		slog.Error(fmt.Sprintf("Failed to parse parameters: %v", err))
		_ = ctx.JSON(response.Fail("PARAM_PARSE_ERROR", err.Error()))
		return
	}

	// Parameter validation
	if err := validateParams(params); err != nil {
		slog.Error(fmt.Sprintf("Failed to validate parameters: %v", err))
		_ = ctx.JSON(response.Fail("PARAM_VALIDATE_ERROR", err.Error()))
		return
	}

	// Business logic processing
	result := f(params)
	if result.Error != nil {
		slog.Error(fmt.Sprintf("Failed to process business logic: %v", result.Error))
		errCode := result.ErrCode
		if errCode == "" {
			errCode = "BUSINESS_ERROR"
		}
		_ = ctx.JSON(response.Fail(errCode, result.Error.Error()))
		return
	}

	_ = ctx.JSON(response.Succeed(result.Data))
}

func validateParams(params any) error {
	if v := validate.Struct(params); !v.Validate() {
		return errors.New(v.Errors.One())
	}
	return nil
}
