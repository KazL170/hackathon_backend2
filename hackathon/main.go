package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/oklog/ulid/v2"
	_ "github.com/oklog/ulid/v2"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type CatalogResForHTTP struct {
	Id                  string `json:"id"`
	Name                string `json:"name"`
	Item_category       string `json:"item_category"`
	Curriculum_category string `json:"curriculum_category"`
	Detail              string `json:"detail"`
	URL1                string `json:"url"`
	Update_time         string `json:"update_time"`
}

type responseMessage struct {
	Id string `json:"id"`
}

// ① GoプログラムからMySQLへ接続
var db *sql.DB

func init() {
	// ①-1
	//err := godotenv.Load("mysql/.env_mysql")
	//if err != nil {
	//	fmt.Print("環境変数の読み込みに失敗")
	//}
	mysqlUser := os.Getenv("MYSQL_USER")
	mysqlUserPwd := os.Getenv("MYSQL_PWD")
	mysqlDatabase := os.Getenv("MYSQL_DATABASE")
	mysqlHost := os.Getenv("MYSQL_HOST")
	// ①-2
	_db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@%s/%s", mysqlUser, mysqlUserPwd, mysqlHost, mysqlDatabase))
	if err != nil {
		log.Fatalf("fail: sql.Open, %v\n", err)
	}
	// ①-3
	if err := _db.Ping(); err != nil {
		log.Fatalf("fail: _db.Ping, %v\n", err)
	}
	db = _db
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	switch r.Method {
	//case http.MethodGet:
	//	// ②-1 nameが空だった場合のエラー処理を施しています
	//	name := r.URL.Query().Get("name") // To be filled
	//	if name == " " {
	//		log.Println("fail: name is empty")
	//		w.WriteHeader(http.StatusBadRequest)
	//		return
	//	}
	//
	//	// ②-2 GETリクエストのクエリパラメータから条件を満たすデータを取得します。
	//	rows, err := db.Query("SELECT id, name, age FROM user WHERE name=?", name)
	//	if err != nil {
	//		log.Printf("fail: db.Query, %v\n", err)
	//		w.WriteHeader(http.StatusInternalServerError)
	//		return
	//	}
	//
	//	// ②-3
	//	users := make([]UserResForHTTPGet, 0)
	//	for rows.Next() {
	//		var u UserResForHTTPGet
	//		if err := rows.Scan(&u.Id, &u.Name, &u.Age); err != nil {
	//			log.Printf("fail: rows.Scan, %v\n", err)
	//
	//			if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
	//				log.Printf("fail: rows.Close(), %v\n", err)
	//			}
	//			w.WriteHeader(http.StatusInternalServerError)
	//			return
	//		}
	//		users = append(users, u)
	//	}
	//
	//	// ②-4 モジュール encoding/json を用いてレスポンス用ユーザースライスをJSONへ変換し、HTTPレスポンスボディに書き込みます。
	//	bytes, err := json.Marshal(users)
	//	if err != nil {
	//		log.Printf("fail: json.Marshal, %v\n", err)
	//		w.WriteHeader(http.StatusInternalServerError)
	//		return
	//	}
	//	w.Header().Set("Content-Type", "application/json")
	//
	//	w.Write(bytes)

	case http.MethodPost:
		//① idを採番
		var user CatalogResForHTTP
		user.Id = ulid.Make().String() //GetNewULIDString()
		user.Update_time = time.Now().String()
		//② SQLにid, name, ageをinsert, 何らかのエラーにはinternal server error(500)
		json.NewDecoder(r.Body).Decode(&user)
		id := user.Id
		name := user.Name
		item_category := user.Item_category
		curriculum_category := user.Curriculum_category
		detail := user.Detail
		url1 := user.URL1
		update_time := user.Update_time

		//nameが空の時BadRequest400
		if name == " " {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		tx, err := db.Begin()

		_, newErr := tx.Exec("INSERT INTO catalog(id, name,item_category, curriculum_category, detail, URL, update_time) VALUES(?,?,?,?,?,?,?)",
			id, name, item_category, curriculum_category, detail, url1, update_time)
		if newErr != nil {
			tx.Rollback()
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Print("error1")
			return
		}
		fmt.Println("Insert", user)

		if err := tx.Commit(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Print("error2")
			return
		}

		//③ insertができたものはstatus code=200と,jsonでidを返す
		bytes, err := json.Marshal(responseMessage{
			Id: user.Id,
		})
		if err != nil {
			log.Printf("fail: json.Marshal, %v\n", err)
			return
		}
		if err == nil {
			w.WriteHeader(http.StatusOK)
			w.Write(bytes)
			return
		}

	default:
		log.Printf("fail: HTTP Method is %s\n", r.Method)
		//	log.Printf("error")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

}

func handler2(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ②-2 GETリクエストのクエリパラメータから条件を満たすデータを取得します。
	rows, err := db.Query("SELECT * FROM catalog  order by update_time")
	if err != nil {
		log.Printf("fail: db.Query, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// ②-3
	users := make([]CatalogResForHTTP, 0)
	for rows.Next() {
		var u CatalogResForHTTP
		if err := rows.Scan(&u.Id, &u.Name, &u.Item_category, &u.Curriculum_category,
			&u.Detail, &u.URL1, &u.Update_time); err != nil {
			log.Printf("fail: rows.Scan, %v\n", err)

			if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
				log.Printf("fail: rows.Close(), %v\n", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	// ②-4 モジュール encoding/json を用いてレスポンス用ユーザースライスをJSONへ変換し、HTTPレスポンスボディに書き込みます。
	bytes, err := json.Marshal(users)
	if err != nil {
		log.Printf("fail: json.Marshal, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func handler3(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	switch r.Method {

	case http.MethodPost:
		//① idを採番
		var user CatalogResForHTTP
		user.Id = ulid.Make().String() //GetNewULIDString()
		user.Update_time = time.Now().String()
		//② SQLにid, name, ageをinsert, 何らかのエラーにはinternal server error(500)
		json.NewDecoder(r.Body).Decode(&user)
		name := user.Name

		//nameが空の時BadRequest400
		if name == " " {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		tx, err := db.Begin()

		_, newErr := tx.Exec("DELETE FROM catalog WHERE name=?", name)
		if newErr != nil {
			tx.Rollback()
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Println("Delete", user)

		if err := tx.Commit(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		//③ insertができたものはstatus code=200と,jsonでidを返す
		bytes, err := json.Marshal(responseMessage{
			Id: user.Id,
		})
		if err != nil {
			log.Printf("fail: json.Marshal, %v\n", err)
			return
		}
		if err == nil {
			w.WriteHeader(http.StatusOK)
			w.Write(bytes)
			return
		}

		//default:
		//	log.Printf("fail: HTTP Method is %s\n", r.Method)
		//	w.WriteHeader(http.StatusBadRequest)
		//	return
	}
}

func handler4(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ②-2 GETリクエストのクエリパラメータから条件を満たすデータを取得します。
	rows, err := db.Query("SELECT * FROM catalog WHERE item_category ='本' order by update_time")
	if err != nil {
		log.Printf("fail: db.Query, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// ②-3
	users := make([]CatalogResForHTTP, 0)
	for rows.Next() {
		var u CatalogResForHTTP
		if err := rows.Scan(&u.Id, &u.Name, &u.Item_category, &u.Curriculum_category,
			&u.Detail, &u.URL1, &u.Update_time); err != nil {
			log.Printf("fail: rows.Scan, %v\n", err)

			if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
				log.Printf("fail: rows.Close(), %v\n", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	// ②-4 モジュール encoding/json を用いてレスポンス用ユーザースライスをJSONへ変換し、HTTPレスポンスボディに書き込みます。
	bytes, err := json.Marshal(users)
	if err != nil {
		log.Printf("fail: json.Marshal, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func handler5(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ②-2 GETリクエストのクエリパラメータから条件を満たすデータを取得します。
	rows, err := db.Query("SELECT * FROM catalog WHERE item_category ='ブログ' order by update_time")
	if err != nil {
		log.Printf("fail: db.Query, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// ②-3
	users := make([]CatalogResForHTTP, 0)
	for rows.Next() {
		var u CatalogResForHTTP
		if err := rows.Scan(&u.Id, &u.Name, &u.Item_category, &u.Curriculum_category,
			&u.Detail, &u.URL1, &u.Update_time); err != nil {
			log.Printf("fail: rows.Scan, %v\n", err)

			if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
				log.Printf("fail: rows.Close(), %v\n", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	// ②-4 モジュール encoding/json を用いてレスポンス用ユーザースライスをJSONへ変換し、HTTPレスポンスボディに書き込みます。
	bytes, err := json.Marshal(users)
	if err != nil {
		log.Printf("fail: json.Marshal, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func handler6(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ②-2 GETリクエストのクエリパラメータから条件を満たすデータを取得します。
	rows, err := db.Query("SELECT * FROM catalog WHERE item_category ='動画' order by update_time")
	if err != nil {
		log.Printf("fail: db.Query, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// ②-3
	users := make([]CatalogResForHTTP, 0)
	for rows.Next() {
		var u CatalogResForHTTP
		if err := rows.Scan(&u.Id, &u.Name, &u.Item_category, &u.Curriculum_category,
			&u.Detail, &u.URL1, &u.Update_time); err != nil {
			log.Printf("fail: rows.Scan, %v\n", err)

			if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
				log.Printf("fail: rows.Close(), %v\n", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	// ②-4 モジュール encoding/json を用いてレスポンス用ユーザースライスをJSONへ変換し、HTTPレスポンスボディに書き込みます。
	bytes, err := json.Marshal(users)
	if err != nil {
		log.Printf("fail: json.Marshal, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func handler7(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ②-2 GETリクエストのクエリパラメータから条件を満たすデータを取得します。
	rows, err := db.Query("SELECT * FROM catalog WHERE curriculum_category ='エディタ' order by update_time")
	if err != nil {
		log.Printf("fail: db.Query, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// ②-3
	users := make([]CatalogResForHTTP, 0)
	for rows.Next() {
		var u CatalogResForHTTP
		if err := rows.Scan(&u.Id, &u.Name, &u.Item_category, &u.Curriculum_category,
			&u.Detail, &u.URL1, &u.Update_time); err != nil {
			log.Printf("fail: rows.Scan, %v\n", err)

			if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
				log.Printf("fail: rows.Close(), %v\n", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	// ②-4 モジュール encoding/json を用いてレスポンス用ユーザースライスをJSONへ変換し、HTTPレスポンスボディに書き込みます。
	bytes, err := json.Marshal(users)
	if err != nil {
		log.Printf("fail: json.Marshal, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func handler8(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ②-2 GETリクエストのクエリパラメータから条件を満たすデータを取得します。
	rows, err := db.Query("SELECT * FROM catalog WHERE curriculum_category ='OSコマンド（とシェル）' order by update_time")
	if err != nil {
		log.Printf("fail: db.Query, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// ②-3
	users := make([]CatalogResForHTTP, 0)
	for rows.Next() {
		var u CatalogResForHTTP
		if err := rows.Scan(&u.Id, &u.Name, &u.Item_category, &u.Curriculum_category,
			&u.Detail, &u.URL1, &u.Update_time); err != nil {
			log.Printf("fail: rows.Scan, %v\n", err)

			if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
				log.Printf("fail: rows.Close(), %v\n", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	// ②-4 モジュール encoding/json を用いてレスポンス用ユーザースライスをJSONへ変換し、HTTPレスポンスボディに書き込みます。
	bytes, err := json.Marshal(users)
	if err != nil {
		log.Printf("fail: json.Marshal, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func handler9(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ②-2 GETリクエストのクエリパラメータから条件を満たすデータを取得します。
	rows, err := db.Query("SELECT * FROM catalog WHERE curriculum_category ='Git'")
	if err != nil {
		log.Printf("fail: db.Query, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// ②-3
	users := make([]CatalogResForHTTP, 0)
	for rows.Next() {
		var u CatalogResForHTTP
		if err := rows.Scan(&u.Id, &u.Name, &u.Item_category, &u.Curriculum_category,
			&u.Detail, &u.URL1, &u.Update_time); err != nil {
			log.Printf("fail: rows.Scan, %v\n", err)

			if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
				log.Printf("fail: rows.Close(), %v\n", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	// ②-4 モジュール encoding/json を用いてレスポンス用ユーザースライスをJSONへ変換し、HTTPレスポンスボディに書き込みます。
	bytes, err := json.Marshal(users)
	if err != nil {
		log.Printf("fail: json.Marshal, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}
func handler10(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ②-2 GETリクエストのクエリパラメータから条件を満たすデータを取得します。
	rows, err := db.Query("SELECT * FROM catalog WHERE curriculum_category ='Github' order by update_time")
	if err != nil {
		log.Printf("fail: db.Query, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// ②-3
	users := make([]CatalogResForHTTP, 0)
	for rows.Next() {
		var u CatalogResForHTTP
		if err := rows.Scan(&u.Id, &u.Name, &u.Item_category, &u.Curriculum_category,
			&u.Detail, &u.URL1, &u.Update_time); err != nil {
			log.Printf("fail: rows.Scan, %v\n", err)

			if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
				log.Printf("fail: rows.Close(), %v\n", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	// ②-4 モジュール encoding/json を用いてレスポンス用ユーザースライスをJSONへ変換し、HTTPレスポンスボディに書き込みます。
	bytes, err := json.Marshal(users)
	if err != nil {
		log.Printf("fail: json.Marshal, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}
func handler11(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ②-2 GETリクエストのクエリパラメータから条件を満たすデータを取得します。
	rows, err := db.Query("SELECT * FROM catalog WHERE curriculum_category ='HTML&CSS' order by update_time")
	if err != nil {
		log.Printf("fail: db.Query, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// ②-3
	users := make([]CatalogResForHTTP, 0)
	for rows.Next() {
		var u CatalogResForHTTP
		if err := rows.Scan(&u.Id, &u.Name, &u.Item_category, &u.Curriculum_category,
			&u.Detail, &u.URL1, &u.Update_time); err != nil {
			log.Printf("fail: rows.Scan, %v\n", err)

			if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
				log.Printf("fail: rows.Close(), %v\n", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	// ②-4 モジュール encoding/json を用いてレスポンス用ユーザースライスをJSONへ変換し、HTTPレスポンスボディに書き込みます。
	bytes, err := json.Marshal(users)
	if err != nil {
		log.Printf("fail: json.Marshal, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func handler12(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ②-2 GETリクエストのクエリパラメータから条件を満たすデータを取得します。
	rows, err := db.Query("SELECT * FROM catalog WHERE curriculum_category ='JavaScript' order by update_time")
	if err != nil {
		log.Printf("fail: db.Query, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// ②-3
	users := make([]CatalogResForHTTP, 0)
	for rows.Next() {
		var u CatalogResForHTTP
		if err := rows.Scan(&u.Id, &u.Name, &u.Item_category, &u.Curriculum_category,
			&u.Detail, &u.URL1, &u.Update_time); err != nil {
			log.Printf("fail: rows.Scan, %v\n", err)

			if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
				log.Printf("fail: rows.Close(), %v\n", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	// ②-4 モジュール encoding/json を用いてレスポンス用ユーザースライスをJSONへ変換し、HTTPレスポンスボディに書き込みます。
	bytes, err := json.Marshal(users)
	if err != nil {
		log.Printf("fail: json.Marshal, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func handler13(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ②-2 GETリクエストのクエリパラメータから条件を満たすデータを取得します。
	rows, err := db.Query("SELECT * FROM catalog WHERE curriculum_category ='React' order by update_time")
	if err != nil {
		log.Printf("fail: db.Query, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// ②-3
	users := make([]CatalogResForHTTP, 0)
	for rows.Next() {
		var u CatalogResForHTTP
		if err := rows.Scan(&u.Id, &u.Name, &u.Item_category, &u.Curriculum_category,
			&u.Detail, &u.URL1, &u.Update_time); err != nil {
			log.Printf("fail: rows.Scan, %v\n", err)

			if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
				log.Printf("fail: rows.Close(), %v\n", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	// ②-4 モジュール encoding/json を用いてレスポンス用ユーザースライスをJSONへ変換し、HTTPレスポンスボディに書き込みます。
	bytes, err := json.Marshal(users)
	if err != nil {
		log.Printf("fail: json.Marshal, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func handler14(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ②-2 GETリクエストのクエリパラメータから条件を満たすデータを取得します。
	rows, err := db.Query("SELECT * FROM catalog WHERE curriculum_category ='ReactxTypescript' order by update_time")
	if err != nil {
		log.Printf("fail: db.Query, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// ②-3
	users := make([]CatalogResForHTTP, 0)
	for rows.Next() {
		var u CatalogResForHTTP
		if err := rows.Scan(&u.Id, &u.Name, &u.Item_category, &u.Curriculum_category,
			&u.Detail, &u.URL1, &u.Update_time); err != nil {
			log.Printf("fail: rows.Scan, %v\n", err)

			if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
				log.Printf("fail: rows.Close(), %v\n", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	// ②-4 モジュール encoding/json を用いてレスポンス用ユーザースライスをJSONへ変換し、HTTPレスポンスボディに書き込みます。
	bytes, err := json.Marshal(users)
	if err != nil {
		log.Printf("fail: json.Marshal, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func handler15(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ②-2 GETリクエストのクエリパラメータから条件を満たすデータを取得します。
	rows, err := db.Query("SELECT * FROM catalog WHERE curriculum_category ='SQL' order by update_time")
	if err != nil {
		log.Printf("fail: db.Query, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// ②-3
	users := make([]CatalogResForHTTP, 0)
	for rows.Next() {
		var u CatalogResForHTTP
		if err := rows.Scan(&u.Id, &u.Name, &u.Item_category, &u.Curriculum_category,
			&u.Detail, &u.URL1, &u.Update_time); err != nil {
			log.Printf("fail: rows.Scan, %v\n", err)

			if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
				log.Printf("fail: rows.Close(), %v\n", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	// ②-4 モジュール encoding/json を用いてレスポンス用ユーザースライスをJSONへ変換し、HTTPレスポンスボディに書き込みます。
	bytes, err := json.Marshal(users)
	if err != nil {
		log.Printf("fail: json.Marshal, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func handler16(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ②-2 GETリクエストのクエリパラメータから条件を満たすデータを取得します。
	rows, err := db.Query("SELECT * FROM catalog WHERE curriculum_category='Docker' order by update_time")
	if err != nil {
		log.Printf("fail: db.Query, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// ②-3
	users := make([]CatalogResForHTTP, 0)
	for rows.Next() {
		var u CatalogResForHTTP
		if err := rows.Scan(&u.Id, &u.Name, &u.Item_category, &u.Curriculum_category,
			&u.Detail, &u.URL1, &u.Update_time); err != nil {
			log.Printf("fail: rows.Scan, %v\n", err)

			if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
				log.Printf("fail: rows.Close(), %v\n", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	// ②-4 モジュール encoding/json を用いてレスポンス用ユーザースライスをJSONへ変換し、HTTPレスポンスボディに書き込みます。
	bytes, err := json.Marshal(users)
	if err != nil {
		log.Printf("fail: json.Marshal, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func handler17(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ②-2 GETリクエストのクエリパラメータから条件を満たすデータを取得します。
	rows, err := db.Query("SELECT * FROM catalog WHERE curriculum_category ='Go' order by update_time")
	if err != nil {
		log.Printf("fail: db.Query, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// ②-3
	users := make([]CatalogResForHTTP, 0)
	for rows.Next() {
		var u CatalogResForHTTP
		if err := rows.Scan(&u.Id, &u.Name, &u.Item_category, &u.Curriculum_category,
			&u.Detail, &u.URL1, &u.Update_time); err != nil {
			log.Printf("fail: rows.Scan, %v\n", err)

			if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
				log.Printf("fail: rows.Close(), %v\n", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	// ②-4 モジュール encoding/json を用いてレスポンス用ユーザースライスをJSONへ変換し、HTTPレスポンスボディに書き込みます。
	bytes, err := json.Marshal(users)
	if err != nil {
		log.Printf("fail: json.Marshal, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func handler18(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ②-2 GETリクエストのクエリパラメータから条件を満たすデータを取得します。
	rows, err := db.Query("SELECT * FROM catalog WHERE curriculum_category ='HTTP Server (Go)' order by update_time")
	if err != nil {
		log.Printf("fail: db.Query, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// ②-3
	users := make([]CatalogResForHTTP, 0)
	for rows.Next() {
		var u CatalogResForHTTP
		if err := rows.Scan(&u.Id, &u.Name, &u.Item_category, &u.Curriculum_category,
			&u.Detail, &u.URL1, &u.Update_time); err != nil {
			log.Printf("fail: rows.Scan, %v\n", err)

			if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
				log.Printf("fail: rows.Close(), %v\n", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	// ②-4 モジュール encoding/json を用いてレスポンス用ユーザースライスをJSONへ変換し、HTTPレスポンスボディに書き込みます。
	bytes, err := json.Marshal(users)
	if err != nil {
		log.Printf("fail: json.Marshal, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func handler19(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ②-2 GETリクエストのクエリパラメータから条件を満たすデータを取得します。
	rows, err := db.Query("SELECT * FROM catalog WHERE curriculum_category ='RDBMS(MySQL)へ接続(Go)' order by update_time")
	if err != nil {
		log.Printf("fail: db.Query, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// ②-3
	users := make([]CatalogResForHTTP, 0)
	for rows.Next() {
		var u CatalogResForHTTP
		if err := rows.Scan(&u.Id, &u.Name, &u.Item_category, &u.Curriculum_category,
			&u.Detail, &u.URL1, &u.Update_time); err != nil {
			log.Printf("fail: rows.Scan, %v\n", err)

			if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
				log.Printf("fail: rows.Close(), %v\n", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	// ②-4 モジュール encoding/json を用いてレスポンス用ユーザースライスをJSONへ変換し、HTTPレスポンスボディに書き込みます。
	bytes, err := json.Marshal(users)
	if err != nil {
		log.Printf("fail: json.Marshal, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func handler20(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ②-2 GETリクエストのクエリパラメータから条件を満たすデータを取得します。
	rows, err := db.Query("SELECT * FROM catalog WHERE curriculum_category = 'Unit Test(Go)' order by update_time")
	if err != nil {
		log.Printf("fail: db.Query, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// ②-3
	users := make([]CatalogResForHTTP, 0)
	for rows.Next() {
		var u CatalogResForHTTP
		if err := rows.Scan(&u.Id, &u.Name, &u.Item_category, &u.Curriculum_category,
			&u.Detail, &u.URL1, &u.Update_time); err != nil {
			log.Printf("fail: rows.Scan, %v\n", err)

			if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
				log.Printf("fail: rows.Close(), %v\n", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	// ②-4 モジュール encoding/json を用いてレスポンス用ユーザースライスをJSONへ変換し、HTTPレスポンスボディに書き込みます。
	bytes, err := json.Marshal(users)
	if err != nil {
		log.Printf("fail: json.Marshal, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}
func handler21(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ②-2 GETリクエストのクエリパラメータから条件を満たすデータを取得します。
	rows, err := db.Query("SELECT * FROM catalog WHERE curriculum_category ='フロントエンドとバックエンドの接続' order by update_time")
	if err != nil {
		log.Printf("fail: db.Query, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// ②-3
	users := make([]CatalogResForHTTP, 0)
	for rows.Next() {
		var u CatalogResForHTTP
		if err := rows.Scan(&u.Id, &u.Name, &u.Item_category, &u.Curriculum_category,
			&u.Detail, &u.URL1, &u.Update_time); err != nil {
			log.Printf("fail: rows.Scan, %v\n", err)

			if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
				log.Printf("fail: rows.Close(), %v\n", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	// ②-4 モジュール encoding/json を用いてレスポンス用ユーザースライスをJSONへ変換し、HTTPレスポンスボディに書き込みます。
	bytes, err := json.Marshal(users)
	if err != nil {
		log.Printf("fail: json.Marshal, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}
func handler22(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ②-2 GETリクエストのクエリパラメータから条件を満たすデータを取得します。
	rows, err := db.Query("SELECT * FROM catalog WHERE curriculum_category ='CI(Continuous Integration)' order by update_time")
	if err != nil {
		log.Printf("fail: db.Query, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// ②-3
	users := make([]CatalogResForHTTP, 0)
	for rows.Next() {
		var u CatalogResForHTTP
		if err := rows.Scan(&u.Id, &u.Name, &u.Item_category, &u.Curriculum_category,
			&u.Detail, &u.URL1, &u.Update_time); err != nil {
			log.Printf("fail: rows.Scan, %v\n", err)

			if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
				log.Printf("fail: rows.Close(), %v\n", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	// ②-4 モジュール encoding/json を用いてレスポンス用ユーザースライスをJSONへ変換し、HTTPレスポンスボディに書き込みます。
	bytes, err := json.Marshal(users)
	if err != nil {
		log.Printf("fail: json.Marshal, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}
func handler23(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ②-2 GETリクエストのクエリパラメータから条件を満たすデータを取得します。
	rows, err := db.Query("SELECT * FROM catalog WHERE curriculum_category ='CD(Conteinuous Delivery / Deployment)' order by update_time")
	if err != nil {
		log.Printf("fail: db.Query, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// ②-3
	users := make([]CatalogResForHTTP, 0)
	for rows.Next() {
		var u CatalogResForHTTP
		if err := rows.Scan(&u.Id, &u.Name, &u.Item_category, &u.Curriculum_category,
			&u.Detail, &u.URL1, &u.Update_time); err != nil {
			log.Printf("fail: rows.Scan, %v\n", err)

			if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
				log.Printf("fail: rows.Close(), %v\n", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	// ②-4 モジュール encoding/json を用いてレスポンス用ユーザースライスをJSONへ変換し、HTTPレスポンスボディに書き込みます。
	bytes, err := json.Marshal(users)
	if err != nil {
		log.Printf("fail: json.Marshal, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}
func handler24(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ②-2 GETリクエストのクエリパラメータから条件を満たすデータを取得します。
	rows, err := db.Query("SELECT * FROM catalog WHERE curriculum_category ='認証' order by update_time")
	if err != nil {
		log.Printf("fail: db.Query, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// ②-3
	users := make([]CatalogResForHTTP, 0)
	for rows.Next() {
		var u CatalogResForHTTP
		if err := rows.Scan(&u.Id, &u.Name, &u.Item_category, &u.Curriculum_category,
			&u.Detail, &u.URL1, &u.Update_time); err != nil {
			log.Printf("fail: rows.Scan, %v\n", err)

			if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
				log.Printf("fail: rows.Close(), %v\n", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	// ②-4 モジュール encoding/json を用いてレスポンス用ユーザースライスをJSONへ変換し、HTTPレスポンスボディに書き込みます。
	bytes, err := json.Marshal(users)
	if err != nil {
		log.Printf("fail: json.Marshal, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func handler25(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ②-2 GETリクエストのクエリパラメータから条件を満たすデータを取得します。
	rows, err := db.Query("SELECT * FROM catalog WHERE curriculum_category ='ハッカソン準備' order by update_time")
	if err != nil {
		log.Printf("fail: db.Query, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// ②-3
	users := make([]CatalogResForHTTP, 0)
	for rows.Next() {
		var u CatalogResForHTTP
		if err := rows.Scan(&u.Id, &u.Name, &u.Item_category, &u.Curriculum_category,
			&u.Detail, &u.URL1, &u.Update_time); err != nil {
			log.Printf("fail: rows.Scan, %v\n", err)

			if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
				log.Printf("fail: rows.Close(), %v\n", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	// ②-4 モジュール encoding/json を用いてレスポンス用ユーザースライスをJSONへ変換し、HTTPレスポンスボディに書き込みます。
	bytes, err := json.Marshal(users)
	if err != nil {
		log.Printf("fail: json.Marshal, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}
func handler26(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ②-2 GETリクエストのクエリパラメータから条件を満たすデータを取得します。
	rows, err := db.Query("SELECT * FROM catalog WHERE curriculum_category ='ハッカソンの概要' order by update_time")
	if err != nil {
		log.Printf("fail: db.Query, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// ②-3
	users := make([]CatalogResForHTTP, 0)
	for rows.Next() {
		var u CatalogResForHTTP
		if err := rows.Scan(&u.Id, &u.Name, &u.Item_category, &u.Curriculum_category,
			&u.Detail, &u.URL1, &u.Update_time); err != nil {
			log.Printf("fail: rows.Scan, %v\n", err)

			if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
				log.Printf("fail: rows.Close(), %v\n", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	// ②-4 モジュール encoding/json を用いてレスポンス用ユーザースライスをJSONへ変換し、HTTPレスポンスボディに書き込みます。
	bytes, err := json.Marshal(users)
	if err != nil {
		log.Printf("fail: json.Marshal, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}
func handler27(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ②-2 GETリクエストのクエリパラメータから条件を満たすデータを取得します。
	rows, err := db.Query("SELECT * FROM catalog WHERE curriculum_category ='不明' order by update_time")
	if err != nil {
		log.Printf("fail: db.Query, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// ②-3
	users := make([]CatalogResForHTTP, 0)
	for rows.Next() {
		var u CatalogResForHTTP
		if err := rows.Scan(&u.Id, &u.Name, &u.Item_category, &u.Curriculum_category,
			&u.Detail, &u.URL1, &u.Update_time); err != nil {
			log.Printf("fail: rows.Scan, %v\n", err)

			if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
				log.Printf("fail: rows.Close(), %v\n", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	// ②-4 モジュール encoding/json を用いてレスポンス用ユーザースライスをJSONへ変換し、HTTPレスポンスボディに書き込みます。
	bytes, err := json.Marshal(users)
	if err != nil {
		log.Printf("fail: json.Marshal, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func main() {
	// ② /userでリクエストされたらnameパラメーターと一致する名前を持つレコードをJSON形式で返す
	http.HandleFunc("/catalog_add", handler)
	http.HandleFunc("/catalogs", handler2)
	http.HandleFunc("/catalog_delete", handler3)
	http.HandleFunc("/catalog_book", handler4)
	http.HandleFunc("/catalog_blog", handler5)
	http.HandleFunc("/catalog_video", handler6)
	http.HandleFunc("/catalog_edita", handler7)
	http.HandleFunc("/catalog_os", handler8)
	http.HandleFunc("/catalog_git", handler9)
	http.HandleFunc("/catalog_github", handler10)
	http.HandleFunc("/catalog_html", handler11)
	http.HandleFunc("/catalog_javascript", handler12)
	http.HandleFunc("/catalog_react", handler13)
	http.HandleFunc("/catalog_react-ts", handler14)
	http.HandleFunc("/catalog_sql", handler15)
	http.HandleFunc("/catalog_docker", handler16)
	http.HandleFunc("/catalog_go", handler17)
	http.HandleFunc("/catalog_http", handler18)
	http.HandleFunc("/catalog_rdbms", handler19)
	http.HandleFunc("/catalog_unit", handler20)
	http.HandleFunc("/catalog_fb", handler21)
	http.HandleFunc("/catalog_ci", handler22)
	http.HandleFunc("/catalog_cd", handler23)
	http.HandleFunc("/catalog_authen", handler24)
	http.HandleFunc("/catalog_ready", handler25)
	http.HandleFunc("/catalog_hackathon", handler26)
	http.HandleFunc("/catalog_mys", handler27)

	// ③ Ctrl+CでHTTPサーバー停止時にDBをクローズする
	closeDBWithSysCall()

	// 8000番ポートでリクエストを待ち受ける
	log.Println("Listening...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

// ③ Ctrl+CでHTTPサーバー停止時にDBをクローズする
func closeDBWithSysCall() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sig
		log.Printf("received syscall, %v", s)

		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
		log.Printf("success: db.Close()")
		os.Exit(0)
	}()
}

// ①idの採番
//func GetNewULIDString() string {
//	t := time.Now()
//	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
//	id := ulid.MustNew(ulid.Timestamp(t), entropy)
//	return strings.ToLower(id.String())
//}
