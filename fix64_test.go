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

type OneArgTestCase struct {
	A           uint64
	Expected    uint64
	err         error
	Description string
}

type TwoArgTestCase struct {
	A           uint64
	B           uint64
	Expected    uint64
	err         error
	Description string
}

type ThreeArgTestCase struct {
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
	successCount int
	failureCount int
}

// A useful function to debug a specific test case. You can just copy/paste the test case values
// from a failing test's log output into this function and then debug it.
func TestDebugOneArgTestCase(t *testing.T) {

	tc := OneArgTestCase{
		A:        0x0000000000000001,
		Expected: 0xffffffff92344596,
		err:      nil,
	}

	a := UFix64(tc.A)
	res, err := a.Ln()

	// Used for debugging clampAngle
	// err := error(nil)
	// if neg {
	// 	res = res.Neg()
	// }

	var errorAmount uint64 = 0

	if uint64(res) > tc.Expected {
		errorAmount = uint64(res) - tc.Expected
	} else {
		errorAmount = tc.Expected - uint64(res)
	}

	t.Logf("(0x%016x) = 0x%016x, %v; want 0x%016x, %v (±%d)",
		tc.A, uint64(res), err, tc.Expected, tc.err, errorAmount)
}

func TestDebugTwoArgTestCase(t *testing.T) {
	t.Skip()

	tc := TwoArgTestCase{
		A:        0x000000001dcd6500,
		B:        0x000000005f5e0fff,
		Expected: 0xd3c21b959f2d0222,
		err:      nil,
	}

	a := UFix64(tc.A)
	b := Fix64(tc.B)
	res, err := a.Pow(b)

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
//go:generate sh -c "stat generators/add64.py > /dev/null"
//go:generate sh -c "stat generators/data64.py > /dev/null"

func OneArgTestChannel(t *testing.T, valType string, operation string) chan OneArgTestCase {
	cmd := exec.Command("uv", "run", "add64.py", valType, operation)
	cmd.Dir = "./generators"

	// Get a pipe to Python's stdout
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	ch := make(chan OneArgTestCase)

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

			ch <- OneArgTestCase{
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

func TwoArgTestChannel(t *testing.T, valType string, operation string) chan TwoArgTestCase {
	cmd := exec.Command("uv", "run", "add64.py", valType, operation)
	cmd.Dir = "./generators"

	// Get a pipe to Python's stdout
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	ch := make(chan TwoArgTestCase)

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

			ch <- TwoArgTestCase{
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

func ThreeArgTestChannel(t *testing.T, valType string, operation string) chan ThreeArgTestCase {
	cmd := exec.Command("uv", "run", "add64.py", valType, operation)
	cmd.Dir = "./generators"

	// Get a pipe to Python's stdout
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	ch := make(chan ThreeArgTestCase)

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

			ch <- ThreeArgTestCase{
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

func OneArgResultCheck(t *testing.T, ts *TestState, tc OneArgTestCase, actualResult uint64, actualErr error) bool {
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

func TwoArgResultCheck(t *testing.T, ts *TestState, tc TwoArgTestCase, actualResult uint64, actualErr error) bool {
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

func ThreeArgResultCheck(t *testing.T, ts *TestState, tc ThreeArgTestCase, actualResult uint64, actualErr error) bool {
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
	testState := &TestState{
		outType:      "UFix64",
		operation:    "Add",
		successCount: 0,
		failureCount: 0,
	}

	t.Parallel()

	for tc := range TwoArgTestChannel(t, testState.outType, testState.operation) {
		a := UFix64(tc.A)
		b := UFix64(tc.B)
		res, err := a.Add(b)

		TwoArgResultCheck(t, testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestAddFix64(t *testing.T) {
	testState := &TestState{
		outType:      "Fix64",
		operation:    "Add",
		successCount: 0,
		failureCount: 0,
	}

	t.Parallel()

	for tc := range TwoArgTestChannel(t, testState.outType, testState.operation) {
		a := Fix64(tc.A)
		b := Fix64(tc.B)
		res, err := a.Add(b)

		TwoArgResultCheck(t, testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestSubUFix64(t *testing.T) {
	testState := &TestState{
		outType:      "UFix64",
		operation:    "Sub",
		successCount: 0,
		failureCount: 0,
	}

	t.Parallel()

	for tc := range TwoArgTestChannel(t, testState.outType, testState.operation) {
		a := UFix64(tc.A)
		b := UFix64(tc.B)
		res, err := a.Sub(b)

		TwoArgResultCheck(t, testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestSubFix64(t *testing.T) {
	testState := &TestState{
		outType:      "Fix64",
		operation:    "Sub",
		successCount: 0,
		failureCount: 0,
	}

	t.Parallel()

	for tc := range TwoArgTestChannel(t, testState.outType, testState.operation) {
		a := Fix64(tc.A)
		b := Fix64(tc.B)
		res, err := a.Sub(b)

		TwoArgResultCheck(t, testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestMulUFix64(t *testing.T) {
	testState := &TestState{
		outType:      "UFix64",
		operation:    "Mul",
		successCount: 0,
		failureCount: 0,
	}

	t.Parallel()

	for tc := range TwoArgTestChannel(t, testState.outType, testState.operation) {
		a := UFix64(tc.A)
		b := UFix64(tc.B)
		res, err := a.Mul(b)

		TwoArgResultCheck(t, testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestMulFix64(t *testing.T) {
	testState := &TestState{
		outType:      "Fix64",
		operation:    "Mul",
		successCount: 0,
		failureCount: 0,
	}

	t.Parallel()

	for tc := range TwoArgTestChannel(t, testState.outType, testState.operation) {
		a := Fix64(tc.A)
		b := Fix64(tc.B)
		res, err := a.Mul(b)

		TwoArgResultCheck(t, testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestDivUFix64(t *testing.T) {
	testState := &TestState{
		outType:      "UFix64",
		operation:    "Div",
		successCount: 0,
		failureCount: 0,
	}

	t.Parallel()

	for tc := range TwoArgTestChannel(t, testState.outType, testState.operation) {
		a := UFix64(tc.A)
		b := UFix64(tc.B)
		res, err := a.Div(b)

		TwoArgResultCheck(t, testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestDivFix64(t *testing.T) {
	testState := &TestState{
		outType:      "Fix64",
		operation:    "Div",
		successCount: 0,
		failureCount: 0,
	}

	t.Parallel()

	for tc := range TwoArgTestChannel(t, testState.outType, testState.operation) {
		a := Fix64(tc.A)
		b := Fix64(tc.B)
		res, err := a.Div(b)

		TwoArgResultCheck(t, testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestFMDUFix64(t *testing.T) {
	testState := &TestState{
		outType:      "UFix64",
		operation:    "FMD",
		successCount: 0,
		failureCount: 0,
	}

	t.Parallel()

	for tc := range ThreeArgTestChannel(t, testState.outType, testState.operation) {
		a := UFix64(tc.A)
		b := UFix64(tc.B)
		c := UFix64(tc.C)
		res, err := a.FMD(b, c)

		ThreeArgResultCheck(t, testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestFMDFix64(t *testing.T) {
	testState := &TestState{
		outType:      "Fix64",
		operation:    "FMD",
		successCount: 0,
		failureCount: 0,
	}

	t.Parallel()

	for tc := range ThreeArgTestChannel(t, testState.outType, testState.operation) {
		a := Fix64(tc.A)
		b := Fix64(tc.B)
		c := Fix64(tc.C)
		res, err := a.FMD(b, c)

		ThreeArgResultCheck(t, testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestSqrtUFix64(t *testing.T) {
	testState := &TestState{
		outType:      "UFix64",
		operation:    "Sqrt",
		successCount: 0,
		failureCount: 0,
	}

	t.Parallel()

	for tc := range OneArgTestChannel(t, testState.outType, testState.operation) {
		a := UFix64(tc.A)
		res, err := a.Sqrt()

		OneArgResultCheck(t, testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestLnFix64(t *testing.T) {
	testState := &TestState{
		outType:      "Fix64",
		operation:    "Ln",
		successCount: 0,
		failureCount: 0,
	}

	t.Parallel()

	for tc := range OneArgTestChannel(t, testState.outType, testState.operation) {
		a := UFix64(tc.A)
		res, err := a.Ln()

		OneArgResultCheck(t, testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestExpFix64(t *testing.T) {
	testState := &TestState{
		outType:      "UFix64",
		operation:    "Exp",
		successCount: 0,
		failureCount: 0,
	}

	t.Parallel()

	for tc := range OneArgTestChannel(t, testState.outType, testState.operation) {
		a := Fix64(tc.A)
		res, err := a.Exp()

		OneArgResultCheck(t, testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestPowFix64(t *testing.T) {
	testState := &TestState{
		outType:      "UFix64",
		operation:    "Pow",
		successCount: 0,
		failureCount: 0,
	}

	t.Parallel()

	for tc := range TwoArgTestChannel(t, testState.outType, testState.operation) {
		a := UFix64(tc.A)
		b := Fix64(tc.B)
		res, err := a.Pow(b)

		TwoArgResultCheck(t, testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

// It turns out that for sin, cos, and tan, simply normalizing the input angle into
// the range [-π, π] is a huge potential source of error. We can test this function
// separately, to make sure we aren't introducing errors before we even get to the
// core sin/cos calculations.
func TestClampFix64(t *testing.T) {
	testState := &TestState{
		outType:      "Fix64",
		operation:    "Clamp",
		successCount: 0,
		failureCount: 0,
	}

	t.Parallel()

	for tc := range OneArgTestChannel(t, testState.outType, testState.operation) {
		a := Fix64(tc.A)
		res, neg := clampAngle64(a)

		if neg {
			res = res.intMul(-1)
		}

		OneArgResultCheck(t, testState, tc, uint64(res), nil)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestSinFix64(t *testing.T) {
	testState := &TestState{
		outType:      "Fix64",
		operation:    "Sin",
		successCount: 0,
		failureCount: 0,
	}

	t.Parallel()

	for tc := range OneArgTestChannel(t, testState.outType, testState.operation) {
		a := Fix64(tc.A)
		res, err := a.Sin()

		OneArgResultCheck(t, testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestCosFix64(t *testing.T) {
	testState := &TestState{
		outType:      "Fix64",
		operation:    "Cos",
		successCount: 0,
		failureCount: 0,
	}

	t.Parallel()

	for tc := range OneArgTestChannel(t, testState.outType, testState.operation) {
		a := Fix64(tc.A)
		res, err := a.Cos()

		OneArgResultCheck(t, testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestTanFix64(t *testing.T) {
	testState := &TestState{
		outType:      "Fix64",
		operation:    "Tan",
		successCount: 0,
		failureCount: 0,
	}

	t.Parallel()

	for tc := range OneArgTestChannel(t, testState.outType, testState.operation) {
		a := Fix64(tc.A)
		res, err := a.Tan()

		OneArgResultCheck(t, testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}
