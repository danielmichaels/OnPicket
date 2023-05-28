package request

import (
	"errors"
	"github.com/danielmichaels/onpicket/internal/validator"
	"github.com/go-playground/form/v4"
	"net/http"
	"net/url"
	"strconv"
)

var decoder = form.NewDecoder()

func DecodePostForm(r *http.Request, dst any) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	err = decoder.Decode(dst, r.PostForm)
	if err != nil {
		var invalidDecoderError *form.InvalidDecoderError

		if errors.As(err, &invalidDecoderError) {
			panic(err)
		}
	}

	return err
}

func DecodeQueryString(r *http.Request, dst any) error {
	err := decoder.Decode(dst, r.PostForm)
	if err != nil {
		var invalidDecoderError *form.InvalidDecoderError

		if errors.As(err, &invalidDecoderError) {
			panic(err)
		}
	}

	return err
}

// ReadInt reads string value from the query string and converts it to an integer
// before returning. If no matching key is found it returns the provided default
// value. If the value couldn't be converted to an int, then we record an error
// message in the provided Validator instance.
func ReadInt(qs url.Values, key string, defaultValue int64, v *validator.Validator) int64 {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	// Try to convert the value to an int. If this fails, add an error message to
	// the validator instance and return the default value.
	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError("must be an integer value")
		return defaultValue
	}
	return int64(i)
}
