package probe

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"

type ChatOpts struct {
	Stream      bool
	Temperature *float64
	MaxTokens   *int
}

type Client struct {
	httpClient *http.Client
	userAgent  string
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 120 * time.Second},
		userAgent:  defaultUserAgent,
	}
}

type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream"`
	Temperature *float64  `json:"temperature,omitempty"`
	MaxTokens   *int      `json:"max_tokens,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	ID      string   `json:"id"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   *Usage   `json:"usage,omitempty"`
	Error   *APIError `json:"error,omitempty"`
}

type Choice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
	Delta        *Message `json:"delta,omitempty"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

type Result struct {
	Content       string
	ResponseModel string
	LatencyMS     int
	TTFTMS        *int
	TPOTMS        *float64
	InputTokens   *int
	OutputTokens  *int
	HTTPStatus    int
	Error         string
	ResponseHash  string
	ContentHash   string
	CacheHeaders  string
}

type Endpoint struct {
	BaseURL string
	APIKey  string
	Backups []string
}

func (c *Client) ChatCompletion(ctx context.Context, ep Endpoint, model, prompt string, stream bool) (*Result, error) {
	return c.ChatCompletionOpts(ctx, ep, model, prompt, ChatOpts{Stream: stream})
}

func (c *Client) ChatCompletionOpts(ctx context.Context, ep Endpoint, model, prompt string, opts ChatOpts) (*Result, error) {
	urls := append([]string{ep.BaseURL}, ep.Backups...)
	var lastErr error
	for _, base := range urls {
		res, err := c.doChat(ctx, base, ep.APIKey, model, prompt, opts)
		if err == nil {
			return res, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func (c *Client) doChat(ctx context.Context, baseURL, apiKey, model, prompt string, opts ChatOpts) (*Result, error) {
	baseURL = strings.TrimRight(baseURL, "/")
	url := baseURL + "/chat/completions"

	body, _ := json.Marshal(ChatRequest{
		Model:       model,
		Messages:    []Message{{Role: "user", Content: prompt}},
		Stream:      opts.Stream,
		Temperature: opts.Temperature,
		MaxTokens:   opts.MaxTokens,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &Result{Error: err.Error(), LatencyMS: int(time.Since(start).Milliseconds())}, nil
	}
	defer resp.Body.Close()

	if opts.Stream {
		return c.parseStream(resp, start)
	}
	return c.parseNonStream(resp, start)
}

func (c *Client) parseNonStream(resp *http.Response, start time.Time) (*Result, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &Result{HTTPStatus: resp.StatusCode, Error: err.Error(), LatencyMS: int(time.Since(start).Milliseconds())}, nil
	}

	latency := int(time.Since(start).Milliseconds())
	hash := hashContent(body)

	if resp.StatusCode >= 400 {
		return &Result{
			HTTPStatus:   resp.StatusCode,
			Error:        string(body),
			LatencyMS:    latency,
			ResponseHash: hash,
		}, nil
	}

	var chat ChatResponse
	if err := json.Unmarshal(body, &chat); err != nil {
		return &Result{HTTPStatus: resp.StatusCode, Error: err.Error(), LatencyMS: latency, ResponseHash: hash}, nil
	}
	if chat.Error != nil {
		return &Result{HTTPStatus: resp.StatusCode, Error: chat.Error.Message, LatencyMS: latency, ResponseHash: hash}, nil
	}

	content := ""
	if len(chat.Choices) > 0 {
		content = chat.Choices[0].Message.Content
	}

	res := &Result{
		Content:       content,
		ResponseModel: chat.Model,
		LatencyMS:     latency,
		HTTPStatus:    resp.StatusCode,
		ResponseHash:  hash,
		ContentHash:   hashContent([]byte(strings.TrimSpace(content))),
		CacheHeaders:  FormatCacheHeaderHints(resp.Header),
	}
	if chat.Usage != nil {
		res.InputTokens = &chat.Usage.PromptTokens
		res.OutputTokens = &chat.Usage.CompletionTokens
	}
	return res, nil
}

func (c *Client) parseStream(resp *http.Response, start time.Time) (*Result, error) {
	latency := int(time.Since(start).Milliseconds())

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return &Result{
			HTTPStatus:   resp.StatusCode,
			Error:        string(body),
			LatencyMS:    latency,
			ResponseHash: hashContent(body),
		}, nil
	}

	var content strings.Builder
	var responseModel string
	var ttft *int
	var outputTokens int
	firstToken := false

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk ChatResponse
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if chunk.Model != "" {
			responseModel = chunk.Model
		}
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta != nil {
			delta := chunk.Choices[0].Delta.Content
			if delta != "" {
				if !firstToken {
					ms := int(time.Since(start).Milliseconds())
					ttft = &ms
					firstToken = true
				}
				content.WriteString(delta)
				outputTokens++
			}
		}
	}

	total := int(time.Since(start).Milliseconds())
	full := content.String()
	hash := hashContent([]byte(full))

	res := &Result{
		Content:       full,
		ResponseModel: responseModel,
		LatencyMS:     total,
		TTFTMS:        ttft,
		HTTPStatus:    resp.StatusCode,
		ResponseHash:  hash,
		ContentHash:   hashContent([]byte(strings.TrimSpace(full))),
		CacheHeaders:  FormatCacheHeaderHints(resp.Header),
	}
	if outputTokens > 0 && ttft != nil {
		tpot := float64(total-*ttft) / float64(outputTokens)
		res.TPOTMS = &tpot
		res.OutputTokens = &outputTokens
	}
	return res, nil
}

func (c *Client) ListModels(ctx context.Context, ep Endpoint) ([]string, error) {
	baseURL := strings.TrimRight(ep.BaseURL, "/")
	url := baseURL + "/models"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+ep.APIKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("models API %d: %s", resp.StatusCode, body)
	}

	var payload struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	models := make([]string, 0, len(payload.Data))
	for _, m := range payload.Data {
		models = append(models, m.ID)
	}
	return models, nil
}

func hashContent(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:8])
}

func Summarize(content string, maxLen int) string {
	content = strings.TrimSpace(content)
	content = strings.ReplaceAll(content, "\n", " ")
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen] + "..."
}
