package app

import (
	cfg "EskgTgBotGo/config"
	"EskgTgBotGo/service"
	"fmt"
	tg "gopkg.in/telebot.v3"
	"log"
	"strings"
	"time"
)

var config *cfg.Config

func Run() error {
	config = cfg.Get()

	pref := tg.Settings{
		Token:  config.Telegram.Token,
		Poller: &tg.LongPoller{Timeout: 10 * time.Second},
	}
	bot, _ := tg.NewBot(pref)

	bot.Handle(tg.OnText, OnTextHandle)
	bot.Start()

	return nil
}

func OnTextHandle(context tg.Context) error {
	msg := context.Text()
	log.Printf("<-- [%s]: %s", context.Sender().Username, msg)
	switch {
	// команда Get для получения
	case hasPrefixIgnoreCase(msg, "get"):
		{
			_ = context.Notify(tg.FindingLocation)
			time.Sleep(5 * time.Second)
			_ = context.Notify(tg.RecordingVideo)
			time.Sleep(5 * time.Second)
			_ = context.Notify(tg.ChoosingSticker)
			time.Sleep(5 * time.Second)

			data := service.GetGitWorkflowResult(config)
			data = formatJson(data)

			_ = context.Send(fmt.Sprintf(data), tg.ModeHTML)
		}
	// команда New для генерации
	case hasPrefixIgnoreCase(msg, "new"):
		{
			_ = context.Send(tg.Typing)
			count := getN(msg)
			data := service.TriggerWorkflow(count, config)
			err := context.Send(fmt.Sprintf(data))
			if err != nil {
				return nil
			}
		}
	}
	return nil
}

func formatJson(jsonData string) string {
	// Преобразуем JSON в строку и удаляем фигурные скобки и кавычки =^_^=
	formattedJSON := jsonData
	formattedJSON = strings.ReplaceAll(formattedJSON, "[", "")
	formattedJSON = strings.ReplaceAll(formattedJSON, "]", "")
	formattedJSON = strings.ReplaceAll(formattedJSON, "{", "")
	formattedJSON = strings.ReplaceAll(formattedJSON, "}", "")
	formattedJSON = strings.ReplaceAll(formattedJSON, "\"", "")
	formattedJSON = strings.ReplaceAll(formattedJSON, ",", "")

	lines := strings.Split(formattedJSON, "\n")
	var result []string

	for _, line := range lines {
		parts := strings.Split(line, ":")
		if len(parts) == 2 {
			if strings.EqualFold(strings.TrimSpace(parts[0]), "Key") {
				line = fmt.Sprintf("<b>%s</b>: <code>%s</code>", strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))

			} else {
				line = fmt.Sprintf("<b>%s</b>: %s", strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
			}
		}
		result = append(result, line)
	}
	finalFormattedJSON := strings.Join(result, "\n")

	return finalFormattedJSON
}

func getN(msg string) string {
	parts := strings.Split(msg, " ")
	if len(parts) > 1 {
		return parts[1]
	}
	return "1"
}

func hasPrefixIgnoreCase(s, prefix string) bool {
	return len(s) >= len(prefix) && strings.EqualFold(s[0:len(prefix)], prefix)
}
