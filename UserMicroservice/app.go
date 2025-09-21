package main

import (
	"UserMicroservice/metrics"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type App struct {
	Database *DbProcessor
}

var jwtSecretKey = []byte("microservices")

func setErrorStatus(statusCode int, w http.ResponseWriter, memStatusCode *int) {
	*memStatusCode = statusCode
	w.WriteHeader(statusCode)
	metrics.ErrorMetrics.Inc()
}

// @Summary        Получить всех пользователей
// @Description    Возвращает список пользователей
// @Tags 		   Пользователи
// @Produce        json
// @Success        200 {object} User  "Успешно получены все пользователи"
// @Failure        422 {object} AppError "Неподдерживаемые данные"
// @Router         /users [get]
func (app *App) showAllHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	statusCode := http.StatusOK

	defer func() {
		metrics.ObserveRequest(time.Since(start), statusCode)
	}()

	w.Header().Set("Content-Type", "application/json")

	slog.Info("Entered GET /users/")

	allUsers, err := app.Database.GetAllUsers()
	if err != nil {
		setErrorStatus(http.StatusInternalServerError, w, &statusCode)
		json.NewEncoder(w).Encode("неподдерживаемые данные")
		return
	}

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(allUsers)
}

// @Summary        Получить пользователя
// @Description    Возвращает пользователя по индексу
// @Tags 		   Пользователи
// @Produce        json
// @Param          id   query  string  true  "Идентификатор пользователя"
// @Success        200 {object} User  "Успешно получен пользователь"
// @Failure        404 {object} AppError "Нет элемента с указанным идентификатором"
// @Failure        422 {object} AppError "Неподдерживаемые данные
// @Router         /users/ [get]
func (app *App) showHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	statusCode := http.StatusOK

	defer func() {
		metrics.ObserveRequest(time.Since(start), statusCode)
	}()

	w.Header().Set("Content-Type", "application/json")

	index := r.URL.Query().Get("id")

	user, err := app.Database.GetUser(index)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			setErrorStatus(http.StatusNotFound, w, &statusCode)
			json.NewEncoder(w).Encode(AppError{"нет элемента с указанным идентификатором"})

		} else {
			setErrorStatus(http.StatusInternalServerError, w, &statusCode)
			json.NewEncoder(w).Encode(AppError{"неподдерживаемые данные"})
		}
		return
	}

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(user)
}

// @Summary        Обновить пользователя
// @Description    Обновляет информацию о пользователе
// @Tags 		   Пользователи
// @Produce        json
// @Param          id   query  string  true  "Идентификатор обновляемого пользователя"
// @Param		   Name formData string false "Имя"
// @Param		   Surname formData string false "Фамилия"
// @Param		   Age formData int false "Возраст"
// @Success        200 "Успешно обновлена"
// @Failure        400 {object} AppError "Неверное значение"
// @Failure        404 {object} AppError "Нет элемента с таким идентификатором"
// @Failure        422 {object} AppError "Неподдерживаемые данные"
// @Router         /users/ [patch]
func (app *App) updateHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	statusCode := http.StatusOK

	defer func() {
		metrics.ObserveRequest(time.Since(start), statusCode)
	}()

	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPatch {
		setErrorStatus(http.StatusMethodNotAllowed, w, &statusCode)
		json.NewEncoder(w).Encode(AppError{"используется неверный метод"})
		return
	}
	index := r.URL.Query().Get("id")

	if r.ParseForm() != nil {
		setErrorStatus(http.StatusBadRequest, w, &statusCode)
		json.NewEncoder(w).Encode(AppError{"неверные аргументы"})
		return
	}
	new_name := r.PostForm.Get("Name")
	new_surname := r.PostForm.Get("Surname")
	new_age := r.PostForm.Get("Age")

	var newValues bson.D

	if len(new_name) != 0 {
		newValues = append(newValues, bson.E{"name", new_name})
	}
	if len(new_surname) != 0 {
		newValues = append(newValues, bson.E{"surname", new_surname})
	}
	if len(new_age) != 0 {
		new_price_int, err := strconv.Atoi(new_age)
		if err != nil || new_price_int < 0 {
			setErrorStatus(http.StatusBadRequest, w, &statusCode)
			json.NewEncoder(w).Encode(AppError{"неверные аргументы"})
			return
		}
		newValues = append(newValues, bson.E{"age", new_price_int})
	}

	err := app.Database.UpdateUser(index, newValues)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			setErrorStatus(http.StatusNotFound, w, &statusCode)
			json.NewEncoder(w).Encode(AppError{"нет элемента с указанным идентификатором"})

		} else {
			setErrorStatus(http.StatusUnprocessableEntity, w, &statusCode)
			json.NewEncoder(w).Encode(AppError{"неподдерживаемые данные"})
		}
		return
	}

	w.WriteHeader(statusCode)
}

