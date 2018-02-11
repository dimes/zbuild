## Go

zbuild assumes Go packages have this structure:

    <package-name>/
        build.yaml
        src/
            <package-name>
                a.go
                b.go
                etc...

The reason for this directory structure is that each zbuild package is its own mini go workspace. At compile time, the transitive closure of dependencies is calculated, and the `GOPATH` environment variable is set to a concatenated list of all the package locations.

### Vendoring

zbuild does its own dependency resolution and conflict management based on the contents of build.yaml files. Therefore, it likely does not play well with conflicts in vendor dependencies (although this has not been tested!)

To solve this, and other problems, there is a utility you can use to import a package into the zbuild system. This utility uses Go's tooling to pull down the transitive dependencies of the target package and publishes them all to the workspace's repository. The command is

    zbuild import -type go -package <package>

The `package` parameter should be the full package name, i.e. the name you provide when doing `go get`
