package utils

import (
	"encoding/json"
)

type InvalidationsOutput []Invalidation

func (i InvalidationsOutput) String() string {
	output := ""
	for idx, inv := range i {
		output += inv.String()
		if idx != len(i)-1 {
			output += "\n"
		}
	}
	return output
}

func (i InvalidationsOutput) Paths() []string {
	output := []string{}
	for _, inv := range i {
		output = append(output, inv.Path)
	}
	return output
}

type Invalidation struct {
	Path         string `json:"path,omitempty"`
	Value        string `json:"value"`
	AnyError     string `json:"any_error,omitempty"`
	FailingRegex string `json:"failing_regex,omitempty"`
}

func (i Invalidation) String() string {
	jsonBytes, _ := json.Marshal(&i)
	return string(jsonBytes)
}
