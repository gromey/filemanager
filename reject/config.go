package reject

type Config struct {
	Paths []string `json:"paths"`
	Mask  struct {
		On         bool     `json:"on"`
		Extensions []string `json:"extensions"`
		Include    bool     `json:"include"`
		Details    bool     `json:"details"`
	} `json:"mask"`
	Delete []string `json:"delete"`
	Space  []string `json:"space"`
}
