# Dynatrace Configuration as Code - Core
[![stability-wip](https://img.shields.io/badge/stability-wip-lightgrey.svg)](https://github.com/mkenney/software-guides/blob/master/STABILITY-BADGES.md#work-in-progress)

Dynatrace Configuration as Code Core provides libraries simplifying development of configuration as code tooling for Dynatrace.

It provides Go libraries for things like API clients, which are shared between several Dynatrace configuration as code tools.


## API Clients

The library provides different kinds of API clients to interact with the Dynatrace API.
Following are some important characteristics of the API clients to keep in mind:

* Each client provides a method set, usually supporting CRUD operations and an Upsert - which will create or update a configuration as needed.
However, the specific interface might differ between clients.
* Payloads to and from the APIs aren't interpreted in any particular way.
Thus, it's the user's responsibility to marshal/unmarshal payloads into/from Go structs.
* API clients typically return `(Response, error)` pairs. Note that any API result (including `4xx`,`5xx`...) will be carried back
in the `Response` return value.
It is the responsibility of the user to check for success or failure of the actual operations by inspecting the 
`Response`. The user can expect `error` to be `!= nil` only for (technical) failures that
happen either prior to making the actual HTTP calls or if the HTTP calls couldn't be carried out (e.g. due to network problems, etc.)

| API Client         | Implemented |
|--------------------|-------------|
| grail buckets      | ✅           |
| automation         | ❌           |
| settings 2.0       | ❌           |
| classic config API | ❌           |

### Usage

To instantiate an API client, it's recommended to create an instance via the provided `clients.Factory()` function:

```go
// create the factory
factory := clients.Factory().
	WithEnvironmentURL("https://dt-environment.com").
	WithOAuthCredentials(credentials)

// get the client
bucketClient, err := factory.BucketClient()
if err != nil {
	// handle error
}

// perform operation
resp, err := bucketClient.Get(context.TODO(), "my bucket")
if err != nil {
	// handle error
}

// inspect response
if !resp.IsSuccess() {
	// handle api error
}

// unmarshal payload
bucketDefinition, err := api.DecodeJSON[BucketDefninition](resp.Response)
if err != nil {
	// handle error
}
```

### Logging

The library uses [logr](https://github.com/go-logr/logr), a simple logging interface for Go.
Hence, it can be used with a wide range of known logging libraries for Go.
Per default, the library does not log anything. If you want to turn logging on you need to carry
a logger to each library method via its context argument.

To do this, use [logr.NewContext](https://pkg.go.dev/github.com/go-logr/logr#NewContext).

For example, if you wish to use [Logrus](https://github.com/sirupsen/logrus) for logging:

```go
ctx := logr.NewContext(context.TODO(), logrusr.New(logrus.New()))
resp, err := ctx.Get(ctx,"...")
```

## Forms of Dynatrace Configuration as Code

* [Dynatrace Configuration as Code CLI Monaco](https://github.com/dynatrace/dynatrace-configuration-as-code)
* [Dynatrace Terraform provider](https://github.com/dynatrace-oss/terraform-provider-dynatrace)
