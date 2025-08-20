package fixedPoint

import (
	"bufio"
	"os/exec"
	"strconv"
	"strings"
	"testing"
)

var errorMap = map[string]error{
	"None":        nil,
	"Overflow":    ErrOverflow,
	"NegOverflow": ErrNegOverflow,
	"Underflow":   ErrUnderflow,
	"DivByZero":   ErrDivByZero,
	"DomainError": ErrDomain,
}

type OneArgTestCase64 struct {
	A           uint64
	Expected    uint64
	err         error
	Description string
}

type TwoArgTestCase64 struct {
	A           uint64
	B           uint64
	Expected    uint64
	err         error
	Description string
}

type ThreeArgTestCase64 struct {
	A           uint64
	B           uint64
	C           uint64
	Expected    uint64
	err         error
	Description string
}

type TestState struct {
	outType      string
	operation    string
	round        string
	successCount int
	failureCount int
}

// A useful function to debug a specific test case. You can just copy/paste the test case values
// from a failing test's log output into this function and then debug it.
func TestDebugOneArgTestCase64(t *testing.T) {

	tc := OneArgTestCase64{
		A:        0x7ffffffffa0a1eff,
		Expected: 0xfffffffba5862754,
		err:      nil,
	}

	a := Fix64(tc.A)
	res, err := a.Cos()

	// Used for debugging clampAngle
	// temp, sign := clampAngle64Test(a)
	// res, err := temp.ApplySign(sign)

	var errorAmount uint64 = 0

	if uint64(res) > tc.Expected {
		errorAmount = uint64(res) - tc.Expected
	} else {
		errorAmount = tc.Expected - uint64(res)
	}

	t.Logf("(0x%016x) = 0x%016x, %v; want 0x%016x, %v (±%d)",
		tc.A, uint64(res), err, tc.Expected, tc.err, errorAmount)
}

func TestDebugTwoArgTestCase64(t *testing.T) {
	// t.Skip()

	tc := TwoArgTestCase64{
		A:        0x0000000000000001,
		B:        0x0000000005f5e101,
		Expected: 0x0000000000000001,
		err:      nil,
	}

	a := Fix64(tc.A)
	b := Fix64(tc.B)

	res, err := a.Div(b, RoundHalfUp)

	// a128 := a.ToUFix128()
	// b128 := b.ToFix128()
	// res128, err := a128.powNearOne(b128)

	// res, _ := res128.ToUFix64()

	var errorAmount uint64 = 0

	if uint64(res) > tc.Expected {
		errorAmount = uint64(res) - tc.Expected
	} else {
		errorAmount = tc.Expected - uint64(res)
	}

	t.Logf("(0x%016x, 0x%016x) = 0x%016x, %v; want 0x%016x, %v (±%d)",
		tc.A, tc.B, res, err, tc.Expected, tc.err, errorAmount)
}

// This line is used to tell the Go toolchain that the code in this file depends the Python scripts
//go:generate sh -c "stat generators/genTestData.py > /dev/null"
//go:generate sh -c "stat generators/data64.py > /dev/null"

func OneArgTestChannel64(t *testing.T, state TestState) chan OneArgTestCase64 {
	if state.round == "" {
		state.round = "ROUND_DOWN"
	}

	cmd := exec.Command("uv", "run", "genTestData.py", state.outType, state.operation, state.round)
	cmd.Dir = "./generators"

	// Get a pipe to Python's stdout
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	ch := make(chan OneArgTestCase64)

	go func() {
		defer close(ch)
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimPrefix(line, "(")
			line = strings.TrimSuffix(line, ")")

			parts := strings.Split(line, ", ")
			if len(parts) != 4 {
				t.Log("Unexpected data format from Python script: ", line)
				t.Fail()
			}

			a, _ := strconv.ParseUint(parts[0], 0, 64)
			expected, _ := strconv.ParseUint(parts[1], 0, 64)
			errorString := parts[2]
			message := strings.Trim(parts[3], "\"")

			ch <- OneArgTestCase64{
				A:           a,
				Expected:    expected,
				err:         errorMap[errorString],
				Description: message,
			}
		}

		if err := scanner.Err(); err != nil {
			t.Log(err)
			t.Fail()
		}
	}()

	return ch
}

