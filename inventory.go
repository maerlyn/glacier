package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glacier"
	"io/ioutil"
	"time"
)

type InventoryCommand struct {
	VaultName string `long:"vault-name" required:"true"`
}

func (c InventoryCommand) Execute(args []string) error {
	debugLog.Printf("executing inventory on vault %s\n", c.VaultName)

	input := glacier.InitiateJobInput{
		AccountId: aws.String("-"),
		VaultName: aws.String(c.VaultName),
		JobParameters: &glacier.JobParameters{
			Type: aws.String("inventory-retrieval"),
		},
	}

	initiateResult, err := glacierService.InitiateJob(&input)
	if err != nil {
		return err
	}

	debugLog.Printf("initiate job response is %+v\n", *initiateResult)
	fmt.Printf("job started as %s\n", *initiateResult.JobId)
	fmt.Println("checking for completion every 10 minutes")

	jobId := *initiateResult.JobId
	completed := false
	timer := time.NewTicker(10 * time.Minute)
	listJobsInput := glacier.ListJobsInput{
		AccountId: aws.String("-"),
		VaultName: aws.String(c.VaultName),
	}

	for !completed {
		<-timer.C

		fmt.Printf("checking... ")
		result, err := glacierService.ListJobs(&listJobsInput)
		if err != nil {
			return err
		}
		debugLog.Printf("ListJobs response is %+v\n", result)

		for _, v := range result.JobList {
			if *v.JobId == jobId {
				completed = *v.Completed

				if completed {
					fmt.Println("completed")
				} else {
					fmt.Println("incomplete")
				}
			}
		}
	}

	timer.Stop()

	getJobOutputInput := glacier.GetJobOutputInput{
		AccountId: aws.String("-"),
		VaultName: aws.String(c.VaultName),
		JobId:     aws.String(jobId),
	}
	getJobResult, err := glacierService.GetJobOutput(&getJobOutputInput)
	defer getJobResult.Body.Close()
	if err != nil {
		return err
	}
	debugLog.Printf("GetJobOutput response is %+v\n", getJobResult)

	output, err := ioutil.ReadAll(getJobResult.Body)
	if err != nil {
		return err
	}

	fmt.Println(string(output))

	return nil
}
