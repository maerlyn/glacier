package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/glacier"
	"github.com/jessevdk/go-flags"
	"log"
	"os"
)

var (
	Defaults struct{}

	glacierService *glacier.Glacier
	debugLog       *log.Logger
)

func init() {
	vaultListCommand := VaultListCommand{}
	uploadCommand := UploadCommand{}
	inventoryCommand := InventoryCommand{}

	parser := flags.NewParser(&Defaults, flags.Default)
	parser.AddCommand("vault-list", "list vaults", "list vaults", &vaultListCommand)
	parser.AddCommand("upload", "upload", "upload", &uploadCommand)
	parser.AddCommand("inventory", "inventory", "inventory", &inventoryCommand)

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-central-1"),
	})
	if err != nil {
		panic(err)
	}

	glacierService = glacier.New(sess)

	debugOut, err := os.OpenFile("debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	debugLog = log.New(debugOut, "", log.LstdFlags|log.Lshortfile)

	_, err = parser.Parse()
	if err != nil {
		panic(err)
	}
}

func main() {
	//intentionally empty
}
