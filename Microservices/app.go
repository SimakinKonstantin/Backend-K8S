package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"lab1/metrics"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type App struct {
	Database *dbProcessor
	Redis    *redis.Client
	Producer *kafka.Writer
}

// @Summary        Получить все программы тренировок
// @Description    Возвращает список программ
// @Tags 		   Программы тренировок
// @Produce        json
// @Success        200 {object} Program  "Успешно получены все программы тренировок"
// @Failure        422 {object} Error "Неподдерживаемые данные"
// @Router         /programs [get]
func (app *App) showAllHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	statusCode := http.StatusOK

	defer func() {
		metrics.ObserveRequest(time.Since(start), statusCode)
	}()

	slog.Info("/programs GET")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	allPrograms, err := app.Database.getAllPrograms()
	if err != nil {
		metrics.ErrorMetrics.Inc()
		statusCode = http.StatusUnprocessableEntity
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(Error{"неподдерживаемые данные"})
		return
	}
	json.NewEncoder(w).Encode(allPrograms)
}

// @Summary        Получить программу тренировок
// @Description    Возвращает программу тренировок по индексу
// @Tags 		   Программы тренировок
// @Produce        json
// @Param          id   query  string  true  "Идентификатор программы"
// @Success        200 {object} Program  "Успешно получена программа тренировок"
// @Failure        404 {object} Error "Нет элемента с указанным идентификатором"
// @Failure        422 {object} Error "Неподдерживаемые данные
// @Router         /programs/ [get]
func (app *App) showHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	statusCode := http.StatusOK

	defer func() {
		metrics.ObserveRequest(time.Since(start), statusCode)
	}()

	slog.Info("/programs/ GET")

	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	index := r.URL.Query().Get("id")

	// Получение из кеша.
	unmarshalledProgram, err := app.Redis.Get(context.TODO(), index).Result()
	if err == nil {
		slog.Info("Получили программу их кеша")

		// Нашлось в кеше. Оно хранится там в виде строки.
		var p CachedProgram
		json.Unmarshal([]byte(unmarshalledProgram), &p)
		json.NewEncoder(w).Encode(p)
		return

	} else {
		program, err := app.Database.getProgram(index)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				statusCode = http.StatusNotFound
				w.WriteHeader(statusCode)
				metrics.ErrorMetrics.Inc()
				json.NewEncoder(w).Encode(Error{fmt.Sprintf("нет элемента с указанным id = %s", index)})

			} else {
				statusCode = http.StatusUnprocessableEntity
				w.WriteHeader(statusCode)
				metrics.ErrorMetrics.Inc()
				json.NewEncoder(w).Encode(Error{"неподдерживаемые данные"})
			}
			return
		}

		// Положили в кеш.
		marshalled_bson, _ := bson.Marshal(program)
		var p CachedProgram
		bson.Unmarshal(marshalled_bson, &p)
		_, err = app.Redis.Set(context.TODO(), index, p, 5*time.Minute).Result()
		if err != nil {
			slog.Info(err.Error())
		}
		json.NewEncoder(w).Encode(program)
	}
	w.WriteHeader(statusCode)
}

// @Summary        Обновить программу тренировок
// @Description    Возвращает программу тренировок по индексу
// @Tags 		   Программы тренировок
// @Produce        json
// @Param          id   query  string  true  "Идентификатор обновляемой программы"
// @Param		   Name formData string false "Название программы"
// @Param		   Description formData string false "Описание программы"
// @Param		   Price formData int false "Стоимость программы"
// @Success        200 "Успешно обновлена"
// @Failure        400 {object} Error "Неверное значение цены"
// @Failure        404 {object} Error "Нет элемента с таким идентификатором"
// @Failure        422 {object} Error "Неподдерживаемые данные"
// @Router         /programs/ [patch]
func (app *App) updateHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	statusCode := http.StatusOK

	defer func() {
		metrics.ObserveRequest(time.Since(start), statusCode)
	}()

	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPatch {
		statusCode = http.StatusMethodNotAllowed
		w.WriteHeader(statusCode)
		metrics.ErrorMetrics.Inc()
		json.NewEncoder(w).Encode(Error{"используется неверный метод"})
		return
	}
	index := r.URL.Query().Get("id")

	if r.ParseForm() != nil {
		statusCode = http.StatusBadRequest
		w.WriteHeader(statusCode)
		metrics.ErrorMetrics.Inc()
		json.NewEncoder(w).Encode(Error{"неверные аргументы ParseForm"})
		return
	}
	new_name := r.PostForm.Get("Name")
	new_description := r.PostForm.Get("Description")
	new_price := r.PostForm.Get("Price")

	slog.Info(fmt.Sprintf("%v", new_name))
	slog.Info(fmt.Sprintf("%v", new_description))
	slog.Info(fmt.Sprintf("%v", new_price))
	slog.Info(fmt.Sprintf("%v", index))

	var newValues bson.D

	if len(new_name) != 0 {
		newValues = append(newValues, bson.E{"name", new_name})
	}
	if len(new_description) != 0 {
		newValues = append(newValues, bson.E{"description", new_description})
	}
	if len(new_price) != 0 {
		new_price_int, err := strconv.Atoi(new_price)
		if err != nil || new_price_int < 0 {
			statusCode = http.StatusBadRequest
			w.WriteHeader(statusCode)
			metrics.ErrorMetrics.Inc()
			json.NewEncoder(w).Encode(Error{"неверные аргументы price"})
			return
		}
		newValues = append(newValues, bson.E{"price", new_price_int})
	}

	// Удаление из кеша.
	app.Redis.Del(context.TODO(), index)

	// Обновление в БД.
	err := app.Database.updateProgram(index, newValues)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			statusCode = http.StatusNotFound
			w.WriteHeader(statusCode)
			metrics.ErrorMetrics.Inc()
			json.NewEncoder(w).Encode(Error{"нет элемента с указанным идентификатором"})

		} else {
			statusCode = http.StatusUnprocessableEntity
			w.WriteHeader(statusCode)
			metrics.ErrorMetrics.Inc()
			json.NewEncoder(w).Encode(Error{"неподдерживаемые данные"})
		}
		return
	}

	w.WriteHeader(statusCode)
}

