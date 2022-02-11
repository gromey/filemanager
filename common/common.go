package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func GetConfig(config string, v interface{}) error {
	data, err := ioutil.ReadFile(config)
	if os.IsNotExist(err) {
		return fmt.Errorf("no config file")
	} else if err != nil {
		return fmt.Errorf("can't read config: %s", err)
	}

	err = json.Unmarshal(data, &v)
	if err != nil {
		return fmt.Errorf("can't unmarshal config: %s", err)
	}

	return nil
}
