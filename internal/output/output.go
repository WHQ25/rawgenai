package output

import (
	"encoding/json"
	"fmt"
	"os"
)

type Response struct {
	Success bool   `json:"success"`
	Error   *Error `json:"error,omitempty"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func Success(data any) {
	output, _ := json.Marshal(data)
	fmt.Println(string(output))
}

func Fail(code, message string) {
	resp := Response{
		Success: false,
		Error: &Error{
			Code:    code,
			Message: message,
		},
	}
	output, _ := json.Marshal(resp)
	fmt.Fprintln(os.Stderr, string(output))
	os.Exit(1)
}
