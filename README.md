# helm unittest

[![CircleCI](https://circleci.com/gh/quintush/helm-unittest.svg?style=svg)](https://circleci.com/gh/quintush/helm-unittest)
[![Go Report Card](https://goreportcard.com/badge/github.com/quintush/helm-unittest)](https://goreportcard.com/report/github.com/quintush/helm-unittest)
[![codecov](https://codecov.io/gh/quintush/helm-unittest/branch/master/graph/badge.svg)](https://codecov.io/gh/quintush/helm-unittest)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=quintush_helm-unittest&metric=alert_status)](https://sonarcloud.io/dashboard?id=quintush_helm-unittest)

Unit test for *helm chart* in YAML to keep your chart consistent and robust!

Feature:
  - write test file in pure YAML
  - render locally with no need of *tiller*
  - create **nothing** on your cluster
  - [define values and release options](./DOCUMENT.md#test-job)
  - [snapshot testing](#snapshot-testing)

## Documentation

If you are ready for writing tests, check the [DOCUMENT](./DOCUMENT.md) for the test API in YAML.

- [Install](#install)
- [Get Started](#get-started)
- [Test Suite File](#test-suite-file)
- [Usage](#usage)
  - [Flags](#flags)
- [Example](#example)
- [Snapshot Testing](#snapshot-testing)
- [Related Projects / Commands](#related-projects--commands)
- [Contributing](#contributing)


## Install

```
$ helm plugin install https://github.com/quintush/helm-unittest
```

It will install the latest version of binary into helm plugin directory.

## Get Started

Add `tests` in `.helmignore` of your chart, and create the following test file at `$YOUR_CHART/tests/deployment_test.yaml`:

```yaml
suite: test deployment
templates:
  - deployment.yaml
tests:
  - it: should work
    set:
      image.tag: latest
    asserts:
      - isKind:
          of: Deployment
      - matchRegex:
          path: metadata.name
          pattern: -my-chart$
      - equal:
          path: spec.template.spec.containers[0].image
          value: nginx:latest
```
and run:

```
$ helm unittest $YOUR_CHART
```

Now there is your first test! ;)  

## Test Suite File

The test suite file is written in pure YAML, and default placed under the `tests/` directory of the chart with suffix `_test.yaml`. You can also have your own suite files arrangement with `-f, --file` option of cli set as the glob patterns of test suite files related to chart directory, like:

```bash
$ helm unittest -f 'my-tests/*.yaml' -f 'more-tests/*.yaml' my-chart
```
Check [DOCUMENT](./DOCUMENT.md) for more details about writing tests.

## Usage

```
$ helm unittest [flags] CHART [...]
```

This renders your charts locally (without tiller) and runs tests
defined in test suite files.

### Flags

```
      --color                  enforce printing colored output even stdout is not a tty. Set to false to disable color
  -f, --file stringArray       glob paths of test files location, default to tests\*_test.yaml (default [tests\*_test.yaml])
  -3, --helm3                  parse helm charts as helm3 charts (default false)
  -h, --help                   help for unittest
  -o, --output-file string     output-type the file-format where testresults are written in, accepted types are (JUnit, NUnit, XUnit) (default XUnit)
  -u, --update-snapshot        update the snapshot cached if needed, make sure you review the change before update
  -s, --with-subchart charts   include tests of the subcharts within charts folder (default true)
```

## Example

Check [`__fixtures__/v2/basic/`](./__fixtures__/v2/basic) for some basic use cases of a simple chart (version < 2).
Check [`__fixtures__/v3/basic/`](./__fixtures__/v3/basic) for some basic use cases of a simple chart (version > 3).

## Snapshot Testing

Sometimes you may just want to keep the rendered manifest not changed between changes without every details asserted. That's the reason for snapshot testing! Check the tests below:

```yaml
templates:
  - deployment.yaml
tests:
  - it: pod spec should match snapshot
    asserts:
      - matchSnapshot:
          path: spec.template.spec
  # or you can snapshot the whole manifest
  - it: manifest should match snapshot
    asserts:
      - matchSnapshot: {}
```

The `matchSnapshot` assertion validate the content rendered the same as cached last time. It fails if the content changed, and you should check and update the cache with `-u, --update-snapshot` option of cli.

```
$ helm unittest -u my-chart
```
The cache files is stored as `__snapshot__/v2/*_test.yaml.snap` at the directory your test file placed, you should add them in version control with your chart.

## Dependend subchart Testing (Requirements.yaml)

If you have dependend subcharts (installed via `helm dependency`) existed in `charts` directory (they don't need to be extracted), it is possible to unittest these from the root chart. This feature can be helpfull to validate if good default values are accidently overwritten within your default helm chart.

```yaml
# $YOUR_CHART/tests/xxx_test.yaml
templates:
  - charts/postgresql/templates/xxx.yaml
tests:
  - it:
    set:
      # this time required to prefix with "postgresql."
      postgresql.somevalue: should_be_scoped
    asserts:
      - ...
```
Note 1: if dependend subcharts uses an alias, use the alias name in the templates.
Note 2: using the folder structure in templates can also be used to unittest templates which are placed in subfolders or unittest subcharts from the rootchart.

Check [`__fixtures__/v2/with-subchart/`](./__fixtures__/v2/with-subchart) or [`__fixtures__/v3/with-subchart/`](./__fixtures__/v3/with-subchart) as an example.

## Tests within subchart

If you have customized subchart (not installed via `helm dependency`) existed in `charts` directory, tests inside would also be executed by default. You can disable this behavior by setting `--with-subchart=false` flag in cli, thus only the tests in root chart will be executed. Notice that the values defined in subchart tests will be automatically scoped, you don't have to add dependency scope yourself:

```yaml
# with-subchart/charts/child-chart/tests/xxx_test.yaml
templates:
  - xxx.yaml
tests:
  - it:
    set:
      # no need to prefix with "child-chart."
      somevalue: should_be_scoped
    asserts:
      - ...
```
Check [`__fixtures__/v2/with-subchart/`](./__fixtures__/v2/with-subchart) or [`__fixtures__/v3/with-subchart/`](./__fixtures__/v3/with-subchart) as an example.

## Related Projects / Commands

This plugin is inspired by [helm-template](https://github.com/technosophos/helm-template), and the idea of snapshot testing and some printing format comes from [jest](https://github.com/facebook/jest).


And there are some other helm commands you might want to use:

- [`helm template`](https://github.com/kubernetes/helm/blob/master/docs/helm/helm_template.md): render the chart and print the output.

- [`helm lint`](https://github.com/kubernetes/helm/blob/master/docs/helm/helm_lint.md): examines a chart for possible issues, useful to validate chart dependencies.

- [`helm test`](https://github.com/kubernetes/helm/blob/master/docs/helm/helm_test.md): test a release with testing pod defined in chart. Note this does create resources on your cluster to verify if your release is correct. Check the [doc](https://github.com/kubernetes/helm/blob/master/docs/chart_tests.md).

## License

MIT

## Contributing

Issues and PRs are welcome!  
Before start developing this plugin, you must have [go] (https://golang.org/doc/install) >= 1.12 installed, and run:

```
git clone git@github.com:quintush/helm-unittest.git
cd helm-unittest
```

And please make CI passed when request a PR which would check following things:

- `gofmt` no changes needed. Please run `gofmt -w -s .` before you commit.
- `go test ./unittest/...` passed.

In some cases you might need to manually fix the tests in `*_test.go`. If the snapshot tests (of the plugin's test code) failed you need to run:

```
UPDATE_SNAPSHOTS=true go test ./unittest/...
```

This update the snapshot cache file and please add them before you commit.
