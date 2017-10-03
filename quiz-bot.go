package main

import (
	"github.com/Syfaro/telegram-bot-api"
	"log"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"strconv"
	"time"
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

func main() {
	classes := make(map[int]string)
	classes[1] = "PHP"
	classes[2] = "Python"
	classes[3] = "Go"

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
	var userId, specialityId, answerNum int

	for update := range updates {
		var msg tgbotapi.MessageConfig

		if update.CallbackQuery != nil {
			tmId := update.CallbackQuery.From.ID

			if userId == 0 {
				insertUser(update, db)
			}

			userId, specialityId = getUserId(tmId, db)

			answerNum = getAnswerNumber(db, userId)
			question, a1, a2, a3, a4, finish := getQuestionById(db, answerNum, specialityId)

			if answerNum == 0 {
				specialityId, _ := strconv.Atoi(update.CallbackQuery.Data)

				editMessage := tgbotapi.NewEditMessageText(
					update.CallbackQuery.Message.Chat.ID,
					update.CallbackQuery.Message.MessageID,
					"Выбраны вопросы по *"+classes[specialityId]+"*")
				editMessage.ParseMode = "markdown"

				bot.Send(editMessage)
			}

			setAnswer(update, db, userId, answerNum)

			if !finish {
				var variants = make(map[int]string)
				variants[1] = a1
				variants[2] = a2
				variants[3] = a3
				variants[4] = a4

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
			} else {
				msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Вы прошли опрос, ваш результат N баллов")
				bot.Send(msg)
			}

			if answerNum != 0 {
				editMessage := tgbotapi.NewEditMessageText(
					update.CallbackQuery.Message.Chat.ID,
					update.CallbackQuery.Message.MessageID,
					"Ответ №*"+strconv.Itoa(answerNum)+"* записан")
				editMessage.ParseMode = "markdown"
				bot.Send(editMessage)
			}

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

		tmId := update.Message.From.ID
		userId, specialityId = getUserId(tmId, db)
		answerNum = getAnswerNumber(db, userId)

		_, _, _, _, _, finish := getQuestionById(db, answerNum, specialityId)

		if finish {
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Вы прошли опрос, ваш результат N баллов")
		}

		if userId == 0 {
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, responseText)

			keyboard := tgbotapi.InlineKeyboardMarkup{}

			for key, value := range classes {
				var row []tgbotapi.InlineKeyboardButton
				btn := tgbotapi.NewInlineKeyboardButtonData(value, strconv.Itoa(key))
				row = append(row, btn)
				keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
			}
			msg.ReplyMarkup = keyboard
		}

		bot.Send(msg)
	}
}

func setAnswer(update tgbotapi.Update, db *sql.DB, userId int, answerId int) (lastInsertId int) {
	err := db.QueryRow("INSERT INTO answers(user_id, speciality_id, question_id, text, created_at) VALUES($1,$2,$3,$4,$5) RETURNING id;",
		userId,
		1,
		answerId,
		update.CallbackQuery.Data,
		time.Now().Format(time.RFC3339),
	).Scan(&lastInsertId)

	checkErr(err)
	return
}

func getQuestionById(db *sql.DB, answerId, specialityId int) (text, a1, a2, a3, a4 string, finish bool) {
	err := db.QueryRow("SELECT text, answer_1, answer_2, answer_3, answer_4  FROM questions WHERE speciality_id = $1 ORDER BY id ASC LIMIT 1 OFFSET $2", specialityId, answerId).Scan(&text, &a1, &a2, &a3, &a4)

	if err == nil {
		finish = false
	} else if err == sql.ErrNoRows {
		finish = true
	} else {
		panic(err)
	}
	return
}

func getAnswerNumber(db *sql.DB, userId int) (count int) {
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

func getUserId(tmid int, db *sql.DB) (user, specialityId int) {
	user = 0

	rows, err := db.Query("SELECT id, speciality_id FROM users WHERE tmid = $1", tmid)

	checkErr(err)

	for rows.Next() {
		rows.Scan(&user, &specialityId)
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
}
