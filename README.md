# OpenTelemetry Tracing Quick Start Tutorial

Every commit that is a milestone is tagged like `step1`, `step2`, `step3`... You need to start from the first tag `git checkout step1` and make your way to the next one. You can compare the steps like `git diff step1..step2`

The tutorial is implemented in golang but other languages have very similar API flavor. 


## `step1` Initial commit

This is the first commit. It is an application that builds and runs. You will not see anything interesting when you run it. Go ahead and start 
```
go run main.go
```
##  `step2` Instrument using the OpenTelemetry API

Another boring step without any output but lots of comments in the code to be read.

##  `step3` Trace using the SDK

Finally we are seeing something interesting!\
We should be seeing 2 spans created and they are linked. They both share the same trace ID and one span is the parent of the other one.

##  `step4` Annotate spans

We are adding extra meta data to spans. It is very useful to add more data to spans to give you more insights of what is happening.\
Run the program and look at the annotations you have added.

##  `step5` Send to an OpenTelemetry Collector

Usually it is a good pattern to send the telemetry data to an `OpenTelemetry Collector` so that the collector can process it before it sends the data to the next destination. In this step, we are going to send data to our local `OpenTelemetry Collector` in the `OTLP` format.

First we need to download the software. As of writing this tutorial I have downloaded `otelcol-contrib_0.74.0_darwin_arm64.tar.gz` from the following link `version 0.74.0` https://github.com/open-telemetry/opentelemetry-collector-releases/releases/tag/v0.74.0

My advice is to download `otelcol-contrib` variant instead of just `otelcol`. `otelcol-contrib` has much more advanced plugins but they are not official. It is good for playing around. 

`trace.debug.yaml` file which is in this repository is a simple `OpenTelemetry Collector` configuration that just dumps the traces on the console, nothing more. 

```
./otelcol-contrib --config=./trace.debug.yaml
```
##  `step6` Sampling

Sampling is the process of deciding if you need the telemetry data or not and it is a tough process. If you sample everything, you will end up paying a lot for the data. If you sample too little, you might miss crucial information. Head based and tail based sampling are two approaches. We are going to look at head based sampling, decision of sampling before we create the trace or span.

##  `step7` Context Propagation

`Trace` can only be created if the `Spans` are connected to each other. The connection is `Context` and effort of moving `Context` in between process boundaries is called `Context Propagation`. This step creates 2 spans that have parent/child relationship. This relationship is set by doing `Context Propagation`.

##  `step8` More annotation with Resources

You can add more annotations to all the `Spans` related to the environment your application is running in. You can even implement custom resource detector. Look at all the extra annotations that are added to the `Spans`.
```
OTEL_RESOURCE_ATTRIBUTES="baz=qux" go run main.go
```

##  `step9` Linking spans

`Spans` that are linked through context propagation have parent-child relationship. There is another way of casually linking spans by just having a reference to each other. `Links` can point to `Span` inside a single `Trace` or across different `Traces`. You can read about the use cases [here](https://opentelemetry.io/docs/reference/specification/overview/#links-between-spans).\
When you run the program, look for `SpanContext` inside `Links` attribute to see the linkage. Remember that we just made up a `TraceID` and `SpanID`.\
Note: Not all observability tools support `Links` yet.