// @Summary        Добавить новую программу тренировок
// @Description    Добавляет в список тренировок новую программу
// @Tags 		   Программы тренировок
// @Param		   Name formData string true "Название программы"
// @Param		   Description formData string true "Описание программы"
// @Param		   Price formData int true "Стоимость программы"
// @Param		   ConfirmUserId formData string true "Id подтверждающего пользователя"
// @Produce        json
// @Success        200 "Успешно добавлена"
// @Failure        400 {object} Error "Ошибки в передаваемых параметрах"
// @Failure        405 {object} Error "Неверный метод"
// @Failure        422 {object} Error "Неподдерживаемые данные"
// @Router         /programs/ [post]
func (app *App) createHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	statusCode := http.StatusOK

	defer func() {
		metrics.ObserveRequest(time.Since(start), statusCode)
	}()

	if r.Method != http.MethodPost {
		statusCode = http.StatusMethodNotAllowed
		w.WriteHeader(statusCode)
		metrics.ErrorMetrics.Inc()
		json.NewEncoder(w).Encode(Error{"используется неверный метод"})
		return
	}
	if r.ParseForm() != nil {
		statusCode = http.StatusBadRequest
		w.WriteHeader(statusCode)
		metrics.ErrorMetrics.Inc()
		json.NewEncoder(w).Encode(Error{"неверные аргументы при парсинге"})
		return
	}

	price, err := strconv.Atoi(r.PostForm.Get("Price"))
	if err != nil || price < 0 {
		statusCode = http.StatusBadRequest
		w.WriteHeader(statusCode)
		metrics.ErrorMetrics.Inc()
		json.NewEncoder(w).Encode(Error{fmt.Sprintf("неверные аргументы: Price: %s"+
			"", r.PostForm.Get("Price"))})
		return
	}

	confirmedUserId, err := bson.ObjectIDFromHex(r.PostForm.Get("ConfirmUserId"))
	if err != nil {
		statusCode = http.StatusBadRequest
		w.WriteHeader(statusCode)
		metrics.ErrorMetrics.Inc()
		json.NewEncoder(w).Encode(Error{"некорректный id подтверждающего пользователя"})
		return
	}

	newProgram := Program{
		r.PostForm.Get("Name"),
		r.PostForm.Get("Description"),
		price,
		"-",
		confirmedUserId,
	}
	insertedId, err := app.Database.addProgram(&newProgram)
	if err != nil {
		statusCode = http.StatusUnprocessableEntity
		w.WriteHeader(statusCode)
		metrics.ErrorMetrics.Inc()
		json.NewEncoder(w).Encode(Error{"неподдерживаемые данные"})
		return
	}

	oid, ok := insertedId.InsertedID.(bson.ObjectID)
	if !ok {
		statusCode = http.StatusUnprocessableEntity
		w.WriteHeader(statusCode)
		metrics.ErrorMetrics.Inc()
		json.NewEncoder(w).Encode(Error{"при добавлении новой программы сгенерированное id не типа ObjectID"})
		return
	}

	err = Write(app.Producer, oid.Hex(), confirmedUserId.Hex())
	if err != nil {
		statusCode = http.StatusUnprocessableEntity
		w.WriteHeader(statusCode)
		metrics.ErrorMetrics.Inc()
		json.NewEncoder(w).Encode(Error{"ошибка при записи в очередь " + err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(oid)
}

// @Summary        Удалить программу тренировок
// @Description    Удаляет программу тренировок из списка
// @Tags 		   Программы тренировок
// @Param          id   query  string  true  "Идентификатор удаляемой программы"
// @Produce        json
// @Success        200 "Успешно удалена"
// @Failure        400 {object} Error "Ошибки в передаваемых параметрах"
// @Failure        404 {object} Error "Нет элемента с указанным идентификатором"
// @Failure        405 {object} Error "Неверный метод"
// @Failure        422 {object} Error "Неподдерживаемые данные"
// @Router         /programs/ [delete]
func (app *App) deleteHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	statusCode := http.StatusOK

	defer func() {
		w.WriteHeader(statusCode)
		metrics.ObserveRequest(time.Since(start), statusCode)
	}()

	if r.Method != http.MethodDelete {
		statusCode = http.StatusMethodNotAllowed
		metrics.ErrorMetrics.Inc()
		json.NewEncoder(w).Encode(Error{"используется неверный метод"})
		return
	}

	id := r.URL.Query().Get("id")

	// Удаление из кеша.
	app.Redis.Del(context.TODO(), id)

	err := app.Database.deleteProgram(id)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			statusCode = http.StatusNotFound
			w.WriteHeader(statusCode)
			metrics.ErrorMetrics.Inc()
			json.NewEncoder(w).Encode(Error{"нет элемента с указанным идентификатором"})

		} else {
			statusCode = http.StatusUnprocessableEntity
			w.WriteHeader(statusCode)
			metrics.ErrorMetrics.Inc()
			json.NewEncoder(w).Encode(Error{"неподдерживаемые данные"})
		}
		return
	}

	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
}

