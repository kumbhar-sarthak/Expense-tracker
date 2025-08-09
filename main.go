package main

import (
	"encoding/json"
	"expense/handlers"
	"expense/models"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/csrf"
	"github.com/joho/godotenv"
)

type DashboardData struct {
	Name     string
	Email    string
	Picture  string
	ID       string
	Expenses []models.Expense
	Total string
	CSRFToken string
}


func showDashboard(w http.ResponseWriter, r *http.Request) {
	session, err := handlers.Store.Get(r, "user-session")
	if err != nil {
		log.Println("Unable to load session:", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	userID, _ := session.Values["google_id"].(string)
	name, _ := session.Values["name"].(string)
	email, _ := session.Values["email"].(string)
	picture, _ := session.Values["picture"].(string)
	cleanPicture := strings.TrimSuffix(picture, "=s96-c") + "=s100"

	_, err = models.RegisterOrGetUser(userID,name,email,picture)

	if err != nil {
		log.Fatal("did't register user",err)
	}

	expenses, err := models.GetData(userID)
	if err != nil {
		log.Println("unable to load your expenses:", err)
		http.Error(w, "Failed to load expenses", http.StatusInternalServerError)
		return
	}
	var total int = 0
	for _, e := range expenses {
    amt, err := strconv.ParseInt(e.Amount, 10, 64)
    if err != nil {
        log.Printf("Invalid amount: %v\n", err)
        continue
    }
    total += int(amt)
	}

	tmpl, err := template.New("dashboard.html").Funcs(template.FuncMap{
		"add" : func (a,b int) int {
			return a + b
		},
	}).ParseFiles("templates/dashboard.html")

	if err != nil {
		log.Println("Template error", " ", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	token := csrf.Token(r)


	data := DashboardData{
		Name:     name,
		Email:    email,
		Picture:  cleanPicture,
		ID:       userID,
		Expenses: expenses,
		Total: strconv.Itoa(total),
		CSRFToken: token,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Println("Template execution error:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}


func logoutUser(w http.ResponseWriter, r *http.Request){
	session, _ := handlers.Store.Get(r,"user-session")

	session.Options.MaxAge = -1
	session.Save(r,w)

	http.Redirect(w,r,"/",http.StatusSeeOther)
}

func addData(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPost {
		http.Error(w,"Invalid Request",http.StatusMethodNotAllowed)
	}

	session,err := handlers.Store.Get(r,"user-session")

	if err != nil {
		http.Error(w,"Session Error: ",http.StatusInternalServerError);;
	}

	userId, ok := session.Values["google_id"].(string)
	if !ok || userId == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	type Input struct {
		Description string `json:"description"`
		Amount string `json:"amount"`
		Category string `json:"category"`
		Ptype string `json:"ptype"`
	}

	var in Input
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		log.Println("Data format is not correct",err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return 
	}

	err = models.Addexpense(userId,in.Description,in.Amount,in.Category,in.Ptype)
	if err != nil {
		log.Println("Failed to add expense:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Expense added successfully",
	})
}

func deleteExpense(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPost {
		http.Error(w,"id required",http.StatusMethodNotAllowed)
	}

	idStr := strings.TrimPrefix(r.URL.Path,"/delete/")
	id,err := strconv.Atoi(idStr)

	if err != nil {
		http.Error(w,"Invalid expense ID" + err.Error(),http.StatusBadRequest)
		return
	}

	session, err := handlers.Store.Get(r,"user-session")

	if err != nil {
		http.Error(w,"Session error",http.StatusInternalServerError)
		return
	}

	userID := session.Values["google_id"].(string)

	err = models.Delete(userID,id)

	if err != nil {
		http.Error(w, "Failed to delete expense", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}


func main() {

	err := godotenv.Load("./.env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	models.InitDB()

	http.HandleFunc("/dashboard", showDashboard)

	http.HandleFunc("/logout",logoutUser)

	http.HandleFunc("/addData",addData)

	http.HandleFunc("/delete/",deleteExpense)

	http.HandleFunc("/auth/google/login", handlers.HandleGoogleLogin)

	http.HandleFunc("/auth/google/callback", handlers.HandleGoogleCallback)

	fs := http.FileServer(http.Dir("./templates"))

	http.Handle("/", fs)

	log.Println("Server running at http://localhost:7878")
	log.Fatal(http.ListenAndServe(":7878", nil))
}
