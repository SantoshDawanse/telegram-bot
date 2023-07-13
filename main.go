package main

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sashabaranov/go-openai"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	telegramToken := os.Getenv("TELEGRAM_API_TOKEN")
	placeholderText := os.Getenv("placeholder")
	if placeholderText == "" {
		placeholderText = "Typing ..."
	}
	systemPrompt := os.Getenv("system_prompt")
	if systemPrompt == "" {
		systemPrompt = "You are a helpful assistant answering questions on Telegram."
	}
	helpMessage := os.Getenv("help_mesg")
	if helpMessage == "" {
		helpMessage = "I am your assistant on Telegram. Ask me any question! To start a new conversation, type the /restart command."
	}

	bot, err := tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message != nil {
			chatID := update.Message.Chat.ID
			log.Printf("Received message from %d", chatID)

			openaiClient := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

			text := update.Message.Text
			if strings.EqualFold(text, "/help") {
				msg := tgbotapi.NewMessage(chatID, helpMessage)
				_, _ = bot.Send(msg)
			} else if strings.EqualFold(text, "/start") {
				msg := tgbotapi.NewMessage(chatID, helpMessage)
				_, _ = bot.Send(msg)
				// Set conversation state to true
				_ = setConversationState(chatID, true)
				log.Printf("Started conversation for %d", chatID)
			} else if strings.EqualFold(text, "/restart") {
				msg := tgbotapi.NewMessage(chatID, "Ok, I am starting a new conversation.")
				_, _ = bot.Send(msg)
				// Set conversation state to true
				_ = setConversationState(chatID, true)
				log.Printf("Restarted conversation for %d", chatID)
			} else {
				placeholderMsg := tgbotapi.NewMessage(chatID, placeholderText)
				placeholder, err := bot.Send(placeholderMsg)
				if err != nil {
					log.Printf("Error sending message to Telegram: %v", err)
				}

				restart := getConversationState(chatID)
				if restart {
					log.Printf("Detected restart = true")
					_ = setConversationState(chatID, false)
				}

				msg, err := openaiClient.CreateChatCompletion(
					context.Background(),
					openai.ChatCompletionRequest{
						Model: openai.GPT3Dot5Turbo,
						Messages: []openai.ChatCompletionMessage{
							{
								Role:    "system",
								Content: systemPrompt,
								//Name:         "",
								//FunctionCall: nil,
							},
							{
								Role:    "user",
								Content: text,
								//Name:         "",
								//FunctionCall: nil,
							},
						},
						MaxTokens:        100,
						Temperature:      0.6,
						TopP:             0,
						N:                0,
						Stream:           false,
						Stop:             []string{"\n"},
						PresencePenalty:  0,
						FrequencyPenalty: 0,
						LogitBias:        nil,
						User:             "",
						Functions:        nil,
						FunctionCall:     nil,
					},
					//&openai.ChatParams{
					//	Messages: []*openai.MessageParams{
					//		&openai.MessageParams{
					//			Role:    "system",
					//			Content: systemPrompt,
					//		},
					//		&openai.MessageParams{
					//			Role:    "user",
					//			Content: text,
					//		},
					//	},
					//	ConversationID: chatIDToString(chatID),
					//	Completion: &openai.CompletionParams{
					//		MaxTokens:   openai.Int(100),
					//		Temperature: openai.Float64(0.6),
					//		Stop:        []string{"\n"},
					//	},
					//},
				)

				if err != nil {
					msg := tgbotapi.NewEditMessageText(chatID, placeholder.MessageID, "Sorry, an error has occurred. Please try again later!")
					_, _ = bot.Send(msg)
					log.Printf("OpenAI returns error: %v", err)
				} else {
					msg := tgbotapi.NewEditMessageText(chatID, placeholder.MessageID, msg.Choices[0].Message.Content)
					_, _ = bot.Send(msg)
				}
			}
		}
	}
}

func setConversationState(chatID int64, state bool) error {
	// Implement your set conversation state logic here
	return nil
}

func getConversationState(chatID int64) bool {
	// Implement your get conversation state logic here
	return false
}

func chatIDToString(chatID int64) string {
	// Convert chat ID to string
	return strconv.FormatInt(chatID, 10)
}
