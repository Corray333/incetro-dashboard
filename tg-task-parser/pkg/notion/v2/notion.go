package v2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
)

const TIME_LAYOUT = "2006-01-02T15:04:05.000-07:00"
const TIME_LAYOUT_IN = "2006-01-02T15:04:05.999Z07:00"

type TableType string

var IsSyncing = false

const (
	TimeTable    TableType = "time"
	TaskTable    TableType = "task"
	ProjectTable TableType = "project"
)

type Client struct {
	transport *http.Client
	auth      string
}

func NewClient() *Client {
	token := "Bearer " + os.Getenv("NOTION_SECRET")

	return &Client{
		transport: getHTTPClient(),
		auth:      token,
	}
}

func getHTTPClient() *http.Client {

	transport := &http.Transport{}

	proxy := os.Getenv("PROXY_URL")
	if proxy != "" {
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			slog.Error("Failed to parse proxy URL", slog.String("error", err.Error()))
			panic(err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	return &http.Client{
		Transport: transport,
	}
}

type searchResponse struct {
	HasMore    bool            `json:"has_more"`
	NextCursor string          `json:"next_cursor"`
	Results    json.RawMessage `json:"results"` // или []map[string]interface{} если хочешь сразу распарсить
}

func (c *Client) SearchPages(dbid string, filter map[string]interface{}) ([]byte, error) {
	urlStr := "https://api.notion.com/v1/databases/" + dbid + "/query"

	var allResults []json.RawMessage
	startCursor := ""

	for {
		// Создаём копию фильтра, чтобы можно было безопасно дополнять его курсором
		reqBody := make(map[string]interface{})
		for k, v := range filter {
			reqBody[k] = v
		}
		if startCursor != "" {
			reqBody["start_cursor"] = startCursor
		}

		data, err := json.Marshal(reqBody)
		if err != nil {
			return nil, err
		}

		req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(data))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", c.auth)
		req.Header.Set("Notion-Version", "2022-06-28")
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.transport.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != 200 {
			slog.Error("notion error while searching pages: " + string(body))
			return nil, fmt.Errorf("notion error %s while searching pages with body %s", string(body), string(data))
		}

		var page searchResponse
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}

		// добавляем текущие результаты
		var pageResults []json.RawMessage
		if err := json.Unmarshal(page.Results, &pageResults); err != nil {
			return nil, err
		}
		allResults = append(allResults, pageResults...)

		if !page.HasMore || page.NextCursor == "" {
			break
		}
		startCursor = page.NextCursor
	}

	// Финальный маршал всех результатов в []byte
	finalJSON, err := json.Marshal(allResults)
	if err != nil {
		return nil, err
	}
	return finalJSON, nil
}

func (c *Client) GetPage(pageid string) ([]byte, error) {
	url := "https://api.notion.com/v1/pages/" + pageid

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", c.auth)
	req.Header.Set("Notion-Version", "2022-06-28")

	resp, err := c.transport.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("notion error while getting page: %s", string(body))
	}

	return body, nil
}

func (c *Client) CreatePage(dbid string, properties interface{}, content interface{}) ([]byte, error) {
	url := "https://api.notion.com/v1/pages"

	reqBody := map[string]interface{}{
		"parent": map[string]interface{}{
			"type":        "database_id",
			"database_id": dbid,
		},
		"properties": properties,
	}

	if content != nil {
		reqBody["children"] = content
	}

	// if icons[icon] != "" {
	// 	reqBody["icon"] = map[string]interface{}{
	// 		"type": "external",
	// 		"external": map[string]interface{}{
	// 			"url": icons[icon],
	// 		},
	// 	}
	// }

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", c.auth)
	req.Header.Set("Notion-Version", "2022-06-28")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.transport.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("notion error %s while creating page with properties %s", string(body), string(data))
	}

	return body, nil
}

func (c *Client) UpdatePage(pageid string, properties interface{}) ([]byte, error) {
	url := "https://api.notion.com/v1/pages/" + pageid

	reqBody := map[string]interface{}{
		"properties": properties,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", c.auth)
	req.Header.Set("Notion-Version", "2022-06-28")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.transport.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("notion error %s while updating page with properties %s", string(body), string(data))
	}

	return body, nil

}

type Schema struct {
	Properties map[string]interface{} `json:"properties"`
}

func (c *Client) GetSchema(dbid string) ([]string, error) {
	url := "https://api.notion.com/v1/databases/" + dbid

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", c.auth)
	req.Header.Set("Notion-Version", "2022-06-28")

	resp, err := c.transport.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("notion error while getting page: %s", string(body))
	}

	schema := Schema{}
	if err := json.Unmarshal(body, &schema); err != nil {
		return nil, err
	}

	res := []string{}
	for k := range schema.Properties {
		res = append(res, k)
	}
	return res, nil
}

func (c *Client) AppendPageContent(pageID string, content interface{}) ([]byte, error) {
	url := "https://api.notion.com/v1/blocks/" + pageID + "/children"

	reqBody := map[string]interface{}{
		"children": content,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", c.auth)
	req.Header.Set("Notion-Version", "2022-06-28")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.transport.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("notion error %s while appending content %s", string(body), string(data))
	}

	return body, nil
}

func (c *Client) AddCommentToPage(pageID string, text string) ([]byte, error) {
	url := "https://api.notion.com/v1/comments"

	reqBody := map[string]interface{}{
		"parent": map[string]string{
			"block_id": pageID,
		},
		"rich_text": []map[string]interface{}{
			{
				"type": "text",
				"text": map[string]string{
					"content": text,
				},
			},
		},
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", c.auth)
	req.Header.Set("Notion-Version", "2022-06-28")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.transport.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("notion error %s while adding comment to page %s", string(body), string(data))
	}

	return body, nil
}
