package auth

type ErrUnauthenticated struct{}

func (*ErrUnauthenticated) Error() string {
	return "unauthenticated"
}

type User struct {
	UserID    int64    `json:"userid"`
	Username  string   `json:"username"`
	Privleges []string `json:"privleges"`
	Features  []string `json:"features"`
}
