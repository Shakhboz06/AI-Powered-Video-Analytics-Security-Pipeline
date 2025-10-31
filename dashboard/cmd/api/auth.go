package api

import (
	"net/http"
	"video-analytics-pipe/dashboard/internal/auth"
	"video-analytics-pipe/dashboard/internal/store"
	"github.com/gin-gonic/gin"
)

type RequestPayload struct {
	Username string `json:"username" validate:"required,max=100"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=6,max=72"`
}
type LoginPayload struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=6,max=72"`
}
type UserPayload struct {
	UserID   int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}
type ResponsePayload struct {
	User  UserPayload `json:"user"`
	Token string      `json:"token"`
}

func UserRegister(s *store.UserStore, a *auth.JWTAuthenticator) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		
		var payload RequestPayload
		if err := ctx.BindJSON(&payload); err != nil {
			ctx.JSON(400, gin.H{"error": "invalid payload"})
			return
		}
		
		user := &store.Users{
			Username: payload.Username,
			Email:    payload.Email,
		}
		
		if err := user.Password.Set(payload.Password); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid credentials"})
			return
		}
		
	
		users, err := s.Create(ctx.Request.Context(), user)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not save user"})
			return
		}

		token_version, err := s.UpdateUserTokenVersion(ctx, users.ID)
		if err != nil{
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update token version"})
			return 
		}

		token, err := a.GenerateToken(user.ID, token_version.Version)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not sign token"})
			return
		}

		ctx.JSON(http.StatusCreated, ResponsePayload{
			User: UserPayload{
				UserID:   int(users.ID),
				Username: users.Username,
				Email:    users.Email,
			},
			Token: token,
		})

	}
}

func UserLogin(s *store.UserStore, a *auth.JWTAuthenticator) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var payload LoginPayload
		if err := ctx.BindJSON(&payload); err != nil {
			ctx.JSON(400, gin.H{"error": "invalid payload"})
			return
		}

		user, err := s.GetUser(ctx.Request.Context(), payload.Email)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}

		if !user.Password.Compare(payload.Password) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "incorrect email or password"})
			return
		}

		token_version, err := s.UpdateUserTokenVersion(ctx, user.ID)
		if err != nil{
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update token version"})
			return 
		}

		token, err := a.GenerateToken(user.ID, token_version.Version)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not sign token"})
			return
		}

		ctx.JSON(http.StatusOK, ResponsePayload{
			User: UserPayload{
				UserID:   int(user.ID),
				Username: user.Username,
				Email:    user.Email,
			},
			Token: token,
		})

	}
}

func LoginToDashboard(s *store.UserStore) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Step 1: Get userID from context (set by middleware)
		userIDRaw, exists := ctx.Get("userID")
		if !exists {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		// Step 2: Convert string to int64
		userID, ok := userIDRaw.(int64)
		if !ok {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
			return
		}

		// Step 3: Use your GetUserByID function from the store
		user, err := s.GetUserByID(ctx.Request.Context(), userID)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}

		// Step 4: Return the user
		ctx.JSON(http.StatusOK, gin.H{
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
			},
		})
	}
}

// type UserCreds struct{
// 	Username string `json:"username"`
// 	Password string `json:"password"`
// }

// func UserLoginToDashboard() gin.HandlerFunc {
// 	return func(ctx *gin.Context){
// 		var creds UserCreds
// 		if err := ctx.BindJSON(&creds); err != nil{
// 			ctx.JSON(400, gin.H{"error": "invalid payload"})
// 			return
// 		}

// 		if creds.Username == "admin" && creds.Password == "111111"{
// 			token := "hardcodedjwtthenwillbereplacedwiththerealoneinprodcution123890"
// 			ctx.JSON(http.StatusOK, gin.H{"token": token})
// 		}else{
// 			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials" })
// 		}
// 	}
// }

// func UserRegister() gin.HandlerFunc {
// 	return func(ctx *gin.Context){
// 		var creds UserCreds
// 		if err := ctx.BindJSON(&creds); err != nil{
// 			ctx.JSON(400, gin.H{"error": "invalid payload"})
// 			return
// 		}

// 		if creds.Username == "admin" && creds.Password == "111111"{
// 			token := "hardcodedjwtthenwillbereplacedwiththerealoneinprodcution123890"
// 			ctx.JSON(http.StatusOK, gin.H{"token": token})
// 		}else{
// 			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials" })
// 		}
// 	}
// }