func TwoArgTestChannel64(t *testing.T, state TestState) chan TwoArgTestCase64 {
	if state.round == "" {
		state.round = "ROUND_DOWN"
	}

	cmd := exec.Command("uv", "run", "genTestData.py", state.outType, state.operation, state.round)
	cmd.Dir = "./generators"

	// Get a pipe to Python's stdout
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	ch := make(chan TwoArgTestCase64)

	go func() {
		defer close(ch)
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimPrefix(line, "(")
			line = strings.TrimSuffix(line, ")")

			parts := strings.Split(line, ", ")
			if len(parts) != 5 {
				t.Log("Unexpected data format from Python script: ", line)
				t.Fail()
			}

			a, _ := strconv.ParseUint(parts[0], 0, 64)
			b, _ := strconv.ParseUint(parts[1], 0, 64)
			expected, _ := strconv.ParseUint(parts[2], 0, 64)
			errorString := parts[3]
			message := strings.Trim(parts[4], "\"")

			ch <- TwoArgTestCase64{
				A:           a,
				B:           b,
				Expected:    expected,
				err:         errorMap[errorString],
				Description: message,
			}
		}

		if err := scanner.Err(); err != nil {
			t.Log(err)
			t.Fail()
		}
	}()

	return ch
}

func ThreeArgTestChannel64(t *testing.T, state TestState) chan ThreeArgTestCase64 {
	if state.round == "" {
		state.round = "ROUND_DOWN"
	}

	cmd := exec.Command("uv", "run", "genTestData.py", state.outType, state.operation, state.round)
	cmd.Dir = "./generators"

	// Get a pipe to Python's stdout
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	ch := make(chan ThreeArgTestCase64)

	go func() {
		defer close(ch)
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimPrefix(line, "(")
			line = strings.TrimSuffix(line, ")")

			parts := strings.Split(line, ", ")
			if len(parts) != 6 {
				t.Log("Unexpected data format from Python script: ", line)
				t.Fail()
			}

			a, _ := strconv.ParseUint(parts[0], 0, 64)
			b, _ := strconv.ParseUint(parts[1], 0, 64)
			c, _ := strconv.ParseUint(parts[2], 0, 64)
			expected, _ := strconv.ParseUint(parts[3], 0, 64)
			errorString := parts[4]
			message := strings.Trim(parts[5], "\"")

			ch <- ThreeArgTestCase64{
				A:           a,
				B:           b,
				C:           c,
				Expected:    expected,
				err:         errorMap[errorString],
				Description: message,
			}
		}

		if err := scanner.Err(); err != nil {
			t.Log(err)
			t.Fail()
		}
	}()

	return ch
}

func OneArgResultCheck64(t *testing.T, ts *TestState, tc OneArgTestCase64, actualResult uint64, actualErr error) bool {
	success := true

	if tc.err != nil || actualErr != nil {
		if actualErr != tc.err {
			t.Errorf("%s (0x%016x) returned error: %v, want: %v (%s)",
				ts.operation+ts.outType, tc.A, actualErr, tc.err, tc.Description)
			success = false
		}
	} else if actualResult != tc.Expected {
		var errorAmount uint64

		if actualResult > tc.Expected {
			errorAmount = actualResult - tc.Expected
		} else {
			errorAmount = tc.Expected - actualResult
		}

		t.Errorf("%s (0x%016x) = 0x%016x, want 0x%016x (±%d) (%s)",
			ts.operation+ts.outType, tc.A, actualResult, tc.Expected, errorAmount, tc.Description)

		success = false
	}

	if success {
		ts.successCount++
	} else {
		ts.failureCount++
	}

	if ts.failureCount >= 10 {
		t.Log("Too many failures, stopping test early")
		t.FailNow()
	}

	return true
}

func TwoArgResultCheck64(t *testing.T, ts *TestState, tc TwoArgTestCase64, actualResult uint64, actualErr error) bool {
	success := true

	if tc.err != nil || actualErr != nil {
		if actualErr != tc.err {
			t.Errorf("%s (0x%016x, 0x%016x) returned error: %v, want: %v (%s)",
				ts.operation+ts.outType, tc.A, tc.B, actualErr, tc.err, tc.Description)
			success = false
		}
	}

	if actualResult != tc.Expected {
		var errorAmount uint64

		if actualResult > tc.Expected {
			errorAmount = actualResult - tc.Expected
		} else {
			errorAmount = tc.Expected - actualResult
		}

		t.Errorf("%s (0x%016x, 0x%016x) = 0x%016x, want 0x%016x (±%d) (%s)",
			ts.operation+ts.outType, tc.A, tc.B, actualResult, tc.Expected, errorAmount, tc.Description)
		success = false
	}

	if success {
		ts.successCount++
	} else {
		ts.failureCount++
	}

	if ts.failureCount >= 10 {
		t.Log("Too many failures, stopping test early")
		t.FailNow()
	}

	return true
}

