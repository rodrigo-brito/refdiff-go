package testdata

type IntAlias int
type ChanType chan struct{}

type (
	IntSlice    []int
	StringSlice []string
	A           struct{}
	iA          interface {
		A() int
	}
)

func (IntSlice) Test() {

}
