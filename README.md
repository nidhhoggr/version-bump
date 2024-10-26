# version-bump

[![CI](https://github.com/joe-at-startupmedia/version-bump/actions/workflows/ci.yml/badge.svg)](https://github.com/joe-at-startupmedia/version-bump/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/joe-at-startupmedia/version-bump/graph/badge.svg?token=HHRUHCX1EJ)](https://codecov.io/gh/joe-at-startupmedia/version-bump)
[![Go Report Card](https://goreportcard.com/badge/github.com/joe-at-startupmedia/version-bump/v2)](https://goreportcard.com/report/github.com/joe-at-startupmedia/version-bump/v2)
[![Release](https://img.shields.io/github/v/release/joe-at-startupmedia/version-bump)](https://github.com/joe-at-startupmedia/version-bump/releases/latest)
[![License](https://img.shields.io/github/license/joe-at-startupmedia/version-bump)](LICENSE.md)


## Configuration

**version-bump** has two modes of operation: automatic / manual.
In automatic mode, **version-bump** will try to identify versions of all supported languages in a root of the project (wherever executed).
In a manual mode, **version-bump** will read a configuration file and modify files according to it. It is expected be executed in a root of the project where the configuration file is.

Some languages, have a constant value in a specific file that contains a version, which are fairly easy to increment.
But some languages are leaving that decision to a developer, thus **version-bump** assumes a constant position/value for them as well.

| Settings      | Expected Values                               | Filename                              |
|:-------------:|:---------------------------------------------:|:-------------------------------------:|
| Docker        | `org.opencontainers.image.version` label      | `Dockerfile`                          |
| Go            | String constant named `Version`/`version`     | `*.go`                                |
| JavaScript    | JSON `version` field                          | `package.json`, `package-lock.json`   |

### Manual

1. Create a configuration `.bump` file in a root of the project.
2. Add project languages and their configuration in a form of:

    ```toml
    [ <language-name> ]
    enabled = true/false
    directories = [ <path>, <path>, ... ]
    exclude_files = [ <path>, <path>, ... ]
    ```

    - `<language-name>` - one of `[ 'docker', 'go', 'javascript' ]`
    - `enabled` - default `false`
    - `directories` - default `['.']`
    - `exclude_files` - default `[]`

3. Run **version-bump** in a root of the project: `version-bump [major|minor|patch]`

*Configuration Example:*

```toml
[docker]
enabled = true
directories = [ '.', 'tools/qa' ]

[go]
enabled = true
directories = [ 'server', 'tools/cli', 'tools/qa' ]
exclude_files = [ 'server/server_test.go', 'tools/qa/main_test.go' ]

[javascript]
enabled = true
directories = [ 'client' ]
```

## Remarks

- Versions are expected to be consistent across all files
- With the absence of a `.bump` configuration file,, **version-bump** has all languages enabled
