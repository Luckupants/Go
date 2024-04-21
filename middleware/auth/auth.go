//go:build !solution

package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"
)

type User struct {
	Name  string
	Email string
}

func ContextUser(ctx context.Context) (*User, bool) {
	user := ctx.Value("User")
	if user == nil {
		return nil, false
	}
	return user.(*User), true
}

var ErrInvalidToken = errors.New("invalid token")

type TokenChecker interface {
	CheckToken(ctx context.Context, token string) (*User, error)
}

func CheckAuth(checker TokenChecker) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			str := r.Header.Get("Authorization")
			token, found := strings.CutPrefix(str, "Bearer ")
			if !found {
				w.WriteHeader(http.StatusUnauthorized)
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
			user, err := checker.CheckToken(ctx, token)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if user == nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			r = r.Clone(context.WithValue(r.Context(), "User", user))
			next.ServeHTTP(w, r)
		})
	}
}
