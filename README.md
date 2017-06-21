# bf-ia-broker
A broker for image archives in support of Beachfront. This component generally stands between a UI (e.g., bf-ui) and one or more imagery providers (e.g., Planet Labs).

## Development setup

Go relies heavily on configuration by convention. These instructions will help
you set up the `bf-ia-broker` repository in a way that best adheres to convention
in order to simplify development.

### 1. Install Go

`bf-ia-broker` is built on **Go 1.8.x**. You can download it here:
https://golang.org/dl/. Make sure the `go` tool is on your path once the install
is done.

### 2. Set up Go environment variables

To function right, Go must have some environment variables set. Run the `go env`
command to list all relevant environment variables. The two most important lines
to look for are `GOROOT` and `GOPATH`:

- `GOROOT` must point to the base directory at which Go is installed
- `GOPATH` must point to a directory that is to serve as your development
  environment. This is where this code and dependencies will live.

### 3. Clone the repository

Create the directory the respository must live in, then clone the repository there:

    $ mkdir -p $GOPATH/src/github.com/venicegeo
    $ cd $GOPATH/src/github.com/venicegeo
    $ git clone git@github.com:venicegeo/bf-ia-broker.git

### 4. Install dependencies

This project manages dependencies by populating a `vendor/` directory using the
[Glide](https://glide.sh) tool. Install it as detailed at its website. Then, in
the code repository, run:

    $ glide install -v

This will retrieve all the relevant dependencies at their appropriate versions
and place them in `vendor/`, which enables Go to use those versions in building
rather than the default (which is the newest revision in Github).

> **Adding new dependencies.** When adding new dependencies, simply installing
  them with `go get <package>` will fetch their latest version and place it in
  `$GOPATH/src`. This is undesirable, since it is not repeatable for others.
  Instead, to add a dependency, use `glide get <package>`, which will place it
  in `vendor/` and update `glide.yaml` and `glide.lock` to remember its version.

### 5. Set up Beachfront environment variables

|Variable|Description|Default|
|---------|-----------|------|
|BF_TIDE_PREDICTION_URL|Location of the tide prediction service|https://bf-tideprediction.int.geointservices.io/tides |
|PL_API_URL|Location of Planet Labs API|https://api.planet.com/ |
|PL_API_KEY|Planet Labs API Key|N/A|

## Building, running, and testing

### Build the project

To build `bf-ia-broker`, run `go install` from the project directory. To build
it from elsewhere, run:

    $ go install github.com/venicegeo/bf-ia-broker

This will build and place a statically-linked Go executable at
`$GOPATH/bin/bf-ia-broker`.

### Run the project

To launch `bf-ia-broker` use:

    $ $GOPATH/bin/bf-ia-broker serve

> You can also set your `PATH` to include `$GOPATH/bin` to only have to run
  `bf-ia-broker serve`.

This starts the `bf-ia-broker` listening on all interfaces on port **8080**.

### Run unit tests

To run `bf-ia-broker`, run the `run-tests.sh` script in the repository. This
will run the unit tests for `bf-ia-broker` and all its subpackages and print
coverage summaries.

### Using
In [handlers.go](planet/handlers.go) there are some REST handlers.

|Endpoint|Command|Description|
|-------|--------|------------|
|/planet/discover/{itemType}|GET|Discover (search), as a GeoJSON feature collection|
|/planet/{itemType}/{id}|GET|Metadata for an ID, as a GeoJSON feature|
|/planet/activate/{itemType}/{id}|POST|Activate a resource|

See the Swagger docs or the source for details on using those handlers.

