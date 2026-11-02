//go:build ignore

// llmstub 是开发用的 OpenAI 兼容桩服务，与宿主 main.go 一样放在模块根目录。
//
// 两者都是 main 包，用 ignore 构建约束保证 `go build .` / `go vet ./...` 只编译宿主；
// 需要 stub 时显式运行：go run llmstub.go
package main

import (
	"encoding/json"
	"flag"
	"hash/fnv"
	"log"
	"math"
	"net/http"
	"strings"
	"time"
)

const (
	defaultDims  = 8
	memoryPrefix = "你是记忆整理器"
)

type chatMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

type chatRequest struct {
	Messages []chatMessage `json:"messages"`
}

type embedRequest struct {
	Input      []string `json:"input"`
	Dimensions int      `json:"dimensions"`
}

func main() {
	addr := flag.String("addr", "127.0.0.1:18100", "监听地址")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("POST /v1/chat/completions", handleChat)
	mux.HandleFunc("POST /v1/embeddings", handleEmbeddings)

	srv := &http.Server{
		Addr:              *addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("llmstub listening on http://%s/v1", *addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func handleChat(w http.ResponseWriter, r *http.Request) {
	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	extract := false
	userText := ""
	for _, m := range req.Messages {
		switch m.Role {
		case "system":
			if strings.HasPrefix(contentText(m.Content), memoryPrefix) {
				extract = true
			}
		case "user":
			if t := contentText(m.Content); t != "" {
				userText = lastUserLine(t)
			}
		}
	}
	content := "stub: " + userText
	if extract {
		content = `{"memories":[]}`
	} else if userText == "" {
		content = "stub reply"
	}
	writeJSON(w, map[string]any{
		"choices": []any{map[string]any{
			"message":       map[string]any{"content": content},
			"finish_reason": "stop",
		}},
	})
}

func handleEmbeddings(w http.ResponseWriter, r *http.Request) {
	var req embedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	dims := req.Dimensions
	if dims <= 0 {
		dims = defaultDims
	}
	data := make([]any, len(req.Input))
	for i, text := range req.Input {
		data[i] = map[string]any{"index": i, "embedding": vector(text, dims)}
	}
	writeJSON(w, map[string]any{"data": data})
}

func contentText(raw json.RawMessage) string {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	var parts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &parts); err != nil {
		return ""
	}
	var b strings.Builder
	for _, p := range parts {
		if p.Type == "text" {
			b.WriteString(p.Text)
		}
	}
	return b.String()
}

func lastUserLine(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	lines := strings.Split(s, "\n")
	line := lines[len(lines)-1]
	if _, rest, ok := strings.Cut(line, ": "); ok {
		line = rest
	}
	return line
}

func vector(text string, dims int) []float64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(text))
	seed := h.Sum64()
	v := make([]float64, dims)
	var norm float64
	for i := range v {
		x := float64((seed>>(uint(i)%64))&0xffff)/65535 - 0.5
		v[i] = x
		norm += x * x
	}
	if norm == 0 {
		v[0] = 1
		return v
	}
	norm = math.Sqrt(norm)
	for i := range v {
		v[i] /= norm
	}
	return v
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
