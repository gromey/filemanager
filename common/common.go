package common

import (
	"encoding/json"
	"fmt"
	"os"
)

func GetConfig(config string, v interface{}) error {
	data, err := os.ReadFile(config)
	if os.IsNotExist(err) {
		return fmt.Errorf("no config file")
	} else if err != nil {
		return fmt.Errorf("can't read config: %s", err)
	}

	if err = json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("can't unmarshal config: %s", err)
	}

	return nil
}