func ThreeArgResultCheck64(t *testing.T, ts *TestState, tc ThreeArgTestCase64, actualResult uint64, actualErr error) bool {
	success := true

	if tc.err != nil || actualErr != nil {
		if actualErr != tc.err {
			t.Errorf("%s (0x%016x, 0x%016x, 0x%016x) returned error: %v, want: %v (%s)",
				ts.operation+ts.outType, tc.A, tc.B, tc.C, actualErr, tc.err, tc.Description)
			success = false
		}
	}

	if actualResult != tc.Expected {
		var errorAmount uint64

		if actualResult > tc.Expected {
			errorAmount = actualResult - tc.Expected
		} else {
			errorAmount = tc.Expected - actualResult
		}

		t.Errorf("%s (0x%016x, 0x%016x, 0x%016x) = 0x%016x, want 0x%016x (±%d) (%s)",
			ts.operation+ts.outType, tc.A, tc.B, tc.C, actualResult, tc.Expected, errorAmount, tc.Description)
		success = false
	}

	if success {
		ts.successCount++
	} else {
		ts.failureCount++
	}

	if ts.failureCount >= 10 {
		t.Log("Too many failures, stopping test early")
		t.FailNow()
	}

	return true
}

func TestAddUFix64(t *testing.T) {
	testState := TestState{
		outType:   "UFix64",
		operation: "Add",
	}

	t.Parallel()

	for tc := range TwoArgTestChannel64(t, testState) {
		a := UFix64(tc.A)
		b := UFix64(tc.B)
		res, err := a.Add(b)

		TwoArgResultCheck64(t, &testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestAddFix64(t *testing.T) {
	testState := TestState{
		outType:   "Fix64",
		operation: "Add",
	}

	t.Parallel()

	for tc := range TwoArgTestChannel64(t, testState) {
		a := Fix64(tc.A)
		b := Fix64(tc.B)
		res, err := a.Add(b)
		TwoArgResultCheck64(t, &testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestSubUFix64(t *testing.T) {
	testState := TestState{
		outType:   "UFix64",
		operation: "Sub",
	}

	t.Parallel()

	for tc := range TwoArgTestChannel64(t, testState) {
		a := UFix64(tc.A)
		b := UFix64(tc.B)
		res, err := a.Sub(b)

		TwoArgResultCheck64(t, &testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestSubFix64(t *testing.T) {
	testState := TestState{
		outType:   "Fix64",
		operation: "Sub",
	}

	t.Parallel()

	for tc := range TwoArgTestChannel64(t, testState) {
		a := Fix64(tc.A)
		b := Fix64(tc.B)
		res, err := a.Sub(b)

		TwoArgResultCheck64(t, &testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestMulUFix64(t *testing.T) {
	testState := TestState{
		outType:   "UFix64",
		operation: "Mul",
	}

	t.Parallel()

	for tc := range TwoArgTestChannel64(t, testState) {
		a := UFix64(tc.A)
		b := UFix64(tc.B)
		res, err := a.Mul(b, RoundTowardZero)

		TwoArgResultCheck64(t, &testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestMulFix64(t *testing.T) {
	testState := TestState{
		outType:   "Fix64",
		operation: "Mul",
	}

	t.Parallel()

	for tc := range TwoArgTestChannel64(t, testState) {
		a := Fix64(tc.A)
		b := Fix64(tc.B)
		res, err := a.Mul(b, RoundTowardZero)

		TwoArgResultCheck64(t, &testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestDivUFix64(t *testing.T) {
	testState := TestState{
		outType:   "UFix64",
		operation: "Div",
	}

	t.Parallel()

	for tc := range TwoArgTestChannel64(t, testState) {
		a := UFix64(tc.A)
		b := UFix64(tc.B)
		res, err := a.Div(b, RoundTowardZero)

		TwoArgResultCheck64(t, &testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestDivFix64(t *testing.T) {
	testState := TestState{
		outType:   "Fix64",
		operation: "Div",
	}

	t.Parallel()

	for tc := range TwoArgTestChannel64(t, testState) {
		a := Fix64(tc.A)
		b := Fix64(tc.B)
		res, err := a.Div(b, RoundTowardZero)

		TwoArgResultCheck64(t, &testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestFMDUFix64(t *testing.T) {
	testState := TestState{
		outType:   "UFix64",
		operation: "FMD",
	}

	t.Parallel()

	for tc := range ThreeArgTestChannel64(t, testState) {
		a := UFix64(tc.A)
		b := UFix64(tc.B)
		c := UFix64(tc.C)
		res, err := a.FMD(b, c, RoundTowardZero)

		ThreeArgResultCheck64(t, &testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestFMDFix64(t *testing.T) {
	testState := TestState{
		outType:   "Fix64",
		operation: "FMD",
	}

	t.Parallel()

	for tc := range ThreeArgTestChannel64(t, testState) {
		a := Fix64(tc.A)
		b := Fix64(tc.B)
		c := Fix64(tc.C)
		res, err := a.FMD(b, c, RoundTowardZero)

		ThreeArgResultCheck64(t, &testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestModUFix64(t *testing.T) {
	testState := TestState{
		outType:   "UFix64",
		operation: "Mod",
	}

	t.Parallel()

	for tc := range TwoArgTestChannel64(t, testState) {
		a := UFix64(tc.A)
		b := UFix64(tc.B)
		res, err := a.Mod(b)

		TwoArgResultCheck64(t, &testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestModFix64(t *testing.T) {
	testState := TestState{
		outType:   "Fix64",
		operation: "Mod",
	}

	t.Parallel()

	for tc := range TwoArgTestChannel64(t, testState) {
		a := Fix64(tc.A)
		b := Fix64(tc.B)
		res, err := a.Mod(b)

		TwoArgResultCheck64(t, &testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestSqrtUFix64(t *testing.T) {
	testState := TestState{
		outType:   "UFix64",
		operation: "Sqrt",
		round:     "ROUND_HALF_UP",
	}

	t.Parallel()

	for tc := range OneArgTestChannel64(t, testState) {
		a := UFix64(tc.A)
		res, err := a.Sqrt()

		OneArgResultCheck64(t, &testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestLnFix64(t *testing.T) {
	testState := TestState{
		outType:   "Fix64",
		operation: "Ln",
		round:     "ROUND_HALF_UP",
	}

	t.Parallel()

	for tc := range OneArgTestChannel64(t, testState) {
		a := UFix64(tc.A)
		res, err := a.Ln()

		OneArgResultCheck64(t, &testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestExpFix64(t *testing.T) {
	testState := TestState{
		outType:   "UFix64",
		operation: "Exp",
		round:     "ROUND_HALF_UP",
	}

	t.Parallel()

	for tc := range OneArgTestChannel64(t, testState) {
		a := Fix64(tc.A)
		res, err := a.Exp()

		OneArgResultCheck64(t, &testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestPowFix64(t *testing.T) {
	testState := TestState{
		outType:   "UFix64",
		operation: "Pow",
		round:     "ROUND_HALF_UP",
	}

	t.Parallel()

	for tc := range TwoArgTestChannel64(t, testState) {
		a := UFix64(tc.A)
		b := Fix64(tc.B)
		res, err := a.Pow(b)

		TwoArgResultCheck64(t, &testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestSinFix64(t *testing.T) {
	testState := TestState{
		outType:   "Fix64",
		operation: "Sin",
		round:     "ROUND_HALF_UP",
	}

	t.Parallel()

	for tc := range OneArgTestChannel64(t, testState) {
		a := Fix64(tc.A)
		res, err := a.Sin()

		OneArgResultCheck64(t, &testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestCosFix64(t *testing.T) {
	testState := TestState{
		outType:   "Fix64",
		operation: "Cos",
		round:     "ROUND_HALF_UP",
	}

	t.Parallel()

	for tc := range OneArgTestChannel64(t, testState) {
		a := Fix64(tc.A)
		res, err := a.Cos()

		OneArgResultCheck64(t, &testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}
