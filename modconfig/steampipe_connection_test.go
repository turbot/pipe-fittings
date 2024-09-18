package modconfig

import "testing"

type connectionEquality struct {
	connection1 *SteampipeConnection
	connection2 *SteampipeConnection
	expectation bool
}

var conn1 = &SteampipeConnection{
	Name:   "connection",
	Config: "hclhelpers",
}

var conn1_duplicate = &SteampipeConnection{
	Name:   "connection",
	Config: "hclhelpers",
}

var other_conn = &SteampipeConnection{
	Name:   "connection2",
	Config: "connection_config2",
}

var equalsCases = map[string]connectionEquality{
	"expected_equal":     {connection1: conn1, connection2: conn1_duplicate, expectation: true},
	"not_expected_equal": {connection1: conn1, connection2: other_conn, expectation: false},
}

func TestConnectionEquals(t *testing.T) {
	for caseName, caseData := range equalsCases {
		isEqual := caseData.connection1.Equals(caseData.connection2)
		if caseData.expectation != isEqual {
			t.Errorf(`Test: '%s' FAILED: expected: %v, actual: %v`, caseName, caseData.expectation, isEqual)
		}
	}
}
