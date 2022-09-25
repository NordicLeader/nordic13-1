
git clone АДРЕС_РЕПОЗИТОРИЯ

//праллельные вседенные
git branch - посмотреть все ветки
git branch НАЗВАНИЕ_НОВОЙ_ВЕТКИ - создать новую ветку клон текущей
git checkout НАЗВАНИЕ_ВЕТКИ - перейти на эту ветку

//точки на оси времени
git add * - подготовить к сохранению
git commit -m "любой комментарий" - сохранить


package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type BigData struct {
	Ok     bool `json:"ok"`
	Result []struct {
		UpdateID int `json:"update_id"`
		Message  struct {
			MessageID int `json:"message_id"`
			From      struct {
				ID           int    `json:"id"`
				IsBot        bool   `json:"is_bot"`
				FirstName    string `json:"first_name"`
				Username     string `json:"username"`
				LanguageCode string `json:"language_code"`
			} `json:"from"`
			Chat struct {
				ID        int    `json:"id"`
				FirstName string `json:"first_name"`
				Username  string `json:"username"`
				Type      string `json:"type"`
			} `json:"chat"`
			Date int    `json:"date"`
			Text string `json:"text"`
		} `json:"message"`
	} `json:"result"`
}

func main() {

	//подключаемся к бд
	db, _ := sql.Open("mysql", "root:nordic123@tcp(database:3306)/bot_data")

	//создаем массив под id юзеров bdfggdfgfdgfdgfdfdgfdgfdgdffdfdfdgfdfgfd
	users := []int{}

	//offset
	offset := 0

	//в цикле отправляем запросы для получения новых сообщений и записи их в бд
	for range time.Tick(time.Second) {

		//отправляем запрос и сохраняем данные полученные из него в перемнную
		resp, _ := http.Get("https://api.telegram.org/bot5327059939:AAGr9otM_gS8FWzzuuHePa93zhnSJPCSnqg/getUpdates?offset=" + strconv.Itoa(offset))

		//получаем тело запроса
		bodyBytes, _ := ioutil.ReadAll(resp.Body)

		//создаем структуру под все сообщения
		data := BigData{}

		//переливаем данные из тела запроса в структуру
		json.Unmarshal(bodyBytes, &data)

		//в цикле бежим по всем сообщениям
		for i := 0; i < len(data.Result); i++ {

			//смотрим id юзера в элементе массива
			userId := data.Result[i].Message.From.ID
			cTime := data.Result[i].Message.Date
			message := data.Result[i].Message.Text
			username := data.Result[i].Message.From.Username
			firstName := data.Result[i].Message.From.FirstName

			//проверяем есть ли он в списке записанных в кэш юзеров
			if !inArray(userId, users) {

				//если его там нет то делаем запись в базу данных
				addUser(userId, cTime, username, firstName, db)

				//добавляем юзера в кэш
				users = append(users, userId)

			}

			//записываем в базу новое сообщение
			addMessage(userId, cTime, message, db)

			//обновляем значение offset
			offset = data.Result[i].UpdateID + 1

			//если есть ключевое слово для рассылки то рассылаем
			if strings.Contains(message, "/send ") {

				fmt.Println("нашли команду")

				//в цикле рассылаем сообщения
				for j := 0; j < len(users); j++ {

					fmt.Println("нашли юзера")

					//отправляем в отдельных потоках
					go sendMessage(users[j], message)
				}

			}
		}
	}
}

//функция для определения есть ли элемент в массиве
func inArray(needle int, haystack []int) bool {

	//смотрим каждый элемент массива
	for i := 0; i < len(haystack); i++ {

		//сравниваем его значенеи с тем что ищем
		if haystack[i] == needle {
			return true
		}
	}

	return false

}

func sendMessage(userId int, message string) {

	fmt.Println("пробуем отправить сообщение")
	fmt.Println(userId, message)
	http.Get("https://api.telegram.org/bot5327059939:AAGr9otM_gS8FWzzuuHePa93zhnSJPCSnqg/sendMessage?chat_id=" + strconv.Itoa(userId) + "&text=" + message)
}

//функция для добаления сообщения в бд
func addMessage(userId int, time int, text string, db *sql.DB) {
	//отправляем запрос
	db.Exec("INSERT INTO `messages`(`time`, `content`, `user_id`) VALUES(?, ?, ?)", time, text, userId)
}

//функция для добаления сообщения в бд
func addUser(userId int, time int, username string, firstName string, db *sql.DB) {
	//отправляем запрос
	db.Exec("INSERT INTO `users`(`registration_time`, `username`, `id`,`first_name`) VALUES(?, ?, ?, ?)", time, username, userId, firstName)
}
