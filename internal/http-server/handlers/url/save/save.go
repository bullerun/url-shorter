package save

import (
	"REST-API-Service/internal/lib/api/response"
	"REST-API-Service/internal/lib/logger/sl"
	"REST-API-Service/internal/lib/random"
	"REST-API-Service/internal/storage"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}
type Response struct {
	response.Response
	Alias string `json:"alias,omitempty"`
}
type URLSaver interface {
	SaveUrl(urlToSave, alias string) (int64, error)
}

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var request Request

		err := render.DecodeJSON(r.Body, &request)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.JSON(w, r, response.Error("failed to decode request"))
			return
		}
		log.Info("request body decoded", slog.Any("request", request))
		if err := validator.New().Struct(request); err != nil {
			validate := err.(validator.ValidationErrors)

			log.Error("invalid request", sl.Err(err))

			render.JSON(w, r, response.ValidationError(validate))

			return
		}
		alias := request.Alias
		if alias == "" {
			alias = random.GetURL()
		}
		id, err := urlSaver.SaveUrl(request.URL, alias)
		if errors.Is(err, storage.ErrURLExists) {
			log.Error("url already exists", sl.Err(err))
			render.JSON(w, r, response.Error("url already exists"))
			return
		}
		if err != nil {
			log.Error("failed to add url", sl.Err(err))
			render.JSON(w, r, response.Error("failed to add url"))
			return
		}
		log.Info("url added", slog.Int64("id", id))
		render.JSON(w, r, Response{
			Response: response.OK(),
			Alias:    alias})
	}
}
