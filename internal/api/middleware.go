package api

import (
	"code-garden-server/config"
	"code-garden-server/internal/database/models"
	"code-garden-server/internal/database/redis"
	"code-garden-server/internal/services/auth"
	"code-garden-server/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

type Middleware struct {
	Handler     func(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request, bool)
	PreHandler  func(path string)
	PostHandler func(w http.ResponseWriter, r *http.Request)
}

func NewLoggerMiddleware() Middleware {
	preHandler := func(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request, bool) {
		return w, r, true
	}

	postHandler := func(w http.ResponseWriter, r *http.Request) {
		status := w.Header().Get("Status")
		message := w.Header().Get("Message")
		path := r.URL.String()
		method := r.Method
		log.Printf("%s %s \t %s(%s)\n", method, path, message, status)
	}

	m := Middleware{PostHandler: postHandler, Handler: preHandler}
	return m
}

func NewCorsMiddleware(s *Server) Middleware {
	registeredPaths := map[string]bool{}
	// sets up options handler for every request
	preHandler := func(path string) {
		if _, ok := registeredPaths[path]; !ok {
			s.mux.HandleFunc(fmt.Sprintf("%s %s", http.MethodOptions, path), func(w http.ResponseWriter, r *http.Request) {
				setCorsHeaders(w, true)
			})
		}
		registeredPaths[path] = true
	}

	// handles every non-option http method
	handler := func(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request, bool) {
		setCorsHeaders(w, false)
		return w, r, true
	}

	return Middleware{Handler: handler, PreHandler: preHandler}
}

func NewAuthMiddleware(s *Server) Middleware {
	handler := func(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request, bool) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.WriteRes(w, utils.Response{Status: 401, Message: "Unauthorized! Invalid Token", Error: "invalid token"})
			return w, r, false
		}
		token := authHeader[len("Bearer "):]
		jwtToken, err := jwt.ParseWithClaims(token, &auth.CustomJWTClaims{}, func(*jwt.Token) (interface{}, error) {
			key := config.GetEnv("JWT_SECRET")
			return []byte(key), nil
		})

		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				utils.WriteRes(w, utils.Response{Status: 401, Message: "Unauthorized! Token Expired", Error: err.Error()})
			} else if errors.Is(err, jwt.ErrTokenMalformed) {
				utils.WriteRes(w, utils.Response{Status: 401, Message: "Unauthorized! Token Malformed", Error: err.Error()})
			} else {
				utils.WriteRes(w, utils.Response{Status: 401, Message: "Unauthorized!", Error: err.Error()})
			}
			return w, r, false
		} else if claims, ok := jwtToken.Claims.(*auth.CustomJWTClaims); !ok {
			utils.WriteRes(w, utils.Response{Status: 401, Message: "Unauthorized! Token Malformed", Error: "token malformed"})
			return w, r, false
		} else {
			q := redis.CacheKey{Entity: redis.UserEntity, Identifier: claims.User.ID.String()}

			var user models.User

			// TODO: Remove this line after the project has been dockerized.
			if s.rdc != nil {
				res, err := s.rdc.Get(context.Background(), q.String()).Result()
				if err != nil {
					log.Println("error getting user from cache")
				}

				err = json.Unmarshal([]byte(res), &user)
				if err != nil {
					log.Println("Failed to unmarshal user from redis cache", res)
				} else {
					log.Println("Got User from redis cache")
					ctx := context.WithValue(r.Context(), "User", &user)
					return w, r.WithContext(ctx), true
				}
			}

			// fetch the user from the database
			user = models.User{BaseModel: models.BaseModel{ID: claims.User.ID}, Email: claims.User.Email}
			if tx := s.db.First(&user, "id = ?", claims.User.ID); tx.Error != nil {
				utils.WriteRes(w, utils.Response{Status: 401, Message: "Unauthorized! Token Malformed", Error: "token malformed"})
				return w, r, false
			} else {
				ctx := context.WithValue(r.Context(), "User", &user)
				return w, r.WithContext(ctx), true
			}
		}
	}

	return Middleware{
		Handler: handler,
	}
}

func setCorsHeaders(w http.ResponseWriter, isOptions bool) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if isOptions {
		w.WriteHeader(http.StatusNoContent)
	}
}
