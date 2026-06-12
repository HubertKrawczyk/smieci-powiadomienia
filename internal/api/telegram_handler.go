package api

import (
	"bytes"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"smieci-sms/internal/model"
	"smieci-sms/internal/repository"
	"smieci-sms/internal/service"
)

const telegramAPIBase = "https://api.telegram.org"

type TelegramHandler struct {
	repo           repository.UserRepository
	garbageService service.GarbageService
	secretToken    string
	botToken       string
}

type ConversationState int

const (
	StateNone ConversationState = iota
	StateAwaitingStreet
	StateAwaitingNumber
	StateAwaitingPostcode
	StateAwaitingLocationConfirmation
	StateAwaitingSchedule
)

type UserSession struct {
	State        ConversationState
	Street       string
	Number       string
	Postcode     string
	LocationID   string
	LocationName string
}

var (
	sessionMutex sync.Mutex // concurrency safety
	sessions     = make(map[int64]*UserSession)
)

func getSession(chatID int64) *UserSession {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	if _, exists := sessions[chatID]; !exists { // create new session if not exists
		sessions[chatID] = &UserSession{State: StateNone}
	}
	return sessions[chatID]
}

func NewTelegramHandler(repo repository.UserRepository, garbageService service.GarbageService, secretToken string, botToken string) *TelegramHandler {
	return &TelegramHandler{repo: repo, garbageService: garbageService, secretToken: secretToken, botToken: botToken}
}

