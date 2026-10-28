package auth

import "net/http"

type contextKey string

const UserIDKey contextKey = "userID"

func GetUserIDFromContext(r *http.Request) (int64, bool) {
	userID, ok := r.Context().Value(UserIDKey).(int64)
	return userID, ok
}
