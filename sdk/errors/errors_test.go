package errors

import (
	"testing"

	"fmt"

	"github.com/stretchr/testify/assert"
)

// TestNewMultiError tests creating a new instance of a MultiError.
func TestNewMultiError(t *testing.T) {
	var testTable = []struct {
		desc   string
		source string
	}{
		{
			desc:   "The source is unspecified",
			source: "",
		},
		{
			desc:   "The source is a simple string",
			source: "test",
		},
		{
			desc:   "The source is a long string",
			source: "a long and complex description of the error source",
		},
	}

	for _, testCase := range testTable {
		merr := NewMultiError(testCase.source)
		assert.IsType(t, &MultiError{}, merr, testCase.desc)
		assert.Equal(t, testCase.source, merr.For, testCase.desc)
		assert.Equal(t, 0, len(merr.Errors), testCase.desc)
	}
}

// TestMultiError_Error tests getting the error string from a MultiError.
func TestMultiError_Error(t *testing.T) {
	var testTable = []struct {
		desc     string
		source   string
		errs     []error
		expected string
	}{
		{
			desc:     "MultiError has no errors",
			source:   "test",
			errs:     []error{},
			expected: "",
		},
		{
			desc:   "MultiError has 1 error, no source",
			source: "",
			errs: []error{
				fmt.Errorf("error 1"),
			},
			expected: "MultiError has 1 error(s) for source: unspecified\nerror 1\n",
		},
		{
			desc:   "MultiError has 1 error, with source",
			source: "test",
			errs: []error{
				fmt.Errorf("error 1"),
			},
			expected: "MultiError has 1 error(s) for source: test\nerror 1\n",
		},
		{
			desc:   "MultiError has multiple errors, no source",
			source: "",
			errs: []error{
				fmt.Errorf("error 1"),
				fmt.Errorf("error 2"),
				fmt.Errorf("error 3"),
			},
			expected: "MultiError has 3 error(s) for source: unspecified\nerror 1\nerror 2\nerror 3\n",
		},
		{
			desc:   "MultiError has multiple errors, with source",
			source: "test",
			errs: []error{
				fmt.Errorf("error 1"),
				fmt.Errorf("error 2"),
				fmt.Errorf("error 3"),
			},
			expected: "MultiError has 3 error(s) for source: test\nerror 1\nerror 2\nerror 3\n",
		},
	}

	for _, testCase := range testTable {
		merr := MultiError{
			Errors: testCase.errs,
			For:    testCase.source,
		}

		errStr := merr.Error()
		assert.Equal(t, testCase.expected, errStr, testCase.desc)
	}
}

// TestMultiError_Add tests adding errors to the MultiError
func TestMultiError_Add(t *testing.T) {
	var testTable = []struct {
		desc        string
		toAdd       []error
		expectedLen int
	}{
		{
			desc:        "Add no errors to a MultiError",
			toAdd:       []error{},
			expectedLen: 0,
		},
		{
			desc: "Add one error to a MultiError",
			toAdd: []error{
				fmt.Errorf("error 1"),
			},
			expectedLen: 1,
		},
		{
			desc: "Add multiple errors to a MultiError",
			toAdd: []error{
				fmt.Errorf("error 1"),
				fmt.Errorf("error 2"),
				fmt.Errorf("error 3"),
			},
			expectedLen: 3,
		},
	}

	for _, testCase := range testTable {
		merr := NewMultiError("test")

		assert.Equal(t, 0, len(merr.Errors), "MultiError should be initialized with no errors")
		for _, e := range testCase.toAdd {
			merr.Add(e)
		}
		assert.Equal(t, testCase.expectedLen, len(merr.Errors), testCase.desc)
	}
}

// TestMultiError_HasErrors tests checking whether or not a MultiError has errors specified.
func TestMultiError_HasErrors(t *testing.T) {
	var testTable = []struct {
		desc     string
		expected bool
		errors   []error
	}{
		{
			desc:     "No errors",
			expected: false,
			errors:   []error{},
		},
		{
			desc:     "Has one error",
			expected: true,
			errors: []error{
				fmt.Errorf("error 1"),
			},
		},
		{
			desc:     "Has multiple errors",
			expected: true,
			errors: []error{
				fmt.Errorf("error 1"),
				fmt.Errorf("error 2"),
				fmt.Errorf("error 3"),
			},
		},
	}

	for _, testCase := range testTable {
		merr := MultiError{
			Errors: testCase.errors,
		}
		actual := merr.HasErrors()
		assert.Equal(t, testCase.expected, actual, testCase.desc)
	}
}

// TestMultiError_Err tests getting an error return from the MultiError.
func TestMultiError_Err(t *testing.T) {
	var testTable = []struct {
		desc    string
		isError bool
		errors  []error
	}{
		{
			desc:    "No errors, should return nil",
			isError: false,
			errors:  []error{},
		},
		{
			desc:    "Has one error, should return MultiError",
			isError: true,
			errors: []error{
				fmt.Errorf("error 1"),
			},
		},
		{
			desc:    "Has multiple errors, should return MultiError",
			isError: true,
			errors: []error{
				fmt.Errorf("error 1"),
				fmt.Errorf("error 2"),
				fmt.Errorf("error 3"),
			},
		},
	}

	for _, testCase := range testTable {
		merr := MultiError{
			Errors: testCase.errors,
		}
		err := merr.Err()

		if testCase.isError {
			assert.Error(t, err, testCase.desc)
			assert.IsType(t, &MultiError{}, err, testCase.desc)
		} else {
			assert.NoError(t, err, testCase.desc)
		}
	}
}
