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
	"errors"
	"os/exec"
	"strconv"
	"strings"
	"testing"
)

type OneArgTestCase128 struct {
	A           raw128
	Expected    raw128
	err         error
	Description string
}

type TwoArgTestCase128 struct {
	A           raw128
	B           raw128
	Expected    raw128
	err         error
	Description string
}

type ThreeArgTestCase128 struct {
	A           raw128
	B           raw128
	C           raw128
	Expected    raw128
	err         error
	Description string
}

// A useful function to debug a specific test case. You can just copy/paste the test case values
// from a failing test's log output into this function and then debug it.
func TestDebugOneArgTestCase128(t *testing.T) {

	t.Parallel()

	tc := OneArgTestCase128{
		// A: raw128{0x25c7a, 0xe56142a41836c0be},
		A:        raw128{0x000000000000d3c2, 0x1bcecceda1000000},
		Expected: raw128{0x0000000000000000, 0x0000000000000000},
		// A:        raw128{0x10fd9, 0xdb44724e2ca06865},
		// Expected: raw128{0x3bf2, 0x70199efc4e0a0e98},
		err: nil,
	}

	a := UFix128(tc.A)
	res, err := a.Ln()

	// Used for debugging clampAngle
	// res, sign := clampAngle128(a)
	// err := error(nil)
	// res = res.intMul(sign)

	// For clampAngle on fix192
	// a := Fix128(tc.A)
	// a192 := a.toFix192()
	// r192, sign := a192.clampAngle()
	// r192 = r192.uintMul(21264757054)
	// res128, err := r192.toFix128()
	// var res Fix128

	// if err == nil {
	// 	res = res128.intMul(sign)
	// }

	var errorAmount raw128
	actualResult := raw128(res)

	if ult128(actualResult, tc.Expected) {
		errorAmount, _ = sub128(tc.Expected, actualResult, 0)
	} else {
		errorAmount, _ = sub128(actualResult, tc.Expected, 0)
	}

	var errorAmountStr string
	if errorAmount.Hi != 0 || errorAmount.Lo > 1000000 {
		errorAmountStr = "±lots"
	} else {
		errorAmountStr = "±" + strconv.FormatUint(uint64(errorAmount.Lo), 10)
	}

	t.Logf("(0x%016x, 0x%016x) = (0x%016x, 0x%016x) %v, want (0x%016x, 0x%016x) %v (%s) (%s)",
		tc.A.Hi, tc.A.Lo, actualResult.Hi, actualResult.Lo, err, tc.Expected.Hi, tc.Expected.Lo, tc.err, errorAmountStr, tc.Description)
}

func TestDebugTwoArgTestCase128(t *testing.T) {
	// t.Skip()

	t.Parallel()

	tc := TwoArgTestCase128{
		A:        raw128{0x000000000000d3c2, 0x1bcecceda0ffffff},
		B:        raw128{0x00000000000069e1, 0x0de76676d07fffff},
		Expected: raw128{0x000000000000d3c2, 0x1bcecceda1000000},
		err:      nil,
	}

	a := UFix128(tc.A)
	b := Fix128(tc.B)
	res, err := a.Pow(b)

	var errorAmount raw128
	actualResult := raw128(res)

	if ult128(actualResult, tc.Expected) {
		errorAmount, _ = sub128(tc.Expected, actualResult, 0)
	} else {
		errorAmount, _ = sub128(actualResult, tc.Expected, 0)
	}

	var errorAmountStr string
	if errorAmount.Hi != 0 || errorAmount.Lo > 1e8 {
		errorAmountStr = "±lots"
	} else {
		errorAmountStr = "±" + strconv.FormatUint(uint64(errorAmount.Lo), 10)
	}

	t.Logf("(0x%016x, 0x%016x)(0x%016x, 0x%016x) = (0x%016x, 0x%016x) %v, want (0x%016x, 0x%016x) %v (%s) (%s)",
		tc.A.Hi, tc.A.Lo, tc.B.Hi, tc.B.Lo, actualResult.Hi, actualResult.Lo, err, tc.Expected.Hi, tc.Expected.Lo, tc.err, errorAmountStr, tc.Description)
}

// This line is used to tell the Go toolchain that the code in this file depends the Python scripts
//go:generate sh -c "stat generators/genTestData.py > /dev/null"
//go:generate sh -c "stat generators/data64.py > /dev/null"

