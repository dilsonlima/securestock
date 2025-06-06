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
	EstoqueEmergencia float64
}

var db *sql.DB
var templates = template.Must(template.ParseGlob("templates/*.html"))

func main() {
	initDB()
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/new", newItemHandler)
	http.HandleFunc("/save", saveItemHandler)
	http.HandleFunc("/view", viewItemHandler)
	http.HandleFunc("/edit", editItemHandler)
	http.HandleFunc("/update", updateItemHandler)
	http.HandleFunc("/delete", deleteItemHandler)
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

	eseg := int((float64(lead) * consumo) * 0.20) // cálculo automático do estoque de emergência

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
	i.EstoqueEmergencia = (float64(i.LeadTimeDias) * i.ConsumoDiario) * 0.20
	i.EstoqueMinimo = i.ConsumoDiario * float64(i.LeadTimeDias)
	i.PontoDeRecompra = i.EstoqueMinimo + float64(i.EstoqueSeguranca)
	templates.ExecuteTemplate(w, "view_item.html", i)
}

func editItemHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	row := db.QueryRow("SELECT id, nome, consumo_diario, lead_time_dias, estoque_seguranca FROM items WHERE id = ?", id)

	var i Item
	err := row.Scan(&i.ID, &i.Nome, &i.ConsumoDiario, &i.LeadTimeDias, &i.EstoqueSeguranca)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	templates.ExecuteTemplate(w, "edit_item.html", i)
}

func updateItemHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.FormValue("id"))
	nome := r.FormValue("nome")
	consumo, _ := strconv.ParseFloat(r.FormValue("consumo"), 64)
	lead, _ := strconv.Atoi(r.FormValue("lead"))

	eseg := int((float64(lead) * consumo) * 0.20) // recálculo automático do estoque de emergência

	_, err := db.Exec("UPDATE items SET nome = ?, consumo_diario = ?, lead_time_dias = ?, estoque_seguranca = ? WHERE id = ?", nome, consumo, lead, eseg, id)
	if err != nil {
		log.Fatal(err)
	}
	http.Redirect(w, r, "/view?id="+strconv.Itoa(id), http.StatusSeeOther)
}

func deleteItemHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	_, err := db.Exec("DELETE FROM items WHERE id = ?", id)
	if err != nil {
		log.Fatal(err)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
