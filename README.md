# version-bump

[![CI](https://github.com/joe-at-startupmedia/version-bump/actions/workflows/ci.yml/badge.svg)](https://github.com/joe-at-startupmedia/version-bump/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/joe-at-startupmedia/version-bump/graph/badge.svg?token=HHRUHCX1EJ)](https://codecov.io/gh/joe-at-startupmedia/version-bump)
[![Go Report Card](https://goreportcard.com/badge/github.com/joe-at-startupmedia/version-bump/v2)](https://goreportcard.com/report/github.com/joe-at-startupmedia/version-bump/v2)
[![Release](https://img.shields.io/github/v/release/joe-at-startupmedia/version-bump)](https://github.com/joe-at-startupmedia/version-bump/releases/latest)
[![License](https://img.shields.io/github/license/joe-at-startupmedia/version-bump)](LICENSE.md)


## Configuration

**version-bump** has two modes of operation: automatic / manual.
In automatic mode, **version-bump** will try to identify versions of all supported languages in a root of the project (wherever executed).
In a manual mode, **version-bump** will read a configuration file to determine which modifications to make. It is expected be executed in the root of the project where the configuration file is located.

Some languages, have a constant value in a specific file that contains a version, which are fairly easy to increment.
But some languages are leaving that decision to a developer, thus **version-bump** assumes a constant position/value for them as well.

| Settings      | Expected Patterns                             | Filename                              |
|:-------------:|:---------------------------------------------:|:-------------------------------------:|
| Docker        | `org.opencontainers.image.version` label      | `Dockerfile`                          |
| Go            | String constant named `Version`/`version`     | `*.go`                                |
| JavaScript    | JSON `version` field                          | `package.json`, `package-lock.json`   |

### Manual

1. Create a configuration `.bump` file in a root of the project.
2. Add project languages and their configuration in a form of:

    ```
    [ language_name ]
    enabled = bool
    directories = [ string, string, ... ]
    exclude_files = [ string, string, ... ]
    files = [ string, string, ... ]
    regex = [string, string, ...]
    ```

    - `[ language_name ]` - one of `[ 'docker', 'go', 'javascript', 'generic' ]`
    - `enabled` - default `false`
    - `directories` - path default `['.']`
    - `exclude_files` - path default `[]`
    - `files` - an array of glob values to overide the settings default `declared in the langs module`
    - `regex` - an array of regex patterns to overide the settings default `declared in the langs module`
      
3. Run **version-bump** in a root of the project: `version-bump [major|minor|patch] [flags]`

*Simple Configuration Example:*

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

## CLI Usage

```
This application helps incrementing a semantic version of a project.
It can bump the version in multiple different files at once,
for example in package.json and a Dockerfile.

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
versions can be optionally specified as an argument along with prerelease flags and metadata. Must be one of the following:

* `major`
* `minor`
* `patch`
  
![Screenshot 2024-10-28 at 21 06 20](https://github.com/user-attachments/assets/eb9fcace-246d-495d-b744-fb1152ddfa76)

When incrementing a prerelease without updating the version, simply omit the version type argument.

## Prerelease Automation

<a name="Prerelease_types"></a>
### Types
Prereleases can be specified as a flag along with metadata. The currently supported prerelease types are the following:

* `alpha`
* `beta`
* `rc`

### Format
conforming to the Semver specification, Prereleases must be in the following format:

`Prerelease-type`.`Prerelease-version`+`Prerelease-metadata`

Where the following criterion must be met:

* `Prerelease-type`: Must be a string value matching one of the [Prerelease types](#Prerelease_types)
* `Prerelease-version`: Must be an integer
* `Prerelease-metadata`: A string without special characters seperated beginning with a `+`


<<<<<<< Updated upstream
<a name="prerelease_alpha"></a>  
### Alpha Prerelease
=======
<a name="Prerelease_alpha"></a>  
### Alpha Prerelease
>>>>>>> Stashed changes

must be released from an existing alpha release whose patch is the same by omitting the [version type](#version_types) argument:

![Screenshot 2024-10-28 at 20 58 55](https://github.com/user-attachments/assets/f1679ed9-7464-430c-9396-704128e94435)

Or from a new version by specifying the [version type](#version_types):

![Screenshot 2024-10-28 at 21 04 21](https://github.com/user-attachments/assets/8ef99cdf-8fa1-43b4-8d27-8be987b6f52b)

<<<<<<< Updated upstream
<a name="prerelease_beta"></a>  
### Beta Prerelease
=======
<a name="Prerelease_beta"></a>  
### Beta Prerelease
>>>>>>> Stashed changes

must be released from an existing alpha or beta release whose patch is the same by omitting the [version type](#version_types) argument:

![Screenshot 2024-10-28 at 21 11 32](https://github.com/user-attachments/assets/cdff9fa2-539a-4c1c-907a-b248d5840ad4)

Attempting to release an [alpha release](#Prerelease_alpha) from a [beta release](#Prerelease_beta) without specifying the [version type](#version_types)  will throw an error:

![Screenshot 2024-10-28 at 21 15 44](https://github.com/user-attachments/assets/44006c13-82e4-4bc5-9403-d865d9868654)

<a name="Prerelease_rc"></a>  
### Release Candidate

Similar to [alpha releases](#Prerelease_alpha) and [beta releases](#Prerelease_beta), the flags must be specified appropriately
![Screenshot 2024-10-28 at 21 22 35](https://github.com/user-attachments/assets/60b6cf57-21e0-4f21-a05b-724714766c5f)

Attempting to release an [alpha release](#Prerelease_alpha) or a [beta release](#Prerelease_beta) and omitting the [version type](#version_types) argument will produce errors:

![Screenshot 2024-10-28 at 21 23 48](https://github.com/user-attachments/assets/58dfe870-9d5c-4a04-87d0-8614b7fb62e3)

### Incrementing aPrereleaseVersion

Simply specify the same Prerelease type of the existing Prerelease while omitting the [version type](#version_types) argument. 
It will automatically increment the Prerelease version.

![Screenshot 2024-10-28 at 21 24 59](https://github.com/user-attachments/assets/5870f006-b41d-4de2-b7bd-df87ab5c545c)

### Promoting a Prerelease

After our Prerelease has been tested and you're ready for rollout you can simply `patch` as the [version type](#version_types) argument. 
It will remove all of the Prerelease versioning and metadata from the version.

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

![Screenshot 2024-10-28 at 21 46 45](https://github.com/user-attachments/assets/db672938-c795-4994-90c1-b822cd8e34ba)

## GPG Signing

If GPG signing is detected from the local or global git configuration, you will be prompted to enter you GPG passphrase in a secure fashion. 
This will allow commits and tags to verified as a result of a successful version increment. In order for GPG passphrase prompts to be enabled you must have [GPG signing configured](https://docs.github.com/en/authentication/managing-commit-signature-verification/signing-commits) correctly.

To disable this behavior you can provide the `--disable-prompts` flag.

## Interactive Mode

Another option automate releases it to use interactive mode by specifying the `--interactive` flag.

![Screenshot 2024-10-28 at 21 58 02](https://github.com/user-attachments/assets/556b3300-8b24-4787-8598-5459718c2600)

## Creating A New Language

This will allow you to specify a new configuration directive in your `.bump` configuration. 
The codebase has been refactored to a point to make this process as simple as possible. 
In the future more refactoring can provide improvements. 

See [issue #2](https://github.com/joe-at-startupmedia/version-bump/issues/2) for instructions and more specifically, this [commit](https://github.com/joe-at-startupmedia/version-bump/commit/505eb8c3f492bfd9f75be63cd3353354bef1dc46).