// @Summary        Добавить нового пользователя
// @Description    Добавляет в список пользователей нового
// @Tags 		   Пользователи
// @Param		   Name formData string true "Имя"
// @Param		   Surname formData string true "Фамилия"
// @Param		   login formData string true "Логин"
// @Param		   Age formData int true "Возвраст"
// @Produce        json
// @Success        200 "Успешно добавлен"
// @Failure        400 {object} AppError "Ошибки в передаваемых параметрах"
// @Failure        405 {object} AppError "Неверный метод"
// @Failure        422 {object} AppError "Неподдерживаемые данные"
// @Router         /users/ [post]
func (app *App) createHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	statusCode := http.StatusOK

	defer func() {
		metrics.ObserveRequest(time.Since(start), statusCode)
	}()

	if r.Method != http.MethodPost {
		setErrorStatus(http.StatusMethodNotAllowed, w, &statusCode)
		json.NewEncoder(w).Encode(AppError{"используется неверный метод"})
		return
	}
	if r.ParseForm() != nil {
		setErrorStatus(http.StatusBadRequest, w, &statusCode)
		json.NewEncoder(w).Encode(AppError{"неверные аргументы"})
		return
	}

	age, err := strconv.Atoi(r.PostForm.Get("Age"))
	if err != nil || age < 0 {
		setErrorStatus(http.StatusBadRequest, w, &statusCode)
		json.NewEncoder(w).Encode(AppError{"неверные аргументы"})
		return
	}

	newUser := User{
		Name:    r.PostForm.Get("Name"),
		Surname: r.PostForm.Get("Surname"),
		Login:   r.PostForm.Get("login"),
		Age:     age,
	}
	insertedId, err := app.Database.AddUser(&newUser)
	if err != nil {
		setErrorStatus(http.StatusUnprocessableEntity, w, &statusCode)
		json.NewEncoder(w).Encode(AppError{"неподдерживаемые данные"})
		return
	}

	oid, ok := insertedId.InsertedID.(bson.ObjectID)
	if !ok {
		setErrorStatus(http.StatusUnprocessableEntity, w, &statusCode)
		json.NewEncoder(w).Encode(AppError{"при добавлении нового пользователя сгенерированное id не типа ObjectID"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(oid)
	w.WriteHeader(statusCode)
}

// @Summary        Удалить пользователя
// @Description    Удаляет пользователя из списка
// @Tags 		   Пользователи
// @Param          id   query  string  true  "Идентификатор удаляемого пользователя"
// @Produce        json
// @Success        200 "Успешн удален"
// @Failure        400 {object} AppError "Ошибки в передаваемых параметрах"
// @Failure        404 {object} AppError "Нет элемента с указанным идентификатором"
// @Failure        405 {object} AppError "Неверный метод"
// @Failure        422 {object} AppError "Неподдерживаемые данные"
// @Router         /users/ [delete]
func (app *App) deleteHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	statusCode := http.StatusOK

	defer func() {
		metrics.ObserveRequest(time.Since(start), statusCode)
	}()

	if r.Method != http.MethodDelete {
		setErrorStatus(http.StatusMethodNotAllowed, w, &statusCode)
		json.NewEncoder(w).Encode(AppError{"используется неверный метод"})
		return
	}

	id := r.URL.Query().Get("id")

	err := app.Database.DeleteUser(id)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			setErrorStatus(http.StatusNotFound, w, &statusCode)
			json.NewEncoder(w).Encode(AppError{"нет элемента с указанным идентификатором"})

		} else {
			setErrorStatus(http.StatusUnprocessableEntity, w, &statusCode)
			json.NewEncoder(w).Encode(AppError{"неподдерживаемые данные"})
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
}

// @Summary        Удалить всех пользователей
// @Description    Удаляет всех пользователей из списка
// @Tags 		   Пользователи
// @Success        200 "Успешно удален"
// @Failure        422 {object} AppError "Неподдерживаемые данные"
// @Router         /users [delete]
func (app *App) deleteAllHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	statusCode := http.StatusOK

	defer func() {
		metrics.ObserveRequest(time.Since(start), statusCode)
	}()

	err := app.Database.DeleteAllUsers()
	if err != nil {
		setErrorStatus(http.StatusUnprocessableEntity, w, &statusCode)
		json.NewEncoder(w).Encode(AppError{"неподдерживаемые данные"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
}

// @Summary        Получить кол-во пользователей
// @Description    Возвращает кол-во пользователей
// @Tags 		   Пользователи
// @Produce        text/plain
// @Success        200 {integer} int "Кол-во элементов"
// @Failure        422 {object} AppError "Неподдерживаемые данные"
// @Router         /users/count [get]
func (app *App) getCount(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	statusCode := http.StatusOK

	defer func() {
		metrics.ObserveRequest(time.Since(start), statusCode)
	}()

	w.Header().Set("Content-Type", "text/plain")
	count, err := app.Database.CountUsers()
	if err != nil {
		setErrorStatus(http.StatusUnprocessableEntity, w, &statusCode)
		json.NewEncoder(w).Encode(AppError{"неподдерживаемые данные"})
		return
	}

	w.WriteHeader(statusCode)
	fmt.Fprintf(w, "%d", count)
}

// @Summary        Проверить наличие пользователя
// @Description    Возвращает 200, или 404
// @Tags 		   Пользователи
// @Param          login   query  string  true  "Логин пользователя"
// @Success        200 "Данный пользователь существует"
// @Failure        404 "Нет элемента с указанным логином"
// @Router         /users/validness [get]
func (app *App) isValidUser(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	statusCode := http.StatusOK

	defer func() {
		metrics.ObserveRequest(time.Since(start), statusCode)
	}()

	login := r.URL.Query().Get("login")
	slog.Info(login)
	if !app.Database.FindUserByLogin(login) {
		slog.Info("user with login wasnt found: " + login)
		setErrorStatus(http.StatusNotFound, w, &statusCode)
		return
	}
	slog.Info("user with login was found: " + login)

	w.WriteHeader(statusCode)
}

func (app *App) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /users", app.showAllHandler)
	mux.HandleFunc("GET /users/", app.showHandler)
	mux.HandleFunc("PATCH /users/", app.updateHandler)
	mux.HandleFunc("POST /users/", app.createHandler)
	mux.HandleFunc("DELETE /users/", app.deleteHandler)
	mux.HandleFunc("DELETE /users", app.deleteAllHandler)
	mux.HandleFunc("GET /users/count", app.getCount)
	mux.HandleFunc("GET /users/validness", app.isValidUser)
	mux.HandleFunc("/swagger/", httpSwagger.Handler(httpSwagger.URL("http://localhost:8080/swagger/doc.json")))
	mux.Handle("/metrics", promhttp.Handler())
	return mux
}

type AppError struct {
	msg string
}

func (e AppError) Error() string {
	return e.msg
}
