package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glacier"
	"github.com/gosuri/uiprogress"
	"github.com/mattetti/filebuffer"
	"io"
	"os"
	"path/filepath"
	"time"
	"strconv"
)

const (
	//LargeFileLimit = 4 * 1024 * 1024 * 1024 //4GB
	//BufferSize     = 100 * 1024 * 1024      //100MB
	LargeFileLimit = 1 * 1024 * 1024 //4GB
	BufferSize     = 1 * 1024 * 1024      //100MB
)

type UploadCommand struct {
	VaultName string `long:"vault-name" required:"true"`
	FileName  string `long:"file-name" required:"true"`
}

func (c UploadCommand) Execute(args []string) error {
	debugLog.Printf("executing UploadCommand on vault %s, file %s\n", c.VaultName, c.FileName)

	stat, err := os.Stat(c.FileName)
	if err != nil {
		return err
	}
	debugLog.Printf("file size is %d\n", stat.Size())

	if stat.Size() < LargeFileLimit {
		debugLog.Println("uploadSmall")
		return c.uploadSmall()
	} else {
		debugLog.Println("uploadLarge")
		return c.uploadLarge()
	}
}

func (c UploadCommand) uploadSmall() error {
	stat, err := os.Stat(c.FileName)
	if err != nil {
		debugLog.Printf("err %+v\n", err)
		return err
	}

	file, err := os.Open(c.FileName)
	if err != nil {
		debugLog.Printf("err %+v\n", err)
		return err
	}

	_, OnlyFileName := filepath.Split(c.FileName)

	input := glacier.UploadArchiveInput{
		AccountId:          aws.String("-"),
		VaultName:          aws.String(c.VaultName),
		ArchiveDescription: aws.String(OnlyFileName),
		Body:               file,
		Checksum:           aws.String(fmt.Sprintf("%x", glacier.ComputeHashes(file).TreeHash)),
	}
	debugLog.Printf("UploadArchive input is %+v\n", input)

	uiprogress.Start()
	pb := uiprogress.AddBar(int(stat.Size()))
	pb.AppendCompleted()
	pb.PrependElapsed()

	ticker := time.NewTicker(time.Second)
	quit := make(chan interface{})

	go func() {
		for {
			select {
			case <-ticker.C:
				pos, _ := file.Seek(0, io.SeekCurrent)
				pb.Set(int(pos))
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	response, err := glacierService.UploadArchive(&input)
	close(quit)
	if err != nil {
		debugLog.Printf("err %+v\n", err)
		return err
	}
	debugLog.Printf("UploadArchive response is %+v\n", response)

	fmt.Printf("uploaded as %s\n", *response.ArchiveId)

	return nil
}

func (c UploadCommand) uploadLarge() error {
	stat, err := os.Stat(c.FileName)
	if err != nil {
		debugLog.Printf("err %+v\n", err)
		return err
	}

	file, err := os.Open(c.FileName)
	if err != nil {
		debugLog.Printf("err %+v\n", err)
		return err
	}

	_, OnlyFileName := filepath.Split(c.FileName)

	input := glacier.InitiateMultipartUploadInput{
		AccountId:          aws.String("-"),
		VaultName:          aws.String(c.VaultName),
		ArchiveDescription: aws.String(OnlyFileName),
		PartSize:           aws.String(strconv.Itoa(BufferSize)),
	}
	debugLog.Printf("InitiateMultipartUpload input is %+v\n", input)
	multipart, err := glacierService.InitiateMultipartUpload(&input)
	if err != nil {
		debugLog.Printf("err %+v\n", err)
		return err
	}
	debugLog.Printf("InitiateMultipartUpload response is %+v\n", multipart)

	buffer := make([]byte, BufferSize)
	fb := filebuffer.New(buffer)
	loopIndex := 0

	uiprogress.Start()
	pb := uiprogress.AddBar(int(stat.Size()))
	pb.AppendCompleted()
	pb.PrependElapsed()

	ticker := time.NewTicker(time.Second)
	quit := make(chan interface{})

	go func() {
		for {
			select {
			case <-ticker.C:
				pos, _ := fb.Seek(0, io.SeekCurrent)
				pb.Set(loopIndex*BufferSize + int(pos))
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	for {
		bytesRead, err := io.ReadFull(file, buffer)
		debugLog.Printf("read %d bytes into buffer\n", bytesRead)

		if bytesRead != len(buffer) {
			fb = filebuffer.New(buffer[:bytesRead])
		}

		if err == io.EOF {
			break
		}

		if err != nil && err != io.ErrUnexpectedEOF  {
			debugLog.Printf("err %+v\n", err)
			return err
		}

		fb.Seek(0, io.SeekStart)
		bytesLower := loopIndex * BufferSize
		bytesUpper := 0
		if BufferSize == bytesRead {
			bytesUpper = (loopIndex+1)*BufferSize - 1
		} else {
			bytesUpper = loopIndex*BufferSize + bytesRead - 1
		}
		debugLog.Printf("current part range is %d-%d\n", bytesLower, bytesUpper)

		multipartInput := glacier.UploadMultipartPartInput{
			AccountId: aws.String("-"),
			VaultName: aws.String(c.VaultName),

			UploadId: multipart.UploadId,
			Body:     fb,
			Range:    aws.String(fmt.Sprintf("bytes %d-%d/*", bytesLower, bytesUpper)),
			Checksum: aws.String(fmt.Sprintf("%x", glacier.ComputeHashes(fb).TreeHash)),
		}
		debugLog.Printf("UploadMultipartPartInput request is %+v\n", multipartInput)
		multipartOutput, err := glacierService.UploadMultipartPart(&multipartInput)
		if err != nil {
			debugLog.Printf("err %+v\n", err)
			return err
		}
		debugLog.Printf("UploadMultipartPartInput response is %+v\n", multipartOutput)

		loopIndex = loopIndex + 1

		if len(buffer) != bytesRead {
			break
		}
	}

	completeMultipartInput := glacier.CompleteMultipartUploadInput{
		AccountId: aws.String("-"),
		VaultName: aws.String(c.VaultName),

		ArchiveSize: aws.String(fmt.Sprintf("%d", stat.Size())),
		Checksum:    aws.String(fmt.Sprintf("%x", glacier.ComputeHashes(file).TreeHash)),
		UploadId:    multipart.UploadId,
	}

	debugLog.Printf("CompleteMultipartUploadInput request is %+v\n", completeMultipartInput)
	completeResponse, err := glacierService.CompleteMultipartUpload(&completeMultipartInput)
	if err != nil {
		debugLog.Printf("err %+v\n", err)
		return err
	}
	debugLog.Printf("CompleteMultipartUploadInput response is %+v\n", completeResponse)

	fmt.Printf("Uploaded as %s\n", *completeResponse.ArchiveId)

	return nil
}
