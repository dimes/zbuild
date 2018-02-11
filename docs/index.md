This page contains a detailed description of zbuild. If you're looking for something a little more consumable, check out the [quick start guide](https://github.com/dimes/zbuild#quick-start).

## Packages

Packages are the atomic unit of zbuild. They are identified by a namespace, name, and version. Valid identifiers are of the format `^[a-z0-9\.\-]{1,40}$`.

### Buildfiles

Packages are identified by directories containing a `build.yaml` file. These buildfiles have the following format:

    namespace: <string>
    name:      <string>
    version:   <string>

    type: [go|java|protobuf]

    dependencies:
      compile:
      - namspace: <string>
        name:     <string>
        version:  <string>
      test:
      - namespace: <string>
        name:      <string>
        version:   <string>

### Versioning

Each package has a version. zbuild operates under the assumption that version number changes are only necessary when backwards incompatible changes are made. Therefore, declaring a dependency on a version is assumed to mean "the latest available in the source set."

### Artifacts

When a package is built and published it becomes an artifact. Artifacts are like packages, but are immutable and have a build number attached.

## Source Sets

Source sets are a collection of artifacts. For each (namespace, name, version) tuple in a source set, there will be exactly one artifact.

## Workspaces

Workspaces are locally directories that contain packages. These are typically under active development or are being built. A workspace is identified by the presence of a workspace metadata directory.

When a local build happens, direct child directories of the workspace directories will be checked for packages. The packages found in the workspace override any packages with the same (namespace, name, version) in the dependency graph.

Each workspace has a source set where it pulls artifacts from and publishes artifacts to.

## CLI

The command-line interface contains useful functionality for zbuild.

### init-workspace

    zbuild init-workspace

This command initializes a workspace in the working directory. It provides an interactive prompt for filling in the required settings

### publish

    zbuild publish

This command should be executed inside a package. It builds and uploads an artifact to the workspace's source set.

### pathfinder

The pathfinder is a separate CLI that handles common build path related operations. For instance, you can list the path for a workspace package by executing this command somewhere in the package's file tree:

    pathfinder -resolver [test|compile]
