//go:build !solution

package testequal

import "sort"

// AssertEqual checks that expected and actual are equal.
//
// Marks caller function as having failed but continues execution.
//
// Returns true if arguments are equal.

func sendMessage(t T, msgAndArgs ...interface{}) {
	t.Helper()
	if msgAndArgs == nil {
		t.Errorf("")
		return
	}
	t.Errorf(msgAndArgs[0].(string), msgAndArgs[1:]...)
}

func sliceFromMap(from map[string]string) []string {
	answer := make([]string, len(from))
	for _, val := range from {
		answer = append(answer, val)
	}
	return answer
}

func equal(expected, actual interface{}) bool {
	switch exp := expected.(type) {
	case int, int8, int16, int32, int64, uint8, uint16, uint32, uint64, string:
		return expected == actual
	case map[string]string:
		act, ok := actual.(map[string]string)
		if !ok || len(exp) != len(act) || ((act == nil) != (exp == nil)) {
			return false
		}
		expValues := sliceFromMap(exp)
		actValues := sliceFromMap(act)
		sort.Strings(expValues)
		sort.Strings(actValues)
		for i := 0; i < len(expValues); i++ {
			if expValues[i] != actValues[i] {
				return false
			}
		}
		return true
	case []int:
		act, ok := actual.([]int)
		if !ok || len(exp) != len(act) || ((act == nil) != (exp == nil)) {
			return false
		}
		for i := 0; i < len(exp); i++ {
			if exp[i] != act[i] {
				return false
			}
		}
		return true
	case []byte:
		act, ok := actual.([]byte)
		if !ok || len(exp) != len(act) || ((act == nil) != (exp == nil)) {
			return false
		}
		for i := 0; i < len(exp); i++ {
			if exp[i] != act[i] {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func AssertEqual(t T, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	t.Helper()
	eq := equal(expected, actual)
	if !eq {
		sendMessage(t, msgAndArgs...)
		return false
	}
	return true
}

// AssertNotEqual checks that expected and actual are not equal.
//
// Marks caller function as having failed but continues execution.
//
// Returns true iff arguments are not equal.
func AssertNotEqual(t T, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	t.Helper()
	eq := equal(expected, actual)
	if eq {
		sendMessage(t, msgAndArgs...)
		return false
	}
	return true
}

// RequireEqual does the same as AssertEqual but fails caller test immediately.
func RequireEqual(t T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	eq := equal(expected, actual)
	if !eq {
		sendMessage(t, msgAndArgs...)
		t.FailNow()
	}
}

// RequireNotEqual does the same as AssertNotEqual but fails caller test immediately.
func RequireNotEqual(t T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	eq := equal(expected, actual)
	if eq {
		sendMessage(t, msgAndArgs...)
		t.FailNow()
	}
}
