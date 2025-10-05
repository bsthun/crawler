package main

import (
	"backend/common/config"
	"context"
	"database/sql"
	"embed"
	"fmt"
	"time"

	"github.com/bsthun/gut"
	"github.com/go-resty/resty/v2"
	"github.com/huantt/plaintext-extractor"
	_ "github.com/lib/pq"
	"go.uber.org/fx"
)

var embedMigrations embed.FS

type CrawlReviser struct {
	config   *config.Config
	database *sql.DB
	client   *resty.Client
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

type FirecrawlRequest struct {
	URL     string   `json:"url"`
	Formats []string `json:"formats"`
}

type FirecrawlResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Markdown string `json:"markdown"`
		Content  string `json:"content"`
	} `json:"data"`
	Error string `json:"error"`
}

func main() {
	fx.New(
		fx.Supply(
			embedMigrations,
		),
		fx.Provide(
			config.Init,
			initDatabase,
			initHttpClient,
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

func initHttpClient() *resty.Client {
	// * create resty client
	client := resty.New()
	client.SetTimeout(60 * time.Second)
	client.SetRetryCount(3)
	client.SetRetryWaitTime(2 * time.Second)

	return client
}

func invoke(
	lifecycle fx.Lifecycle,
	config *config.Config,
	db *sql.DB,
	client *resty.Client,
) {
	// * create reviser instance
	reviser := &CrawlReviser{
		config:   config,
		database: db,
		client:   client,
	}

	reviser.reviseCrawl()
}

func (r *CrawlReviser) reviseCrawl() {
	ctx := context.Background()

	// * query tasks with category_id = 1 and type = web
	rows, err := r.database.QueryContext(ctx, `
		SELECT id, user_id, category_id, type, source, status, content, token_count, created_at, remark
		FROM _task_revises2
		WHERE category_id = 1
		AND type = 'web'
		AND content = ''
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

	gut.Debug("found %d web tasks to crawl", len(tasks))

	// * process each task
	processedCount := 0
	for _, task := range tasks {
		if task.Source == nil || *task.Source == "" {
			continue
		}

		// * crawl content from url
		content, err := r.crawlContent(*task.Source)
		if err != nil {
			gut.Debug("task %d: failed to crawl content: %v", *task.Id, err)
			continue
		}

		// * update task with crawled content
		_, err = r.database.ExecContext(ctx, `
			UPDATE _task_revises2
			SET content = $1
			WHERE id = $2
		`, content, task.Id)
		if err != nil {
			gut.Debug("task %d: failed to update content: %v", *task.Id, err)
			continue
		}

		gut.Debug("task %d: successfully crawled and updated content (%d chars)", *task.Id, len(content))
		processedCount++
	}

	gut.Debug("processed %d tasks", processedCount)
}

func (r *CrawlReviser) crawlContent(url string) (string, error) {
	// * prepare request payload
	payload := FirecrawlRequest{
		URL:     url,
		Formats: []string{"markdown"},
	}

	// * make request to firecrawl api
	var response FirecrawlResponse
	resp, err := r.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer fc-1").
		SetBody(payload).
		SetResult(&response).
		Post("http://10.5.30.161:3002/v2/scrape")

	if err != nil {
		return "", gut.Err(false, "failed to make request to firecrawl", err)
	}

	// * check http status
	if resp.StatusCode() != 200 {
		return "", gut.Err(false, fmt.Sprintf("firecrawl returned status %d", resp.StatusCode()))
	}

	// * check api response
	if !response.Success {
		return "", gut.Err(false, fmt.Sprintf("firecrawl api error: %s", response.Error))
	}

	// * get markdown content
	markdownContent := response.Data.Markdown
	if markdownContent == "" {
		markdownContent = response.Data.Content
	}

	// * convert markdown to plain text
	extractor := plaintext.NewMarkdownExtractor()
	plainText, err := extractor.PlainText(markdownContent)
	if err != nil {
		return "", gut.Err(false, "failed to convert markdown to plain text", err)
	}

	if plainText == nil {
		return "", nil
	}

	return *plainText, nil
}
