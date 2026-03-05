package routes

import (
	"context"
	"net/http"

	"github.com/AlexL70/linkshortener/backend/web/viewmodels"
	"github.com/danielgtaylor/huma/v2"
)

func RegisterHello(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "hello",
		Method:      http.MethodGet,
		Path:        "/hello",
		Summary:     "Hello World",
	}, helloHandler)
}

func helloHandler(_ context.Context, _ *struct{}) (*viewmodels.HelloResponse, error) {
	return &viewmodels.HelloResponse{Body: &viewmodels.HelloResponseBody{Message: "Hello World"}}, nil
}
