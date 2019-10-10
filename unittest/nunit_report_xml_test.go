package unittest_test

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	. "github.com/lrills/helm-unittest/unittest"
	"github.com/stretchr/testify/assert"
)

//var tmpNunitTestDir, _ = ioutil.TempDir("", "_suite_tests")
var tmpNunitTestDir = "./"

func TestWriteTestOutputAsNUnitMinimalSuccess(t *testing.T) {
	assert := assert.New(t)
	outputFile := path.Join(tmpNunitTestDir, "Test_Output.xml")
	testSuiteDisplayName := "TestingSuite"
	testCaseDisplayName := "TestCaseSucces"

	expected := NUnitTestResults{
		Environment: NUnitEnvironment{},
		CultureInfo: NUnitCultureInfo{},
		TestSuite: []NUnitTestSuite{
			{
				Type:        TestFixture,
				Name:        testSuiteDisplayName,
				Description: outputFile,
				Success:     "true",
				Executed:    "true",
				Result:      "Success",
				TestCases: []NUnitTestCase{
					{
						Failure:     nil,
						Name:        testCaseDisplayName,
						Description: fmt.Sprintf("%s.%s", testSuiteDisplayName, testCaseDisplayName),
						Success:     "true",
						Executed:    "true",
						Asserts:     "0",
						Result:      "Success",
					},
				},
			},
		},
		Name:   "Helm-Unittest",
		Total:  1,
		Errors: 0,
	}

	given := []*TestSuiteResult{
		{
			DisplayName: testSuiteDisplayName,
			FilePath:    outputFile,
			Passed:      true,
			TestsResult: []*TestJobResult{
				{
					DisplayName: testCaseDisplayName,
					Passed:      true,
				},
			},
		},
	}

	writer, cerr := os.Create(outputFile)
	assert.Nil(cerr)

	// Test the formatter
	sut := NewNUnitReportXML()
	sut.WriteTestOutput(given, false, writer)

	// Don't defer, as we want to close it before stopping the test.
	writer.Close()

	assert.FileExists(outputFile)

	// Unmarshall and validate the output with expected.
	testResult, rerr := os.Open(outputFile)
	assert.Nil(rerr)
	bytevalue, _ := ioutil.ReadAll(testResult)

	var actual NUnitTestResults
	xml.Unmarshal(bytevalue, &actual)

	assert.Equal(expected.Total, actual.Total)
	assert.Equal(expected.Errors, actual.Errors)
	assert.Equal(expected.Failures, actual.Failures)
	validateNUnitTestSuite(assert, expected.TestSuite, actual.TestSuite)

	testResult.Close()
	os.Remove(outputFile)
}

