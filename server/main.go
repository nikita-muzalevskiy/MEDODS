package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"os"
)

type ErrorMessage struct {
	Error string `json:"error"`
}

type Token struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type Auth struct {
	UserGuid string `json:"userGuid"`
}

type MyClaims struct {
	jwt.RegisteredClaims
	UserGuid string `json:"userGuid"`
	UserIp   string `json:"userIp"`
}

func init() {
	if err := godotenv.Load("environment.env"); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/auth", createToken)
	router.HandleFunc("/refresh", refresh)
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatalf("Произошла ошибка при запуске сервера: %v", err)
	}
}

func createToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var a Auth
	err := json.NewDecoder(r.Body).Decode(&a)
	if err != nil {
		log.Fatalf("Произошла ошибка при чтении JSON во время создания токенов по ID: %v", err)
	}
	if !checkGuid(a.UserGuid) {
		err = json.NewEncoder(w).Encode(ErrorMessage{"Такого guid нет в БД."})
		if err != nil {
			log.Fatalf("Произошла ошибка формировании ответа во время отпраки сообщения об отсутсвии guid: %v", err)
		}
		return
	}
	aToken, rToken := generateAccess(a.UserGuid, getIP(r)), generateRefresh(a.UserGuid)
	err = json.NewEncoder(w).Encode(Token{
		aToken,
		rToken,
	})
	if err != nil {
		log.Fatalf("Произошла ошибка формировании ответа во время создания токенов по ID: %v", err)
	}
}

func refresh(w http.ResponseWriter, r *http.Request) {
	var jwtKey, _ = os.LookupEnv("JWT_KEY")
	w.Header().Set("Content-Type", "application/json")
	var t Token
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		log.Fatalf("Произошла ошибка при чтении JSON во время обновления токенов: %v", err)
	}
	claims := jwt.MapClaims{}
	_, err = jwt.ParseWithClaims(t.AccessToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtKey), nil
	})
	if err != nil {
		log.Fatalf("Произошла ошибка при формировании JWT-токена: %v", err)
	}
	var (
		oldIP   = fmt.Sprintf("%v", claims["userIp"])
		curIP   = getIP(r)
		curGuid = fmt.Sprintf("%v", claims["userGuid"])
	)

	// сравниваем старый и новый IP
	if oldIP != curIP {
		log.Fatalf("IP не совпадают! Отправляем письмо на почту.")
		//TODO отправляем пиьсмо
	}

	// Проверяем refresh-токены
	decoded, err := base64.StdEncoding.DecodeString(t.RefreshToken)
	if err != nil {
		log.Fatalf("Произошла ошибка при кодировании в base64: %v", err)
	}
	err = bcrypt.CompareHashAndPassword([]byte(getHash(curGuid)), decoded)
	if err != nil {
		fmt.Println("Не совпали refresh-токены: ", err)
		err = json.NewEncoder(w).Encode(ErrorMessage{"Неверный refreshToken"})
		if err != nil {
			log.Fatalf("Произошла ошибка при отправке сообщении о несовпадении токенов")
		}
		return
	}

	// генерируем и отправляем новые токены
	aToken, rToken := generateAccess(curGuid, getIP(r)), generateRefresh(curGuid)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(Token{
		aToken,
		rToken,
	})
	if err != nil {
		log.Fatalf("Произошла ошибка при отправке новых токенов: %v", err)
	}
}

func generateAccess(guidParam string, ipParam string) string {
	var jwtKey, _ = os.LookupEnv("JWT_KEY")
	claims := MyClaims{
		RegisteredClaims: jwt.RegisteredClaims{},
		UserGuid:         guidParam,
		UserIp:           ipParam,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	accessTok, err := token.SignedString([]byte(jwtKey))
	if err != nil {
		log.Fatalf("Произошла ошибка при генерации access-токена: %v", err)
	}
	return accessTok
}

func generateRefresh(userGuid string) string {
	token := []byte(uuid.New().String())
	encoded := base64.StdEncoding.EncodeToString(token)
	bcryptHash, err := bcrypt.GenerateFromPassword(token, 14)
	if err != nil {
		log.Fatalf("Произошла ошибка при формировании bcryptHash: %v", err)
	}
	setHash(userGuid, string(bcryptHash))
	return encoded
}

func getIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}
