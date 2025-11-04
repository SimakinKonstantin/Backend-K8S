package main

import (
	"Gateway/metrics"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type settings struct {
	microserviceURL string
	userServiceURL  string
}

func setErrorStatus(statusCode int, w http.ResponseWriter) {
	w.WriteHeader(statusCode)
	metrics.ErrorMetrics.Inc()
}

// @Summary		   Liveness Probe
// @Description	   Проверка готовности сервиса
// @Success		   200 "Все зависимости сервиса инициализированы, он готов к работе"
// @Router		   /health
func (app *settings) Health(w http.ResponseWriter, r *http.Request) {
	microserviceResp, err := http.Get(app.microserviceURL + "/health")
	if err != nil {
		slog.Error("Ошибка в /health microservice: ", err.Error())
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	userserviceResp, err := http.Get(app.userServiceURL + "/health")
	if err != nil {
		slog.Error("Ошибка в /health users service: ", err.Error())
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	if microserviceResp.StatusCode != http.StatusOK {
		slog.Error("health сервиса програм тренировок НЕ вернул 200!")
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	if userserviceResp.StatusCode != http.StatusOK {
		slog.Error("health сервиса пользователей НЕ вернул 200!")
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Security BearerAuth
// @Summary        Получить все программы тренировок
// @Description    Возвращает список программ
// @Tags 		   Программы тренировок
// @Produce        json
// @Success        200 {object} Program  "Успешно получены все программы тренировок"
// @Failure        401 {object} Error "Нет доступа"
// @Failure        421 {object} Error "Не удалось перенаправить запрос"
// @Failure        422 {object} Error "Неподдерживаемые данные"
// @Router         /gateway/programs [get]
func (s *settings) objectShowAllHandler(w http.ResponseWriter, r *http.Request) {
	loginJWT, err := s.parseJWT(r.Header.Get("Authorization"))
	if err != nil {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	isValid, err := s.isRealUser(loginJWT)
	if err != nil {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(Error{err.Error()})
		return
	}

	if !isValid {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(Error{"невалидный логин"})
		return
	}

	s.forwardResp(&w, r, s.microserviceURL)
}

// @Security BearerAuth
// @Summary        Получить программу тренировок
// @Description    Возвращает программу тренировок по индексу
// @Tags 		   Программы тренировок
// @Produce        json
// @Param          id   query  string  true  "Идентификатор программы"
// @Success        200 {object} Program  "Успешно получена программа тренировок"
// @Failure        401 {object} Error "Нет доступа"
// @Failure        404 {object} Error "Нет элемента с указанным идентификатором"
// @Failure        422 {object} Error "Неподдерживаемые данные
// @Router         /gateway/programs/ [get]
func (s *settings) objectShowHandler(w http.ResponseWriter, r *http.Request) {

	loginJWT, err := s.parseJWT(r.Header.Get("Authorization"))
	if err != nil {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	isValid, err := s.isRealUser(loginJWT)
	if err != nil {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(Error{err.Error()})
		return
	}

	if !isValid {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(Error{"невалидный логин"})
		return
	}

	s.forwardResp(&w, r, s.microserviceURL)
}

// @Security BearerAuth
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
// @Failure        401 {object} Error "Нет доступа"
// @Failure        404 {object} Error "Нет элемента с таким идентификатором"
// @Failure        422 {object} Error "Неподдерживаемые данные"
// @Router         /gateway/programs/ [patch]
func (s *settings) objectUpdateHandler(w http.ResponseWriter, r *http.Request) {

	loginJWT, err := s.parseJWT(r.Header.Get("Authorization"))
	if err != nil {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	isValid, err := s.isRealUser(loginJWT)
	if err != nil {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(Error{err.Error()})
		return
	}

	if !isValid {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(Error{"невалидный логин"})
		return
	}

	s.forwardResp(&w, r, s.microserviceURL)
}

// @Security BearerAuth
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
// @Failure        401 {object} Error "Нет доступа"
// @Failure        405 {object} Error "Неверный метод"
// @Failure        422 {object} Error "Неподдерживаемые данные"
// @Router         /gateway/programs/ [post]
func (s *settings) objectCreateHandler(w http.ResponseWriter, r *http.Request) {

	loginJWT, err := s.parseJWT(r.Header.Get("Authorization"))
	if err != nil {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	isValid, err := s.isRealUser(loginJWT)
	if err != nil {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(Error{err.Error()})
		return
	}

	if !isValid {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(Error{"невалидный логин"})
		return
	}

	s.forwardResp(&w, r, s.microserviceURL)
}

// @Security BearerAuth
// @Summary        Удалить программу тренировок
// @Description    Удаляет программу тренировок из списка
// @Tags 		   Программы тренировок
// @Param          id   query  string  true  "Идентификатор удаляемой программы"
// @Produce        json
// @Success        200 "Успешно удалена"
// @Failure        400 {object} Error "Ошибки в передаваемых параметрах"
// @Failure        401 {object} Error "Нет доступа"
// @Failure        404 {object} Error "Нет элемента с указанным идентификатором"
// @Failure        405 {object} Error "Неверный метод"
// @Failure        422 {object} Error "Неподдерживаемые данные"
// @Router         /gateway/programs/ [delete]
func (s *settings) objectDeleteHandler(w http.ResponseWriter, r *http.Request) {

	loginJWT, err := s.parseJWT(r.Header.Get("Authorization"))
	if err != nil {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	isValid, err := s.isRealUser(loginJWT)
	if err != nil {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(Error{err.Error()})
		return
	}

	if !isValid {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(Error{"невалидный логин"})
		return
	}

	s.forwardResp(&w, r, s.microserviceURL)
}

// @Security BearerAuth
// @Summary        Удалить все программы тренировок
// @Description    Удаляет все программы тренировок из списка
// @Tags 		   Программы тренировок
// @Success        200 "Успешно удалены"
// @Failure        401 {object} Error "Нет доступа"
// @Failure        422 {object} Error "Неподдерживаемые данные"
// @Router         /gateway/programs [delete]
func (s *settings) objectDeleteAllHandler(w http.ResponseWriter, r *http.Request) {

	loginJWT, err := s.parseJWT(r.Header.Get("Authorization"))
	if err != nil {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	isValid, err := s.isRealUser(loginJWT)
	if err != nil {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(Error{err.Error()})
		return
	}

	if !isValid {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(Error{"невалидный логин"})
		return
	}

	s.forwardResp(&w, r, s.microserviceURL)
}

// @Security BearerAuth
// @Summary        Получить кол-во программ
// @Description    Возвращает кол-во документов
// @Tags 		   Программы тренировок
// @Produce        json
// @Success        200 {integer} int "Кол-во элементов"
// @Failure        401 {object} Error "Нет доступа"
// @Failure        422 {object} Error "Неподдерживаемые данные"
// @Router         /gateway/programs/count [get]
func (s *settings) objectGetCount(w http.ResponseWriter, r *http.Request) {

	loginJWT, err := s.parseJWT(r.Header.Get("Authorization"))
	if err != nil {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	isValid, err := s.isRealUser(loginJWT)
	if err != nil {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(Error{err.Error()})
		return
	}

	if !isValid {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(Error{"невалидный логин"})
		return
	}

	s.forwardResp(&w, r, s.microserviceURL)
}

// @Security BearerAuth
// @Summary        Получить всех пользователей
// @Description    Возвращает список пользователей
// @Tags 		   Пользователи
// @Produce        json
// @Success        200 {object} User  "Успешно получены все пользователи"
// @Failure        401 {object} Error "Не авторизован"
// @Failure        422 {object} Error "Неподдерживаемые данные"
// @Router         /gateway/users [get]
func (s *settings) userShowAllHandler(w http.ResponseWriter, r *http.Request) {

	loginJWT, err := s.parseJWT(r.Header.Get("Authorization"))
	if err != nil {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	isValid, err := s.isRealUser(loginJWT)
	if err != nil {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(Error{err.Error()})
		return
	}

	if !isValid {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(Error{"невалидный логин"})
		return
	}

	s.forwardResp(&w, r, s.userServiceURL)
}

// @Security BearerAuth
// @Summary        Получить пользователя
// @Description    Возвращает пользователя по индексу
// @Tags 		   Пользователи
// @Produce        json
// @Param          id   query  string  true  "Идентификатор пользователя"
// @Success        200 {object} User  "Успешно получен пользователь"
// @Failure        401 {object} Error "Не авторизован"
// @Failure        404 {object} Error "Нет элемента с указанным идентификатором"
// @Failure        422 {object} Error "Неподдерживаемые данные
// @Router         /gateway/users/ [get]
func (s *settings) userShowHandler(w http.ResponseWriter, r *http.Request) {

	loginJWT, err := s.parseJWT(r.Header.Get("Authorization"))
	if err != nil {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	isValid, err := s.isRealUser(loginJWT)
	if err != nil {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(Error{err.Error()})
		return
	}

	if !isValid {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(Error{"невалидный логин"})
		return
	}

	s.forwardResp(&w, r, s.userServiceURL)
}

// @Security BearerAuth
// @Summary        Обновить пользователя
// @Description    Обновляет информацию о пользователе
// @Tags 		   Пользователи
// @Produce        json
// @Param          id   query  string  true  "Идентификатор обновляемого пользователя"
// @Param		   Name formData string false "Имя"
// @Param		   Surname formData string false "Фамилия"
// @Param		   Age formData int false "Возраст"
// @Success        200 "Успешно обновлена"
// @Failure        400 {object} Error "Неверное значение"
// @Failure        401 {object} Error "Не авторизован"
// @Failure        404 {object} Error "Нет элемента с таким идентификатором"
// @Failure        422 {object} Error "Неподдерживаемые данные"
// @Router         /gateway/users/ [patch]
func (s *settings) userUpdateHandler(w http.ResponseWriter, r *http.Request) {

	loginJWT, err := s.parseJWT(r.Header.Get("Authorization"))
	if err != nil {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	isValid, err := s.isRealUser(loginJWT)
	if err != nil {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(Error{err.Error()})
		return
	}

	if !isValid {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(Error{"невалидный логин"})
		return
	}

	s.forwardResp(&w, r, s.userServiceURL)
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
// @Failure        400 {object} Error "Ошибки в передаваемых параметрах"
// @Failure        401 {object} Error "Не авторизован"
// @Failure        405 {object} Error "Неверный метод"
// @Failure        422 {object} Error "Неподдерживаемые данные"
// @Router         /gateway/users/ [post]
func (s *settings) userCreateHandler(w http.ResponseWriter, r *http.Request) {
	s.forwardResp(&w, r, s.userServiceURL)
}

// @Security BearerAuth
// @Summary        Удалить пользователя
// @Description    Удаляет пользователя из списка
// @Tags 		   Пользователи
// @Param          id   query  string  true  "Идентификатор удаляемого пользователя"
// @Produce        json
// @Success        200 "Успешн удален"
// @Failure        400 {object} Error "Ошибки в передаваемых параметрах"
// @Failure        401 {object} Error "Не авторизован"
// @Failure        404 {object} Error "Нет элемента с указанным идентификатором"
// @Failure        405 {object} Error "Неверный метод"
// @Failure        422 {object} Error "Неподдерживаемые данные"
// @Router         /gateway/users/ [delete]
func (s *settings) userDeleteHandler(w http.ResponseWriter, r *http.Request) {
	loginJWT, err := s.parseJWT(r.Header.Get("Authorization"))
	if err != nil {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	isValid, err := s.isRealUser(loginJWT)
	if err != nil {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(Error{err.Error()})
		return
	}

	if !isValid {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(Error{"невалидный логин"})
		return
	}

	s.forwardResp(&w, r, s.userServiceURL)
}

// @Security BearerAuth
// @Summary        Удалить всех пользователей
// @Description    Удаляет всех пользователей из списка
// @Tags 		   Пользователи
// @Success        200 "Успешно удален"
// @Failure        401 {object} Error "Не авторизован"
// @Failure        422 {object} Error "Неподдерживаемые данные"
// @Router         /gateway/users [delete]
func (s *settings) userDeleteAllHandler(w http.ResponseWriter, r *http.Request) {
	loginJWT, err := s.parseJWT(r.Header.Get("Authorization"))
	if err != nil {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	isValid, err := s.isRealUser(loginJWT)
	if err != nil {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(Error{err.Error()})
		return
	}

	if !isValid {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(Error{"невалидный логин"})
		return
	}

	s.forwardResp(&w, r, s.userServiceURL)
}

// @Security BearerAuth
// @Summary        Получить кол-во пользователей
// @Description    Возвращает кол-во пользователей
// @Tags 		   Пользователи
// @Produce        json
// @Success        200 {integer} int "Кол-во элементов"
// @Failure        401 {object} Error "Не авторизован"
// @Failure        422 {object} Error "Неподдерживаемые данные"
// @Router         /gateway/users/count [get]
func (s *settings) userGetCount(w http.ResponseWriter, r *http.Request) {
	loginJWT, err := s.parseJWT(r.Header.Get("Authorization"))
	if err != nil {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	isValid, err := s.isRealUser(loginJWT)
	if err != nil {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(Error{err.Error()})
		return
	}

	if !isValid {
		setErrorStatus(http.StatusUnauthorized, w)
		json.NewEncoder(w).Encode(Error{"невалидный логин"})
		return
	}

	s.forwardResp(&w, r, s.userServiceURL)
}

// @Summary        Залогиниться
// @Description    Аутентифицирует пользователя
// @Tags 		   Аутентификация
// @Produce        json
// @Param		   login query string true "Логин"
// @Success        200 {object} string "JWT-токен"
// @Failure        400 {object} Error "Некорректный запрос"
// @Failure        405 {object} Error "Неверный метод"
// @Failure        422 {object} Error "Неподдерживаемые данные"
// @Router         /login [get]
func (s *settings) loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		setErrorStatus(http.StatusMethodNotAllowed, w)
		json.NewEncoder(w).Encode(Error{"используется неверный метод"})
		return
	}
	if r.ParseForm() != nil {
		setErrorStatus(http.StatusBadRequest, w)
		json.NewEncoder(w).Encode(Error{"неверные аргументы"})
		return
	}

	login := r.URL.Query().Get("login")
	slog.Info(fmt.Sprintf("loginHandler login : %s", login))

	isValid, err := s.isRealUser(login)
	if err != nil {
		setErrorStatus(http.StatusBadRequest, w)
		json.NewEncoder(w).Encode(Error{err.Error()})
		return
	}

	if !isValid {
		setErrorStatus(http.StatusBadRequest, w)
		json.NewEncoder(w).Encode(Error{"невалидный логин"})
		return
	}

	payload := jwt.MapClaims{
		"sub": login,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	slog.Info(fmt.Sprintf("token: %+v", token))

	signedToken, err := token.SignedString([]byte("microservices"))
	if err != nil {
		setErrorStatus(http.StatusBadRequest, w)
		json.NewEncoder(w).Encode("не удалось подписать токен")
		return
	}

	json.NewEncoder(w).Encode(signedToken)
	w.WriteHeader(http.StatusOK)
}

// Перенаправляет запрос на url.
func (s *settings) forwardResp(w *http.ResponseWriter, r *http.Request, baseUrl string) {

	newURL, _ := url.Parse(baseUrl)

	// Обрезаем, чтобы получить независимый от base url.
	cutFrom := len("gateway/")
	r.URL.Path = r.URL.Path[cutFrom:]

	proxy := httputil.NewSingleHostReverseProxy(newURL)
	proxy.ServeHTTP(*w, r)
}

func (s *settings) parseJWT(tokenInfo string) (string, error) {
	token, err := jwt.Parse(tokenInfo, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неизвестный метод подписи: %v", token.Header["alg"])
		}
		return []byte("microservices"), nil
	})
	if err != nil {
		return "", fmt.Errorf("не получилось распарсить токен " + err.Error())
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("не получилось распарсить токен")
	}

	return claims["sub"].(string), nil
}

// По логину определяет: находится ли пользователь в базе.
func (s *settings) isRealUser(login string) (bool, error) {
	slog.Info(login)
	resp, err := http.Get(s.userServiceURL + "/users/validness?login=" + login)
	if err != nil {
		slog.Info("isRealUser false")
		return false, fmt.Errorf("ошибка при обращении к userService при авторизации: " + err.Error())
	}

	if resp.StatusCode == http.StatusNotFound {
		slog.Info("isRealUser false")
		return false, nil
	}

	slog.Info(fmt.Sprintf("StatusCode: %v", resp.StatusCode))

	slog.Info("isRealUser true")
	return true, nil
}

func (s *settings) routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /gateway/programs", s.objectShowAllHandler)
	mux.HandleFunc("GET /gateway/programs/", s.objectShowHandler)
	mux.HandleFunc("PATCH /gateway/programs/", s.objectUpdateHandler)
	mux.HandleFunc("POST /gateway/programs/", s.objectCreateHandler)
	mux.HandleFunc("DELETE /gateway/programs/", s.objectDeleteHandler)
	mux.HandleFunc("DELETE /gateway/programs", s.objectDeleteAllHandler)
	mux.HandleFunc("GET /gateway/programs/count", s.objectGetCount)
	mux.HandleFunc("GET /gateway/users", s.userShowAllHandler)
	mux.HandleFunc("GET /gateway/users/", s.userShowHandler)
	mux.HandleFunc("PATCH /gateway/users/", s.userUpdateHandler)
	mux.HandleFunc("POST /gateway/users/", s.userCreateHandler)
	mux.HandleFunc("DELETE /gateway/users/", s.userDeleteHandler)
	mux.HandleFunc("DELETE /gateway/users", s.userDeleteAllHandler)
	mux.HandleFunc("GET /gateway/users/count", s.userGetCount)
	mux.HandleFunc("GET /login", s.loginHandler)
	mux.HandleFunc("/swagger/", httpSwagger.Handler(httpSwagger.URL("http://localhost:8083/swagger/doc.json")))
	mux.Handle("/metrics", promhttp.Handler())
	return mux
}
