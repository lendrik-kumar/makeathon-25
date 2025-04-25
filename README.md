# Second-Hand Product Verification System Implementation

This document provides a detailed implementation of a web platform to verify the background history, ownership, and details of second-hand products using Go (with Gin and Gorm) for the backend and React for the frontend. The system uses a centralized PostgreSQL database with hash chains for data integrity, avoiding blockchain complexity. It supports product registration, ownership transfer, event logging, history verification, and warranty management, with role-based access control.

## Backend Implementation (Go with Gin and Gorm)

### Project Structure
```
backend/
├── main.go
├── models/
│   ├── user.go
│   ├── product.go
│   ├── event.go
│   ├── pending_transfer.go
├── controllers/
│   ├── user_controller.go
│   ├── product_controller.go
│   ├── event_controller.go
├── middlewares/
│   ├── auth.go
├── utils/
│   ├── hash.go
├── go.mod
```

### Database Setup
Use PostgreSQL as the database. Install dependencies:
```bash
go get -u github.com/gin-gonic/gin
go get -u gorm.io/gorm
go get -u gorm.io/driver/postgres
go get -u golang.org/x/crypto/bcrypt
go get -u github.com/golang-jwt/jwt/v4
```

### Database Models
Define Gorm models for users, products, events, and pending transfers.

**models/user.go**
```go
package models

import "gorm.io/gorm"

type User struct {
    gorm.Model
    Username     string `gorm:"unique"`
    PasswordHash string
    Role         string // regular, brand, repair_shop, admin
}
```

**models/product.go**
```go
package models

import "gorm.io/gorm"

type Product struct {
    gorm.Model
    SerialNumber  string `gorm:"unique"`
    Manufacturer  string
    Model         string
}
```

**models/event.go**
```go
package models

import "gorm.io/gorm"

type Event struct {
    gorm.Model
    ProductID         uint
    EventType        string // registration, ownership_transfer, repair, warranty_update
    EventData        string // JSON or text
    PreviousEventHash string
    EventHash        string
    CreatedBy        uint // User ID
}
```

**models/pending_transfer.go**
```go
package models

import "gorm.io/gorm"

type PendingTransfer struct {
    gorm.Model
    ProductID   uint
    NewOwnerID  uint
}
```

### Utility Functions
Implement a function to compute event hashes using SHA-256.

**utils/hash.go**
```go
package utils

import (
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
)

type EventHashData struct {
    ProductID         uint
    EventType        string
    EventData        string
    CreatedAt        time.Time
    CreatedBy        uint
    PreviousEventHash string
}

func ComputeEventHash(data EventHashData) (string, error) {
    jsonData, err := json.Marshal(data)
    if err != nil {
        return "", err
    }
    hash := sha256.Sum256(jsonData)
    return hex.EncodeToString(hash[:]), nil
}
```

### Authentication Middleware
Implement JWT-based authentication.

**middlewares/auth.go**
```go
package middlewares

import (
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v4"
    "net/http"
    "strings"
)

var jwtSecret = []byte("your-secret-key")

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }

        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header"})
            c.Abort()
            return
        }

        token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
            return jwtSecret, nil
        })

        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }

        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
            c.Abort()
            return
        }

        c.Set("user_id", uint(claims["user_id"].(float64)))
        c.Set("role", claims["role"].(string))
        c.Next()
    }
}
```

### API Controllers
Implement handlers for user management, product operations, and event logging.

**controllers/user_controller.go**
```go
package controllers

import (
    "github.com/gin-gonic/gin"
    "golang.org/x/crypto/bcrypt"
    "gorm.io/gorm"
    "net/http"
    "your_project/models"
    "github.com/golang-jwt/jwt/v4"
    "time"
)

var db *gorm.DB
var jwtSecret = []byte("your-secret-key")

func InitUserController(database *gorm.DB) {
    db = database
}

type RegisterInput struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
    Role     string `json:"role" binding:"required"`
}

func RegisterUser(c *gin.Context) {
    var input RegisterInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
        return
    }

    user := models.User{
        Username:     input.Username,
        PasswordHash: string(hash),
        Role:         input.Role,
    }

    if err := db.Create(&user).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "User registered"})
}

type LoginInput struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}

func LoginUser(c *gin.Context) {
    var input LoginInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var user models.User
    if err := db.Where("username = ?", input.Username).First(&user).Error; err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id": user.ID,
        "role":    user.Role,
        "exp":     time.Now().Add(time.Hour * 24).Unix(),
    })

    tokenString, err := token.SignedString(jwtSecret)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"token": tokenString})
}
```

