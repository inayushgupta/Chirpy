package auth

import (
	"errors"
	"net/http"
	"testing"
)

type RetVal struct {
	value string
	err error
}

type Field struct {
	value http.Header
	expected RetVal
}

var passingTests []Field = []Field{
	Field{
		http.Header{
			"Authorization": []string{"Bearer tokenvalue"},
		},
		RetVal{
			"tokenvalue",
			nil,
		},
	},Field{
		http.Header{
			"Authorization": []string{"tokenvalue"},
		},
		RetVal{
			"",
			errors.New("Prefix absent"),
		},
	},Field{
		http.Header{
			"Authorization": []string{""},
		},
		RetVal{
			"",
			errors.New("Header not supplied"),
		},
	},
}

func TestGetBearerToken(t *testing.T) {

	for _, tx := range passingTests {
		got, err := GetBearerToken(tx.value)
		expcted_value, expected_error := tx.expected.value, tx.expected.err

		if expcted_value != got {
			t.Fail()
		}

		if err != nil && expected_error == nil {
			t.Fail()
		}

		if expected_error != nil && err != nil {
			if expected_error.Error() != err.Error() {
				t.Fail()
			}			
		}

		}
	}
