# Dynatrace Configuration as Code - Core
[![stability-wip](https://img.shields.io/badge/stability-wip-lightgrey.svg)](https://github.com/mkenney/software-guides/blob/master/STABILITY-BADGES.md#work-in-progress)

**Dynatrace Configuration as Code Core** provides libraries simplifying development of configuration as code tooling for Dynatrace.

ℹ️ **DISCLAIMER:** This library is **not** meant to be a general purpose client library for Dynatrace but rather contains functionality that is
tailored to simplify the development of Configuration as Code tools like Monaco, the Terraform Provider, and similar tools.

## API Clients

The library provides different kinds of clients to interact with Dynatrace in two different packages:

* **api/clients**: "Basic" clients that offers a one-to-one mapping to the Dynatrace API and do not contain any other special logic.
* **clients**: "Smarter" clients that build upon the basic clients to offer additional logic and operations tailored to simplify configuration as code use cases.
  * Each client provides a method set, usually supporting CRUD operations and an Upsert - which will create or update a configuration as needed.
  However, the specific interface might differ between clients.
  * Payloads to and from the APIs aren't interpreted in any particular way.
  Thus, it's the user's responsibility to marshal/unmarshal payloads into/from Go structs.


  | API Client          | Implemented |
  |---------------------|-------------|
  | Classic config APIs | ❌           |
  | Settings 2.0        | ❌           |
  | Automation          | ✅           |
  | Grail buckets       | ✅           |
  | Documents           | ✅           |
  | OpenPipeline        | ✅           |
  | Segments            | ✅           |
  | SLO's               | ✅           |

### Usage

To instantiate an API client, it's recommended to create an instance via the provided `clients.Factory()` function:

```go
// create the factory
factory := clients.Factory().
	WithEnvironmentURL("https://dt-environment.com").
	WithOAuthCredentials(credentials)

// request any client from the factory, e.g. bucket api client
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
bucketDefinition, err := api.DecodeJSON[BucketDefinition](resp.Response)
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

### Tracking and logging HTTP requests/responses
If you want to keep track or just log all HTTP requests/responses happening as part of the execution of the clients, you can implement an `HTTPListener` and attach it to the client.
All you need to do is implement a custom callback function and pass the `HTTPListener` when constructing a client.
The underlying `rest.Client` will then call your callback function with information about each HTTP request or response.

For example, in order to just print out all HTTP requests that are happening under the hood you can do the following:

```go
requestPrinter := &rest.HTTPListener{Callback: func(r rest.RequestResponse) {
	if req, ok := r.IsRequest(); ok {
		fmt.Printf("There was an HTTP %s request to %s\n", req.Method, req.URL.String())
	}
}}

factory := clients.Factory().WithEnvironmentURL("https://dt-environment.com").
	WithOAuthCredentials(credentials).
	WithHTTPListener(requestPrinter)

// request a client from the factory and use it
```

## Forms of Dynatrace Configuration as Code

* [Dynatrace Configuration as Code CLI Monaco](https://github.com/dynatrace/dynatrace-configuration-as-code)
* [Dynatrace Terraform provider](https://github.com/dynatrace-oss/terraform-provider-dynatrace)