**controllers/product_controller.go**
```go
package controllers

import (
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "net/http"
    "your_project/models"
)

func InitProductController(database *gorm.DB) {
    db = database
}

type ProductInput struct {
    SerialNumber string `json:"serial_number" binding:"required"`
    Manufacturer string `json:"manufacturer" binding:"required"`
    Model        string `json:"model" binding:"required"`
}

func RegisterProduct(c *gin.Context) {
    role, _ := c.Get("role")
    if role != "brand" {
        c.JSON(http.StatusForbidden, gin.H{"error": "Only brands can register products"})
        return
    }

    var input ProductInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    product := models.Product{
        SerialNumber: input.SerialNumber,
        Manufacturer: input.Manufacturer,
        Model:        input.Model,
    }

    if err := db.Create(&product).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
        return
    }

    userID, _ := c.Get("user_id")
    event := models.Event{
        ProductID:  product.ID,
        EventType:  "registration",
        EventData:  `{"details": "Product registered"}`,
        CreatedBy:  userID.(uint),
    }

    if err := createEvent(&event); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log registration event"})
        return
    }

    c.JSON(http.StatusOK, product)
}

func GetProduct(c *gin.Context) {
    id := c.Param("id")
    var product models.Product
    if err := db.First(&product, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
        return
    }

    var events []models.Event
    db.Where("product_id = ?", id).Order("created_at asc").Find(&events)

    c.JSON(http.StatusOK, gin.H{
        "product": product,
        "history": events,
    })
}
```

**controllers/event_controller.go**
```go
package controllers

import (
    "encoding/json"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "net/http"
    "your_project/models"
    "your_project/utils"
    "errors"
)

func InitEventController(database *gorm.DB) {
    db = database
}

type EventInput struct {
    EventType string `json:"event_type" binding:"required"`
    EventData string `json:"event_data" binding:"required"`
}

func CreateEvent(c *gin.Context) {
    productID := c.Param("id")
    role, _ := c.Get("role")

    var input EventInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if input.EventType == "repair" && role != "repair_shop" {
        c.JSON(http.StatusForbidden, gin.H{"error": "Only repair shops can log repairs"})
        return
    }

    var lastEvent models.Event
    err := db.Where("product_id = ?", productID).Order("created_at desc").First(&lastEvent).Error
    previousHash := ""
    if err != nil {
        if !errors.Is(err, gorm.ErrRecordNotFound) {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
    } else {
        previousHash = lastEvent.EventHash
    }

    userID, _ := c.Get("user_id")
    event := models.Event{
        ProductID:         productID,
        EventType:        input.EventType,
        EventData:        input.EventData,
        PreviousEventHash: previousHash,
        CreatedBy:        userID.(uint),
    }

    if err := db.Create(&event).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event"})
        return
    }

    hashData := utils.EventHashData{
        ProductID:         event.ProductID,
        EventType:        event.EventType,
        EventData:        event.EventData,
        CreatedAt:        event.CreatedAt,
        CreatedBy:        event.CreatedBy,
        PreviousEventHash: event.PreviousEventHash,
    }

    eventHash, err := utils.ComputeEventHash(hashData)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to compute hash"})
        return
    }

    event.EventHash = eventHash
    if err := db.Save(&event).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update event hash"})
        return
    }

    c.JSON(http.StatusOK, event)
}

func VerifyProductHistory(c *gin.Context) {
    productID := c.Param("id")
    var events []models.Event
    if err := db.Where("product_id = ?", productID).Order("created_at asc").Find(&events).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    for i, event := range events {
        if i == 0 && event.PreviousEventHash != "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid previous hash for first event"})
            return
        } else if i > 0 && event.PreviousEventHash != events[i-1].EventHash {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Hash chain broken at event " + string(event.ID)})
            return
        }

        hashData := utils.EventHashData{
            ProductID:         event.ProductID,
            EventType:        event.EventType,
            EventData:        event.EventData,
            CreatedAt:        event.CreatedAt,
            CreatedBy:        event.CreatedBy,
            PreviousEventHash: event.PreviousEventHash,
        }

        expectedHash, err := utils.ComputeEventHash(hashData)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to compute hash"})
            return
        }

        if event.EventHash != expectedHash {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hash for event " + string(event.ID)})
            return
        }
    }

    c.JSON(http.StatusOK, gin.H{"message": "History is valid"})
}
```

