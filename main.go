package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

type Item struct {
	ID                int
	Nome              string
	ConsumoDiario     float64
	LeadTimeDias      int
	EstoqueSeguranca  int
	EstoqueMinimo     float64
	PontoDeRecompra   float64
}

var db *sql.DB
var templates = template.Must(template.ParseGlob("templates/*.html"))

func main() {
	initDB()
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/new", newItemHandler)
	http.HandleFunc("/save", saveItemHandler)
	http.HandleFunc("/view", viewItemHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	fmt.Println("Servidor rodando em http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "estoque.db")
	if err != nil {
		log.Fatal(err)
	}

	query := `CREATE TABLE IF NOT EXISTS items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nome TEXT,
		consumo_diario REAL,
		lead_time_dias INTEGER,
		estoque_seguranca INTEGER
	);`
	_, err = db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	rows, _ := db.Query("SELECT id, nome FROM items")
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var i Item
		rows.Scan(&i.ID, &i.Nome)
		items = append(items, i)
	}
	templates.ExecuteTemplate(w, "index.html", items)
}

func newItemHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "new_item.html", nil)
}

func saveItemHandler(w http.ResponseWriter, r *http.Request) {
	nome := r.FormValue("nome")
	consumo, _ := strconv.ParseFloat(r.FormValue("consumo"), 64)
	lead, _ := strconv.Atoi(r.FormValue("lead"))
	eseg, _ := strconv.Atoi(r.FormValue("seguranca"))

	_, err := db.Exec("INSERT INTO items (nome, consumo_diario, lead_time_dias, estoque_seguranca) VALUES (?, ?, ?, ?)", nome, consumo, lead, eseg)
	if err != nil {
		log.Fatal(err)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func viewItemHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	row := db.QueryRow("SELECT id, nome, consumo_diario, lead_time_dias, estoque_seguranca FROM items WHERE id = ?", id)

	var i Item
	err := row.Scan(&i.ID, &i.Nome, &i.ConsumoDiario, &i.LeadTimeDias, &i.EstoqueSeguranca)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	i.EstoqueMinimo = i.ConsumoDiario * float64(i.LeadTimeDias)
	i.PontoDeRecompra = i.EstoqueMinimo + float64(i.EstoqueSeguranca)
	templates.ExecuteTemplate(w, "view_item.html", i)
}
