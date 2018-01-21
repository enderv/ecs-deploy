package main

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
)

// ECSClient ECS Client for updates
type ECSClient struct {
	svc          *ecs.ECS
	logger       *log.Logger
	pollInterval time.Duration
}

// New Initialize a client
func NewECSClient(region *string, logger *log.Logger, profileName *string) *ECSClient {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Profile: *profileName,
		Config:  aws.Config{Region: aws.String(*region)},
	}))
	svc := ecs.New(sess)
	return &ECSClient{
		svc:          svc,
		pollInterval: time.Second * 5,
		logger:       logger,
	}
}

// RegisterTaskDefinition updates the existing task definition's image.
func (c *ECSClient) RegisterTaskDefinition(family, task, name, image *string) (string, error) {
	taskDef, err := c.GetTaskDefinition(task)
	if err != nil {
		return "", err
	}

	//TODO maybe default if only one?
	for _, d := range taskDef.ContainerDefinitions {
		if strings.Contains(*d.Name, *name) {
			d.Image = image
		}
	}

	//Copy over everything from the existing task
	input := &ecs.RegisterTaskDefinitionInput{
		Memory:                  taskDef.Memory,
		Cpu:                     taskDef.Cpu,
		TaskRoleArn:             taskDef.TaskRoleArn,
		Family:                  family,
		NetworkMode:             taskDef.NetworkMode,
		PlacementConstraints:    taskDef.PlacementConstraints,
		RequiresCompatibilities: taskDef.RequiresCompatibilities,
		Volumes:                 taskDef.Volumes,
		ContainerDefinitions:    taskDef.ContainerDefinitions,
	}
	resp, err := c.svc.RegisterTaskDefinition(input)
	if err != nil {
		return "", err
	}
	return *resp.TaskDefinition.TaskDefinitionArn, nil
}

// UpdateService updates the service to use the new task definition.
func (c *ECSClient) UpdateService(cluster, service *string, count *int64, arn *string) error {
	input := &ecs.UpdateServiceInput{
		Cluster: cluster,
		Service: service,
	}
	if *count != -1 {
		input.DesiredCount = count
	}
	if arn != nil {
		input.TaskDefinition = arn
	}
	_, err := c.svc.UpdateService(input)
	return err
}

// Wait waits for the service to finish being updated.
func (c *ECSClient) Wait(cluster, service, arn *string) error {
	t := time.NewTicker(c.pollInterval)
	for {
		select {
		case <-t.C:
			s, err := c.GetDeployment(cluster, service, arn)
			if err != nil {
				return err
			}
			c.logger.Printf("[info] --> desired: %d, pending: %d, running: %d", *s.DesiredCount, *s.PendingCount, *s.RunningCount)
			if *s.RunningCount == *s.DesiredCount {
				return nil
			}
		}
	}
}

// GetDeployment gets the deployment for the arn.
func (c *ECSClient) GetDeployment(cluster, service, arn *string) (*ecs.Deployment, error) {
	input := &ecs.DescribeServicesInput{
		Cluster:  cluster,
		Services: []*string{service},
	}
	output, err := c.svc.DescribeServices(input)
	if err != nil {
		return nil, err
	}
	ds := output.Services[0].Deployments
	for _, d := range ds {
		if *d.TaskDefinition == *arn {
			return d, nil
		}
	}
	return nil, nil
}

// GetTaskDefinition get container definitions of the service.
func (c *ECSClient) GetTaskDefinition(task *string) (*ecs.TaskDefinition, error) {
	output, err := c.svc.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
		TaskDefinition: task,
	})
	if err != nil {
		return nil, err
	}
	fmt.Print(output)
	return output.TaskDefinition, nil
}

// GetCurrentTaskDefinition get current task arn for a service
func (c *ECSClient) GetCurrentTaskDefinition(cluster *string, service *string) (*string, error) {
	fmt.Print(*cluster)
	output, err := c.svc.DescribeServices(&ecs.DescribeServicesInput{
		Cluster: cluster,
		Services: []*string{
			service,
		},
	})
	fmt.Print(output)
	if err != nil {
		return nil, err
	}
	if len(output.Failures) > 0 {
		fmt.Print(output.Failures)
		return nil, errors.New("ECS Returned Failures")
	}
	return output.Services[0].TaskDefinition, nil
}
