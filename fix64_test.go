/*
 * Copyright Flow Foundation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
	successCount int
	failureCount int
}

func TestRounding64(t *testing.T) {

	t.Parallel()

	a := UFix64(10 * Fix64Scale)
	b := UFix64(11 * Fix64Scale)

	res, err := a.Div(b)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if res != UFix64(90909091) {
		t.Fatalf("Expected 90909091, got %v", res)
	}
}

// A useful function to debug a specific test case. You can just copy/paste the test case values
// from a failing test's log output into this function and then debug it.
func TestDebugOneArgTestCase64(t *testing.T) {

	t.Parallel()

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

	t.Parallel()

	tc := TwoArgTestCase64{
		A:        0x0000000005f5e101,
		B:        0x01b69b4ba630f34e,
		Expected: 0x0000000005f5e0ee,
		err:      nil,
	}

	a := UFix64(tc.A)
	b := Fix64(tc.B)

	res, err := a.Pow(b)

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

func OneArgTestChannel64(t *testing.T, valType string, operation string) chan OneArgTestCase64 {
	cmd := exec.Command("uv", "run", "genTestData.py", valType, operation)
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

func TwoArgTestChannel64(t *testing.T, valType string, operation string) chan TwoArgTestCase64 {
	cmd := exec.Command("uv", "run", "genTestData.py", valType, operation)
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

func ThreeArgTestChannel64(t *testing.T, valType string, operation string) chan ThreeArgTestCase64 {
	cmd := exec.Command("uv", "run", "genTestData.py", valType, operation)
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

	t.Parallel()

	testState := &TestState{
		outType:      "UFix64",
		operation:    "Add",
		successCount: 0,
		failureCount: 0,
	}

	for tc := range TwoArgTestChannel64(t, testState.outType, testState.operation) {
		a := UFix64(tc.A)
		b := UFix64(tc.B)
		res, err := a.Add(b)

		TwoArgResultCheck64(t, testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestAddFix64(t *testing.T) {

	t.Parallel()

	testState := &TestState{
		outType:      "Fix64",
		operation:    "Add",
		successCount: 0,
		failureCount: 0,
	}

	for tc := range TwoArgTestChannel64(t, testState.outType, testState.operation) {
		a := Fix64(tc.A)
		b := Fix64(tc.B)
		res, err := a.Add(b)

		TwoArgResultCheck64(t, testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestSubUFix64(t *testing.T) {

	t.Parallel()

	testState := &TestState{
		outType:      "UFix64",
		operation:    "Sub",
		successCount: 0,
		failureCount: 0,
	}

	for tc := range TwoArgTestChannel64(t, testState.outType, testState.operation) {
		a := UFix64(tc.A)
		b := UFix64(tc.B)
		res, err := a.Sub(b)

		TwoArgResultCheck64(t, testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestSubFix64(t *testing.T) {

	t.Parallel()

	testState := &TestState{
		outType:      "Fix64",
		operation:    "Sub",
		successCount: 0,
		failureCount: 0,
	}

	for tc := range TwoArgTestChannel64(t, testState.outType, testState.operation) {
		a := Fix64(tc.A)
		b := Fix64(tc.B)
		res, err := a.Sub(b)

		TwoArgResultCheck64(t, testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestMulUFix64(t *testing.T) {

	t.Parallel()

	testState := &TestState{
		outType:      "UFix64",
		operation:    "Mul",
		successCount: 0,
		failureCount: 0,
	}

	for tc := range TwoArgTestChannel64(t, testState.outType, testState.operation) {
		a := UFix64(tc.A)
		b := UFix64(tc.B)
		res, err := a.Mul(b)

		TwoArgResultCheck64(t, testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestMulFix64(t *testing.T) {

	t.Parallel()

	testState := &TestState{
		outType:      "Fix64",
		operation:    "Mul",
		successCount: 0,
		failureCount: 0,
	}

	for tc := range TwoArgTestChannel64(t, testState.outType, testState.operation) {
		a := Fix64(tc.A)
		b := Fix64(tc.B)
		res, err := a.Mul(b)

		TwoArgResultCheck64(t, testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestDivUFix64(t *testing.T) {

	t.Parallel()

	testState := &TestState{
		outType:      "UFix64",
		operation:    "Div",
		successCount: 0,
		failureCount: 0,
	}

	for tc := range TwoArgTestChannel64(t, testState.outType, testState.operation) {
		a := UFix64(tc.A)
		b := UFix64(tc.B)
		res, err := a.Div(b)

		TwoArgResultCheck64(t, testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestDivFix64(t *testing.T) {

	t.Parallel()

	testState := &TestState{
		outType:      "Fix64",
		operation:    "Div",
		successCount: 0,
		failureCount: 0,
	}

	for tc := range TwoArgTestChannel64(t, testState.outType, testState.operation) {
		a := Fix64(tc.A)
		b := Fix64(tc.B)
		res, err := a.Div(b)

		TwoArgResultCheck64(t, testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestFMDUFix64(t *testing.T) {

	t.Parallel()

	testState := &TestState{
		outType:      "UFix64",
		operation:    "FMD",
		successCount: 0,
		failureCount: 0,
	}

	for tc := range ThreeArgTestChannel64(t, testState.outType, testState.operation) {
		a := UFix64(tc.A)
		b := UFix64(tc.B)
		c := UFix64(tc.C)
		res, err := a.FMD(b, c)

		ThreeArgResultCheck64(t, testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestFMDFix64(t *testing.T) {

	t.Parallel()

	testState := &TestState{
		outType:      "Fix64",
		operation:    "FMD",
		successCount: 0,
		failureCount: 0,
	}

	for tc := range ThreeArgTestChannel64(t, testState.outType, testState.operation) {
		a := Fix64(tc.A)
		b := Fix64(tc.B)
		c := Fix64(tc.C)
		res, err := a.FMD(b, c)

		ThreeArgResultCheck64(t, testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestModUFix64(t *testing.T) {

	t.Parallel()

	testState := &TestState{
		outType:      "UFix64",
		operation:    "Mod",
		successCount: 0,
		failureCount: 0,
	}

	for tc := range TwoArgTestChannel64(t, testState.outType, testState.operation) {
		a := UFix64(tc.A)
		b := UFix64(tc.B)
		res, err := a.Mod(b)

		TwoArgResultCheck64(t, testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestModFix64(t *testing.T) {

	t.Parallel()

	testState := &TestState{
		outType:      "Fix64",
		operation:    "Mod",
		successCount: 0,
		failureCount: 0,
	}

	for tc := range TwoArgTestChannel64(t, testState.outType, testState.operation) {
		a := Fix64(tc.A)
		b := Fix64(tc.B)
		res, err := a.Mod(b)

		TwoArgResultCheck64(t, testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestSqrtUFix64(t *testing.T) {

	t.Parallel()

	testState := &TestState{
		outType:      "UFix64",
		operation:    "Sqrt",
		successCount: 0,
		failureCount: 0,
	}

	for tc := range OneArgTestChannel64(t, testState.outType, testState.operation) {
		a := UFix64(tc.A)
		res, err := a.Sqrt()

		OneArgResultCheck64(t, testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestLnFix64(t *testing.T) {

	t.Parallel()

	testState := &TestState{
		outType:      "Fix64",
		operation:    "Ln",
		successCount: 0,
		failureCount: 0,
	}

	for tc := range OneArgTestChannel64(t, testState.outType, testState.operation) {
		a := UFix64(tc.A)
		res, err := a.Ln()

		OneArgResultCheck64(t, testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestExpFix64(t *testing.T) {

	t.Parallel()

	testState := &TestState{
		outType:      "UFix64",
		operation:    "Exp",
		successCount: 0,
		failureCount: 0,
	}

	for tc := range OneArgTestChannel64(t, testState.outType, testState.operation) {
		a := Fix64(tc.A)
		res, err := a.Exp()

		OneArgResultCheck64(t, testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestPowFix64(t *testing.T) {

	t.Parallel()

	testState := &TestState{
		outType:      "UFix64",
		operation:    "Pow",
		successCount: 0,
		failureCount: 0,
	}

	for tc := range TwoArgTestChannel64(t, testState.outType, testState.operation) {
		a := UFix64(tc.A)
		b := Fix64(tc.B)
		res, err := a.Pow(b)

		TwoArgResultCheck64(t, testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

// It turns out that for sin, cos, and tan, simply normalizing the input angle into
// the range [-π, π] is a huge potential source of error. We can test this function
// separately, to make sure we aren't introducing errors before we even get to the
// core sin/cos calculations.
// func TestClampFix64(t *testing.T) {
// 	testState := &TestState{
// 		outType:      "Fix64",
// 		operation:    "Clamp",
// 		successCount: 0,
// 		failureCount: 0,
// 	}

// 	t.Parallel()

// 	for tc := range OneArgTestChannel64(t, testState.outType, testState.operation) {
// 		a := Fix64(tc.A)
// 		res, sign := clampAngle64Test(a)

// 		resSigned := Fix64(res).intMul(sign)

// 		OneArgResultCheck64(t, testState, tc, uint64(resSigned), nil)
// 	}
// 	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
// }

func TestSinFix64(t *testing.T) {

	t.Parallel()

	testState := &TestState{
		outType:      "Fix64",
		operation:    "Sin",
		successCount: 0,
		failureCount: 0,
	}

	for tc := range OneArgTestChannel64(t, testState.outType, testState.operation) {
		a := Fix64(tc.A)
		res, err := a.Sin()

		OneArgResultCheck64(t, testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestCosFix64(t *testing.T) {

	t.Parallel()

	testState := &TestState{
		outType:      "Fix64",
		operation:    "Cos",
		successCount: 0,
		failureCount: 0,
	}

	for tc := range OneArgTestChannel64(t, testState.outType, testState.operation) {
		a := Fix64(tc.A)
		res, err := a.Cos()

		OneArgResultCheck64(t, testState, tc, uint64(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}
