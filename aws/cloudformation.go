package aws

import (
	"fmt"
	"os"

	"github.com/JeroenSoeters/wheel/templates"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

type AwsClient struct {
}

func ReadTemplate() (string, error) {
	data, err := templates.Asset("templates/single-master.cloudformation.json")
	if err != nil {
		fmt.Errorf("Error loading CloudFormation single-master template: %v", err)
		os.Exit(1)
	}

	return string(data), nil
}

func (c AwsClient) ProvisionBuildEnvironment() error {
	return CreateStack("us-west-2", "dcos-build", map[string]string{"KeyName": "dcos-bootstrap"})
}

func CreateStack(region string, name string, parameters map[string]string) error {
	template, err := ReadTemplate()
	if err != nil {
		fmt.Printf("Error loading template: %v", err)
	}

	s := session.New(&aws.Config{Region: aws.String(region)})
	cf := cloudformation.New(s)

	_, _ = cf.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(name),
	})

	capabilities := []*string{}
	capabilities = append(capabilities, aws.String(cloudformation.CapabilityCapabilityIam))

	params := []*cloudformation.Parameter{}
	for key, value := range parameters {
		params = append(params, &cloudformation.Parameter{
			ParameterKey:   aws.String(key),
			ParameterValue: aws.String(value),
		})
	}

	_, err = cf.CreateStack(&cloudformation.CreateStackInput{
		StackName:    aws.String(name),
		TemplateBody: &template,
		Capabilities: capabilities,
		Parameters:   params,
	})

	return err
}
