package request

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/bravpei/webtools/external/pkg/response"
	"github.com/gookit/validate"
	"github.com/kataras/iris/v12"
)

// ControllerTemplate is a template function for handling requests with JSON parameters.
func ControllerTemplate[Params any](ctx iris.Context, f func(p Params) (any, error)) {
	var params Params

	// Parameter parsing
	if err := ctx.ReadBody(&params); err != nil {
		slog.Error(fmt.Sprintf("Failed to parse parameters: %v", err))
		_ = ctx.JSON(response.Fail(err.Error(), nil))
		return
	}

	// Parameter validation
	if err := validateParams(params); err != nil {
		slog.Error(fmt.Sprintf("Failed to validate parameters: %v", err))
		_ = ctx.JSON(response.Fail(err.Error(), nil))
		return
	}

	// Business logic processing
	data, err := f(params)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to process business logic: %v", err))
		_ = ctx.JSON(response.Fail(err.Error(), data))
		return
	}

	_ = ctx.JSON(response.Succeed(data))
}

func validateParams(params any) error {
	if v := validate.Struct(params); !v.Validate() {
		return errors.New(v.Errors.One())
	}
	return nil
}
