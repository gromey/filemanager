package synchronize

const (
	actMatch = act(iota)
	actCreate
	actReplace
	actDelete
	actProblem
)

type Action interface {
	Apply() error
	String() string
}

type act int

func (a act) String() string {
	switch a {
	case actMatch:
		return "Match"
	case actCreate:
		return "Create"
	case actReplace:
		return "Replace"
	case actDelete:
		return "Delete"
	case actProblem:
		return "Problem"
	}
	return "Unknown action"
}
