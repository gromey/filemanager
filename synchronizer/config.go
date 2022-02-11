package synchronizer

type Config struct {
	FirstPath  string `json:"first_path"`
	SecondPath string `json:"second_path"`
	Mask       struct {
		On      bool     `json:"on"`
		Ext     []string `json:"ext"`
		Include bool     `json:"include"`
		Details bool     `json:"details"`
	} `json:"mask"`
	GetHash bool `json:"get_hash"`
}

//func ConfigsReader(configsFile string) ([]Config, error) {
//	data, err := ioutil.ReadFile(configsFile)
//	if os.IsNotExist(err) {
//		return nil, fmt.Errorf("no config file")
//	} else if err != nil {
//		return nil, fmt.Errorf("could not read config %v", err)
//	}
//
//	configs := make([]Config, 0)
//
//	if err = json.Unmarshal(data, &configs); err != nil {
//		return nil, fmt.Errorf("could not unmarshal config %v", err)
//	}
//
//	return configs, nil
//}
