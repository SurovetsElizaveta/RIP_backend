package repository

import (
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type Repository struct {
}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

type Route struct { // вот наша новая структура
	ID          int    // поля структур, которые передаются в шаблон
	Title       string // ОБЯЗАТЕЛЬНО должны быть написаны с заглавной буквы (то есть публичными)
	Distance    int
	Description string
	Image       string
	Delay       time.Duration
}

func (r *Repository) GetRoutes() ([]Route, error) {
	// имитируем работу с БД. Типа мы выполнили sql запрос и получили эти строки из БД
	routes := []Route{ // массив элементов из наших структур
		{
			ID:          1,
			Title:       "Владивосток - Наньша",
			Distance:    1598,
			Description: "FESCO Korea Soviet Direct Line перевозит генеральные, опасные и рефрижераторные контейнерные грузы, а также доставляет крупногабаритное и специализированное оборудование из Пусана во Владивосток.",
			Image:       "http://127.0.0.1:9000/test/image_1.svg",
			Delay:       time.Duration(48) * time.Hour,
		},
		{
			ID:          2,
			Title:       "Владивосток - Хайфон",
			Distance:    3407,
			Description: "FESCO Korea Soviet Direct Line перевозит генеральные, опасные и рефрижераторные контейнерные грузы, а также доставляет крупногабаритное и специализированное оборудование из Пусана во Владивосток.",
			Image:       "http://127.0.0.1:9000/test/image_2.svg",
			Delay:       time.Duration(48) * time.Hour,
		},
		{
			ID:          3,
			Title:       "Владивосток - Сямынь",
			Distance:    2453,
			Description: "FESCO Korea Soviet Direct Line перевозит генеральные, опасные и рефрижераторные контейнерные грузы, а также доставляет крупногабаритное и специализированное оборудование из Пусана во Владивосток.",
			Image:       "http://127.0.0.1:9000/test/image_3.svg",
			Delay:       time.Duration(48) * time.Hour,
		},
		{
			ID:          4,
			Title:       "Владивосток - Пусан",
			Distance:    943,
			Description: "FESCO Korea Soviet Direct Line перевозит генеральные, опасные и рефрижераторные контейнерные грузы, а также доставляет крупногабаритное и специализированное оборудование из Пусана во Владивосток.",
			Image:       "http://127.0.0.1:9000/test/image_5.svg",
			Delay:       time.Duration(48) * time.Hour,
		},
	}
	// обязательно проверяем ошибки, и если они появились - передаем выше, то есть хендлеру
	// тут я снова искусственно обработаю "ошибку" чисто чтобы показать вам как их передавать выше
	if len(routes) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}

	return routes, nil
}

func (r *Repository) GetRoute(id int) (Route, error) {
	// тут у вас будет логика получения нужной услуги, тоже наверное через цикл в первой лабе, и через запрос к БД начиная со второй
	routes, err := r.GetRoutes()
	if err != nil {
		return Route{}, err // тут у нас уже есть кастомная ошибка из нашего метода, поэтому мы можем просто вернуть ее
	}

	for _, route := range routes {
		if route.ID == id {
			return route, nil // если нашли, то просто возвращаем найденный заказ (услугу) без ошибок
		}
	}
	return Route{}, fmt.Errorf("заказ не найден") // тут нужна кастомная ошибка, чтобы понимать на каком этапе возникла ошибка и что произошло
}

func (r *Repository) GetRoutesByTitle(title string) ([]Route, error) {
	routes, err := r.GetRoutes()
	if err != nil {
		return []Route{}, err
	}

	var result []Route
	for _, route := range routes {
		if strings.Contains(strings.ToLower(route.Title), strings.ToLower(title)) {
			result = append(result, route)
		}
	}

	return result, nil
}

type RouteRequest struct {
	RouteId     int
	ArrivalDate time.Time
	ShipSpeed   int
}

type Request struct { // вот наша новая структура
	ID            int // поля структур, которые передаются в шаблон
	DepartureDate time.Time
	RouteInfo     []RouteRequest
}

// время прибытия = дата отправления + время доставки
// время доставки = 1 день на погрузку + путь (со скоростью 23 узла) + 1 день на разгрузку

func (r *Repository) GetRequestDraft() ([]Request, error) {
	// имитируем работу с БД. Типа мы выполнили sql запрос и получили эти строки из БД
	requestRoutes := []Request{ // массив элементов из наших структур
		{
			ID:            1,
			DepartureDate: time.Date(2025, time.March, 14, 10, 0, 0, 0, time.UTC),
			RouteInfo: []RouteRequest{
				{
					RouteId:     1,
					ArrivalDate: time.Date(2025, time.March, 21, 12, 0, 0, 0, time.UTC),
				},
			},
		},
		{
			ID:            2,
			DepartureDate: time.Date(2025, time.March, 25, 10, 0, 0, 0, time.UTC),
			RouteInfo: []RouteRequest{
				{
					RouteId:     2,
					ArrivalDate: time.Date(2025, time.April, 5, 12, 0, 0, 0, time.UTC),
				},
			},
		},
	}
	// обязательно проверяем ошибки, и если они появились - передаем выше, то есть хендлеру
	// тут я снова искусственно обработаю "ошибку" чисто чтобы показать вам как их передавать выше
	if len(requestRoutes) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}

	var result []Request
	for _, req := range requestRoutes {
		// Создаем копию запроса, которую будем модифицировать
		newReq := Request{
			ID:            req.ID,
			DepartureDate: req.DepartureDate,
			RouteInfo:     make([]RouteRequest, len(req.RouteInfo)),
		}

		// Копируем и модифицируем RouteInfo
		for i, routeReq := range req.RouteInfo {
			route, err := r.GetRoute(routeReq.RouteId)
			if err != nil {
				logrus.Errorf("Ошибка получения маршрута %d: %v", routeReq.RouteId, err)
				continue
			}

			// Расчет скорости
			shipSpeed := calculateShipSpeed(req.DepartureDate, routeReq.ArrivalDate, route.Delay, route.Distance)

			// Создаем новый RouteRequest с рассчитанной скоростью
			newReq.RouteInfo[i] = RouteRequest{
				RouteId:     routeReq.RouteId,
				ArrivalDate: routeReq.ArrivalDate,
				ShipSpeed:   shipSpeed,
			}
		}

		result = append(result, newReq)
	}

	return result, nil
}

// Функция расчета времени прибытия
func calculateShipSpeed(departureDate time.Time, arrivalDate time.Time, delay time.Duration, distance int) int {
	travelDays := arrivalDate.Sub(departureDate.Add(delay))

	speedKmh := distance / int(travelDays.Hours())

	speedKnots := float64(speedKmh) / 1.852

	return int(speedKnots)
}
