package unittest_test

import (
	"io/ioutil"
	"path"
	"testing"
	"time"

	"github.com/bradleyjkemp/cupaloy/v2"
	. "github.com/lrills/helm-unittest/unittest"
	"github.com/lrills/helm-unittest/unittest/snapshot"
	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/chart/loader"
	v2util "k8s.io/helm/pkg/chartutil"
)

var tmpdir, _ = ioutil.TempDir("", "_suite_tests")

func makeTestSuiteResultSnapshotable(result *TestSuiteResult) *TestSuiteResult {

	for _, test := range result.TestsResult {
		test.Duration, _ = time.ParseDuration("0s")
	}

	return result
}

func TestV2ParseTestSuiteFileOk(t *testing.T) {
	a := assert.New(t)
	suite, err := ParseTestSuiteFile("../__fixtures__/v2/basic/tests/deployment_test.yaml", "basic")

	a.Nil(err)
	a.Equal("test deployment", suite.Name)
	a.Equal([]string{"deployment.yaml"}, suite.Templates)
	a.Equal("should pass all kinds of assertion", suite.Tests[0].Name)
}

func TestV2ParseTestSuiteFileInSubfolderOk(t *testing.T) {
	a := assert.New(t)
	suite, err := ParseTestSuiteFile("../__fixtures__/v2/with-subfolder/tests/service_test.yaml", "with-subfolder")

	a.Nil(err)
	a.Equal("test service", suite.Name)
	a.Equal([]string{"webserver/service.yaml"}, suite.Templates)
	a.Equal("should pass", suite.Tests[0].Name)
	a.Equal("should render right if values given", suite.Tests[1].Name)
}

func TestV2RunSuiteWithMultipleTemplatesWhenPass(t *testing.T) {
	c, _ := v2util.Load("../__fixtures__/v2/basic")
	suiteDoc := `
suite: validate metadata
templates:
  - deployment.yaml
  - ingress.yaml
  - service.yaml
tests:
  - it: should pass all metadata
    set:
      ingress.enabled: true
    asserts:
      - matchRegex:
          path: metadata.name
          pattern: ^RELEASE-NAME-basic
      - equal:
          path: metadata.labels.app
          value: basic
      - matchRegex:
          path: metadata.labels.chart
          pattern: ^basic-
      - equal:
          path: metadata.labels.release
          value: RELEASE-NAME
      - equal:
          path: metadata.labels.heritage
          value: Tiller
      - matchSnapshot: {}
`
	testSuite := TestSuite{}
	yaml.Unmarshal([]byte(suiteDoc), &testSuite)

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "my_test.yaml"), false)
	suiteResult := testSuite.RunV2(c, cache, &TestSuiteResult{})

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestSuiteResultSnapshotable(suiteResult))

	a.True(suiteResult.Passed)
	a.Nil(suiteResult.ExecError)
	a.Equal(1, len(suiteResult.TestsResult))
	a.Equal("validate metadata", suiteResult.DisplayName)

	a.Equal(uint(4), suiteResult.SnapshotCounting.Created)
	a.Equal(uint(4), suiteResult.SnapshotCounting.Total)
	a.Equal(uint(0), suiteResult.SnapshotCounting.Failed)
	a.Equal(uint(0), suiteResult.SnapshotCounting.Vanished)
}

func TestV2RunSuiteWhenPass(t *testing.T) {
	c, _ := v2util.Load("../__fixtures__/v2/basic")
	suiteDoc := `
suite: test suite name
templates:
  - deployment.yaml
tests:
  - it: should pass
    asserts:
      - equal:
          path: kind
          value: Deployment
      - matchSnapshot: {}
`
	testSuite := TestSuite{}
	yaml.Unmarshal([]byte(suiteDoc), &testSuite)

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "my_test.yaml"), false)
	suiteResult := testSuite.RunV2(c, cache, &TestSuiteResult{})

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestSuiteResultSnapshotable(suiteResult))

	a.True(suiteResult.Passed)
	a.Nil(suiteResult.ExecError)
	a.Equal(1, len(suiteResult.TestsResult))
	a.Equal("test suite name", suiteResult.DisplayName)

	a.Equal(uint(2), suiteResult.SnapshotCounting.Created)
	a.Equal(uint(2), suiteResult.SnapshotCounting.Total)
	a.Equal(uint(0), suiteResult.SnapshotCounting.Failed)
	a.Equal(uint(0), suiteResult.SnapshotCounting.Vanished)
}

