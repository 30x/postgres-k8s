package cmd

import "bytes"

//InputError the struct to hold all errors
type InputErrors struct {
	errors []string
}

//Add add an error to the input error
func (i *InputErrors) Add(inputError string) {
	if i.errors == nil {
		i.errors = []string{}
	}

	i.errors = append(i.errors, inputError)

}

//HasErrors return true if we have errors
func (i *InputErrors) HasErrors() bool {
	return len(i.errors) > 0
}

func (i *InputErrors) Error() string {
	var buffer bytes.Buffer

	for _, inputError := range i.errors {
		buffer.WriteString(inputError)
		buffer.WriteString("\n")
	}

	return buffer.String()
}
