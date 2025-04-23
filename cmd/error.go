package cmd

import (
	"errors"
	"fmt"

	"github.com/aws/smithy-go"
)

func extractMessage(err error) string {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		msg := fmt.Sprintf("[%s] %s", apiErr.ErrorCode(), apiErr.ErrorMessage())
		if msg != "" {
			return msg
		}
		//return apiErr.ErrorCode() // fallback to code like "ParameterNotFound"
	}
	return err.Error() // fallback
}
