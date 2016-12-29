package function

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

func UnmarshalFile(fileName string, result interface{}) error {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, result)
}
