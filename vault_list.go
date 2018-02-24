package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glacier"
)

type VaultListCommand struct{}

func (c VaultListCommand) Execute(args []string) error {
	debugLog.Println("executing VaultListCommand")

	output, err := glacierService.ListVaults(&glacier.ListVaultsInput{
		AccountId: aws.String("-"),
	})
	if err != nil {
		return err
	}

	debugLog.Printf("ListVaults response is %+v\n", output)
	fmt.Printf("Vaults:\n\n")

	for _, v := range output.VaultList {
		fmt.Println(*v.VaultName)
	}

	return nil
}