func TestV2RunSuiteWhenFail(t *testing.T) {
	c, _ := v2util.Load("../__fixtures__/v2/basic")
	suiteDoc := `
suite: test suite name
templates:
  - deployment.yaml
tests:
  - it: should fail
    asserts:
      - equal:
          path: kind
          value: Pod
`
	testSuite := TestSuite{}
	yaml.Unmarshal([]byte(suiteDoc), &testSuite)

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "my_test.yaml"), false)
	suiteResult := testSuite.RunV2(c, cache, &TestSuiteResult{})

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestSuiteResultSnapshotable(suiteResult))

	a.False(suiteResult.Passed)
	a.Nil(suiteResult.ExecError)
	a.Equal(1, len(suiteResult.TestsResult))
	a.Equal("test suite name", suiteResult.DisplayName)
}

func TestV2RunSuiteWithSubfolderWhenPass(t *testing.T) {
	c, _ := v2util.Load("../__fixtures__/v2/with-subfolder")
	suiteDoc := `
suite: test suite name
templates:
  - db/deployment.yaml
  - webserver/deployment.yaml
tests:
  - it: should pass
    asserts:
      - equal:
          path: kind
          value: Deployment
      - matchSnapshot: {}
`
	testSuite := TestSuite{}
	yaml.Unmarshal([]byte(suiteDoc), &testSuite)

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "my_test.yaml"), false)
	suiteResult := testSuite.RunV2(c, cache, &TestSuiteResult{})

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestSuiteResultSnapshotable(suiteResult))

	a.True(suiteResult.Passed)
	a.Nil(suiteResult.ExecError)
	a.Equal(1, len(suiteResult.TestsResult))
	a.Equal("test suite name", suiteResult.DisplayName)

	a.Equal(uint(2), suiteResult.SnapshotCounting.Created)
	a.Equal(uint(2), suiteResult.SnapshotCounting.Total)
	a.Equal(uint(0), suiteResult.SnapshotCounting.Failed)
	a.Equal(uint(0), suiteResult.SnapshotCounting.Vanished)
}

func TestV3ParseTestSuiteFileOk(t *testing.T) {
	a := assert.New(t)
	suite, err := ParseTestSuiteFile("../__fixtures__/v3/basic/tests/deployment_test.yaml", "basic")

	a.Nil(err)
	a.Equal(suite.Name, "test deployment")
	a.Equal(suite.Templates, []string{"deployment.yaml"})
	a.Equal(suite.Tests[0].Name, "should pass all kinds of assertion")
}

func TestV3RunSuiteWithMultipleTemplatesWhenPass(t *testing.T) {
	c, _ := loader.Load("../__fixtures__/v3/basic")
	suiteDoc := `
suite: validate metadata
templates:
  - deployment.yaml
  - ingress.yaml
  - service.yaml
tests:
  - it: should pass all metadata
    set:
      ingress.enabled: true
    asserts:
      - matchRegex:
          path: metadata.name
          pattern: ^RELEASE-NAME-basic
      - equal:
          path: metadata.labels.app
          value: basic
      - matchRegex:
          path: metadata.labels.chart
          pattern: ^basic-
      - equal:
          path: metadata.labels.release
          value: RELEASE-NAME
      - equal:
          path: metadata.labels.heritage
          value: Helm
      - matchSnapshot: {}
`
	testSuite := TestSuite{}
	yaml.Unmarshal([]byte(suiteDoc), &testSuite)

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "my_test.yaml"), false)
	suiteResult := testSuite.RunV3(c, cache, &TestSuiteResult{})

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestSuiteResultSnapshotable(suiteResult))

	a.True(suiteResult.Passed)
	a.Nil(suiteResult.ExecError)
	a.Equal(1, len(suiteResult.TestsResult))
	a.Equal("validate metadata", suiteResult.DisplayName)

	a.Equal(uint(4), suiteResult.SnapshotCounting.Created)
	a.Equal(uint(4), suiteResult.SnapshotCounting.Total)
	a.Equal(uint(0), suiteResult.SnapshotCounting.Failed)
	a.Equal(uint(0), suiteResult.SnapshotCounting.Vanished)
}