func TestWriteTestOutputAsNUnitWithFailures(t *testing.T) {
	assert := assert.New(t)
	outputFile := path.Join(tmpNunitTestDir, "Test_Failure_Output.xml")
	testSuiteDisplayName := "TestingSuite"
	testCaseSuccessDisplayName := "TestCaseSuccess"
	testCaseFailureDisplayName := "TestCaseFailure"
	assertionFailure := "AssertionFailure"
	assertionType := "equal"
	assertIndex := 0
	failureContent := fmt.Sprintf("\t\t - asserts[%d]%s `%s` fail \n\t\t\t %s \n", assertIndex, "", assertionType, assertionFailure)

	expected := NUnitTestResults{
		Environment: NUnitEnvironment{},
		CultureInfo: NUnitCultureInfo{},
		TestSuite: []NUnitTestSuite{
			{
				Type:        TestFixture,
				Name:        testSuiteDisplayName,
				Description: outputFile,
				Success:     "false",
				Executed:    "true",
				Result:      "Failure",
				TestCases: []NUnitTestCase{
					{
						Failure:     nil,
						Name:        testCaseSuccessDisplayName,
						Description: fmt.Sprintf("%s.%s", testSuiteDisplayName, testCaseSuccessDisplayName),
						Success:     "true",
						Executed:    "true",
						Asserts:     "0",
						Result:      "Success",
					},
					{
						Failure: &NUnitFailure{
							Message:    "Failed",
							StackTrace: failureContent,
						},
						Name:        testCaseFailureDisplayName,
						Description: fmt.Sprintf("%s.%s", testSuiteDisplayName, testCaseFailureDisplayName),
						Success:     "false",
						Executed:    "true",
						Asserts:     "0",
						Result:      "Failure",
					},
				},
			},
		},
		Name:     "Helm-Unittest",
		Total:    2,
		Errors:   0,
		Failures: 1,
	}

	given := []*TestSuiteResult{
		{
			DisplayName: testSuiteDisplayName,
			FilePath:    outputFile,
			Passed:      false,
			TestsResult: []*TestJobResult{
				{
					DisplayName: testCaseSuccessDisplayName,
					Passed:      true,
				},
				{
					DisplayName: testCaseFailureDisplayName,
					Passed:      false,
					AssertsResult: []*AssertionResult{
						{
							Index: 0,
							FailInfo: []string{
								assertionFailure,
							},
							Passed:     false,
							AssertType: assertionType,
							Not:        false,
						},
					},
				},
			},
		},
	}

	writer, cerr := os.Create(outputFile)
	assert.Nil(cerr)

	// Test the formatter
	sut := NewNUnitReportXML()
	sut.WriteTestOutput(given, false, writer)

	// Don't defer, as we want to close it before stopping the test.
	writer.Close()

	assert.FileExists(outputFile)

	// Unmarshall and validate the output with expected.
	testResult, rerr := os.Open(outputFile)
	assert.Nil(rerr)
	bytevalue, _ := ioutil.ReadAll(testResult)

	var actual NUnitTestResults
	xml.Unmarshal(bytevalue, &actual)

	assert.Equal(expected.Total, actual.Total)
	assert.Equal(expected.Errors, actual.Errors)
	assert.Equal(expected.Failures, actual.Failures)
	validateNUnitTestSuite(assert, expected.TestSuite, actual.TestSuite)

	testResult.Close()
	os.Remove(outputFile)
}

func validateNUnitTestSuite(assert *assert.Assertions, expected, actual []NUnitTestSuite) {

	if expected != nil && actual != nil {
		actualLength := len(actual)
		assert.Equal(len(expected), actualLength)

		for i := 0; i < actualLength; i++ {
			assert.Equal(expected[i].Name, actual[i].Name)
			assert.Equal(expected[i].Description, actual[i].Description)
			assert.Equal(expected[i].Success, actual[i].Success)
			assert.Equal(expected[i].Executed, actual[i].Executed)
			assert.Equal(expected[i].Result, actual[i].Result)

			// Validate the testcases
			validatNUnitTestCase(assert, expected[i].TestCases, actual[i].TestCases)

			// Recursive validation loop.
			validateNUnitTestSuite(assert, expected[i].TestSuites, actual[i].TestSuites)
		}
	} else {
		// Verify if both are nil, otherwise it's still a failure.
		assert.True(expected == nil && actual == nil)
	}
}

func validatNUnitTestCase(assert *assert.Assertions, expected, actual []NUnitTestCase) {

	if expected != nil && actual != nil {
		actualLength := len(actual)
		assert.Equal(len(expected), actualLength)

		for i := 0; i < actualLength; i++ {
			assert.Equal(expected[i].Name, actual[i].Name)
			assert.Equal(expected[i].Description, actual[i].Description)
			assert.Equal(expected[i].Success, actual[i].Success)
			assert.Equal(expected[i].Asserts, actual[i].Asserts)
			assert.Equal(expected[i].Executed, actual[i].Executed)
			assert.Equal(expected[i].Result, actual[i].Result)

			if expected[i].Failure != nil || actual[i].Failure != nil {
				assert.Equal(expected[i].Failure.Message, actual[i].Failure.Message)
				assert.Equal(expected[i].Failure.StackTrace, actual[i].Failure.StackTrace)
			} else {
				// Verify if both are nil, otherwise it's still a failure.
				assert.True(expected[i].Failure == nil && actual[i].Failure == nil)
			}
		}
	} else {
		// Verify if both are nil, otherwise it's still a failure.
		assert.True(expected == nil && actual == nil)
	}
}