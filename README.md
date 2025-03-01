# version-bump

[![CI](https://github.com/nidhhoggr/version-bump/actions/workflows/ci.yml/badge.svg)](https://github.com/nidhhoggr/version-bump/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/nidhhoggr/version-bump/graph/badge.svg?token=HHRUHCX1EJ)](https://codecov.io/gh/nidhhoggr/version-bump)
[![Go Report Card](https://goreportcard.com/badge/github.com/nidhhoggr/version-bump)](https://goreportcard.com/report/github.com/nidhhoggr/version-bump)
[![Release](https://img.shields.io/github/v/release/nidhhoggr/version-bump)](https://github.com/nidhhoggr/version-bump/releases/latest)
[![License](https://img.shields.io/github/license/nidhhoggr/version-bump)](LICENSE.md)


## Configuration

**version-bump** has two modes of operation: automatic / manual.
In automatic mode, **version-bump** will try to identify versions of all supported languages in the root of a project (wherever executed).
In a manual mode, **version-bump** will read a configuration file to determine which modifications to make. It is expected be executed in the root of the project where the configuration file is located.


### Default Settings

| Language      | Expected Patterns                             | Filename                              |
|:-------------:|:---------------------------------------------:|:-------------------------------------:|
| Docker        | `org.opencontainers.image.version` label      | `Dockerfile`                          |
| Go            | String constant named `Version`/`version`     | `*.go`                                |
| JavaScript    | JSON `version` field                          | `package.json`, `package-lock.json`   |

### Manual

1. Create a configuration `.bump` file in the root of a project.
2. Add project languages and their configuration in a form of:

    ```
    [ language_name ]
    enabled = bool
    directories = [ string, string, ... ]
    exclude_files = [ string, string, ... ]
    files = [ string, string, ... ]
    regex = [string, string, ...]
    ```

    - `[ language_name ]` - one of `[ 'docker', 'go', 'javascript' ]`
    - `enabled` - default `false`
    - `directories` - path default `['.']`
    - `exclude_files` - path default `[]`
    - `files` - an array of glob values to overide the settings default `declared in the langs module`
    - `regex` - an array of regex patterns to overide the settings default `declared in the langs module`
      
3. Run **version-bump** in the root of a project: `version-bump [major|minor|patch] [flags]`

### Generic Language

You can also add additional supported languages by using the `[[generic]]` directive.

```
[[generic]]
name = string
enabled = bool
directories = [ string, string, ... ]
exclude_files = [ string, string, ... ]
files = [ string, string, ... ]
regex = [string, string, ...]
```
 
### Example

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

[[generic]]
name = "markdown"
enabled = true
files = [ "*.md" ]

[[generic]]
name = "yaml"
enabled = false
directories = [ "config" ]
files = [ "*.yml" ]
regex = [ '^version: (?P<version>{{SEMVER_REGEX}})' ]
```

Note: the convenient `{{SEMVER_REGEX}})` variable is substituted for an actual regex pattern matching a semver string.

## Installation

```
go install github.com/nidhhoggr/version-bump/cmd/version-bump@latest
```

## CLI Usage

```
This application increments the semantic version of a project.
It can bump semantic versions in multiple different files at once,
as well as automate prerelease versioning and promotion.

Usage:
  version-bump [major|minor|patch] [flags]

Flags:
      --alpha               alpha Prerelease
      --auto-confirm        disable confirmation prompts and automatically confirm
      --beta                beta Prerelease
      --debug               output debug information to the console
      --disable-prompts     disable passphrase and confirmation prompts. Caution: this will result in unsigned commits, tags and releases!
      --dry-run             perform a dry run without modifying any files or interacting with git
  -h, --help                help for version-bump
      --interactive         enable interactive mode
      --metadata string     provide metadata for the Prerelease
      --passphrase string   provide gpg passphrase as a flag instead of a secure prompt. Caution!
      --rc                  release candidate Prerelease
  -v, --version             version for version-bump
```

<a name="version_types"></a>
## Version Types
Versions can be optionally specified as an argument along with prerelease flags and metadata. Must be one of the following:

* `major`
* `minor`
* `patch`

### Bump A Major Version
```
➜ version-bump major
```  

### Bump A Minor Version
```
➜ version-bump minor
``` 

### Bump A Patch Version
```
➜ version-bump patch
```

### Screenshot
  
![Screenshot 2024-10-28 at 21 06 20](https://github.com/user-attachments/assets/eb9fcace-246d-495d-b744-fb1152ddfa76)

When incrementing a prerelease without updating the version, simply omit the version type argument.

## Prerelease Automation

<a name="prerelease_types"></a>
### Types
Prereleases can be specified as a flag along with metadata. The currently supported prerelease types are the following:

* `alpha`
* `beta`
* `rc`

### Format
Conforming to the [Semver specification](https://semver.org/), Prereleases must be in the following format:

`prerelease-type`.`prerelease-version`+`prerelease-metadata`

Where the following criterion must be met:

* `prerelease-type`: Must be a string value matching one of the [prerelease types](#prerelease_types)
* `prerelease-version`: Must be an integer
* `prerelease-metadata`: An alphanumberic string without special characters beginning with a `+`

<a name="prerelease_alpha"></a>  
### Alpha Prerelease


Must be released from an existing alpha release whose patch is the same by omitting the [version type](#version_types) argument:

```
➜ version-bump --alpha
``` 

![Screenshot 2024-10-28 at 20 58 55](https://github.com/user-attachments/assets/f1679ed9-7464-430c-9396-704128e94435)

Or from a new version by specifying the [version type](#version_types):

```
➜ version-bump [major|minor|patch] --alpha
``` 

![Screenshot 2024-10-28 at 21 04 21](https://github.com/user-attachments/assets/8ef99cdf-8fa1-43b4-8d27-8be987b6f52b)

<a name="prerelease_beta"></a>  
### Beta Prerelease

Must be released from an existing alpha or beta release whose patch is the same by omitting the [version type](#version_types) argument:

```
➜ version-bump --beta
``` 

![Screenshot 2024-10-28 at 21 11 32](https://github.com/user-attachments/assets/cdff9fa2-539a-4c1c-907a-b248d5840ad4)

Attempting to release an [alpha release](#prerelease_alpha) from a [beta release](#prerelease_beta) without specifying the [version type](#version_types) will throw an error:

```
➜ version-bump [major|minor|patch] --beta
``` 

![Screenshot 2024-10-28 at 21 15 44](https://github.com/user-attachments/assets/44006c13-82e4-4bc5-9403-d865d9868654)

<a name="Prerelease_rc"></a>  
### Release Candidate

Similar to [alpha releases](#prerelease_alpha) and [beta releases](#prerelease_beta), the flags must be specified appropriately.

```
➜ version-bump --rc
``` 

![Screenshot 2024-10-28 at 21 22 35](https://github.com/user-attachments/assets/60b6cf57-21e0-4f21-a05b-724714766c5f)

Attempting to release an [alpha release](#prerelease_alpha) or a [beta release](#prerelease_beta) and omitting the [version type](#version_types) argument will produce errors:

```
➜ version-bump [major|minor|patch] --rc
``` 

![Screenshot 2024-10-28 at 21 23 48](https://github.com/user-attachments/assets/58dfe870-9d5c-4a04-87d0-8614b7fb62e3)

### Increment Prerelease Version

Simply specify the same prerelease type of the existing prerelease while omitting the [version type](#version_types) argument. 
It will automatically increment the prerelease version.

```
➜ version-bump [--alpha | --beta | --rc]
``` 

![Screenshot 2024-10-28 at 21 24 59](https://github.com/user-attachments/assets/5870f006-b41d-4de2-b7bd-df87ab5c545c)

### Promote Prerelease

After our Prerelease has been tested and ready for rollout, you can simply provide `patch` as the [version type](#version_types) argument. 
It will remove all of the Prerelease versioning and metadata from the version.

```
➜ version-bump patch
``` 

![Screenshot 2024-10-28 at 21 30 13](https://github.com/user-attachments/assets/18a0e8e2-f351-4dac-82c6-d84e34ddcfd7)

## Version Inconsistencies

Before any modifications are made to the repository, if any version consistencies are detected, `version-bump` will prematurely exit.
This frees you from the hassle of having to run `git stash`.

No modifications will be made with or without the [auto confirmation](#autoconfirm) flag specified. 
This screenshot demonstrates the inconsistent versioning error being triggered without [auto confirmation](#autoconfirm) enabled.

![Screenshot 2024-10-28 at 21 37 56](https://github.com/user-attachments/assets/52018ef3-4b56-40c2-a7cf-4b57969358db)

<a name="autoconfirm"></a>
## Auto Confirmation

By default you have to confirm each change to a pattern matched version instance. 
If the program prematurely exits before completion of the cormation prompts, no modification will be made to the repository.
If you want to skip this bevahior, simply provide the `--auto-confirm` flags.

```
➜ version-bump [major|minor|patch] [--alpha | --beta | --rc] --auto-confirm
``` 

![Screenshot 2024-10-28 at 21 46 45](https://github.com/user-attachments/assets/db672938-c795-4994-90c1-b822cd8e34ba)

## GPG Signing

If GPG signing is detected from the local or global git configuration, you will be prompted to enter you GPG passphrase in a secure fashion. 
This will allow commits and tags to verified as a result of a successful version increment. In order for GPG passphrase prompts to be enabled you must have [GPG signing configured](https://docs.github.com/en/authentication/managing-commit-signature-verification/signing-commits) correctly.

To disable this behavior you can provide the `--disable-prompts` flag.

```
➜ version-bump [major|minor|patch] [--alpha | --beta | --rc] --disable-prompts
``` 

## Interactive Mode

Another option automate releases it to use interactive mode by specifying the `--interactive` flag.

```
➜ version-bump `--interactive`
``` 

![Screenshot 2024-10-28 at 21 58 02](https://github.com/user-attachments/assets/556b3300-8b24-4787-8598-5459718c2600)

## Creating A New Language

This will allow you to specify a new configuration directive in your `.bump` configuration. 
The codebase has been refactored to a point to make this process as simple as possible. 
In the future more refactoring can provide improvements. 

See [issue #2](https://github.com/nidhhoggr/version-bump/issues/2) for instructions and more specifically, this [commit](https://github.com/nidhhoggr/version-bump/commit/505eb8c3f492bfd9f75be63cd3353354bef1dc46).
