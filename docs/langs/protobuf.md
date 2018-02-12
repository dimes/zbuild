## Protocol Buffers

zbuild assumes protocol buffer packages have the following format:

    <package-name>/
        build.yaml
        proto/
            <folder-name>
                a.proto
                b.proto
                etc...

zbuild will automatically add any `proto` directories to your `PROTO_PATH`, so you can import and proto files in your dependency graphs relative to that package's `proto` directory. The build output is the entire proto directory copied over to the build directory.
