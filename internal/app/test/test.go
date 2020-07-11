package test

import (
	"context"
	"net/http"
	"siody.home/om-like/internal/appmain"
)

func Bind(p *appmain.Params, b *appmain.Bindings) error {
	b.AddHealthCheckFunc(func(ctx context.Context) error {
		return nil
	})
	b.AddHandleFunc(
		func(mux *http.ServeMux) {
			mux.HandleFunc("/ping", func(writer http.ResponseWriter, _ *http.Request) {
				writer.Write([]byte("pong"))
			})
		})
	return nil
}