func TestV3RunSuiteWhenPass(t *testing.T) {
	c, _ := loader.Load("../__fixtures__/v3/basic")
	suiteDoc := `
suite: test suite name
templates:
  - deployment.yaml
tests:
  - it: should pass
    asserts:
      - equal:
          path: kind
          value: Deployment
      - matchSnapshot: {}
`
	testSuite := TestSuite{}
	yaml.Unmarshal([]byte(suiteDoc), &testSuite)

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "my_test.yaml"), false)
	suiteResult := testSuite.RunV3(c, cache, &TestSuiteResult{})

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestSuiteResultSnapshotable(suiteResult))

	a.True(suiteResult.Passed)
	a.Nil(suiteResult.ExecError)
	a.Equal(1, len(suiteResult.TestsResult))
	a.Equal("test suite name", suiteResult.DisplayName)

	a.Equal(uint(2), suiteResult.SnapshotCounting.Created)
	a.Equal(uint(2), suiteResult.SnapshotCounting.Total)
	a.Equal(uint(0), suiteResult.SnapshotCounting.Failed)
	a.Equal(uint(0), suiteResult.SnapshotCounting.Vanished)
}

func TestV3RunSuiteWhenFail(t *testing.T) {
	c, _ := loader.Load("../__fixtures__/v3/basic")
	suiteDoc := `
suite: test suite name
templates:
  - deployment.yaml
tests:
  - it: should fail
    asserts:
      - equal:
          path: kind
          value: Pod
`
	testSuite := TestSuite{}
	yaml.Unmarshal([]byte(suiteDoc), &testSuite)

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "my_test.yaml"), false)
	suiteResult := testSuite.RunV3(c, cache, &TestSuiteResult{})

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestSuiteResultSnapshotable(suiteResult))

	a.False(suiteResult.Passed)
	a.Nil(suiteResult.ExecError)
	a.Equal(1, len(suiteResult.TestsResult))
	a.Equal("test suite name", suiteResult.DisplayName)
}

func TestV3RunSuiteWithSubfolderWhenPass(t *testing.T) {
	c, _ := loader.Load("../__fixtures__/v3/with-subfolder")
	suiteDoc := `
suite: test suite name
templates:
  - db/deployment.yaml
  - webserver/deployment.yaml
tests:
  - it: should pass
    asserts:
      - equal:
          path: kind
          value: Deployment
      - matchSnapshot: {}
`
	testSuite := TestSuite{}
	yaml.Unmarshal([]byte(suiteDoc), &testSuite)

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "my_test.yaml"), false)
	suiteResult := testSuite.RunV3(c, cache, &TestSuiteResult{})

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestSuiteResultSnapshotable(suiteResult))

	a.True(suiteResult.Passed)
	a.Nil(suiteResult.ExecError)
	a.Equal(1, len(suiteResult.TestsResult))
	a.Equal("test suite name", suiteResult.DisplayName)

	a.Equal(uint(2), suiteResult.SnapshotCounting.Created)
	a.Equal(uint(2), suiteResult.SnapshotCounting.Total)
	a.Equal(uint(0), suiteResult.SnapshotCounting.Failed)
	a.Equal(uint(0), suiteResult.SnapshotCounting.Vanished)
}
