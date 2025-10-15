package auth

type CurrentUser struct {
	UserID int
}

var currentUser = &CurrentUser{UserID: 1}

func GetCurrentUser() *CurrentUser {
	return currentUser
}

func SetCurrentUser(userID int) {
	currentUser.UserID = userID
}
