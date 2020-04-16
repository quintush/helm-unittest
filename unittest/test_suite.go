package unittest

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/lrills/helm-unittest/unittest/snapshot"
	"gopkg.in/yaml.v2"
	v3chart "helm.sh/helm/v3/pkg/chart"
	v2chart "k8s.io/helm/pkg/proto/hapi/chart"
)

// ParseTestSuiteFile parse a suite file at path and returns TestSuite
func ParseTestSuiteFile(suiteFilePath, chartRoute string) (*TestSuite, error) {
	suite := TestSuite{chartRoute: chartRoute}
	content, err := ioutil.ReadFile(suiteFilePath)
	if err != nil {
		return &suite, err
	}

	cwd, _ := os.Getwd()
	absPath, _ := filepath.Abs(suiteFilePath)
	suite.definitionFile, err = filepath.Rel(cwd, absPath)
	if err != nil {
		return &suite, err
	}

	if err := yaml.Unmarshal(content, &suite); err != nil {
		return &suite, err
	}

	return &suite, nil
}

const partialTemplatePrefix string = "_"
const templatePrefix string = "templates"
const subchartPrefix string = "charts"

func findV2TemplateChart(fileName, path string, templates []*v2chart.Template) bool {
	relativeFilePath := getTemplateFileName(fileName)
	for _, template := range templates {
		validateName := template.Name
		if len(path) > 0 {
			validateName = filepath.ToSlash(filepath.Join(path, validateName))
		}
		if validateName == relativeFilePath {
			return true
		}
	}
	return false
}

func findV3TemplateChart(fileName, path string, templates []*v3chart.File) bool {
	relativeFilePath := getTemplateFileName(fileName)
	for _, template := range templates {
		validateName := template.Name
		if len(path) > 0 {
			validateName = filepath.ToSlash(filepath.Join(path, validateName))
		}
		if validateName == relativeFilePath {
			return true
		}
	}
	return false
}

// TestSuite defines scope and templates to render and tests to run
type TestSuite struct {
	Name      string `yaml:"suite"`
	Templates []string
	Tests     []*TestJob
	// where the test suite file located
	definitionFile string
	// route indicate which chart in the dependency hierarchy
	// like "parant-chart", "parent-charts/charts/child-chart"
	chartRoute string
}

// RunV2 runs all the test jobs defined in TestSuite.
func (s *TestSuite) RunV2(
	targetChart *v2chart.Chart,
	snapshotCache *snapshot.Cache,
	result *TestSuiteResult,
) *TestSuiteResult {
	s.polishTestJobsPathInfo()

	result.DisplayName = s.Name
	result.FilePath = s.definitionFile

	preparedChart, err := s.validateV2Chart(targetChart)
	if err != nil {
		result.ExecError = err
		return result
	}

	result.Passed, result.TestsResult = s.runV2TestJobs(
		preparedChart,
		snapshotCache,
	)

	result.countSnapshot(snapshotCache)
	return result
}

// RunV3 runs all the test jobs defined in TestSuite.
func (s *TestSuite) RunV3(
	targetChart *v3chart.Chart,
	snapshotCache *snapshot.Cache,
	result *TestSuiteResult,
) *TestSuiteResult {
	s.polishTestJobsPathInfo()

	result.DisplayName = s.Name
	result.FilePath = s.definitionFile

	preparedChart, err := s.validateV3Chart(targetChart)
	if err != nil {
		result.ExecError = err
		return result
	}

	result.Passed, result.TestsResult = s.runV3TestJobs(
		preparedChart,
		snapshotCache,
	)

	result.countSnapshot(snapshotCache)
	return result
}

// fill file path related info of TestJob
func (s *TestSuite) polishTestJobsPathInfo() {
	for _, test := range s.Tests {
		test.chartRoute = s.chartRoute
		test.definitionFile = s.definitionFile
		if len(s.Templates) > 0 {
			test.defaultTemplatesToAssert = s.Templates
		}
	}
}

func (s *TestSuite) validateV2Chart(targetChart *v2chart.Chart) (*v2chart.Chart, error) {
	suiteIsFromRootChart := len(strings.Split(s.chartRoute, string(filepath.Separator))) <= 1

	if len(s.Templates) == 0 && suiteIsFromRootChart {
		return targetChart, nil
	}

	// check templates and add them in chart dependencies, if from subchart leave it empty
	if suiteIsFromRootChart {
		for _, fileName := range s.Templates {
			found := findV2TemplateChart(fileName, "", targetChart.Templates)

			// If first time not found, check if fileName is found in dependencies.
			if !found {
				for _, dependency := range targetChart.Dependencies {
					chartPath := filepath.ToSlash(filepath.Join(subchartPrefix, dependency.Metadata.Name))
					found = findV2TemplateChart(fileName, chartPath, dependency.Templates)
					if found {
						// If found, break out of the loop.
						break
					}
				}
			}

			// Second time found, the chart is not found.
			if !found {
				return &v2chart.Chart{}, fmt.Errorf(
					"template file `%s` not found in chart",
					getTemplateFileName(fileName),
				)
			}
		}
	}

	return targetChart, nil
}

func (s *TestSuite) validateV3Chart(targetChart *v3chart.Chart) (*v3chart.Chart, error) {
	suiteIsFromRootChart := len(strings.Split(s.chartRoute, string(filepath.Separator))) <= 1

	if len(s.Templates) == 0 && suiteIsFromRootChart {
		return targetChart, nil
	}

	// check templates and add them in chart dependencies, if from subchart leave it empty
	if suiteIsFromRootChart {
		for _, fileName := range s.Templates {
			found := findV3TemplateChart(fileName, "", targetChart.Templates)

			// If first time not found, check if fileName is found in dependencies.
			if !found {
				for _, dependency := range targetChart.Dependencies() {
					chartPath := filepath.ToSlash(filepath.Join(subchartPrefix, dependency.Metadata.Name))
					found = findV3TemplateChart(fileName, chartPath, dependency.Templates)
					if found {
						// If found, break out of the loop.
						break
					}
				}
			}

			if !found {
				return &v3chart.Chart{}, fmt.Errorf(
					"template file `templates/%s` not found in chart",
					fileName,
				)
			}
		}
	}

	return targetChart, nil
}

func (s *TestSuite) runV2TestJobs(
	chart *v2chart.Chart,
	cache *snapshot.Cache,
) (bool, []*TestJobResult) {
	suitePass := true
	jobResults := make([]*TestJobResult, len(s.Tests))

	for idx, testJob := range s.Tests {
		jobResult := testJob.RunV2(chart, cache, &TestJobResult{Index: idx})
		jobResults[idx] = jobResult

		if !jobResult.Passed {
			suitePass = false
		}
	}
	return suitePass, jobResults
}

func (s *TestSuite) runV3TestJobs(
	chart *v3chart.Chart,
	cache *snapshot.Cache,
) (bool, []*TestJobResult) {
	suitePass := true
	jobResults := make([]*TestJobResult, len(s.Tests))

	for idx, testJob := range s.Tests {
		jobResult := testJob.RunV3(chart, cache, &TestJobResult{Index: idx})
		jobResults[idx] = jobResult

		if !jobResult.Passed {
			suitePass = false
		}
	}
	return suitePass, jobResults
}
