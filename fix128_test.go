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
		A:        raw128{0x0000000000000000, 0x0000000000000001},
		Expected: raw128{0x0000000000000000, 0x0000000000000001},
		// A:        raw128{0x10fd9, 0xdb44724e2ca06865},
		// Expected: raw128{0x3bf2, 0x70199efc4e0a0e98},
		err: nil,
	}

	a := Fix128(tc.A)
	res, err := a.Sin()

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
		A:        raw128{0x000000000000d3c2, 0x1bcecceda1000001},
		B:        raw128{0x00000000000069e1, 0x0de76676d0800000},
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

func OneArgTestChannel128(t *testing.T, valType string, operation string) chan OneArgTestCase128 {
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

func TwoArgTestChannel128(t *testing.T, valType string, operation string) chan TwoArgTestCase128 {
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

func ThreeArgTestChannel128(t *testing.T, valType string, operation string) chan ThreeArgTestCase128 {
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
		if actualErr != tc.err {
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
		if actualErr != tc.err {
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
		if actualErr != tc.err {
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

func TestAddUFix128(t *testing.T) {

	t.Parallel()

	testState := &TestState{
		outType:      "UFix128",
		operation:    "Add",
		successCount: 0,
		failureCount: 0,
	}

	for tc := range TwoArgTestChannel128(t, testState.outType, testState.operation) {
		a := UFix128(tc.A)
		b := UFix128(tc.B)
		res, err := a.Add(b)

		TwoArgResultCheck128(t, testState, tc, raw128(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestAddFix128(t *testing.T) {

	t.Parallel()

	testState := &TestState{
		outType:      "Fix128",
		operation:    "Add",
		successCount: 0,
		failureCount: 0,
	}

	for tc := range TwoArgTestChannel128(t, testState.outType, testState.operation) {
		a := Fix128(tc.A)
		b := Fix128(tc.B)
		res, err := a.Add(b)

		TwoArgResultCheck128(t, testState, tc, raw128(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestSubUFix128(t *testing.T) {

	t.Parallel()

	testState := &TestState{
		outType:      "UFix128",
		operation:    "Sub",
		successCount: 0,
		failureCount: 0,
	}

	for tc := range TwoArgTestChannel128(t, testState.outType, testState.operation) {
		a := UFix128(tc.A)
		b := UFix128(tc.B)
		res, err := a.Sub(b)

		TwoArgResultCheck128(t, testState, tc, raw128(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestSubFix128(t *testing.T) {

	t.Parallel()

	testState := &TestState{
		outType:      "Fix128",
		operation:    "Sub",
		successCount: 0,
		failureCount: 0,
	}

	for tc := range TwoArgTestChannel128(t, testState.outType, testState.operation) {
		a := Fix128(tc.A)
		b := Fix128(tc.B)
		res, err := a.Sub(b)

		TwoArgResultCheck128(t, testState, tc, raw128(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestMulUFix128(t *testing.T) {

	t.Parallel()

	testState := &TestState{
		outType:      "UFix128",
		operation:    "Mul",
		successCount: 0,
		failureCount: 0,
	}

	for tc := range TwoArgTestChannel128(t, testState.outType, testState.operation) {
		a := UFix128(tc.A)
		b := UFix128(tc.B)
		res, err := a.Mul(b)

		TwoArgResultCheck128(t, testState, tc, raw128(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestMulFix128(t *testing.T) {

	t.Parallel()

	testState := &TestState{
		outType:      "Fix128",
		operation:    "Mul",
		successCount: 0,
		failureCount: 0,
	}

	for tc := range TwoArgTestChannel128(t, testState.outType, testState.operation) {
		a := Fix128(tc.A)
		b := Fix128(tc.B)
		res, err := a.Mul(b)

		TwoArgResultCheck128(t, testState, tc, raw128(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestDivUFix128(t *testing.T) {

	t.Parallel()

	testState := &TestState{
		outType:      "UFix128",
		operation:    "Div",
		successCount: 0,
		failureCount: 0,
	}

	for tc := range TwoArgTestChannel128(t, testState.outType, testState.operation) {
		a := UFix128(tc.A)
		b := UFix128(tc.B)
		res, err := a.Div(b)

		TwoArgResultCheck128(t, testState, tc, raw128(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestDivFix128(t *testing.T) {

	t.Parallel()

	testState := &TestState{
		outType:      "Fix128",
		operation:    "Div",
		successCount: 0,
		failureCount: 0,
	}

	for tc := range TwoArgTestChannel128(t, testState.outType, testState.operation) {
		a := Fix128(tc.A)
		b := Fix128(tc.B)
		res, err := a.Div(b)

		TwoArgResultCheck128(t, testState, tc, raw128(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestFMDUFix128(t *testing.T) {

	t.Parallel()

	testState := &TestState{
		outType:      "UFix128",
		operation:    "FMD",
		successCount: 0,
		failureCount: 0,
	}

	for tc := range ThreeArgTestChannel128(t, testState.outType, testState.operation) {
		a := UFix128(tc.A)
		b := UFix128(tc.B)
		c := UFix128(tc.C)
		res, err := a.FMD(b, c)

		ThreeArgResultCheck128(t, testState, tc, raw128(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestFMDFix128(t *testing.T) {

	t.Parallel()

	testState := &TestState{
		outType:      "Fix128",
		operation:    "FMD",
		successCount: 0,
		failureCount: 0,
	}

	for tc := range ThreeArgTestChannel128(t, testState.outType, testState.operation) {
		a := Fix128(tc.A)
		b := Fix128(tc.B)
		c := Fix128(tc.C)
		res, err := a.FMD(b, c)

		ThreeArgResultCheck128(t, testState, tc, raw128(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestModUFix128(t *testing.T) {

	t.Parallel()

	testState := &TestState{
		outType:      "UFix128",
		operation:    "Mod",
		successCount: 0,
		failureCount: 0,
	}

	for tc := range TwoArgTestChannel128(t, testState.outType, testState.operation) {
		a := UFix128(tc.A)
		b := UFix128(tc.B)
		res, err := a.Mod(b)

		TwoArgResultCheck128(t, testState, tc, raw128(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestModFix128(t *testing.T) {

	t.Parallel()

	testState := &TestState{
		outType:      "Fix128",
		operation:    "Mod",
		successCount: 0,
		failureCount: 0,
	}

	for tc := range TwoArgTestChannel128(t, testState.outType, testState.operation) {
		a := Fix128(tc.A)
		b := Fix128(tc.B)
		res, err := a.Mod(b)

		TwoArgResultCheck128(t, testState, tc, raw128(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestSqrtUFix128(t *testing.T) {

	t.Parallel()

	testState := &TestState{
		outType:      "UFix128",
		operation:    "Sqrt",
		successCount: 0,
		failureCount: 0,
	}

	for tc := range OneArgTestChannel128(t, testState.outType, testState.operation) {
		a := UFix128(tc.A)
		res, err := a.Sqrt()

		OneArgResultCheck128(t, testState, tc, raw128(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestLnFix128(t *testing.T) {

	t.Parallel()

	testState := &TestState{
		outType:      "Fix128",
		operation:    "Ln",
		successCount: 0,
		failureCount: 0,
	}

	for tc := range OneArgTestChannel128(t, testState.outType, testState.operation) {
		a := UFix128(tc.A)
		res, err := a.Ln()

		OneArgResultCheck128(t, testState, tc, raw128(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestExpFix128(t *testing.T) {

	t.Parallel()

	testState := &TestState{
		outType:      "UFix128",
		operation:    "Exp",
		successCount: 0,
		failureCount: 0,
	}

	for tc := range OneArgTestChannel128(t, testState.outType, testState.operation) {
		a := Fix128(tc.A)
		res, err := a.Exp()

		OneArgResultCheck128(t, testState, tc, raw128(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestPowFix128(t *testing.T) {

	t.Parallel()

	testState := &TestState{
		outType:      "UFix128",
		operation:    "Pow",
		successCount: 0,
		failureCount: 0,
	}

	for tc := range TwoArgTestChannel128(t, testState.outType, testState.operation) {
		a := UFix128(tc.A)
		b := Fix128(tc.B)
		res, err := a.Pow(b)

		TwoArgResultCheck128(t, testState, tc, raw128(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestSinFix128(t *testing.T) {

	t.Parallel()

	testState := &TestState{
		outType:      "Fix128",
		operation:    "Sin",
		successCount: 0,
		failureCount: 0,
	}

	for tc := range OneArgTestChannel128(t, testState.outType, testState.operation) {
		a := Fix128(tc.A)
		res, err := a.Sin()

		OneArgResultCheck128(t, testState, tc, raw128(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}

func TestCosFix128(t *testing.T) {

	t.Parallel()

	testState := &TestState{
		outType:      "Fix128",
		operation:    "Cos",
		successCount: 0,
		failureCount: 0,
	}

	for tc := range OneArgTestChannel128(t, testState.outType, testState.operation) {
		a := Fix128(tc.A)
		res, err := a.Cos()

		OneArgResultCheck128(t, testState, tc, raw128(res), err)
	}
	t.Log(testState.operation+testState.outType, testState.successCount, "passed,", testState.failureCount, "failed")
}
