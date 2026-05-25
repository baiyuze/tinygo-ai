package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

func main() {
	deviceURL := flag.String("device", "http://192.168.4.1", "ESP32-S3 base URL")
	baseURL := flag.String("base", "https://api.deepseek.com/v1", "OpenAI-compatible API base URL")
	model := flag.String("model", "deepseek-chat", "chat model")
	prompt := flag.String("prompt", "", "prompt text; stdin is used when empty")
	timeout := flag.Duration("timeout", 45*time.Second, "request timeout")
	flag.Parse()

	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		fatal("set DEEPSEEK_API_KEY first")
	}

	text := strings.TrimSpace(*prompt)
	if text == "" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fatal("read stdin: %v", err)
		}
		text = strings.TrimSpace(string(data))
	}
	if text == "" {
		fatal("empty prompt")
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	answer, err := askDeepSeek(ctx, apiKey, *baseURL, *model, text)
	if err != nil {
		fatal("chat completion: %v", err)
	}

	if err := pushToDevice(ctx, *deviceURL, answer); err != nil {
		fatal("push to device: %v", err)
	}

	fmt.Println(answer)
}

func askDeepSeek(ctx context.Context, apiKey, baseURL, model, prompt string) (string, error) {
	cfg := openai.DefaultConfig(apiKey)
	cfg.BaseURL = strings.TrimRight(baseURL, "/")
	client := openai.NewClientWithConfig(cfg)

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty response choices")
	}
	answer := strings.TrimSpace(resp.Choices[0].Message.Content)
	if answer == "" {
		return "", fmt.Errorf("empty response content")
	}
	return answer, nil
}

func pushToDevice(ctx context.Context, deviceURL, text string) error {
	endpoint := strings.TrimRight(deviceURL, "/") + "/ai"
	form := url.Values{}
	form.Set("text", text)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return fmt.Errorf("%s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
