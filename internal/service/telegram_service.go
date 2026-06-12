package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"smieci-sms/internal/model"
)

const telegramAPIBase = "https://api.telegram.org"

type TelegramService interface {
	SendMessage(ctx context.Context, chatID int64, text string, markup *model.TelegramInlineMenu) error
	AnswerCallbackQuery(ctx context.Context, callbackQueryID string) error
	EditMessageText(ctx context.Context, chatID int64, messageID int64, text string) error
	EditMessageReplyMarkup(ctx context.Context, chatID int64, messageID int64, markup *model.TelegramInlineMenu) error
}

type telegramService struct {
	botToken string
}

func NewTelegramService(botToken string) TelegramService {
	return &telegramService{botToken: botToken}
}

func (s *telegramService) SendMessage(ctx context.Context, chatID int64, text string, markup *model.TelegramInlineMenu) error {
	apiURL := fmt.Sprintf("%s/bot%s/sendMessage", telegramAPIBase, s.botToken)

	payload := model.SendMessagePayload{
		ChatID: chatID,
		Text:   text,
	}
	if markup != nil {
		payload.ReplyMarkup = markup
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal message payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to post message to Telegram API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned non-OK status: %d", resp.StatusCode)
	}
	return nil
}

func (s *telegramService) AnswerCallbackQuery(ctx context.Context, callbackQueryID string) error {
	apiURL := fmt.Sprintf("%s/bot%s/answerCallbackQuery", telegramAPIBase, s.botToken)

	payload := model.AnswerCallbackPayload{
		CallbackQueryID: callbackQueryID,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal callback acknowledgment: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to post answerCallbackQuery: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned non-OK status: %d", resp.StatusCode)
	}
	return nil
}

func (s *telegramService) EditMessageText(ctx context.Context, chatID int64, messageID int64, text string) error {
	apiURL := fmt.Sprintf("%s/bot%s/editMessageText", telegramAPIBase, s.botToken)

	payload := model.EditMessagePayload{
		ChatID:    chatID,
		MessageID: messageID,
		Text:      text,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal edit message payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to post editMessageText: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned non-OK status: %d", resp.StatusCode)
	}
	return nil
}

func (s *telegramService) EditMessageReplyMarkup(ctx context.Context, chatID int64, messageID int64, markup *model.TelegramInlineMenu) error {
	apiURL := fmt.Sprintf("%s/bot%s/editMessageReplyMarkup", telegramAPIBase, s.botToken)

	payload := struct {
		ChatID      int64                     `json:"chat_id"`
		MessageID   int64                     `json:"message_id"`
		ReplyMarkup *model.TelegramInlineMenu `json:"reply_markup"`
	}{
		ChatID:      chatID,
		MessageID:   messageID,
		ReplyMarkup: markup,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal edit reply markup payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to post editMessageReplyMarkup: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned non-OK status: %d", resp.StatusCode)
	}
	return nil
}
