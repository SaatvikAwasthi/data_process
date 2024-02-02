package main

import (
	"data/constants"
	"data/service"
	"data/utils"
	"flag"
	"log"
	"runtime"
	"sync"
)

func main() {
	defer utils.Timer("main")()

	filePath := flag.String(constants.FlagFilePath, constants.EmptyString, "provide the filepath for processing.")
	flag.Parse()

	if filePath == nil || *filePath == constants.EmptyString {
		log.Fatalf("file path not present")
		return
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	file := utils.NewFile(*filePath)
	defer file.GracefullyFileClosing()

	err := ProcessFile(file)
	if err != nil {
		log.Printf("error occured. %v", err)
	}

	//utils.ShowInfo()

}

func ProcessFile(fileUtil utils.FileUtil) error {
	var wg sync.WaitGroup

	err := fileUtil.OpenFileReadWrite()
	if err != nil {
		return err
	}
	jsonPsr := service.NewJSONProcessor(fileUtil, &wg)

	fileSize, err := fileUtil.GetFileSize()
	if err != nil {
		return err
	}

	_, mapErr := fileUtil.GetFileMap()
	if mapErr != nil {
		return mapErr
	}

	// Reading and updating the file in smaller chunks
	for offset := 0; offset < int(fileSize); offset += constants.ChunkSize {
		offset := service.AccessPtr{
			Start:     offset,
			ChunkSize: constants.ChunkSize,
		}

		wg.Add(1)
		go jsonPsr.Process(offset)
	}

	wg.Wait()

	syncErr := fileUtil.SyncToFile()
	if syncErr != nil {
		return syncErr
	}

	log.Println("File processed successfully.")

	return nil
}
