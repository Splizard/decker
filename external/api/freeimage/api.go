package freeimage

import (
	"context"

	"runtime.link/api"
)

type API struct {
	api.Specification `www:"https://freeimage.host"`

	Upload func(context.Context, Data) (Upload, error) `rest:"POST(multipart/form-data) /api/1/upload image"`
}

type Data struct {
	Key    string `json:"key"`
	Action string `json:"action"`
	Source string `json:"source"`
	Format string `json:"format,omitempty"`
}

type Upload struct {
	Image struct {
		URL string `json:"url"`
	}
}
