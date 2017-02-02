package aws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

type StackEventWatcher struct {
	Service   *cloudformation.CloudFormation
	StackName string

	events map[string]struct{}
}

func NewStackEventWatcher(service *cloudformation.CloudFormation, stackName string) (*StackEventWatcher, error) {
	sw := StackEventWatcher{
		Service:   service,
		StackName: stackName,
		events:    map[string]struct{}{},
	}

	err := sw.Service.DescribeStackEventsPages(&cloudformation.DescribeStackEventsInput{
		StackName: aws.String(sw.StackName),
	}, func(p *cloudformation.DescribeStackEventsOutput, _ bool) bool {
		for _, stackEvent := range p.StackEvents {
			sw.events[*stackEvent.EventId] = struct{}{}
		}
		return true
	})
	if err != nil {
		return nil, err
	}
	return &sw, nil
}

func (sw *StackEventWatcher) Watch() error {
	fmt.Print("Watching...")
	if sw.events == nil {
		fmt.Print("Events is nil")
		sw.events = map[string]struct{}{}
	}
	fmt.Print("Done...")
	lastStackStatus := ""
	for {
		// print the events for the stack
		sw.Service.DescribeStackEventsPages(&cloudformation.DescribeStackEventsInput{
			StackName: aws.String(sw.StackName),
		}, func(p *cloudformation.DescribeStackEventsOutput, _ bool) bool {
			for _, stackEvent := range p.StackEvents {
				fmt.Printf("event: %s", *stackEvent.EventId)
				if _, ok := sw.events[*stackEvent.EventId]; ok {
					continue
				}

				/*
					wrapStrPtr := func(s *string) string {
						if s == nil {
							return ""
						}
						return *s
					}

					l := log.WithField("Status", *stackEvent.ResourceStatus)
					if stackEvent.ResourceType != nil {
						l = l.WithField("Type", *stackEvent.ResourceType)
					}
					if stackEvent.ResourceType != nil {
						l = l.WithField("Type", *stackEvent.ResourceType)
					}
					if stackEvent.PhysicalResourceId != nil {
						l = l.WithField("PhysicalID", *stackEvent.PhysicalResourceId)
					}
					if stackEvent.LogicalResourceId != nil {
						l = l.WithField("LogicalID", *stackEvent.LogicalResourceId)
					}
					if strings.Contains(*stackEvent.ResourceStatus, "FAIL") {
						l.Error(wrapStrPtr(stackEvent.ResourceStatusReason))
					} else {
						l.Info(wrapStrPtr(stackEvent.ResourceStatusReason))
					}
				*/
				sw.events[*stackEvent.EventId] = struct{}{}
			}
			return true
		})

		// monitor the status of the stack
		describeStacksResponse, err := sw.Service.DescribeStacks(&cloudformation.DescribeStacksInput{
			StackName: aws.String(sw.StackName),
		})
		if err != nil {
			// the stack might not exist yet
			fmt.Errorf("DescribeStacks: %s", err)
			time.Sleep(time.Second)
			continue
		}

		stackStatus := *describeStacksResponse.Stacks[0].StackStatus
		if stackStatus != lastStackStatus {
			fmt.Printf("Stack: %s\n", stackStatus)
			lastStackStatus = stackStatus
		}
		switch stackStatus {
		case cloudformation.StackStatusCreateComplete:
			return nil
		case cloudformation.StackStatusCreateFailed:
			return fmt.Errorf("%s", stackStatus)
		case cloudformation.StackStatusRollbackComplete:
			return fmt.Errorf("%s", stackStatus)
		case cloudformation.StackStatusUpdateRollbackComplete:
			return fmt.Errorf("%s", stackStatus)
		case cloudformation.StackStatusRollbackFailed:
			return fmt.Errorf("%s", stackStatus)
		case cloudformation.StackStatusUpdateComplete:
			return nil
		case cloudformation.StackStatusUpdateRollbackFailed:
			return fmt.Errorf("%s", stackStatus)
		default:
			time.Sleep(time.Second * 5)
			continue
		}
	}
}
