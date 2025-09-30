package main

import (
	"backend/common/config"
	oai "backend/common/openai"
	"bytes"
	"context"
	"database/sql"
	"embed"
	"encoding/base64"
	"fmt"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/bsthun/gut"
	"github.com/gen2brain/go-fitz"
	_ "github.com/lib/pq"
	"github.com/openai/openai-go"
	"go.uber.org/fx"
)

var embedMigrations embed.FS

type OcrReviser struct {
	config   *config.Config
	database *sql.DB
	openai   *openai.Client
}

type TaskRevise struct {
	Id         *int64
	UserId     *int64
	CategoryId *int64
	Type       *string
	Source     *string
	Status     *string
	Content    *string
	TokenCount *int32
	CreatedAt  *time.Time
	Remark     *string
}

func main() {
	fx.New(
		fx.Supply(
			embedMigrations,
		),
		fx.Provide(
			config.Init,
			oai.Init,
			initDatabase,
		),
		fx.Invoke(
			invoke,
		),
	).Run()
}

func initDatabase(config *config.Config) (*sql.DB, error) {
	// * connect to postgres database
	db, err := sql.Open("postgres", *config.PostgresDsn)
	if err != nil {
		return nil, gut.Err(false, "unable to connect to postgres database", err)
	}

	// * ping database
	if err = db.Ping(); err != nil {
		return nil, gut.Err(false, "unable to ping database", err)
	}

	return db, nil
}

func invoke(
	lifecycle fx.Lifecycle,
	config *config.Config,
	openai *openai.Client,
	db *sql.DB,
) {
	// * create reviser instance
	reviser := &OcrReviser{
		config:   config,
		database: db,
		openai:   openai,
	}

	reviser.reviseOcr()
}

