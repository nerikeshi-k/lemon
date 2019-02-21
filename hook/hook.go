package hook

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nerikeshi-k/lemon/config"
)

// Serve APIサーバ向けのWebhookを開く
func Serve() {
	handler := func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		id := query.Get("room")
		if id == "" || len(id) > 128 {
			w.WriteHeader(404)
			w.Write([]byte("error"))
			return
		}
		w.Write([]byte("ok"))
	}

	rtr := mux.NewRouter()
	rtr.HandleFunc("/", handler)
	hdlr := http.NewServeMux()
	hdlr.Handle("/", rtr)
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", config.BoundAddress, config.WebhookPort), hdlr)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
