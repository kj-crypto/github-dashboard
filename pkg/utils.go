package contribution

import "os"

func GetToken() string {
	return os.Getenv("GITHUB_TOKEN")
}
