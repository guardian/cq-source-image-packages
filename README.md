# CloudQuery Image Packages Source Plugin

[![test](https://github.com/guardian/cq-source-image-packages/actions/workflows/test.yaml/badge.svg)](https://github.com/guardian/cq-source-image-packages/actions/workflows/test.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/guardian/cq-source-image-packages)](https://goreportcard.com/report/github.com/guardian/cq-source-image-packages)

An image packages source plugin for CloudQuery that loads data from S3 and Dynamo DB to any database, data warehouse or data lake supported by [CloudQuery](https://www.cloudquery.io/), such as PostgreSQL, BigQuery, Athena, and many more.

This plugin is currently used to collect data about image packages from the Amigo bake service.

It was created using the CloudQuery skeleton generator: `cq-scaffold source guardian image-packages`.


## Links

 - [CloudQuery Quickstart Guide](https://www.cloudquery.io/docs/quickstart)
 - [Source Spec Reference](https://docs.cloudquery.io/docs/reference/source-spec)
 - [Creating a New CloudQuery Source Integration in Go](https://docs.cloudquery.io/docs/developers/creating-new-integration/go-source)
 - [Supported Tables](docs/tables/README.md)


## Configuration

The following source configuration file will sync to a PostgreSQL database. See [the CloudQuery Quickstart](https://www.cloudquery.io/docs/quickstart) for more information on how to configure the source and destination.

```yaml
kind: source
spec:
  name: "image-packages"
  path: "guardian/image-packages"
  version: "${VERSION}"
  destinations:
    - "postgresql"
  spec:
    # Name of S3 bucket holding Amigo bake package data
    bucket: ...
    # Name of Dynamo table holding bake data
    bakes_table: ...
    # Name of Dynamo table holding recipe data
    recipes_table: ...
    # Name of Dynamo table holding base image data
    base_images_table: ...
```

## Development

### Code structure

The entry point is [main.go](main.go).

Package [client](client) contains the code to interact with the AWS services.

Package [store](store) contains the code to interact with the AWS data sources.

Package [plugin](resources/plugin) contains the code to interact with the CloudQuery source plugin.

Package [services](resources/services) contains the code to generate target tables.


### Run tests

```bash
make test
```

### Run linter

To do this, you will need to have [golangci-lint](https://golangci-lint.run/usage/install/) installed.

```bash
make lint
```

### Generate docs

```bash
make gen-docs
```

### Debugging

To debug the plugin so that code can be stepped through with breakpoints, you will need:
1. [Delve](https://github.com/go-delve/delve) debugger installed.
2. If using Intellij, you will need a run/debug configuration set up.
To set this up, add an Intellij run/debug configuration:
    1. `Run` > `Edit Configurations...` > `+` > `Go Remote`
    2. Fill in `Host: localhost` and `Port: 7777`
3. A local spec file that specifies a local gRPC process as the plugin installed in a directory called `local` in this project root:
```yaml
kind: source
spec:
  name: image-packages
  registry: grpc
  path: localhost:7777
  tables: ['amigo_bake_packages']
  destinations:
    - sqlite
  spec:
    base_images_table: <name of base images table>
    recipes_table: <name of recipes table>
    bakes_table: <name of bakes table>
    bucket: <name of packages bucket>
---
kind: destination
spec:
  name: sqlite
  path: cloudquery/sqlite
  registry: cloudquery
  version: v2.9.18
  spec:
    connection_string: local/db.sql
```

Then:
1. In a terminal window, start up the plugin in debug mode:
```bash
make serve-debug
```
2. Click on `Debug` in the Intellij run/debug configuration, or the green triangle in the `Debug plugin` in the `Run and debug` pane of Visual Studio.
3. Insert breakpoints into the code where required.
4. In another terminal window, run the CloudQuery sync command:
```bash
make run
```


### Release a new version

Before releasing a new version, test the integration locally by building a local SQLite database following these steps:
1. Build the binary:
```bash
make build
```
2. Set up a local spec file with the required configuration:
```yaml
kind: source
spec:
  name: image-packages
  registry: local
  path: ./cq-source-image-packages
  tables: ['amigo_bake_packages']
  destinations:
    - sqlite
  spec:
    base_images_table: <name of base images table>
    recipes_table: <name of recipes table>
    bakes_table: <name of bakes table>
    bucket: <name of packages bucket>
---
kind: destination
spec:
  name: sqlite
  path: cloudquery/sqlite
  registry: cloudquery
  version: v2.9.18
  spec:
    connection_string: local/db.sql
```
3. Run the binary with the spec file:
```bash
make run
```
4. Check the [log](cloudquery.log) and the output in the destination SQLite database.

Then, to release a new version:
1. Run `git tag v1.0.0` to create a new tag for the release (replace `v1.0.0` with the new version number)
2. Run `git push origin v1.0.0` to push the tag to GitHub  

Once the tag is pushed, a new GitHub Actions workflow will be triggered to build the release binaries.
