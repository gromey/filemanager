package synchronize

type Config struct {
	FirstPath  string `json:"first_path"`
	SecondPath string `json:"second_path"`
	Mask       struct {
		On        bool     `json:"on"`
		Extension []string `json:"extension"`
		Include   bool     `json:"include"`
		Details   bool     `json:"details"`
	} `json:"mask"`
	GetHash bool `json:"get_hash"`
}
