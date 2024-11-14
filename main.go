package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	_ "github.com/lib/pq"
	"github.com/rs/xid"
)

type Url struct {
	orgURL     string    `json."orgURL"`
	shortURL   string    `json."shortUrl"`
	createDate time.Time `json.createDate`
}

var dbObj sql.DB

func connectDB(dbObj *sql.DB) {
	connStr := "postgres://postgres:dev12@localhost:5432/postgres?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	// var category_id string
	// var name string
	// type obj struct {
	// 	category_id string
	// 	name        string
	// }
	// var data []obj
	*dbObj = *db
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Connected to Database successfully....")

	// rows, qyErr := db.Query("SELECT * FROM public.\"url \"")
	// if qyErr != nil {
	// 	log.Fatalln("errror", qyErr)
	// }
	// fmt.Println(rows.Columns())
	// for rows.Next() {
	// 	rows.Scan(&category_id, &name)
	// 	var respones obj = obj{category_id: category_id, name: name}

	// 	data = append(data, respones)
	// }
	// defer rows.Close()
	// fmt.Println(data)
}

func generateShortURL(orgURL string) string {
	hasher := md5.New()
	hasher.Write([]byte(orgURL))
	data := hasher.Sum(nil)
	hash := hex.EncodeToString(data)
	return hash[:10]
}
func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to URL Shortner")
}
func createShortURL(w http.ResponseWriter, r *http.Request) {
	var data struct {
		URL string `json.url`
	}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid request body", 400)
		return
	}
	if data.URL == "" {
		http.Error(w, "URL is required", 400)
		return
	}
	guid := xid.New()
	shortURL := generateShortURL(data.URL)
	url_id := guid.String()
	insert_query := fmt.Sprintf("INSERT INTO public.\"url \" (url_id, short_url, org_url, created_date) VALUES ('%s', '%s', '%s', '%s')",
		url_id,
		shortURL,
		data.URL,
		time.Now().Format("2006-01-02 15:04:05"))
	_, error := dbObj.Exec(insert_query)
	if error != nil {
		http.Error(w, "Error creating short URL", 500)
		return
	}
	fmt.Println("Data Inserted")
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"msg":         "Shorten Url created successfully",
		"shorten_url": "http://localhost:8000/redirect/" + url_id,
	}
	json.NewEncoder(w).Encode(response)
}

func redirectUrl(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/redirect/"):]
	var query string = fmt.Sprintf("select * from public.\"url \" where url_id='%s'", id)
	result, err := dbObj.Query(query)
	// fmt.Println(query)

	if err != nil {
		fmt.Println(err)
		return
	}

	var url_id string
	var short_url string
	var org_url string
	var created_date string

	// fmt.Println(result.Columns())
	for result.Next() {
		result.Scan(&url_id, &short_url, &org_url, &created_date)
	}
	// fmt.Println("Org : ",org_url)
	// fmt.Println("Org : ", short_url)
	http.Redirect(w, r, org_url, http.StatusFound)

}

func main() {
	connectDB(&dbObj)
	// http.HandleFunc("/", handler)
	http.HandleFunc("/create", createShortURL)
	http.HandleFunc("/redirect/", redirectUrl)
	fmt.Println("Server is running : http://localhost:8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		fmt.Println("Error while starting the server", err)
	}
}
