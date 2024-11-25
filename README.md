# CloudQuery Image Packages Source Plugin

[![test](https://github.com/guardian/cq-source-image-packages/actions/workflows/test.yaml/badge.svg)](https://github.com/guardian/cq-source-image-packages/actions/workflows/test.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/guardian/cq-source-image-packages)](https://goreportcard.com/report/github.com/guardian/cq-source-image-packages)

An image packages source plugin for CloudQuery that loads data from S3 and Dynamo DB to any database, data warehouse or data lake supported by [CloudQuery](https://www.cloudquery.io/), such as PostgreSQL, BigQuery, Athena, and many more.

## Links

 - [CloudQuery Quickstart Guide](https://www.cloudquery.io/docs/quickstart)
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

### Release a new version

1. Run `git tag v1.0.0` to create a new tag for the release (replace `v1.0.0` with the new version number)
2. Run `git push origin v1.0.0` to push the tag to GitHub  

Once the tag is pushed, a new GitHub Actions workflow will be triggered to build the release binaries.
