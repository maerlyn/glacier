package main

import (
	"github.com/jessevdk/go-flags"
	//"os"
	//"github.com/mattetti/filebuffer"
	//"io"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/glacier"
)

var (
	Defaults struct{}

	glacierService *glacier.Glacier
)

func init() {
	vaultListCommand := VaultListCommand{}

	parser := flags.NewParser(&Defaults, flags.Default)
	parser.AddCommand("vault-list", "list vaults", "list vaults", &vaultListCommand)

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-central-1"),
	})
	if err != nil {
		panic(err)
	}

	glacierService = glacier.New(sess)

	parser.Parse()
}

func main() {

	//file, _ := os.Open("randomfile")
	//stat, _ := file.Stat()
	//buffer := make([]byte, stat.Size())
	//io.ReadFull(file, buffer)
	//
	//fb := filebuffer.New(buffer)
	//
	//fmt.Printf("%x\n", glacier.ComputeHashes(fb).TreeHash)
}
