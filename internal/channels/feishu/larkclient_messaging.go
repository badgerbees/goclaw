package feishu

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"strconv"
)

// --- IM API: Messages ---

type SendMessageResp struct {
	MessageID string `json:"message_id"`
}

func (c *LarkClient) SendMessage(ctx context.Context, receiveIDType, receiveID, msgType, content string) (*SendMessageResp, error) {
	path := "/open-apis/im/v1/messages?receive_id_type=" + receiveIDType
	body := map[string]string{
		"receive_id": receiveID,
		"msg_type":   msgType,
		"content":    content,
	}
	resp, err := c.doJSON(ctx, "POST", path, body)
	if err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("send message: code=%d msg=%s", resp.Code, resp.Msg)
	}
	var data SendMessageResp
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal SendMessageResp: %w", err)
	}
	return &data, nil
}

type MessageItem struct {
	MessageID   string `json:"message_id"`
	RootID      string `json:"root_id"`
	ParentID    string `json:"parent_id"`
	ThreadID    string `json:"thread_id"`
	MsgType     string `json:"msg_type"`
	CreateTime  string `json:"create_time"`
	UpdateTime  string `json:"update_time"`
	Deleted     bool   `json:"deleted"`
	Updated     bool   `json:"updated"`
	ChatID      string `json:"chat_id"`
	Sender      struct {
		ID         string `json:"id"`
		IDType     string `json:"id_type"`
		SenderType string `json:"sender_type"`
		TenantKey  string `json:"tenant_key"`
	} `json:"sender"`
	Body struct {
		Content string `json:"content"`
	} `json:"body"`
	Mentions []struct {
		Key string `json:"key"`
		ID  string `json:"id"`
	} `json:"mentions"`
}

func (c *LarkClient) GetMessage(ctx context.Context, messageID string) (*MessageItem, error) {
	path := "/open-apis/im/v1/messages/" + messageID
	resp, err := c.doJSON(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("get message: code=%d msg=%s", resp.Code, resp.Msg)
	}
	var data struct {
		Items []MessageItem `json:"items"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal MessageItem: %w", err)
	}
	if len(data.Items) == 0 {
		return nil, fmt.Errorf("message %s not found", messageID)
	}
	return &data.Items[0], nil
}

func (c *LarkClient) ListMessages(ctx context.Context, containerIDType, containerID string, pageSize int, pageToken string) ([]MessageItem, string, error) {
	path := fmt.Sprintf("/open-apis/im/v1/messages?container_id_type=%s&container_id=%s",
		url.QueryEscape(containerIDType), url.QueryEscape(containerID))
	if pageSize > 0 {
		path += fmt.Sprintf("&page_size=%d", pageSize)
	}
	if pageToken != "" {
		path += "&page_token=" + pageToken
	}

	resp, err := c.doJSON(ctx, "GET", path, nil)
	if err != nil {
		return nil, "", err
	}
	if resp.Code != 0 {
		return nil, "", fmt.Errorf("list messages: code=%d msg=%s", resp.Code, resp.Msg)
	}
	var data struct {
		Items     []MessageItem `json:"items"`
		PageToken string        `json:"page_token"`
		HasMore   bool          `json:"has_more"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, "", fmt.Errorf("unmarshal ListMessages response: %w", err)
	}
	return data.Items, data.PageToken, nil
}

// --- IM API: Images ---

func (c *LarkClient) DownloadImage(ctx context.Context, imageKey string) ([]byte, error) {
	path := "/open-apis/im/v1/images/" + imageKey
	data, _, err := c.doDownload(ctx, path)
	return data, err
}

func (c *LarkClient) UploadImage(ctx context.Context, data io.Reader) (string, error) {
	resp, err := c.doMultipart(ctx, "/open-apis/im/v1/images",
		map[string]string{"image_type": "message"},
		"image", data, "image.png")
	if err != nil {
		return "", err
	}
	if resp.Code != 0 {
		return "", fmt.Errorf("upload image: code=%d msg=%s", resp.Code, resp.Msg)
	}
	var result struct {
		ImageKey string `json:"image_key"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return "", fmt.Errorf("unmarshal UploadImage response: %w", err)
	}
	return result.ImageKey, nil
}

// --- IM API: Files ---

func (c *LarkClient) UploadFile(ctx context.Context, data io.Reader, fileName, fileType string, durationMs int) (string, error) {
	fields := map[string]string{
		"file_type": fileType,
		"file_name": fileName,
	}
	if durationMs > 0 {
		fields["duration"] = strconv.Itoa(durationMs)
	}
	resp, err := c.doMultipart(ctx, "/open-apis/im/v1/files", fields, "file", data, fileName)
	if err != nil {
		return "", err
	}
	if resp.Code != 0 {
		return "", fmt.Errorf("upload file: code=%d msg=%s", resp.Code, resp.Msg)
	}
	var result struct {
		FileKey string `json:"file_key"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return "", fmt.Errorf("unmarshal UploadFile response: %w", err)
	}
	return result.FileKey, nil
}

// --- IM API: Get Message ---


// --- IM API: Message Resources ---

