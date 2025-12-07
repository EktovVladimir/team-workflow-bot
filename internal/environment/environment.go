package environment

import "os"

var (
	IsDev  bool
	IsProd bool
	Env    string
)

func InitGlobal() {
	Env = os.Getenv("APP_ENV")

	if Env == "" {
		Env = "development"
	}

	switch Env {
	case "production":
		IsProd = true
	case "development":
		IsDev = true
	}
}