func OneArgTestChannel128(t *testing.T, state *TestState) chan OneArgTestCase128 {
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

	ch := make(chan OneArgTestCase128)

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

			aHi, _ := strconv.ParseUint(parts[0], 0, 64)
			aLo, _ := strconv.ParseUint(parts[1], 0, 64)
			expectedHi, _ := strconv.ParseUint(parts[2], 0, 64)
			expectedLo, _ := strconv.ParseUint(parts[3], 0, 64)
			errorString := parts[4]
			message := strings.Trim(parts[5], "\"")

			ch <- OneArgTestCase128{
				A:           raw128{raw64(aHi), raw64(aLo)},
				Expected:    raw128{raw64(expectedHi), raw64(expectedLo)},
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

func TwoArgTestChannel128(t *testing.T, state *TestState) chan TwoArgTestCase128 {
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

	ch := make(chan TwoArgTestCase128)

	go func() {
		defer close(ch)
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimPrefix(line, "(")
			line = strings.TrimSuffix(line, ")")

			parts := strings.Split(line, ", ")
			if len(parts) != 8 {
				t.Log("Unexpected data format from Python script: ", line)
				t.Fail()
			}

			aHi, _ := strconv.ParseUint(parts[0], 0, 64)
			aLo, _ := strconv.ParseUint(parts[1], 0, 64)
			bHi, _ := strconv.ParseUint(parts[2], 0, 64)
			bLo, _ := strconv.ParseUint(parts[3], 0, 64)
			expectedHi, _ := strconv.ParseUint(parts[4], 0, 64)
			expectedLo, _ := strconv.ParseUint(parts[5], 0, 64)
			errorString := parts[6]
			message := strings.Trim(parts[7], "\"")

			ch <- TwoArgTestCase128{
				A:           raw128{raw64(aHi), raw64(aLo)},
				B:           raw128{raw64(bHi), raw64(bLo)},
				Expected:    raw128{raw64(expectedHi), raw64(expectedLo)},
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

func ThreeArgTestChannel128(t *testing.T, state *TestState) chan ThreeArgTestCase128 {
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

	ch := make(chan ThreeArgTestCase128)

	go func() {
		defer close(ch)
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimPrefix(line, "(")
			line = strings.TrimSuffix(line, ")")

			parts := strings.Split(line, ", ")
			if len(parts) != 10 {
				t.Log("Unexpected data format from Python script: ", line)
				t.Fail()
			}

			aHi, _ := strconv.ParseUint(parts[0], 0, 64)
			aLo, _ := strconv.ParseUint(parts[1], 0, 64)
			bHi, _ := strconv.ParseUint(parts[2], 0, 64)
			bLo, _ := strconv.ParseUint(parts[3], 0, 64)
			cHi, _ := strconv.ParseUint(parts[4], 0, 64)
			cLo, _ := strconv.ParseUint(parts[5], 0, 64)
			expectedHi, _ := strconv.ParseUint(parts[6], 0, 64)
			expectedLo, _ := strconv.ParseUint(parts[7], 0, 64)
			errorString := parts[8]
			message := strings.Trim(parts[9], "\"")

			ch <- ThreeArgTestCase128{
				A:           raw128{raw64(aHi), raw64(aLo)},
				B:           raw128{raw64(bHi), raw64(bLo)},
				C:           raw128{raw64(cHi), raw64(cLo)},
				Expected:    raw128{raw64(expectedHi), raw64(expectedLo)},
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

func OneArgResultCheck128(t *testing.T, ts *TestState, tc OneArgTestCase128, actualResult raw128, actualErr error) bool {
	success := true

	if tc.err != nil || actualErr != nil {
		if !errors.Is(actualErr, tc.err) {
			t.Errorf("%s (0x%016x, 0x%016x) returned error: %v, want: %v (%s)",
				ts.operation+ts.outType, tc.A.Hi, tc.A.Lo, actualErr, tc.err, tc.Description)
			success = false
		}
	} else if actualResult != tc.Expected {
		var errorAmount raw128

		if slt128(actualResult, tc.Expected) {
			errorAmount, _ = sub128(tc.Expected, actualResult, 0)
		} else {
			errorAmount, _ = sub128(actualResult, tc.Expected, 0)
		}

		var errorAmountStr string
		if errorAmount.Hi != 0 || errorAmount.Lo > 1000000 {
			errorAmountStr = "±lots"
		} else {
			errorAmountStr = "±" + strconv.FormatUint(uint64(errorAmount.Lo), 10)
		}

		t.Errorf("%s (0x%016x, 0x%016x) = (0x%016x, 0x%016x), want (0x%016x, 0x%016x) (%s) (%s)",
			ts.operation+ts.outType, tc.A.Hi, tc.A.Lo, actualResult.Hi, actualResult.Lo, tc.Expected.Hi, tc.Expected.Lo, errorAmountStr, tc.Description)

		success = false
	}

	if success {
		ts.successCount++
	} else {
		ts.failureCount++
	}

	if ts.failureCount >= 20 {
		t.Log("Too many failures, stopping test early")
		t.FailNow()
	}

	return true
}

func TwoArgResultCheck128(t *testing.T, ts *TestState, tc TwoArgTestCase128, actualResult raw128, actualErr error) bool {
	success := true

	if tc.err != nil || actualErr != nil {
		if !errors.Is(actualErr, tc.err) {
			t.Errorf("%s ((0x%016x, 0x%016x), (0x%016x, 0x%016x)) returned error: %v, want: %v (%s)",
				ts.operation+ts.outType, tc.A.Hi, tc.A.Lo, tc.B.Hi, tc.B.Lo, actualErr, tc.err, tc.Description)
			success = false
		}
	} else if actualResult != tc.Expected {
		var errorAmount raw128

		if ult128(actualResult, tc.Expected) {
			errorAmount, _ = sub128(tc.Expected, actualResult, 0)
		} else {
			errorAmount, _ = sub128(actualResult, tc.Expected, 0)
		}

		var errorAmountStr string
		if errorAmount.Hi != 0 || errorAmount.Lo > 1000000 {
			errorAmountStr = "±lots"
		} else {
			errorAmountStr = "±" + strconv.FormatUint(uint64(errorAmount.Lo), 10)
		}

		t.Errorf("%s ((0x%016x, 0x%016x), (0x%016x, 0x%016x)) = (0x%016x, 0x%016x), want (0x%016x, 0x%016x) (%s) (%s)",
			ts.operation+ts.outType, tc.A.Hi, tc.A.Lo, tc.B.Hi, tc.B.Lo, actualResult.Hi, actualResult.Lo, tc.Expected.Hi, tc.Expected.Lo, errorAmountStr, tc.Description)

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

func ThreeArgResultCheck128(t *testing.T, ts *TestState, tc ThreeArgTestCase128, actualResult raw128, actualErr error) bool {
	success := true

	if tc.err != nil || actualErr != nil {
		if !errors.Is(actualErr, tc.err) {
			t.Errorf("%s ((0x%016x, 0x%016x), (0x%016x, 0x%016x), (0x%016x, 0x%016x)) returned error: %v, want: %v (%s)",
				ts.operation+ts.outType, tc.A.Hi, tc.A.Lo, tc.B.Hi, tc.B.Lo, tc.C.Hi, tc.C.Lo, actualErr, tc.err, tc.Description)
			success = false
		}
	} else if actualResult != tc.Expected {
		var errorAmount raw128

		if ult128(actualResult, tc.Expected) {
			errorAmount, _ = sub128(tc.Expected, actualResult, 0)
		} else {
			errorAmount, _ = sub128(actualResult, tc.Expected, 0)
		}

		var errorAmountStr string
		if errorAmount.Hi != 0 || errorAmount.Lo > 1000000 {
			errorAmountStr = "±lots"
		} else {
			errorAmountStr = "±" + strconv.FormatUint(uint64(errorAmount.Lo), 10)
		}

		t.Errorf("%s ((0x%016x, 0x%016x), (0x%016x, 0x%016x), (0x%016x, 0x%016x)) = (0x%016x, 0x%016x), want (0x%016x, 0x%016x) (%s) (%s)",
			ts.operation+ts.outType, tc.A.Hi, tc.A.Lo, tc.B.Hi, tc.B.Lo, tc.C.Hi, tc.C.Lo,
			actualResult.Hi, actualResult.Lo, tc.Expected.Hi, tc.Expected.Lo, errorAmountStr, tc.Description)

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

func b2i128(b bool) raw128 {
	if b {
		return raw128(UFix128One)
	}
	return raw128{}
}

func TestLtFix128(t *testing.T) {

	t.Parallel()

	testState := TestState{
		outType:   "Fix128",
		operation: "LessThan",
	}

	for tc := range TwoArgTestChannel128(t, &testState) {
		a := Fix128(tc.A)
		b := Fix128(tc.B)
		res := a.Lt(b)
		TwoArgResultCheck128(t, &testState, tc, b2i128(res), nil)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestLtUFix128(t *testing.T) {

	t.Parallel()

	testState := TestState{
		outType:   "UFix128",
		operation: "LessThan",
	}

	for tc := range TwoArgTestChannel128(t, &testState) {
		a := UFix128(tc.A)
		b := UFix128(tc.B)
		res := a.Lt(b)
		TwoArgResultCheck128(t, &testState, tc, b2i128(res), nil)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestLteFix128(t *testing.T) {

	t.Parallel()

	testState := TestState{
		outType:   "Fix128",
		operation: "LessThanEqual",
	}

	for tc := range TwoArgTestChannel128(t, &testState) {
		a := Fix128(tc.A)
		b := Fix128(tc.B)
		res := a.Lte(b)
		TwoArgResultCheck128(t, &testState, tc, b2i128(res), nil)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestLteUFix128(t *testing.T) {

	t.Parallel()

	testState := TestState{
		outType:   "UFix128",
		operation: "LessThanEqual",
	}

	for tc := range TwoArgTestChannel128(t, &testState) {
		a := UFix128(tc.A)
		b := UFix128(tc.B)
		res := a.Lte(b)
		TwoArgResultCheck128(t, &testState, tc, b2i128(res), nil)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestGtFix128(t *testing.T) {

	t.Parallel()

	testState := TestState{
		outType:   "Fix128",
		operation: "GreaterThan",
	}

	for tc := range TwoArgTestChannel128(t, &testState) {
		a := Fix128(tc.A)
		b := Fix128(tc.B)
		res := a.Gt(b)
		TwoArgResultCheck128(t, &testState, tc, b2i128(res), nil)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestGtUFix128(t *testing.T) {

	t.Parallel()

	testState := TestState{
		outType:   "UFix128",
		operation: "GreaterThan",
	}

	for tc := range TwoArgTestChannel128(t, &testState) {
		a := UFix128(tc.A)
		b := UFix128(tc.B)
		res := a.Gt(b)
		TwoArgResultCheck128(t, &testState, tc, b2i128(res), nil)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestGteFix128(t *testing.T) {

	t.Parallel()

	testState := TestState{
		outType:   "Fix128",
		operation: "GreaterThanEqual",
	}

	for tc := range TwoArgTestChannel128(t, &testState) {
		a := Fix128(tc.A)
		b := Fix128(tc.B)
		res := a.Gte(b)
		TwoArgResultCheck128(t, &testState, tc, b2i128(res), nil)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestGteUFix128(t *testing.T) {

	t.Parallel()

	testState := TestState{
		outType:   "UFix128",
		operation: "GreaterThanEqual",
	}

	for tc := range TwoArgTestChannel128(t, &testState) {
		a := UFix128(tc.A)
		b := UFix128(tc.B)
		res := a.Gte(b)
		TwoArgResultCheck128(t, &testState, tc, b2i128(res), nil)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestAddUFix128(t *testing.T) {

	t.Parallel()

	testState := TestState{
		outType:   "UFix128",
		operation: "Add",
	}

	for tc := range TwoArgTestChannel128(t, &testState) {
		a := UFix128(tc.A)
		b := UFix128(tc.B)
		res, err := a.Add(b)

		TwoArgResultCheck128(t, &testState, tc, raw128(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestAddFix128(t *testing.T) {

	t.Parallel()

	testState := TestState{
		outType:   "Fix128",
		operation: "Add",
	}

	for tc := range TwoArgTestChannel128(t, &testState) {
		a := Fix128(tc.A)
		b := Fix128(tc.B)
		res, err := a.Add(b)

		TwoArgResultCheck128(t, &testState, tc, raw128(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestSubUFix128(t *testing.T) {

	t.Parallel()

	testState := TestState{
		outType:   "UFix128",
		operation: "Sub",
	}

	for tc := range TwoArgTestChannel128(t, &testState) {
		a := UFix128(tc.A)
		b := UFix128(tc.B)
		res, err := a.Sub(b)

		TwoArgResultCheck128(t, &testState, tc, raw128(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestSubFix128(t *testing.T) {

	t.Parallel()

	testState := TestState{
		outType:   "Fix128",
		operation: "Sub",
	}

	for tc := range TwoArgTestChannel128(t, &testState) {
		a := Fix128(tc.A)
		b := Fix128(tc.B)
		res, err := a.Sub(b)

		TwoArgResultCheck128(t, &testState, tc, raw128(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestMulUFix128(t *testing.T) {

	t.Parallel()

	testState := TestState{
		outType:   "UFix128",
		operation: "Mul",
	}

	for tc := range TwoArgTestChannel128(t, &testState) {
		a := UFix128(tc.A)
		b := UFix128(tc.B)
		res, err := a.Mul(b, RoundTowardZero)

		TwoArgResultCheck128(t, &testState, tc, raw128(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestMulFix128(t *testing.T) {

	t.Parallel()

	testState := TestState{
		outType:   "Fix128",
		operation: "Mul",
	}

	for tc := range TwoArgTestChannel128(t, &testState) {
		a := Fix128(tc.A)
		b := Fix128(tc.B)
		res, err := a.Mul(b, RoundTowardZero)

		TwoArgResultCheck128(t, &testState, tc, raw128(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestDivUFix128(t *testing.T) {

	t.Parallel()

	testState := TestState{
		outType:   "UFix128",
		operation: "Div",
	}

	for tc := range TwoArgTestChannel128(t, &testState) {
		a := UFix128(tc.A)
		b := UFix128(tc.B)
		res, err := a.Div(b, RoundTowardZero)

		TwoArgResultCheck128(t, &testState, tc, raw128(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestDivFix128(t *testing.T) {

	t.Parallel()

	testState := TestState{
		outType:   "Fix128",
		operation: "Div",
	}

	for tc := range TwoArgTestChannel128(t, &testState) {
		a := Fix128(tc.A)
		b := Fix128(tc.B)
		res, err := a.Div(b, RoundTowardZero)

		TwoArgResultCheck128(t, &testState, tc, raw128(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestFMDUFix128(t *testing.T) {

	t.Parallel()

	testState := TestState{
		outType:   "UFix128",
		operation: "FMD",
	}

	for tc := range ThreeArgTestChannel128(t, &testState) {
		a := UFix128(tc.A)
		b := UFix128(tc.B)
		c := UFix128(tc.C)
		res, err := a.FMD(b, c, RoundTowardZero)

		ThreeArgResultCheck128(t, &testState, tc, raw128(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestFMDFix128(t *testing.T) {

	t.Parallel()

	testState := TestState{
		outType:   "Fix128",
		operation: "FMD",
	}

	for tc := range ThreeArgTestChannel128(t, &testState) {
		a := Fix128(tc.A)
		b := Fix128(tc.B)
		c := Fix128(tc.C)
		res, err := a.FMD(b, c, RoundTowardZero)

		ThreeArgResultCheck128(t, &testState, tc, raw128(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestModUFix128(t *testing.T) {

	t.Parallel()

	testState := TestState{
		outType:   "UFix128",
		operation: "Mod",
	}

	for tc := range TwoArgTestChannel128(t, &testState) {
		a := UFix128(tc.A)
		b := UFix128(tc.B)
		res, err := a.Mod(b)

		TwoArgResultCheck128(t, &testState, tc, raw128(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestModFix128(t *testing.T) {

	t.Parallel()

	testState := TestState{
		outType:   "Fix128",
		operation: "Mod",
	}

	for tc := range TwoArgTestChannel128(t, &testState) {
		a := Fix128(tc.A)
		b := Fix128(tc.B)
		res, err := a.Mod(b)

		TwoArgResultCheck128(t, &testState, tc, raw128(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestSqrtUFix128(t *testing.T) {

	t.Parallel()

	testState := TestState{
		outType:   "UFix128",
		operation: "Sqrt",
		round:     "ROUND_HALF_UP",
	}

	for tc := range OneArgTestChannel128(t, &testState) {
		a := UFix128(tc.A)
		res, err := a.Sqrt(RoundHalfUp)

		OneArgResultCheck128(t, &testState, tc, raw128(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestLnFix128(t *testing.T) {

	t.Parallel()

	testState := TestState{
		outType:   "Fix128",
		operation: "Ln",
		round:     "ROUND_HALF_UP",
	}

	for tc := range OneArgTestChannel128(t, &testState) {
		a := UFix128(tc.A)
		res, err := a.Ln()

		OneArgResultCheck128(t, &testState, tc, raw128(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestExpFix128(t *testing.T) {

	t.Parallel()

	testState := TestState{
		outType:   "UFix128",
		operation: "Exp",
		round:     "ROUND_HALF_UP",
	}

	for tc := range OneArgTestChannel128(t, &testState) {
		a := Fix128(tc.A)
		res, err := a.Exp()

		OneArgResultCheck128(t, &testState, tc, raw128(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestPowFix128(t *testing.T) {

	t.Parallel()

	testState := TestState{
		outType:   "UFix128",
		operation: "Pow",
		round:     "ROUND_HALF_UP",
	}

	// The following inputs produce known off-by-one errors, the actual error on these inputs is
	// <1e-40, but due to rounding, the error propagates all the way up to the 24th decimal point
	// and becomes significant. We check for and REQUIRE the off-by-one behaviour for these inputs
	// so that all implementations produce the same bit-pattern (as required for reproducibility and
	// compatability with hashing algorithms).
	knownOffByOneCases := []TwoArgTestCase128{
		{
			A:        raw128{0x000000000000d3c2, 0x1bcecceda0ffffff},
			B:        raw128{0x00000000000069e1, 0x0de76676d07fffff},
			Expected: raw128{0x000000000000d3c2, 0x1bcecceda0ffffff},
			err:      nil, Description: ""},
		{
			A:        raw128{0x000000000000d3c2, 0x1bcecceda0ffffff},
			B:        raw128{0xffffffffffff961e, 0xf21899892f800001},
			Expected: raw128{0x000000000000d3c2, 0x1bcecceda1000001},
			err:      nil, Description: ""},
		{
			A:        raw128{0x00000000000034f0, 0x86f3b33b68400001},
			B:        raw128{0x000000000001a784, 0x379d99db42000000},
			Expected: raw128{0x0000000000000d3c, 0x21bcecceda100000},
			err:      nil, Description: ""},
		{
			A:        raw128{0x00000000000034f0, 0x86f3b33b683fffff},
			B:        raw128{0x000000000001a784, 0x379d99db42000000},
			Expected: raw128{0x0000000000000d3c, 0x21bcecceda0fffff},
			err:      nil, Description: ""},
		{
			A:        raw128{0x00000000000069e1, 0x0de76676d0800001},
			B:        raw128{0x0000000000034f08, 0x6f3b33b684000000},
			Expected: raw128{0x0000000000000d3c, 0x21bcecceda100000},
			err:      nil, Description: ""},
		{
			A:        raw128{0x00000000000069e1, 0x0de76676d07fffff},
			B:        raw128{0x0000000000034f08, 0x6f3b33b684000000},
			Expected: raw128{0x0000000000000d3c, 0x21bcecceda0fffff},
			err:      nil, Description: ""},
	}

	for tc := range TwoArgTestChannel128(t, &testState) {
		a := UFix128(tc.A)
		b := Fix128(tc.B)
		res, err := a.Pow(b)
		rawRes := raw128(res)

		if err == nil && rawRes != tc.Expected {
			foundKnownOffByOne := false

			// Check for off-by-one errors that match the known list encoded above
			for _, known := range knownOffByOneCases {
				if known.A == raw128(a) && known.B == raw128(b) {
					var errorAmount raw128

					if ult128(rawRes, tc.Expected) {
						errorAmount, _ = sub128(tc.Expected, rawRes, 0)
					} else {
						errorAmount, _ = sub128(rawRes, tc.Expected, 0)
					}

					// If the returned result is exactly off-by-one AND matches the result in the
					// off-by-one list, we log that we found an expected mismatch and continue the
					// loop (skipping over the result check)
					if isEqual128(errorAmount, raw128{0, 1}) && raw128(res) == known.Expected {
						t.Logf("Known off-by-one case matched for Pow((0x%016x, 0x%016x), (0x%016x, 0x%016x)) = (0x%016x, 0x%016x)",
							raw128(a).Hi, raw128(a).Lo, raw128(b).Hi, raw128(b).Lo, raw128(res).Hi, raw128(res).Lo)

						foundKnownOffByOne = true
					}

					break
				}
			}

			if foundKnownOffByOne {
				continue
			}
		}

		TwoArgResultCheck128(t, &testState, tc, raw128(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestSinFix128(t *testing.T) {

	t.Parallel()

	testState := TestState{
		outType:   "Fix128",
		operation: "Sin",
		round:     "ROUND_HALF_UP",
	}

	for tc := range OneArgTestChannel128(t, &testState) {
		a := Fix128(tc.A)
		res, err := a.Sin()

		OneArgResultCheck128(t, &testState, tc, raw128(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestCosFix128(t *testing.T) {

	t.Parallel()

	testState := TestState{
		outType:   "Fix128",
		operation: "Cos",
		round:     "ROUND_HALF_UP",
	}

	for tc := range OneArgTestChannel128(t, &testState) {
		a := Fix128(tc.A)
		res, err := a.Cos()

		OneArgResultCheck128(t, &testState, tc, raw128(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}
