package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/alyu/configparser"
)

func main() {
	profile := flag.String("profile", "default", "Credential Profile To Use")
	cluster := flag.String("cluster", "cluster", "ECS Cluster")
	region := flag.String("region", "us-east-1", "AWS Region")
	task := flag.String("taskName", "", "Task ARN to update (optional)")
	family := flag.String("family", "", "Task Family - if not provided service name is used")
	service := flag.String("service", "service", "ECS Service to update")
	container := flag.String("container", "", "container to update (for use with multi container tasks)")
	image := flag.String("image", "image", "Image to run")
	nowait := flag.Bool("noWait", false, "Disable waiting")
	count := flag.Int64("count", -1, "Number of desired -1 means ignored")
	credFile := flag.String("credFile", filepath.Join(getCredentialPath(), ".aws", "credentials"), "Full path to credentials file")
	flag.Parse()

	//Get Current Credentials
	exists, err := checkProfileExists(credFile, profile)
	if err != nil || !exists {
		fmt.Println(err.Error())
		return
	}

	prefix := fmt.Sprintf("%s/%s ", *cluster, *service)
	logger := log.New(os.Stderr, prefix, log.LstdFlags)
	c := NewECSClient(region, logger, profile)

	arn := ""
	if *family == "" {
		family = service
	}

	if *task == "" {
		currentTask, err := c.GetCurrentTaskDefinition(cluster, service)
		if err != nil {
			logger.Printf("[error] getting current task definition: %s\n", err)
			return
		}
		task = currentTask
	}
	if image != nil {
		arn, err = c.RegisterTaskDefinition(family, task, container, image)
		if err != nil {
			logger.Printf("[error] register task definition: %s\n", err)
			return
		}
	}

	err = c.UpdateService(cluster, service, count, &arn)
	if err != nil {
		logger.Printf("[error] update service: %s\n", err)
		return
	}

	if *nowait == false {
		err := c.Wait(cluster, service, &arn)
		if err != nil {
			logger.Printf("[error] wait: %s\n", err)
			return
		}
	}

	logger.Printf("[info] update service success")
}

// getCredentialPath returns the users home directory path as a string
func getCredentialPath() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return usr.HomeDir
}

func checkProfileExists(credFile *string, profileName *string) (bool, error) {
	config, err := configparser.Read(*credFile)
	if err != nil {
		fmt.Println("Could not find credentials file")
		fmt.Println(err.Error())
		return false, err
	}
	section, err := config.Section(*profileName)
	if err != nil {
		fmt.Println("Could not find profile in credentials file")
		return false, nil
	}
	if !section.Exists("aws_access_key_id") {
		fmt.Println("Could not find access key in profile")
		return false, nil
	}

	return true, nil
}
