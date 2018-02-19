## Protocol Buffers

Packages containing protocol buffer definitions should have `type: proto` set in their build.yaml file. zbuild assumes protocol buffer packages have the following format:

    <package-name>/
        build.yaml
        proto/
            <folder-name>
                a.proto
                b.proto
                etc...

zbuild will automatically add any `proto` directories to your `PROTO_PATH`, so you can import and proto files in your dependency graphs relative to that package's `proto` directory. The build output is the entire proto directory copied over to the build directory.

### Code Generation

Code generation is done by a separate package type: `type: protogen`. These options are valid for packages of `type: protogen`:

    # build.yaml
    namespace: <namespace>
    name: <name>-go
    version: <version>
    type: protogen
    protogen:
      lang: go # java, go, etc.
      source:
        namespace: <namespace>
        name: <name>
        version: <version>

* **lang** specifies the language of the generated output
* **source** specifies a package of `type: proto`. Each `.proto` file in the source package will be passed into `protoc`.

#### Transitive Dependencies

Imagine we have a package `A` that contains proto files. This package doesn't have any dependencies and we've declared a package, say `A-go` that generates go files. These packages might look like this:

    # package: A             # package A-go
    name: A                  name: A-go
    type: proto              type: protogen
                             protogen:
                               lang: go
                               source:
                                 name: A

This setup is all well and good, but becomes a little more complicated when transitive dependencies are introduced. For instance, say we would like to add a package `B` that depends on `A`. This package would look like this:

    # package B
    name: B
    type: proto

    dependencies:
      compile:
      - name: A

If we naïvely try to declare a package `B-go`, we're going to run into trouble:

    # package: B-go
    name: B-go
    type: protogen
    protogen:
      lang: go
      source:
        name: A

Notice that `B-go` has an implicit dependency on `A-go`! Without explicitly stating this dependency, the compiler wont be able to find the generated definitions from package A. The final build file should look like:

    # package: B-go
    name: B-go
    type: protogen
    protogen:
      lang: go
      source:
        name: A
    dependencies:
      compile:
      - name: A-go

##### Important options

In order for transitive dependencies to work properly, it is important to set a few options in the proto files. For instance, when generating Go code, it is important to specify `option go_package = "my.package.name"`. This is because **only** proto files directly inside `source` will be turned into Go code as part of the build. If the source package has any transitive dependencies, the proto compiler will need to know what the correct import path is. Similarly `option java_package`, `option csharp_namespace`, etc. should be set in all proto files.
