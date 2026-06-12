package model

type TelegramRequest struct {
	UpdateID      int64                  `json:"update_id"`
	Message       *TelegramMessage       `json:"message,omitempty"`
	CallbackQuery *TelegramCallbackQuery `json:"callback_query,omitempty"`
}

type TelegramMessage struct {
	MessageID   int64               `json:"message_id"`
	From        TelegramUser        `json:"from"`
	Chat        TelegramChat        `json:"chat"`
	Text        string              `json:"text"`
	ReplyMarkup *TelegramInlineMenu `json:"reply_markup,omitempty"`
}

type TelegramChat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

type TelegramUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

type TelegramCallbackQuery struct {
	ID      string           `json:"id"`
	From    TelegramUser     `json:"from"`
	Message *TelegramMessage `json:"message,omitempty"`
	Data    string           `json:"data"`
}

// responses

type SendMessagePayload struct {
	ChatID      int64               `json:"chat_id"`
	Text        string              `json:"text"`
	ReplyMarkup *TelegramInlineMenu `json:"reply_markup,omitempty"`
}

type EditMessagePayload struct {
	ChatID    int64  `json:"chat_id"`
	MessageID int64  `json:"message_id"`
	Text      string `json:"text"`
}

type AnswerCallbackPayload struct {
	CallbackQueryID string `json:"callback_query_id"`
}

type TelegramInlineMenu struct {
	InlineKeyboard [][]TelegramInlineButton `json:"inline_keyboard"`
}

type TelegramInlineButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}
