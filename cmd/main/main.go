package main

import (
	"fmt"
	"net/http"
	middlewares "stargazer/video-recording/internal/api/MiddleWares"
	"stargazer/video-recording/internal/api/handlers"
	"stargazer/video-recording/internal/enums"
	initializers "stargazer/video-recording/internal/initilizers"
	"strconv"

	"github.com/go-chi/chi/v5"
)

var ListenPort = 4001

func init() {
	initializers.Initialize_backend("../../config.env")
}

func main() {
	router := chi.NewRouter()

	router.Use(middlewares.RequestLogger)
	router.Use(middlewares.CORS)
	router.Use(middlewares.Timeout)
	router.Use(middlewares.PanicHandler)

	router.Group(func(r chi.Router) {
		r.Get("/ping", handlers.Pong)
	})

	router.Group(func(r chi.Router) {
		r.Post("/create_room", handlers.CreateRoom)
		r.Post("/close_room", handlers.CloseRoom)
		r.Post("/delete_all_rooms", handlers.DeleteAllRoom)
		r.Get("/get_all_rooms", handlers.GetAllRooms)
		r.Post("/get_all_recordings", handlers.GetAllRecordings)
		r.Post("/media_convert_callback", handlers.MediaConvertCallback)
	})

	router.Group(func(r chi.Router) {
		r.Post("/status_call_back", handlers.RecordingStatusCallback)
	})

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("route does not exist"))
	})

	router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(405)
		w.Write([]byte("method is not valid"))
	})

	fmt.Println(enums.Terminalbold + enums.TerminalcCyan + "Stargazer Video Service Listening on port: " + enums.Terminalitalic + strconv.Itoa(ListenPort) + enums.Terminalreset)
	fmt.Println(enums.Terminalbold + enums.TerminalcRed + "Check Directory:" + enums.TerminalcWhite + enums.Terminalitalic + " logs/" + enums.Terminalreset + " for detailed logging and Error reports" + enums.Terminalreset)

	http.ListenAndServe(":"+strconv.Itoa(ListenPort), router)
}
