package handlers_test

import (
"bytes"
"context"
"encoding/json"
"net/http"
"net/http/httptest"
"os"
"testing"

"github.com/gin-gonic/gin"

"app/internal/auth"
"app/internal/http/dto"
"app/internal/http/handlers"
"app/internal/store/postgres"
)

var testStore *postgres.Store
var testJWTSecret = "test-secret-key-for-testing-only"

func TestMain(m *testing.M) {
dbURL := os.Getenv("DATABASE_URL")
if dbURL == "" {
dbURL = "postgres://finance:finance@localhost:5488/finance?sslmode=disable"
}

var err error
testStore, err = postgres.NewStore(dbURL)
if err != nil {
panic(err)
}

cleanupDB()

code := m.Run()

cleanupDB()
testStore.Close()
os.Exit(code)
}

func cleanupDB() {
ctx := context.Background()
testStore.DB().ExecContext(ctx, "DELETE FROM user_settings")
testStore.DB().ExecContext(ctx, "DELETE FROM users")
}

func setupRouter(store *postgres.Store, jwtSecret string) *gin.Engine {
gin.SetMode(gin.TestMode)
router := gin.New()

authHandler := handlers.NewAuthHandler(store, jwtSecret)

router.POST("/api/auth/register", authHandler.Register)
router.POST("/api/auth/login", authHandler.Login)
router.POST("/api/auth/logout", authHandler.Logout)

return router
}

func TestRegisterHappyPath(t *testing.T) {
cleanupDB()
router := setupRouter(testStore, testJWTSecret)

reqBody := dto.RegisterRequest{
Email:    "test@example.com",
Password: "password123",
}
body, _ := json.Marshal(reqBody)

req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewReader(body))
req.Header.Set("Content-Type", "application/json")
w := httptest.NewRecorder()

router.ServeHTTP(w, req)

if w.Code != http.StatusCreated {
t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
}

var response map[string]interface{}
json.Unmarshal(w.Body.Bytes(), &response)

if response["user"] == nil {
t.Error("Expected user in response")
}

user := response["user"].(map[string]interface{})
if user["email"] != "test@example.com" {
t.Errorf("Expected email test@example.com, got %v", user["email"])
}

cookies := w.Result().Cookies()
var authToken string
for _, cookie := range cookies {
if cookie.Name == "auth_token" {
authToken = cookie.Value
break
}
}
if authToken == "" {
t.Error("Expected auth_token cookie to be set")
}
}

func TestRegisterDuplicateEmail(t *testing.T) {
cleanupDB()
router := setupRouter(testStore, testJWTSecret)

reqBody := dto.RegisterRequest{
Email:    "duplicate@example.com",
Password: "password123",
}
body, _ := json.Marshal(reqBody)

req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewReader(body))
req.Header.Set("Content-Type", "application/json")
w := httptest.NewRecorder()
router.ServeHTTP(w, req)

if w.Code != http.StatusCreated {
t.Fatalf("First registration failed: %s", w.Body.String())
}

req, _ = http.NewRequest("POST", "/api/auth/register", bytes.NewReader(body))
req.Header.Set("Content-Type", "application/json")
w = httptest.NewRecorder()
router.ServeHTTP(w, req)

if w.Code != http.StatusConflict {
t.Errorf("Expected status %d for duplicate email, got %d", http.StatusConflict, w.Code)
}
}

func TestLoginHappyPath(t *testing.T) {
cleanupDB()
router := setupRouter(testStore, testJWTSecret)

email := "login@example.com"
password := "password123"

hashedPassword, _ := auth.HashPassword(password)
testStore.CreateUser(context.Background(), email, hashedPassword)

reqBody := dto.LoginRequest{
Email:    email,
Password: password,
}
body, _ := json.Marshal(reqBody)

req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
req.Header.Set("Content-Type", "application/json")
w := httptest.NewRecorder()

router.ServeHTTP(w, req)

if w.Code != http.StatusOK {
t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
}

var response map[string]interface{}
json.Unmarshal(w.Body.Bytes(), &response)

if response["user"] == nil {
t.Error("Expected user in response")
}

cookies := w.Result().Cookies()
var authToken string
for _, cookie := range cookies {
if cookie.Name == "auth_token" {
authToken = cookie.Value
break
}
}
if authToken == "" {
t.Error("Expected auth_token cookie to be set after login")
}
}

func TestLoginInvalidCredentials(t *testing.T) {
cleanupDB()
router := setupRouter(testStore, testJWTSecret)

email := "invalid@example.com"
password := "password123"
hashedPassword, _ := auth.HashPassword(password)
testStore.CreateUser(context.Background(), email, hashedPassword)

reqBody := dto.LoginRequest{
Email:    email,
Password: "wrongpassword",
}
body, _ := json.Marshal(reqBody)

req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
req.Header.Set("Content-Type", "application/json")
w := httptest.NewRecorder()

router.ServeHTTP(w, req)

if w.Code != http.StatusUnauthorized {
t.Errorf("Expected status %d for invalid credentials, got %d", http.StatusUnauthorized, w.Code)
}
}

func TestLogoutHappyPath(t *testing.T) {
router := setupRouter(testStore, testJWTSecret)

req, _ := http.NewRequest("POST", "/api/auth/logout", nil)
w := httptest.NewRecorder()

router.ServeHTTP(w, req)

if w.Code != http.StatusOK {
t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
}

cookies := w.Result().Cookies()
var authToken string
for _, cookie := range cookies {
if cookie.Name == "auth_token" {
authToken = cookie.Value
break
}
}
if authToken != "" {
t.Error("Expected auth_token cookie to be cleared")
}
}

func TestPasswordNotStoredInPlainText(t *testing.T) {
cleanupDB()

email := "plaintext@example.com"
password := "password123"

hashedPassword, _ := auth.HashPassword(password)
user, err := testStore.CreateUser(context.Background(), email, hashedPassword)
if err != nil {
t.Fatal(err)
}

var storedHash string
err = testStore.DB().QueryRow("SELECT password_hash FROM users WHERE id = $1", user.ID).Scan(&storedHash)
if err != nil {
t.Fatal(err)
}

if storedHash == password {
t.Error("Password should not be stored in plain text")
}

err = auth.CheckPasswordHash(password, storedHash)
if err != nil {
t.Errorf("Password hash verification failed: %v", err)
}
}