func (h *TelegramHandler) Start(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	receivedToken := r.Header.Get("X-Telegram-Bot-Api-Secret-Token")
	log.Printf("New webhook request for start.")

	if subtle.ConstantTimeCompare([]byte(receivedToken), []byte(h.secretToken)) != 1 {
		log.Printf("Unauthorized webhook request blocked: invalid or missing secret token.")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var payload model.TelegramRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if payload.CallbackQuery != nil {
		cb := payload.CallbackQuery
		chatID := cb.Message.Chat.ID
		session := getSession(chatID)

		if session.State == StateAwaitingLocationConfirmation {
			// Acknowledge click immediately to clear loading icon on user screen
			h.sendCallbackAcknowledgment(cb.ID)

			if cb.Data == "loc_cancel" {
				session.State = StateNone
				h.sendTelegramMessage(chatID, "Registration canceled. Please type /start to try again.")
				return
			}

			// Extract your AddressPointID from the payload string
			// Format was: "loc_YOUR_ID"
			selectedLocationID := strings.TrimPrefix(cb.Data, "loc_")
			selectedLocationName := "Selected Location"
			if cb.Message.ReplyMarkup != nil {
				for _, row := range cb.Message.ReplyMarkup.InlineKeyboard {
					for _, btn := range row {
						if btn.CallbackData == cb.Data {
							selectedLocationName = btn.Text
							break
						}
					}
				}
			}

			// Save selected variant option to your persistent storage layer
			if err := h.repo.DeleteUserLocationByChatID(r.Context(), chatID); err != nil {
				log.Printf("WARNING: failed to delete existing user location: %v", err)
			}
			err := h.repo.SaveUserLocation(r.Context(), model.UserLocation{
				ChatID:      chatID,
				LocationID:  selectedLocationID,
				Name:        selectedLocationName, // Or fetch full record name mapping matching ID
				Phone:       "123456789",
				AddressName: session.Postcode,
			})

			if err != nil {
				log.Printf("ERROR: failed to save user location: %v", err)
				h.sendTelegramMessage(chatID, "Something went wrong saving your location option.")
			} else {
				// Edit original button grid text away into a plain confirmed message string
				confirmationMsg := fmt.Sprintf("Location choice confirmed (%s)! You have been successfully registered.", selectedLocationName)
				h.sendTelegramEditMessage(chatID, cb.Message.MessageID, confirmationMsg)
			}

			session.State = StateNone
			return
		}
	}

	if payload.Message != nil && payload.Message.Text == "/start" {
		chatID := payload.Message.Chat.ID
		fmt.Printf("User on Chat ID %d wants to START the process!\n", chatID)

		session := getSession(chatID)
		session.State = StateAwaitingStreet
		h.sendTelegramMessage(chatID, "Welcome! Please reply with your street name.")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		return
	}
	if payload.Message != nil {
		chatID := payload.Message.Chat.ID
		session := getSession(chatID)
		text := payload.Message.Text

		switch session.State {
		case StateAwaitingStreet:
			session.Street = text
			session.State = StateAwaitingNumber
			h.sendTelegramMessage(chatID, "Got it. What is the street number?")
		case StateAwaitingNumber:
			session.Number = text
			session.State = StateAwaitingPostcode
			h.sendTelegramMessage(chatID, "Thanks! What is your postcode?")
		case StateAwaitingPostcode:
			session.Postcode = text

			items, err := h.garbageService.GetLocationID(r.Context(), session.Street, session.Number, session.Postcode)
			if err != nil {
				session.State = StateNone
				h.sendTelegramMessage(chatID, "Error finding location ID.")
				return
			}
			// 1. Dynamically build the rows of inline buttons from your items array
			var buttons [][]model.TelegramInlineButton
			for _, item := range items {
				btn := model.TelegramInlineButton{
					Text:         item.FullName,
					CallbackData: fmt.Sprintf("loc_%s", item.AddressPointID), // Prefix prevents collision
				}
				buttons = append(buttons, []model.TelegramInlineButton{btn})
			}

			// Add a fallback cancel button at the bottom
			cancelBtn := model.TelegramInlineButton{Text: "❌ None of these", CallbackData: "loc_cancel"}
			buttons = append(buttons, []model.TelegramInlineButton{cancelBtn})

			inlineMenu := &model.TelegramInlineMenu{
				InlineKeyboard: buttons,
			}
			h.sendTelegramMessage(chatID,
				"Multiple locations found. Please select yours from the options below:",
				inlineMenu,
			)

			session.State = StateAwaitingLocationConfirmation
		case StateAwaitingLocationConfirmation:
			session.State = StateNone
			h.sendTelegramMessage(chatID, "Sorry, something went wrong. Please type /start to begin the registration process.")
		default:
			h.sendTelegramMessage(chatID, "Please type /start to begin the registration process.")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *TelegramHandler) sendTelegramMessage(chatID int64, text string, markup ...*model.TelegramInlineMenu) {
	apiURL := fmt.Sprintf("%s/bot%s/sendMessage", telegramAPIBase, h.botToken)

	payload := model.SendMessagePayload{
		ChatID: chatID,
		Text:   text,
	}
	if len(markup) > 0 && markup[0] != nil {
		payload.ReplyMarkup = markup[0]
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal message payload: %v", err)
		return
	}

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Failed to post message to Telegram API: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Telegram API returned non-OK status: %d", resp.StatusCode)
	}
}

// sendCallbackAcknowledgment tells Telegram to clear the button loading state
func (h *TelegramHandler) sendCallbackAcknowledgment(callbackQueryID string) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/answerCallbackQuery", h.botToken)

	payload := model.AnswerCallbackPayload{
		CallbackQueryID: callbackQueryID,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal callback acknowledgment: %v", err)
		return
	}

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Failed to post answerCallbackQuery: %v", err)
		return
	}
	defer resp.Body.Close()
}

// sendTelegramEditMessage swaps out the button markup with a plain string confirmation
func (h *TelegramHandler) sendTelegramEditMessage(chatID int64, messageID int64, text string) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/editMessageText", h.botToken)

	payload := model.EditMessagePayload{
		ChatID:    chatID,
		MessageID: messageID,
		Text:      text,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal edit message payload: %v", err)
		return
	}

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Failed to post editMessageText: %v", err)
		return
	}
	defer resp.Body.Close()
}
