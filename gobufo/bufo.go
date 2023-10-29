package main

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"

	"bufo.zone/dbufo"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var tables string

//go:embed templates/*
var templates embed.FS
var t = template.Must(template.ParseFS(templates, "templates/*"))

func GetDb(ctx context.Context) *dbufo.Queries {
	dbPath := os.Getenv("DB_PATH")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}

	if _, err := db.ExecContext(ctx, tables); err != nil {
		panic(err)
	}

	return dbufo.New(db)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	type BufoViewData struct {
		Name  string
		Score int
		Frogs string
		Url   string
	}

	ctx := context.Background()
	queries := GetDb(ctx)

	data := struct {
		Bufos []BufoViewData
	}{
		Bufos: []BufoViewData{},
	}

	res, err := queries.ListBufos(ctx)
	if err != nil {
		panic(err)
	}

	for _, bufo := range res {
		score := -1
		if bufo.Rating.Valid {
			score = int(bufo.Rating.Float64)
		}

		url, _ := url.JoinPath(os.Getenv("BUFO_URL"), bufo.Name)
		viewData := BufoViewData{
			Name:  bufo.Name,
			Score: score,
			Frogs: strings.Repeat("🐸", int(bufo.Rating.Float64)),
			Url:   url,
		}
		data.Bufos = append(data.Bufos, viewData)
	}

	err = t.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		panic(err)
	}
}

func voteHandler(w http.ResponseWriter, r *http.Request) {
	type BufoVoteInputs struct {
		Name  string
		Value int
	}

	type BufoVoteSuccessData struct {
		Name  string `json:"name"`
		Score int    `json:"score"`
	}

	type BufoVoteErrorData struct {
		Error       string `json:"error"`
		Description string `json:"description"`
	}

	inputs := BufoVoteInputs{}
	json.NewDecoder(r.Body).Decode(&inputs)

	store := sessions.NewCookieStore([]byte(os.Getenv("SECRET_KEY")))

	session, _ := store.Get(r, "bufo-go-votes")
	if session.IsNew {
		session.Options.MaxAge = int(time.Hour) * 12
	}

	if session.Values[inputs.Name] != nil {
		log.Println("vote: already voted today", inputs)
		errorData, err := json.Marshal(BufoVoteErrorData{
			Error:       "already_voted",
			Description: "Already rated this Bufo today, try again tomorrow!",
		})
		if err != nil {
			panic(err)
		}
		w.Header().Add("Content-Type", "application/json")
		w.Write(errorData)
		return
	}

	log.Println("vote: adding vote for", inputs)
	session.Values[inputs.Name] = true
	err := session.Save(r, w)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	queries := GetDb(ctx)
	_, err = queries.CreateVote(ctx, dbufo.CreateVoteParams{
		Value:   int64(inputs.Value),
		Created: time.Now(),
		BufoID:  inputs.Name,
	})
	if err != nil {
		panic(err)
	}

	vote, err := queries.GetBufo(ctx, inputs.Name)
	if err != nil {
		panic(err)
	}

	data, err := json.Marshal(BufoVoteSuccessData{
		Name:  vote.Name,
		Score: int(vote.Rating.Float64),
	})
	if err != nil {
		panic(err)
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(data)
}

func main() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("Error loading .env file")
	}

	router := httprouter.New()
	router.HandlerFunc(http.MethodGet, "/", indexHandler)
	router.HandlerFunc(http.MethodPost, "/vote", voteHandler)
	router.ServeFiles("/static/*filepath", http.Dir("static"))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}
	log.Println("running on", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
