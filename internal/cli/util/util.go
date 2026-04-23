package util

import (
	"encoding/json"
	"fmt"
	"io"
)

// WriteEnvFile writes secrets in .env format
func WriteEnvFile(w io.Writer, secrets map[string]interface{}) {
	if secrets == nil {
		return
	}
	for k, v := range secrets {
		fmt.Fprintf(w, "%s=%v\n", k, v) //nolint:errcheck
	}
}

// WriteJSONFile writes secrets in JSON format
func WriteJSONFile(w io.Writer, secrets map[string]interface{}) error {
	data, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}
