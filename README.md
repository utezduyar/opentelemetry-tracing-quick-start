# OpenTelemetry Tracing Quick Start Tutorial

Every commit that is a milestone is tagged. 
You need to start from the first tag and follow the text here. 
You can compare the milestones like `git diff step1..step2'
The tutorial is implemented in golang but other languages have very similar flavor. 


## `step1` Initial commit

This is the first commit. It is an application that builds and runs. It also has a library part. For the sake of simplicity the library code is in the same file as the main code. It is important to have a library part because there are some restrictions regarding what a library can or cannot do.

Go ahead and run it with `go run main.go`

##  `step2` Instrument using the OpenTelemetry API

We are instrumenting the library to create a span. Libraries do this by linking to the `OpenTelemetry API`. 

```
_, span := otel.Tracer(name).Start(context.Background(), "InsertUser")
defer span.End()
```
The `Start` API creates a span called `InsertUser`. The first parameter to it is `Context`. This parameters defines if the created span will be child of a parent span. `defer` is a keyword in `golang` which instructs the CPU to execute the statement after it. This way, we are ensuring that the created span is terminated just before we leave the function. 

However as long as the application is only linked with `OpenTelemetry API` the span created is almost like a no operation. 

![OpenTelemetry API](https://github.com/utezduyar/opentelemetry-tracing-quick-start/raw/step2/OpenTelemetryAPI.jpeg)

Span creation should not fail and should not impact the performance. 

##  `step3` ........