func (r *OcrReviser) reviseOcr() {
	ctx := context.Background()

	// * query tasks with google drive links
	rows, err := r.database.QueryContext(ctx, `
		SELECT id, user_id, category_id, type, source, status, content, token_count, created_at, remark
		FROM _task_revises2
		WHERE source LIKE 'https://drive.google.com/%'
		ORDER BY id
	`)
	if err != nil {
		gut.Fatal("failed to query task revises", err)
	}
	defer rows.Close()

	// * iterate through tasks
	var tasks []TaskRevise
	for rows.Next() {
		var task TaskRevise
		err := rows.Scan(
			&task.Id,
			&task.UserId,
			&task.CategoryId,
			&task.Type,
			&task.Source,
			&task.Status,
			&task.Content,
			&task.TokenCount,
			&task.CreatedAt,
			&task.Remark,
		)
		if err != nil {
			gut.Debug("failed to scan task: %v", err)
			continue
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		gut.Fatal("error iterating rows", err)
	}

	gut.Debug("found %d tasks with google drive links", len(tasks))

	// * process each task
	processedCount := 0
	for _, task := range tasks {
		if task.Source == nil {
			continue
		}

		// * extract file id from google drive url
		fileId := r.extractFileId(*task.Source)
		if fileId == "" {
			gut.Debug("task %d: could not extract file id from %s", *task.Id, *task.Source)
			continue
		}

		// * download pdf from google drive
		pdfData, err := r.downloadFromGoogleDrive(fileId)
		if err != nil {
			gut.Debug("task %d: failed to download pdf: %v", *task.Id, err)
			continue
		}

		// * extract text from pdf
		extractedText, err := r.extractTextFromPdf(pdfData)
		if err != nil {
			gut.Debug("task %d: failed to extract text from pdf: %v", *task.Id, err)
			continue
		}

		// * update task with extracted content
		_, err = r.database.ExecContext(ctx, `
			UPDATE _task_revises2
			SET content = $1
			WHERE id = $2
		`, extractedText, task.Id)
		if err != nil {
			gut.Debug("task %d: failed to update content: %v", *task.Id, err)
			continue
		}

		gut.Debug("task %d: successfully extracted and updated content (%d chars)", *task.Id, len(extractedText))
		processedCount++
	}

	gut.Debug("processed %d tasks", processedCount)
}

func (r *OcrReviser) extractFileId(url string) string {
	// * regex patterns for different google drive url formats
	patterns := []string{
		`/file/d/([a-zA-Z0-9_-]+)`,
		`id=([a-zA-Z0-9_-]+)`,
		`/d/([a-zA-Z0-9_-]+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(url)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}

func (r *OcrReviser) downloadFromGoogleDrive(fileId string) ([]byte, error) {
	// * construct direct download url
	downloadUrl := fmt.Sprintf("https://drive.google.com/uc?export=download&id=%s", fileId)

	// * create http client with timeout
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	// * make request
	resp, err := client.Get(downloadUrl)
	if err != nil {
		return nil, gut.Err(false, "failed to download file", err)
	}
	defer resp.Body.Close()

	// * check response status
	if resp.StatusCode != http.StatusOK {
		// * try alternative download url
		downloadUrl = fmt.Sprintf("https://drive.google.com/u/0/uc?id=%s&export=download", fileId)
		resp, err = client.Get(downloadUrl)
		if err != nil {
			return nil, gut.Err(false, "failed to download file with alternative url", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, gut.Err(false, fmt.Sprintf("failed to download file: status %d", resp.StatusCode))
		}
	}

	// * read response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, gut.Err(false, "failed to read response body", err)
	}

	return data, nil
}

func (r *OcrReviser) extractTextFromPdf(pdfData []byte) (string, error) {
	// * create temp directory if not exists
	tempDir := ".local/temp"
	err := os.MkdirAll(tempDir, os.ModePerm)
	if err != nil {
		return "", gut.Err(false, "failed to create temp directory", err)
	}

	// * create temp file for pdf
	tempFile := filepath.Join(tempDir, fmt.Sprintf("temp_%d.pdf", time.Now().UnixNano()))
	err = os.WriteFile(tempFile, pdfData, 0644)
	if err != nil {
		return "", gut.Err(false, "failed to write temp pdf file", err)
	}
	defer os.Remove(tempFile)

	// * open pdf with fitz for rendering pages as images
	doc, err := fitz.New(tempFile)
	if err != nil {
		return "", gut.Err(false, "failed to open pdf with fitz", err)
	}
	defer doc.Close()

	// * prepare for concurrent processing
	totalPages := doc.NumPage()
	type pageResult struct {
		pageNo int
		text   string
		err    error
	}

	// * create channel for results and semaphore for concurrency control
	resultChan := make(chan pageResult, totalPages)
	semaphore := make(chan struct{}, 10) // * max 10 concurrent goroutines

	// * waitgroup to track all goroutines
	var wg sync.WaitGroup

	// * spawn goroutines for each page
	for pageNo := 0; pageNo < totalPages; pageNo++ {
		wg.Add(1)
		go func(pageNum int) {
			defer wg.Done()

			// * acquire semaphore slot
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// * render page as image
			img, err := doc.Image(pageNum)
			if err != nil {
				resultChan <- pageResult{
					pageNo: pageNum,
					text:   "",
					err:    fmt.Errorf("failed to render page %d as image: %v", pageNum+1, err),
				}
				return
			}

			// * create in-memory buffer for image
			imgBuffer := new(bytes.Buffer)

			// * encode image to jpeg in memory
			err = jpeg.Encode(imgBuffer, img, &jpeg.Options{Quality: 95})
			if err != nil {
				resultChan <- pageResult{
					pageNo: pageNum,
					text:   "",
					err:    fmt.Errorf("failed to encode page image for page %d: %v", pageNum+1, err),
				}
				return
			}

			// * get image data from buffer
			imageData := imgBuffer.Bytes()

			// * encode image to base64
			base64Image := base64.StdEncoding.EncodeToString(imageData)

			// * extract text using openai
			pageText, err := r.extractTextWithOpenAI(base64Image, pageNum+1)
			if err != nil {
				resultChan <- pageResult{
					pageNo: pageNum,
					text:   "",
					err:    fmt.Errorf("failed to extract text with openai for page %d: %v", pageNum+1, err),
				}
				return
			}

			println(pageText)

			// * send successful result
			resultChan <- pageResult{
				pageNo: pageNum,
				text:   pageText,
				err:    nil,
			}
		}(pageNo)
	}

	// * close result channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// * collect results in order
	pageTexts := make(map[int]string)
	for result := range resultChan {
		if result.err != nil {
			gut.Debug("%v", result.err)
			continue
		}
		pageTexts[result.pageNo] = result.text
	}

	// * build final text in correct page order
	var textBuilder strings.Builder
	for pageNo := 0; pageNo < totalPages; pageNo++ {
		text, exists := pageTexts[pageNo]
		if !exists {
			continue
		}

		// * add page separator if not first page
		if pageNo > 0 && textBuilder.Len() > 0 {
			textBuilder.WriteString("\n\n--- Page ")
			textBuilder.WriteString(fmt.Sprintf("%d", pageNo+1))
			textBuilder.WriteString(" ---\n\n")
		}

		// * add extracted text
		textBuilder.WriteString(text)
	}

	return textBuilder.String(), nil
}

func (r *OcrReviser) extractTextWithOpenAI(base64Image string, pageNo int) (string, error) {
	// * create openai messages for text extraction
	var messages []openai.ChatCompletionMessageParamUnion

	// * add system message
	systemContent := "Extract all text from this image exactly as it appears. Maintain the original formatting, structure, and language. Do not translate, summarize, or modify the content. Return only the extracted text without any additional commentary."
	messages = append(messages, openai.SystemMessage(systemContent))

	// * add user content parts
	var userContentParts []openai.ChatCompletionContentPartUnionParam
	userContentParts = append(userContentParts, openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
		URL: fmt.Sprintf("data:image/jpeg;base64,%s", base64Image),
	}))

	// * add user message with content parts
	messages = append(messages, openai.ChatCompletionMessageParamUnion{
		OfUser: &openai.ChatCompletionUserMessageParam{
			Role: "user",
			Content: openai.ChatCompletionUserMessageParamContentUnion{
				OfArrayOfContentParts: userContentParts,
			},
		},
	})

	// * prepare chat completion request params
	chatParams := openai.ChatCompletionNewParams{
		Messages:  messages,
		Model:     *r.config.OpenaiModel,
		MaxTokens: openai.Int(2048),
	}

	// * call openai api with retry
	ctx := context.Background()
	maxRetries := 3
	var chatCompletion *openai.ChatCompletion
	var err error

	for i := 0; i < maxRetries; i++ {
		chatCompletion, err = r.openai.Chat.Completions.New(ctx, chatParams)
		if err == nil {
			break
		}
		if i < maxRetries-1 {
			gut.Debug("retry %d for page %d due to error: %v", i+1, pageNo, err)
			time.Sleep(2 * time.Second)
		}
	}

	if err != nil {
		return "", gut.Err(false, fmt.Sprintf("failed to extract text for page %d after %d retries", pageNo, maxRetries), err)
	}

	// * get extracted text
	extractedText := chatCompletion.Choices[0].Message.Content

	return extractedText, nil
}
