package main

import (
	"github.com/Syfaro/telegram-bot-api"
	"log"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"strconv"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "your-password"
	dbname   = "quiz"
)

var numericKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("1"),
		tgbotapi.NewKeyboardButton("2"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("3"),
		tgbotapi.NewKeyboardButton("4"),
	),
)

//var classes = [3]string{"PHP", "Python", "Go"}

func main() {
	classes := make(map[int]string)
	classes[0] = "PHP"
	classes[1] = "Python"
	classes[2] = "Go"

	bot, err := tgbotapi.NewBotAPI("TOKEN")

	psqlInfo := fmt.Sprintf("host=%s port=%d "+
		"dbname=%s sslmode=disable",
		host, port, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	checkErr(err)
	defer db.Close()

	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	var incomeText string
	var responseText string
	lastId := 0

	for update := range updates {
		if update.CallbackQuery != nil {
			user := getUserId(update, db)
			var answerNum int

			if user == 0 {
				insertUser(update, db)

				specialityId, _ := strconv.Atoi(update.CallbackQuery.Data)

				editMessage := tgbotapi.NewEditMessageText(
					update.CallbackQuery.Message.Chat.ID,
					update.CallbackQuery.Message.MessageID,
					"Выбраны вопросы по *"+classes[specialityId]+"*")
				editMessage.ParseMode = "markdown"

				bot.Send(editMessage)
			}
			answerNum = getAnswerNumber(update, db, user)

			if answerNum == 0 {

			}

			question, a1, a2, a3, a4 := getQuestionById(db, answerNum)

			var variants = make(map[int]string)
			variants[0] = a1
			variants[1] = a2
			variants[2] = a3
			variants[3] = a4

			keyboard := tgbotapi.InlineKeyboardMarkup{}
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, question)

			for key, value := range variants {
				var row []tgbotapi.InlineKeyboardButton
				btn := tgbotapi.NewInlineKeyboardButtonData(value, strconv.Itoa(key))
				row = append(row, btn)
				keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
			}
			msg.ReplyMarkup = keyboard

			bot.Send(msg)

			config := tgbotapi.CallbackConfig{
				CallbackQueryID: update.CallbackQuery.ID,
				Text:            "Done!",
				ShowAlert:       false,
			}
			bot.AnswerCallbackQuery(config)
		}

		if update.Message == nil {
			continue
		}

		incomeText = update.Message.Text

		if incomeText == "/start" {
			responseText = "Выбери свой любимый язык программирования"
		} else {
			responseText = incomeText
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, responseText)

		// NewInlineKeyboardButtonData
		keyboard := tgbotapi.InlineKeyboardMarkup{}

		for key, value := range classes {
			var row []tgbotapi.InlineKeyboardButton
			btn := tgbotapi.NewInlineKeyboardButtonData(value, strconv.Itoa(key))
			row = append(row, btn)
			keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
		}
		msg.ReplyMarkup = keyboard

		//if (questions) {
		//	msg.ReplyMarkup = numericKeyboard
		//} else {
		//	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		//}

		sm, _ := bot.Send(msg)
		lastId = sm.MessageID
		fmt.Println(lastId)
	}
}

func getQuestionById(db *sql.DB, i int) (text, a1, a2, a3, a4 string) {
	row := db.QueryRow("SELECT text, answer_1, answer_2, answer_3, answer_4  FROM questions LIMIT 1 OFFSET $1", i)
	err := row.Scan(&text, &a1, &a2, &a3, &a4)
	checkErr(err)
	return
}

func getAnswerNumber(update tgbotapi.Update, db *sql.DB, userId int) (count int) {
	row := db.QueryRow("SELECT COUNT(*) AS count FROM  answers WHERE user_id = $1", userId)
	err := row.Scan(&count)
	checkErr(err)
	return
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func getUserId(update tgbotapi.Update, db *sql.DB) (user int) {
	user = 0

	rows, err := db.Query("SELECT id FROM users WHERE tmid = $1", update.CallbackQuery.From.ID)

	checkErr(err)

	for rows.Next() {
		rows.Scan(&user)
	}

	return
}

func insertUser(update tgbotapi.Update, db *sql.DB) {
	var lastInsertId int
	err := db.QueryRow("INSERT INTO users(tmid,user_name,first_name,last_name, speciality_id) VALUES($1,$2,$3,$4,$5) RETURNING id;",
		update.CallbackQuery.From.ID,
		update.CallbackQuery.From.UserName,
		update.CallbackQuery.From.FirstName,
		update.CallbackQuery.From.LastName,
		update.CallbackQuery.Data,
	).Scan(&lastInsertId)
	checkErr(err)
	fmt.Println("last inserted id =", lastInsertId)
}
