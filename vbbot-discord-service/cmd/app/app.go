package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"os"
	"syscall"
	"vb-bot/internal/config"
	mysheets "vb-bot/internal/sheets"
	"vb-bot/internal/vb"
	"vb-bot/pkg/logging"
	"vb-bot/pkg/shutdown"
	"vb-bot/pkg/utilities"
)

func main() {
	logging.Init()
	logger := logging.GetLogger()
	logger.Println("logger inited")

	logger.Println("config initing")
	cfg := config.GetConfig()

	logger.Println("bot session initing")
	bot, _ := vb.NewBotSession(cfg.Bot, logger)

	logger.Println("sheet session initing")
	// TODO firstly load from sheets server to sync data
	sheetData, _ := mysheets.NewSpreadsheetDataFromJSON(cfg.Filepath.SpreadsheetDataPath, logger)
	logger.Println("Loaded data from json: %v", sheetData)
	sheet, _ := mysheets.NewSheet(sheetData)

	logger.Println("start bot")
	start(bot, sheet, cfg)

	select {}
	return
}

// start app here
func start(bot vb.Bot, sheet mysheets.SheetCallWrapper, config *config.Config) {
	bot.Session.AddHandler(func(session *discordgo.Session, evt *discordgo.MessageCreate) {
		bot.Logger.Println("Message create event")
		if bot.ID == evt.Author.ID {
			bot.Logger.Println("handled bot msg, return")
			return
		}

		log := mysheets.LogMessage{
			Number:     sheet.Data.LastEntryId,
			UserID:     evt.Author.Username,
			Time:       evt.Timestamp.Format("2006-01-02 15:04:05"),
			ContentMsg: evt.Content,
			ImageURL:   "",
		}
		bot.Logger.Println("Log struct: %v", log)

		if len(evt.Attachments) > 0 {
			log.ImageURL = "=IMAGE(\"" + evt.Attachments[0].URL + "\"" + ";4;100;100)"
			sheet.SetColumnWidth(4, 100)
			sheet.SetRowHeight(log.Number, 100)

			for _, v := range evt.Attachments {
				filename := config.Filepath.DownloadsPath + v.Filename
				utilities.DownloadFile(filename, v.URL)
				fileToSend, _ := os.Open(filename)
				session.ChannelFileSend(evt.ChannelID, v.Filename, fileToSend)
			}
			session.ChannelMessageSend(evt.ChannelID, "All files send!")
		}

		sheet.AddValuesUpdateRequest("USER_ENTERED",
			fmt.Sprintf("English!A%d:E%d", log.Number+1, log.Number+1),
			[][]interface{}{{log.Number}, {log.UserID}, {log.Time}, {log.ContentMsg}, {log.ImageURL}},
			"COLUMNS")
		sheet.RunRequests()

		sheet.Data.LastEntryId++
		session.ChannelMessageSend(evt.ChannelID, "Done")
		bot.Logger.Println("Message handle done")
	})

	err := bot.Session.Open()
	if err != nil {
		bot.Logger.Println(fmt.Sprintf("Error in opening bot session"))
		return
	}

	go shutdown.Graceful([]os.Signal{syscall.SIGABRT, syscall.SIGQUIT, syscall.SIGHUP, os.Interrupt, syscall.SIGTERM}, bot.Session, &sheet.Data)

	bot.Logger.Println("bot inited and started")
}
