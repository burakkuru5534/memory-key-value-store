package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"io/ioutil"

	//	"encoding/json"
	//	"errors"
	"fmt"
	"log"
	"net/http"
	//	"time"

	_ "github.com/lib/pq"

	"github.com/google/uuid"

	//	"github.com/go-redis/redis"
	//"github.com/gorilla/mux"
	//	"github.com/patrickmn/go-cache"
)

//var myCache CacheItf
var localDB *sql.DB
var addr = flag.String("addr", ":8080", "http server address")

func main() {
	InitDB()
	//InitRedisCache() // comment if want to use app cache
	//InitCache()      // comment if want to use redis cache

	//r := mux.NewRouter()

	h1 := func(w http.ResponseWriter, r *http.Request) {
		SetRecord(w,r)
	}
	h2 := func(w http.ResponseWriter, r *http.Request) {
		GetRecord(w,r)
	}
	http.HandleFunc("/set/record", h1)
	http.HandleFunc("/get/record", h2)


//	http.HandleFunc("/set/record", SetRecord)
//	http.HandleFunc("/get/record", GetRecord)
	http.HandleFunc("/flush/records", FlushRecord)

	//srv := &http.Server{
	//	Handler: r,
	//	Addr:    "127.0.0.1:8080",
	//}

	log.Fatal(http.ListenAndServe(*addr,nil))
}

//type CacheItf interface {
//	Set(key string, data interface{}, expiration time.Duration) error
//	Get(key string) ([]byte, error)
//}
//
//type RedisCache struct {
//	client *redis.Client
//}
//
//type AppCache struct {
//	client *cache.Cache
//}
//
//func (r *RedisCache) Set(key string, data interface{}, expiration time.Duration) error {
//	b, err := json.Marshal(data)
//	if err != nil {
//		return err
//	}
//
//	return r.client.Set(key, b, expiration).Err()
//}
//
//func (r *RedisCache) Get(key string) ([]byte, error) {
//	result, err := r.client.Get(key).Bytes()
//	if err == redis.Nil {
//		return nil, nil
//	}
//
//	return result, err
//}
//
//func (r *AppCache) Set(key string, data interface{}, expiration time.Duration) error {
//	b, err := json.Marshal(data)
//	if err != nil {
//		return err
//	}
//
//	r.client.Set(key, b, expiration)
//	return nil
//}
//
//func (r *AppCache) Get(key string) ([]byte, error) {
//	res, exist := r.client.Get(key)
//	if !exist {
//		return nil, nil
//	}
//
//	resByte, ok := res.([]byte)
//	if !ok {
//		return nil, errors.New("Format is not arr of bytes")
//	}
//
//	return resByte, nil
//}
//
//type ToDo struct {
//	UserID int    `json:"userId"`
//	ID     int    `json:"id"`
//	Title  string `json:"title"`
//	Body   string `json:"body"`
//}
//
//func GetPost(w http.ResponseWriter, r *http.Request) {
//	start := time.Now()
//
//	var result ToDo
//
//	b, err := myCache.Get("todo")
//	if err != nil {
//		// error
//		log.Fatal(err)
//	}
//
//	if b != nil {
//		// cache exist
//		err := json.Unmarshal(b, &result)
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		b, _ := json.Marshal(map[string]interface{}{
//			"data":    result,
//			"elapsed": time.Since(start).Microseconds(),
//		})
//		w.Write([]byte(b))
//		return
//	}
//
//	// Get from DB
//	err = localDB.QueryRow(`SELECT id, user_id, title, body FROM posts WHERE id = $1`, 1).Scan(&result.ID, &result.UserID, &result.Title, &result.Body)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	err = myCache.Set("todo", result, 1*time.Minute)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	b, err = json.Marshal(map[string]interface{}{
//		"data":    result,
//		"elapsed": time.Since(start).Microseconds(),
//	})
//
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	w.Write(b)
//}
//
//func InitRedisCache() {
//	myCache = &RedisCache{
//		client: redis.NewClient(&redis.Options{
//			Addr:     "localhost:6379",
//			Password: "", // no password set
//			DB:       0,  // use default DB
//		}),
//	}
//
//}
//
//func InitCache() {
//	myCache = &AppCache{
//		client: cache.New(5*time.Minute, 10*time.Minute),
//	}
//}

//Db initialization function
func InitDB() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		"localhost", 5432, "postgres", "tayitkan", "postgres")

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	sqlStmt := `	
		CREATE EXTENSION if not exists hstore;
		CREATE TABLE if not exists my_table (
                       id serial primary key,
                       my_key VARCHAR (255),
                       my_value hstore
);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatal("%q: %s\n", err, sqlStmt)
	}

	localDB = db
}

//set key value
func SetRecord(w http.ResponseWriter, r *http.Request) {

	var data struct {
		Key   string
		Value string
	}

	err := BodyToJsonReq(r, &data)
	if err != nil {
		log.Fatal("%q", err)
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		"localhost", 5432, "postgres", "tayitkan", "postgres")

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	sq := fmt.Sprintf(`	
		INSERT INTO my_table (my_key, my_value)
					VALUES
					(
						'%s',
						'%s'
					);
	`, data.Key,data.Value)

	_, err = db.Exec(sq)
	if err != nil {
		log.Fatal("%q: %s\n", err, sq)
	}

	ResponseSuccess(w,data)

}

//get value by key
func GetRecord(w http.ResponseWriter, r *http.Request) {

	key:= r.FormValue("key")
	type dataS struct {
		Key   string
		Value string
	}

	var valueData []dataS

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		"localhost", 5432, "postgres", "tayitkan", "postgres")

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	sq := fmt.Sprintf(`SELECT
				my_key,
				my_value -> '%s'
			FROM my_table;
			`, key)

	rows, err := db.Query(sq)
	if err != nil {
		log.Fatal(err)
	}


	defer rows.Close()
	for rows.Next() {
		var valueDataOne dataS
		rows.Scan(&valueDataOne.Key, &valueDataOne.Value)
		valueData = append(valueData, valueDataOne)
	}

	ResponseSuccess(w,valueData)

}

//fluch all data
func FlushRecord(w http.ResponseWriter, r *http.Request) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		"localhost", 5432, "postgres", "tayitkan", "postgres")

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	sq := `	
		delete from my_table;
	`
	_, err = db.Exec(sq)
	if err != nil {
		log.Fatal("%q: %s\n", err, sq)
	}

	ResponseSuccess(w,"OK")
}

//--helper functions--//

// return response helper functions
type Response struct {
	ID      string
	Success bool
	Message string
	Data    interface{}
}
func (m *Response) SendWithStatus(w http.ResponseWriter, statusCode int) error {
	encjson, err := json.Marshal(m)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, err = w.Write(encjson)
	return err
}
func (m *Response) Send(w http.ResponseWriter) error {
	return m.SendWithStatus(w, http.StatusOK)
}
func NewResponse(success bool, message string, data interface{}) (*Response, error) {
	u, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	m := &Response{
		ID:      u.String(),
		Success: success,
		Message: message,
		Data:    data,
	}

	return m, nil
}
func ResponseSuccess(w http.ResponseWriter, data interface{}) {
	ar, err := NewResponse(true, "", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = ar.Send(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}
// return response helper functions

//body to json function
func BodyToJsonReq(r *http.Request, data interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return err
	}

	defer r.Body.Close()

	return nil
}
//body to json function

//--helper functions--//
