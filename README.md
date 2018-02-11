# zbuild

[![Build Status](https://travis-ci.org/dimes/zbuild.svg?branch=master)](https://travis-ci.org/dimes/zbuild)

Handling dependencies in large, cross-team code bases is notoriously difficult. Each language has at least one dependency management solution (sometimes many!), and these different solutions rarely work well together. Additionally, most dependency managers rely on public repositories, and setting up private repositories can be challenging or impossible.
 
**zbuild** aims to solve these problems in the following ways:

* Easy setup
* Cross-language dependency support
* Integration with other popular dependency managers

## Quick Start

Before starting, please take a few minutes to familiarize yourself with the [core concepts](https://dimes.github.io/zbuild/concepts) of zbuild.

### Installation

Building from source is the only supported installation mechanism and requires [Go 1.9+](https://golang.org/dl/)

    > mkdir -p zbuild-workspace/src/github.com/dimes
    > cd zbuild-workspace/src/github.com/dimes
    > git clone git@github.com:dimes/zbuild.git
    > cd zbuild
    > GOPATH=$(cd ../../../.. && pwd); go install ./...

After this, the binary will be located at `zbuild-workspace/bin/zbuild`

### Creating a repository

Built artifacts are stored in a package repository. These package repositories are stored on a remote service so they can be shared.

After installing the CLI, this command will get you started:

    zbuild init-workspace

Specific cloud providers may need additional setup. See the provider-specific documentation for more information

* [AWS](https://dimes.github.io/zbuild/providers/aws)
* [Google Cloud](https://dimes.github.io/zbuild/providers/gcloud)

### Creating a package

The `build.yaml` file is the heart of a package.

    # build.yaml
    namespace: my_company_name
    name:      my_package_name
    version:   1.0

    type: go

    dependencies:
      compile:
      - namespace: a_namespace
        name:      a_name
        version:   2.3
      test:
      - namespace: other_namespace
        name:      other_name
        version:   1.1

To understand the impact of the `type` parameter, see the language specific guides:

* [Go](https://dimes.github.io/zbuild/langs/go)
* [Java](https://dimes.github.io/zbuild/langs/java)

### Sharing your package

Publishing a package updates your source set with the newest version of that package. The publish command is 

    zbuild publish

This command should be executed in the directory containing the package's `build.yaml` file or a subdirectory.

## Further reading

See [the docs](https://dimes.github.io/zbuild/) for more detailed information on the inner workings of zbuild.
