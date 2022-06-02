package fileio

type FHandler struct {
	wd string // Working directory

}

// MakeEmpties creates the empty files from the
// FHandlers directory structure. Files are
// initialized with zero bytes.
func (fh *FHandler) MakeEmpties() error {

	return nil
}
