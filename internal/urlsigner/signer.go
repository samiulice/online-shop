package urlsigner

import (
	"fmt"
	"strings"
	"time"

	goalone "github.com/bwmarrin/go-alone"
)

type Signer struct {
	Secret []byte
}

// GenerateTokenFromString generates a token for url signer
func (s *Signer) GenerateTokenFromString(data string) string {
	var urlToSign string

	crypt := goalone.New(s.Secret, goalone.Timestamp)

	if strings.Contains(data, "?") {
		urlToSign = fmt.Sprintf("%s&hash=", data)
	} else {
		urlToSign = fmt.Sprintf("%s?hash=", data)
	}

	tokenBytes := crypt.Sign([]byte(urlToSign))
	token := string(tokenBytes)
	return token
}

// VerifyToken verify the token whether it has changed or not
func (s *Signer) VerifyToken(token string) bool {
	crypt := goalone.New(s.Secret, goalone.Timestamp)

	_, err := crypt.Unsign([]byte(token))
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

// VerifyToken verify the token whether it has changed or not
func (s *Signer) Expired(token string, minutesUntilExpied int) bool {
	crypt := goalone.New(s.Secret, goalone.Timestamp)

	ts := crypt.Parse([]byte(token))
	return time.Since(ts.Timestamp) > time.Duration(minutesUntilExpied)*time.Minute
}
