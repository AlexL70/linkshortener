package viewmodels

type HelloResponse struct {
	Body *HelloResponseBody
}

type HelloResponseBody struct {
	Message string `json:"message"`
}
