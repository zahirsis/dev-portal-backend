package errors

import "encoding/json"

type InputError struct {
	Messages []string
	Input    string
}

func NewInputError(input string, errors []string) error {
	return &InputError{
		Messages: errors,
		Input:    input,
	}
}

func (i *InputError) Error() string {
	jsonError, _ := json.Marshal(i)
	return string(jsonError)
}
