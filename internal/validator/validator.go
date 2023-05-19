package validator

type Validator struct {
	Errors      []string       `json:",omitempty"`
	FieldErrors map[string]any `json:",omitempty"`
}

// HasErrors checks that the Validator struct contains no errors
// Example:
//
//	v := Validator{}
//	v.Check(foo == 0, "field_name", "error message")
//	if v.HasErrors() {
//		return err
//	}
func (v *Validator) HasErrors() bool {
	return len(v.Errors) != 0 || len(v.FieldErrors) != 0
}

func (v *Validator) AddError(message string) {
	if v.Errors == nil {
		v.Errors = []string{}
	}

	v.Errors = append(v.Errors, message)
}

func (v *Validator) AddFieldError(key, message string) {
	if v.FieldErrors == nil {
		v.FieldErrors = map[string]any{}
	}

	if _, exists := v.FieldErrors[key]; !exists {
		v.FieldErrors[key] = message
	}
}

func (v *Validator) Check(ok bool, message string) {
	if !ok {
		v.AddError(message)
	}
}

func (v *Validator) CheckField(ok bool, key, message string) {
	if !ok {
		v.AddFieldError(key, message)
	}
}
