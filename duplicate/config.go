package duplicate

type Config struct {
	Paths []string `json:"paths"`
	Mask  struct {
		On        bool     `json:"on"`
		Extension []string `json:"extension"`
		Include   bool     `json:"include"`
		Details   bool     `json:"details"`
	} `json:"mask"`
}