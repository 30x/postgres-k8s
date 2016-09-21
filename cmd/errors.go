package cmd

import "bytes"

//InputErrors a type for creating all input errors
type InputErrors struct {
	errors []error
}

//Append append an error to the list
func (e *InputErrors) Append(err error) {

	if e.errors == nil {
		e.errors = []error{err}
	}

	e.errors = append(e.errors, err)
}

//Error print the error
func (e *InputErrors) Error() string {

	var buffer bytes.Buffer

	//iterate all the errors and append them together
	for _, e := range e.errors {
		buffer.WriteString(e.Error())
		buffer.WriteString("\n")
	}

	buffer.WriteString("See command usage below\n\n\n")

	return buffer.String()
}

//HasErrors true if there are errors
func (e *InputErrors) HasErrors() bool {
	return len(e.errors) > 0
}
