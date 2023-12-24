package tasks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	client *asynq.Client
	once   sync.Once
)
var StandardWidths = []uint{16, 32, 128, 240, 320, 480, 540, 640, 800, 1024}

const TypeResizeImage = "image:resize"

type ResizeImagePayload struct {
	ImageData []byte
	Width     uint
	Height    uint
	FileName  string
}

func Init(redisAddress string) {
	once.Do(func() {
		client = asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddress})

	})
}

func Close() {
	if client != nil {
		client.Close()
	}
}

func GetClient() *asynq.Client {
	return client
}

func NewImageResizerTask(imageData []byte, fileName string) ([]*asynq.Task, error) {
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, err
	}

	originalBounds := img.Bounds()
	originalWidths := uint(originalBounds.Dx())
	originalHeight := uint(originalBounds.Dy())
	var tasks []*asynq.Task
	for _, width := range StandardWidths {
		height := (width * originalHeight) / originalWidths
		payload := ResizeImagePayload{
			ImageData: imageData,
			Width:     width,
			Height:    height,
			FileName:  fileName,
		}
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		task := asynq.NewTask(TypeResizeImage, payloadBytes)
		_ = append(tasks, task)
	}
	return tasks, nil
}

func HandleResizeImageTask(ctx context.Context, t asynq.Task) error {
	var payload ResizeImagePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to parse resize image task payload: %v", err)
	}
	img, _, err := image.Decode(bytes.NewReader(payload.ImageData))
	if err != nil {
		return fmt.Errorf("image decode failed: %v", err)
	}
	resizeImg := resize.Resize(payload.Width, payload.Height, img, resize.Lanczos3)
	outputUuid := uuid.New()
	outputFileName := fmt.Sprintf("images/%s/%s%s", time.Now().Format("2006-01-02"), outputUuid.String(), filepath.Ext(payload.FileName))
	outputDir := filepath.Dir(outputFileName)
	if _, err := os.Stat(outputDir); err != nil {
		err := os.MkdirAll(outputDir, 0755)
		if err != nil {
			return err
		}
	}
	outputFile, err := os.Create(outputFileName)
	if err != nil {
		return err
	}
	defer outputFile.Close()
	if err := jpeg.Encode(outputFile, resizeImg, nil); err != nil {
		return err
	}
	fmt.Printf("Output UUID for the processed image: %s\b", outputUuid.String())
	return nil
}
