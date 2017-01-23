package run

import "testing"

func TestScraperTypeFromSting(t *testing.T) {
	testCases := []struct {
		input    string
		expected ScraperType
	}{

		{"repubblica", Unknown},
		{"finanza.repubblica", FinanzaRepubblica},
		{"milanofinanza", MilanoFinanza},
	}

	for _, test := range testCases {
		actual, _ := ScraperTypeFromString(test.input)
		if actual != test.expected {
			t.Errorf("found %q, expected %q", actual, test.expected)
		}

	}
}