### Ownership Transfer
Implement a two-step ownership transfer process.

**controllers/product_controller.go (continued)**
```go
type TransferInput struct {
    NewOwnerUsername string `json:"new_owner_username" binding:"required"`
}

func InitiateTransfer(c *gin.Context) {
    productID := c.Param("id")
    userID, _ := c.Get("user_id")

    var product models.Product
    if err := db.First(&product, productID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
        return
    }

    var lastEvent models.Event
    if err := db.Where("product_id = ? AND event_type = ?", productID, "ownership_transfer").Order("created_at desc").First(&lastEvent).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    var currentOwnerID uint
    if lastEvent.ID != 0 {
        var eventData map[string]interface{}
        json.Unmarshal([]byte(lastEvent.EventData), &eventData)
        currentOwnerID = uint(eventData["new_owner_id"].(float64))
    } else {
        // Assume initial owner is the one who registered it
        var regEvent models.Event
        if err := db.Where("product_id = ? AND event_type = ?", productID, "registration").First(&regEvent).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "No registration event found"})
            return
        }
        currentOwnerID = regEvent.CreatedBy
    }

    if currentOwnerID != userID.(uint) {
        c.JSON(http.StatusForbidden, gin.H{"error": "Only the current owner can initiate a transfer"})
        return
    }

    var input TransferInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var newOwner models.User
    if err := db.Where("username = ?", input.NewOwnerUsername).First(&newOwner).Error; err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "New owner not found"})
        return
    }

    pendingTransfer := models.PendingTransfer{
        ProductID:  product.ID,
        NewOwnerID: newOwner.ID,
    }

    if err := db.Create(&pendingTransfer).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate transfer"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Transfer initiated"})
}

func ConfirmTransfer(c *gin.Context) {
    productID := c.Param("id")
    userID, _ := c.Get("user_id")

    var pendingTransfer models.PendingTransfer
    if err := db.Where("product_id = ? AND new_owner_id = ?", productID, userID).First(&pendingTransfer).Error; err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "No pending transfer found"})
        return
    }

    event := models.Event{
        ProductID:  pendingTransfer.ProductID,
        EventType:  "ownership_transfer",
        EventData:  fmt.Sprintf(`{"new_owner_id": %d}`, pendingTransfer.NewOwnerID),
        CreatedBy:  userID.(uint),
    }

    if err := createEvent(&event); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log transfer event"})
        return
    }

    if err := db.Delete(&pendingTransfer).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete pending transfer"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Transfer confirmed"})
}
```

### Main Application
Tie everything together in the main file.

**main.go**
```go
package main

import (
    "github.com/gin-gonic/gin"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "your_project/controllers"
    "your_project/middlewares"
    "your_project/models"
)

var db *gorm.DB

func main() {
    dsn := "host=localhost user=gorm password=gorm dbname=gorm port=5432 sslmode=disable"
    var err error
    db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }

    db.AutoMigrate(&models.User{}, &models.Product{}, &models.Event{}, &models.PendingTransfer{})

    r := gin.Default()

    controllers.InitUserController(db)
    controllers.InitProductController(db)
    controllers.InitEventController(db)

    r.POST("/api/users/register", controllers.RegisterUser)
    r.POST("/api/users/login", controllers.LoginUser)

    authorized := r.Group("/").Use(middlewares.AuthMiddleware())
    {
        authorized.POST("/api/products", controllers.RegisterProduct)
        authorized.GET("/api/products/:id", controllers.GetProduct)
        authorized.POST("/api/products/:id/events", controllers.CreateEvent)
        authorized.POST("/api/products/:id/transfer", controllers.InitiateTransfer)
        authorized.POST("/api/products/:id/transfer/confirm", controllers.ConfirmTransfer)
        authorized.GET("/api/products/:id/verify", controllers.VerifyProductHistory)
    }

    r.Run(":8080")
}
```

