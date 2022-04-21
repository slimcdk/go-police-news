package police

import (
	"log"

	"github.com/go-resty/resty/v2"
)

type client struct {
	resty *resty.Client
	log   *log.Logger
}

// Base struct
func New() client {
	r := resty.New()
	r.SetBaseURL(basePath)
	return client{resty: r}
}
