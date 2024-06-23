package main

import (
	"context"
	"embed"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"text/template"
	"time"

	dbhelper "bufo.zone"
	"bufo.zone/dbufo"
	"github.com/gorilla/sessions"
	_ "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
)

//go:embed templates/*
var templates embed.FS
var t = template.Must(template.ParseFS(templates, "templates/*"))

func indexHandler(w http.ResponseWriter, r *http.Request) {
	type BufoViewData struct {
		Name  string
		Score int
		Frogs string
		Url   string
	}

	ctx := context.Background()
	helpers := dbhelper.GetDb(ctx)
	queries := helpers.Queries
	defer helpers.Close()

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
		score := int(bufo.Rating)
		frogs := ""
		if bufo.Rating > 0 {
			frogs = strings.Repeat("🐸", int(bufo.Rating))
		}

		url, _ := url.JoinPath(os.Getenv("BUFO_URL"), bufo.Name)
		viewData := BufoViewData{
			Name:  bufo.Name,
			Score: score,
			Frogs: frogs,
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

	name, ok := session.Values[inputs.Name].(int64)
	futureTime := time.Unix(name, 0).Add(time.Hour * 12)
	if ok && time.Now().Compare(futureTime) < 0 {
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
	session.Values[inputs.Name] = time.Now().Unix()
	err := session.Save(r, w)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	helpers := dbhelper.GetDb(ctx)
	queries := helpers.Queries
	defer helpers.Close()
	_, err = queries.CreateVote(ctx, dbufo.CreateVoteParams{
		Value:   int32(inputs.Value),
		Created: pgtype.Timestamptz{Time: time.Now(), Valid: true},
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
		Score: int(vote.Rating),
	})
	if err != nil {
		panic(err)
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(data)
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Error loading .env file")
	}

	router := httprouter.New()
	router.Handler(http.MethodGet, "/", http.HandlerFunc(indexHandler))
	router.HandlerFunc(http.MethodPost, "/vote", voteHandler)
	router.ServeFiles("/static/*filepath", http.Dir("static"))

	port := os.Getenv("BUFO_PORT")
	if port == "" {
		port = "8001"
	}
	log.Println("running on", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
