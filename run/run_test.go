package run

import (
	"strings"
	"testing"
)

func checkErr(t *testing.T, err error, prefix string) {
	if err == nil {
		t.Errorf("Expecting error with prefix %q, found no error.", prefix)
		return
	}
	if !strings.HasPrefix(err.Error(), prefix) {
		t.Errorf("Expecting error with prefix %q, found %q.", prefix, err.Error())
	}
}

func TestArgs(t *testing.T) {
	err := checkArgs(nil, nil)
	checkErr(t, err, "Scrapers must not be nil.")

}
