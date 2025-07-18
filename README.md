# Dynatrace Configuration as Code - Core
[![stability-wip](https://img.shields.io/badge/stability-wip-lightgrey.svg)](https://github.com/mkenney/software-guides/blob/master/STABILITY-BADGES.md#work-in-progress)

**Dynatrace Configuration as Code Core** provides libraries simplifying development of configuration as code tooling for Dynatrace.

ℹ️ **DISCLAIMER:** This library is **not** meant to be a general purpose client library for Dynatrace but rather contains functionality that is
tailored to simplify the development of Configuration as Code tools like Monaco, the Terraform Provider, and similar tools.

## API Clients

* **Client operations**: Each client provides a set of methods, typically supporting CRUD operations. However, the exact interface may vary between clients.
* **Payload handling**: The library interprets API payloads only in specific clients, such as buckets or documents. It is the user's responsibility to marshal and unmarshal payloads into or from Go structs as needed.


  | API Client               | Implemented |
  |--------------------------|-------------|
  | Classic config APIs      | ❌           |
  | Settings 2.0             | ❌           |
  | Settings 2.0 permissions | ✅           |
  | Automation               | ✅           |
  | Grail buckets            | ✅           |
  | Documents                | ✅           |
  | OpenPipeline             | ✅           |
  | Segments                 | ✅           |
  | SLO's                    | ✅           |
  | Account management       | ✅           |

### Usage
To instantiate a client, it's recommended to create an instance via the provided `clients.Factory()` function.

#### Platform clients
Platform clients are designed to interact with Dynatrace platform APIs.

Ensure that you are using the correct environment URL, which must include `.apps.dynatrace.com`.
Authentication can be handled using either OAuth or a platform token`.
```go
// create the factory
ctx := context.TODO()
factory := clients.Factory().
    WithEnvironmentURL("https://<dt-environment>.apps.dynatrace.com").
	WithOAuthCredentials(credentials)
    // or if you want to use a platform token
    WithPlatformToken("<YOUR_PLATFORM_TOKEN>")

// request any client from the factory, e.g. bucket client
bucketClient, err := factory.BucketClient(ctx)
if err != nil {
	// handle error
}

// perform operation
resp, err := bucketClient.Get(ctx, "my bucket")

if err != nil {
    // handle error. See Error handling section.
}

// unmarshal payload
bucketDefinition, err := api.DecodeJSON[BucketDefinition](resp.Response)
if err != nil {
	// handle error
}
```

#### Classic rest client
Unlike [Platform clients](#platform-clients), classic clients do not include dedicated resource clients.
Instead, only a general-purpose REST client is available for interacting with the API.

```go
// create the factory
ctx := context.TODO()
factory := clients.Factory().
    WithClassicURL("https://<dt-environment>.live.dynatrace.com").
    WithAccessToken("<YOUR_ACCESS_TOKEN>")

// create a classic client from the factory
client, err := cFactory.CreateClassicClient()
if err != nil {
	// handle error
}

// perform operation
httpResp, err := client.GET(ctx, "/your-endpoint", rest.RequestOptions{})

if err != nil {
    // handle client error
}

resp, err := api.NewResponseFromHTTPResponse(resp)

if err != nil {
    // handle error. See Error handling section.
}

// unmarshal payload
data, err := api.DecodeJSON[YourExpectedStruct](resp.Response)
if err != nil {
    // handle error
}
```

#### Error handling
The library provides custom error structs tailored to specific error scenarios.

Below is an example of the potential custom errors that may occur during usage.
````go
resp, err := segmentClient.Get(ctx, "my-segment-id")
// handle error
if err != nil {
    // response status code validation failed
    var apiErr api.ApiError
    if errors.As(err, &apiErr) {
        // e.g., handle differently if apiErr.StatusCode is 404
    }
    
    // request failed (no response received)
    var clientErr api.ClientError
    if errors.As(err, &clientErr) {
        
    }

    // validation failed (e.g., provided segment ID is empty or missing properties in response data)
    var validationErr api.ValidationError
    if errors.As(err, &validationErr) {

    }
    
    // pre- or post-processing of data failed
    var runtimeErr api.RuntimeError
    if errors.As(err, &runtimeErr) {

    }
}
````

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
