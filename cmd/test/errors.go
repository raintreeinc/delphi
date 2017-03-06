package test

type Errors struct {
	Context string
	List    []error
}

func (errs Errors) Error() string {
	r := errs.Context + ": "
	for i, err := range errs.List {
		if i > 0 {
			r += "; "
		}
		r += err.Error()
	}
	return r
}

func NewErrors(context string, errs ...error) error {
	set := Errors{Context: context}
	for _, err := range errs {
		if err != nil {
			set.List = append(set.List, err)
		}
	}
	if len(set.List) == 0 {
		return nil
	}
	return set
}