## Frontend Implementation (React)

### Project Structure
```
frontend/
├── public/
│   ├── index.html
├── src/
│   ├── components/
│   │   ├── Navbar.js
│   ├── pages/
│   │   ├── Login.js
│   │   ├── Register.js
│   │   ├── ProductDetails.js
│   ├── services/
│   │   ├── api.js
│   ├── context/
│   │   ├── AuthContext.js
│   ├── App.js
│   ├── index.js
├── package.json
```

### Setup
Create a React project and install dependencies:
```bash
npx create-react-app frontend
cd frontend
npm install axios react-router-dom tailwindcss
npx tailwindcss init
```

Configure Tailwind in `tailwind.config.js`:
```js
module.exports = {
  content: ["./src/**/*.{js,jsx}"],
  theme: { extend: {} },
  plugins: [],
}
```

Include Tailwind in `src/index.css`:
```css
@tailwind base;
@tailwind components;
@tailwind utilities;
```

### Main HTML
**public/index.html**
```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Second-Hand Product Verification</title>
    <script src="https://cdn.jsdelivr.net/npm/react@17/umd/react.production.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/react-dom@17/umd/react-dom.production.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/babel-standalone@6.26.0/babel.min.js"></script>
</head>
<body>
    <div id="root"></div>
</body>
</html>
```

### Authentication Context
Manage user authentication state.

**src/context/AuthContext.js**
```js
import React, { createContext, useState, useEffect } from 'react';

export const AuthContext = createContext();

export const AuthProvider = ({ children }) => {
    const [user, setUser] = useState(null);

    useEffect(() => {
        const token = localStorage.getItem('token');
        if (token) {
            // Decode token or fetch user info
            setUser({ token });
        }
    }, []);

    const login = (token) => {
        localStorage.setItem('token', token);
        setUser({ token });
    };

    const logout = () => {
        localStorage.removeItem('token');
        setUser(null);
    };

    return (
        <AuthContext.Provider value={{ user, login, logout }}>
            {children}
        </AuthContext.Provider>
    );
};
```

### API Service
Handle API requests with Axios.

**src/services/api.js**
```js
import axios from 'axios';

const api = axios.create({
    baseURL: 'http://localhost:8080/api',
});

api.interceptors.request.use((config) => {
    const token = localStorage.getItem('token');
    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
});

export const register = (data) => api.post('/users/register', data);
export const login = (data) => api.post('/users/login', data);
export const registerProduct = (data) => api.post('/products', data);
export const getProduct = (id) => api.get(`/products/${id}`);
export const createEvent = (id, data) => api.post(`/products/${id}/events`, data);
export const initiateTransfer = (id, data) => api.post(`/products/${id}/transfer`, data);
export const confirmTransfer = (id) => api.post(`/products/${id}/transfer/confirm`);
export const verifyHistory = (id) => api.get(`/products/${id}/verify`);
```

### Main App
Set up routing.

**src/App.js**
```js
import React from 'react';
import { BrowserRouter as Router, Route, Switch } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import Navbar from './components/Navbar';
import Login from './pages/Login';
import Register from './pages/Register';
import ProductDetails from './pages/ProductDetails';

function App() {
    return (
        <AuthProvider>
            <Router>
                <div className="min-h-screen bg-gray-100">
                    <Navbar />
                    <Switch>
                        <Route path="/login" component={Login} />
                        <Route path="/register" component={Register} />
                        <Route path="/products/:id" component={ProductDetails} />
                    </Switch>
                </div>
            </Router>
        </AuthProvider>
    );
}

export default App;
```

### Navbar Component
**src/components/Navbar.js**
```js
import React, { useContext } from 'react';
import { Link } from 'react-router-dom';
import { AuthContext } from '../context/AuthContext';

function Navbar() {
    const { user, logout } = useContext(AuthContext);

    return (
        <nav className="bg-blue-600 p-4">
            <div className="
