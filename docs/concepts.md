## Concepts

zbuild makes use of a few simple concepts to create a powerful build and dependency system.

### Package

The package is the atomic unit of zbuild. It represents something that can be "depended on", such as a shared library. The source of a package will typically be stored as a single git repository.

### Source Set

A source set is a stable set of packages. They are similar to lock files in other languages, but are shared across all packages inside the source set.

The easiest way to understand source sets may be a motivating example. Imagine you are on a team in a large company and have taken a dependency on another team's library. Your team will share a single source set that will contain a stable version of the library. This allows your team to work without having to worry about potentially breaking changes to the other team's library. The team that develops the library will use a separate source set. In the other team's source set, the library package will always track the absolute latest version because their day-to-day work involves using the latest changes.

### Workspace

Often times, development will involve multiple packages simultaneously. For instance, you might be making changes to a library and want to use those changes in one of your services.

A workspace is a local directory that contains the packages you are working on. When you attempt to build or run your code locally, your workspace will be used to override any dependencies in your dependency graph with their local versions. If you don't have a package checked out locally, the package inside the current source set will be used.