// @Summary        Удалить все программы тренировок
// @Description    Удаляет все программы тренировок из списка
// @Tags 		   Программы тренировок
// @Success        200 "Успешно удалены"
// @Failure        422 {object} Error "Неподдерживаемые данные"
// @Router         /programs [delete]
func (app *App) deleteAllHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	statusCode := http.StatusOK

	defer func() {
		metrics.ObserveRequest(time.Since(start), statusCode)
	}()

	err := app.Database.deleteAllPrograms()
	if err != nil {
		statusCode = http.StatusUnprocessableEntity
		w.WriteHeader(statusCode)
		metrics.ErrorMetrics.Inc()
		json.NewEncoder(w).Encode(Error{"неподдерживаемые данные"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
}

// @Summary        Получить кол-во программ
// @Description    Возвращает кол-во документов
// @Tags 		   Программы тренировок
// @Produce        text/plain
// @Success        200 {integer} int "Кол-во элементов"
// @Failure        422 {object} Error "Неподдерживаемые данные"
// @Router         /programs/count [get]
func (app *App) getCount(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	statusCode := http.StatusOK

	defer func() {
		metrics.ObserveRequest(time.Since(start), statusCode)
	}()

	w.Header().Set("Content-Type", "text/plain")
	count, err := app.Database.countPrograms()
	if err != nil {
		statusCode = http.StatusNotFound
		w.WriteHeader(statusCode)
		metrics.ErrorMetrics.Inc()
		json.NewEncoder(w).Encode(Error{"неподдерживаемые данные"})
		return
	}
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, "%d", count)
}

func (app App) UpdateStatus(programId bson.ObjectID, newStatus string) error {
	collection := app.Database.Client.Database(app.Database.DbName).Collection(app.Database.ColName)
	collection.UpdateOne(
		context.TODO(),
		bson.D{{"_id", programId}},
		bson.D{{"$set", bson.D{{"wasconfirmed", newStatus}}}},
	)

	// Удаление из кеша.
	app.Redis.Del(context.TODO(), programId.Hex())
	return nil
}

func (app *App) routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /programs", app.showAllHandler)
	mux.HandleFunc("GET /programs/", app.showHandler)
	mux.HandleFunc("PATCH /programs/", app.updateHandler)
	mux.HandleFunc("POST /programs/", app.createHandler)
	mux.HandleFunc("DELETE /programs/", app.deleteHandler)
	mux.HandleFunc("DELETE /programs", app.deleteAllHandler)
	mux.HandleFunc("GET /programs/count", app.getCount)
	mux.HandleFunc("/swagger/", httpSwagger.Handler(httpSwagger.URL("http://localhost:8081/swagger/doc.json")))
	mux.Handle("/metrics", promhttp.Handler())
	return mux
}