func (c *LarkClient) DownloadMessageResource(ctx context.Context, messageID, fileKey, resourceType string) ([]byte, string, error) {
	path := fmt.Sprintf("/open-apis/im/v1/messages/%s/resources/%s?type=%s", messageID, fileKey, resourceType)
	return c.doDownload(ctx, path)
}

// --- CardKit API ---

func (c *LarkClient) CreateCard(ctx context.Context, cardType, data string) (string, error) {
	resp, err := c.doJSON(ctx, "POST", "/open-apis/cardkit/v1/cards", map[string]string{
		"type": cardType,
		"data": data,
	})
	if err != nil {
		return "", err
	}
	if resp.Code != 0 {
		return "", fmt.Errorf("create card: code=%d msg=%s", resp.Code, resp.Msg)
	}
	var result struct {
		CardID string `json:"card_id"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return "", fmt.Errorf("unmarshal CreateCard response: %w", err)
	}
	return result.CardID, nil
}

func (c *LarkClient) UpdateCardSettings(ctx context.Context, cardID, settings string, seq int, uuid string) error {
	path := "/open-apis/cardkit/v1/cards/" + cardID
	resp, err := c.doJSON(ctx, "PATCH", path, map[string]any{
		"settings": settings,
		"sequence": seq,
		"uuid":     uuid,
	})
	if err != nil {
		return err
	}
	if resp.Code != 0 {
		return fmt.Errorf("update card settings: code=%d msg=%s", resp.Code, resp.Msg)
	}
	return nil
}

func (c *LarkClient) UpdateCardElement(ctx context.Context, cardID, elementID, content string, seq int, uuid string) error {
	path := fmt.Sprintf("/open-apis/cardkit/v1/cards/%s/elements/%s", cardID, elementID)
	resp, err := c.doJSON(ctx, "PATCH", path, map[string]any{
		"content":  content,
		"sequence": seq,
		"uuid":     uuid,
	})
	if err != nil {
		return err
	}
	if resp.Code != 0 {
		slog.Debug("lark update card element failed", "code", resp.Code, "msg", resp.Msg)
		return fmt.Errorf("update card element: code=%d msg=%s", resp.Code, resp.Msg)
	}
	return nil
}

// --- IM API: Reactions ---

// AddMessageReaction adds an emoji reaction to a message.
// Returns the reaction_id for later removal. emojiType: e.g. "Typing", "THUMBSUP".
// Lark API: POST /open-apis/im/v1/messages/{message_id}/reactions
func (c *LarkClient) AddMessageReaction(ctx context.Context, messageID, emojiType string) (string, error) {
	path := fmt.Sprintf("/open-apis/im/v1/messages/%s/reactions", messageID)
	body := map[string]any{
		"reaction_type": map[string]string{
			"emoji_type": emojiType,
		},
	}
	resp, err := c.doJSON(ctx, "POST", path, body)
	if err != nil {
		return "", err
	}
	if resp.Code != 0 {
		return "", fmt.Errorf("add reaction: code=%d msg=%s", resp.Code, resp.Msg)
	}
	var result struct {
		ReactionID string `json:"reaction_id"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return "", fmt.Errorf("unmarshal AddMessageReaction response: %w", err)
	}
	return result.ReactionID, nil
}

// DeleteMessageReaction removes a reaction from a message.
// Lark API: DELETE /open-apis/im/v1/messages/{message_id}/reactions/{reaction_id}
func (c *LarkClient) DeleteMessageReaction(ctx context.Context, messageID, reactionID string) error {
	path := fmt.Sprintf("/open-apis/im/v1/messages/%s/reactions/%s", messageID, reactionID)
	resp, err := c.doJSON(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}
	if resp.Code != 0 {
		return fmt.Errorf("delete reaction: code=%d msg=%s", resp.Code, resp.Msg)
	}
	return nil
}

// --- Bot API ---

// GetBotInfo fetches the bot's identity from /open-apis/bot/v3/info.
// Returns the bot's open_id which is needed for mention detection in groups.
func (c *LarkClient) GetBotInfo(ctx context.Context) (string, error) {
	resp, err := c.doJSON(ctx, "GET", "/open-apis/bot/v3/info", nil)
	if err != nil {
		return "", err
	}
	if resp.Code != 0 {
		return "", fmt.Errorf("get bot info: code=%d msg=%s", resp.Code, resp.Msg)
	}
	var result struct {
		Bot struct {
			OpenID string `json:"open_id"`
		} `json:"bot"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return "", fmt.Errorf("unmarshal GetBotInfo response: %w", err)
	}
	return result.Bot.OpenID, nil
}

// --- Contact API ---

func (c *LarkClient) GetUser(ctx context.Context, userID, userIDType string) (string, error) {
	path := fmt.Sprintf("/open-apis/contact/v3/users/%s?user_id_type=%s", userID, userIDType)
	resp, err := c.doJSON(ctx, "GET", path, nil)
	if err != nil {
		return "", err
	}
	if resp.Code != 0 {
		return "", fmt.Errorf("get user: code=%d msg=%s", resp.Code, resp.Msg)
	}
	var result struct {
		User struct {
			Name string `json:"name"`
		} `json:"user"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return "", fmt.Errorf("unmarshal GetUser response: %w", err)
	}
	return result.User.Name, nil
}
