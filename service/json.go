package service

import (
	"data/utils"
	"log"
	"sync"
)

type AccessPtr struct {
	Start     int
	ChunkSize int
}

type jsonProcessor struct {
	fileUtil utils.FileUtil
	//mutex       *sync.Mutex
	wg *sync.WaitGroup
}

type JSONProcessor interface {
	Process(offset AccessPtr)
}

func NewJSONProcessor(fileUtil utils.FileUtil, wg *sync.WaitGroup) JSONProcessor {
	return &jsonProcessor{
		fileUtil: fileUtil,
		//mutex:       &sync.Mutex{},
		wg: wg,
	}
}

func (j jsonProcessor) Process(offset AccessPtr) {
	defer j.wg.Done()

	//j.mutex.Lock()
	chunk := j.fileUtil.GetChunkFromFileMap(offset.Start, offset.ChunkSize)
	//j.mutex.Unlock()
	for i := 0; i < len(chunk); i++ {
		if ';' == chunk[i] {
			log.Println("; encountered. fixing it")
			chunk[i] = ':'
		}
	}
	j.fileUtil.UpdateChunkToFileMap(offset.Start, offset.Start+offset.ChunkSize, chunk)
}
