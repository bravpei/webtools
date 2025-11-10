package request

import (
	"errors"
	"github.com/gookit/validate"
	"github.com/kataras/iris/v12"
	"log/slog"
	"github.com/bravpei/webtools/external/pkg/response"
)

// ControllerTemplate is a template function for handling requests with JSON parameters.
func ControllerTemplate[Params interface{}](ctx iris.Context, f func(p Params) (interface{}, error)) {
	var params Params

	// Parameter parsing
	if err := ctx.ReadBody(&params); err != nil {
		slog.Error("Failed to parse parameters", "error", err)
		_ = ctx.JSON(response.Fail(err.Error()))
		return
	}

	// Parameter validation
	if err := validateParams(params); err != nil {
		slog.Error("Failed to validate parameters", "error", err)
		_ = ctx.JSON(response.Fail(err.Error()))
		return
	}

	// Business logic processing
	data, err := f(params)
	if err != nil {
		slog.Error("Failed to process business logic", "error", err)
		_ = ctx.JSON(response.Fail(err.Error()))
		return
	}

	_ = ctx.JSON(response.Succeed(data))
}

func validateParams(params interface{}) error {
	if v := validate.Struct(params); !v.Validate() {
		return errors.New(v.Errors.One())
	}
	return nil
}
