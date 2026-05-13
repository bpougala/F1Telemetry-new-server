package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"
)

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewF1TelemetryStack(app, "F1TelemetryStack", &awscdk.StackProps{
		Env: &awscdk.Environment{
			Region: jsii.String("eu-west-1"),
		},
	})

	app.Synth(nil)
}
