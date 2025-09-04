package repository

import (
	"fmt"
	"strings"
)

type Repository struct {
}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

type Order struct { // вот наша новая структура
	ID          int    // поля структур, которые передаются в шаблон
	Title       string // ОБЯЗАТЕЛЬНО должны быть написаны с заглавной буквы (то есть публичными)
	Distance    int
	Description string
	Image       string
}

func (r *Repository) GetOrders() ([]Order, error) {
	// имитируем работу с БД. Типа мы выполнили sql запрос и получили эти строки из БД
	orders := []Order{ // массив элементов из наших структур
		{
			ID:          1,
			Title:       "Владивосток - Наньша",
			Distance:    1598,
			Description: "FESCO Korea Soviet Direct Line перевозит генеральные, опасные и рефрижераторные контейнерные грузы, а также доставляет крупногабаритное и специализированное оборудование из Пусана во Владивосток.",
			Image:       "http://127.0.0.1:9000/test/image_1.svg",
		},
		{
			ID:          2,
			Title:       "Владивосток - Хайфон",
			Distance:    3407,
			Description: "FESCO Korea Soviet Direct Line перевозит генеральные, опасные и рефрижераторные контейнерные грузы, а также доставляет крупногабаритное и специализированное оборудование из Пусана во Владивосток.",
			Image:       "http://127.0.0.1:9000/test/image_2.svg",
		},
		{
			ID:          3,
			Title:       "Владивосток - Сямынь",
			Distance:    2453,
			Description: "FESCO Korea Soviet Direct Line перевозит генеральные, опасные и рефрижераторные контейнерные грузы, а также доставляет крупногабаритное и специализированное оборудование из Пусана во Владивосток.",
			Image:       "http://127.0.0.1:9000/test/image_3.svg",
		},
		{
			ID:          4,
			Title:       "Владивосток - Пусан",
			Distance:    943,
			Description: "FESCO Korea Soviet Direct Line перевозит генеральные, опасные и рефрижераторные контейнерные грузы, а также доставляет крупногабаритное и специализированное оборудование из Пусана во Владивосток.",
			Image:       "http://127.0.0.1:9000/test/image_5.svg",
		},
	}
	// обязательно проверяем ошибки, и если они появились - передаем выше, то есть хендлеру
	// тут я снова искусственно обработаю "ошибку" чисто чтобы показать вам как их передавать выше
	if len(orders) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}

	return orders, nil
}

func (r *Repository) GetOrder(id int) (Order, error) {
	// тут у вас будет логика получения нужной услуги, тоже наверное через цикл в первой лабе, и через запрос к БД начиная со второй
	orders, err := r.GetOrders()
	if err != nil {
		return Order{}, err // тут у нас уже есть кастомная ошибка из нашего метода, поэтому мы можем просто вернуть ее
	}

	for _, order := range orders {
		if order.ID == id {
			return order, nil // если нашли, то просто возвращаем найденный заказ (услугу) без ошибок
		}
	}
	return Order{}, fmt.Errorf("заказ не найден") // тут нужна кастомная ошибка, чтобы понимать на каком этапе возникла ошибка и что произошло
}

func (r *Repository) GetOrdersByTitle(title string) ([]Order, error) {
	orders, err := r.GetOrders()
	if err != nil {
		return []Order{}, err
	}

	var result []Order
	for _, order := range orders {
		if strings.Contains(strings.ToLower(order.Title), strings.ToLower(title)) {
			result = append(result, order)
		}
	}

	return result, nil
}

type Request struct { // вот наша новая структура
	ID          int    // поля структур, которые передаются в шаблон
	Title       string // ОБЯЗАТЕЛЬНО должны быть написаны с заглавной буквы (то есть публичными)
	Distance    int
	Description string
	Image       string
}

func (r *Repository) GetRequests() ([]Request, error) {
	// имитируем работу с БД. Типа мы выполнили sql запрос и получили эти строки из БД
	requests := []Request{ // массив элементов из наших структур
		{
			ID:          2,
			Title:       "Владивосток - Хайфон",
			Distance:    3407,
			Description: "FESCO Korea Soviet Direct Line перевозит генеральные, опасные и рефрижераторные контейнерные грузы, а также доставляет крупногабаритное и специализированное оборудование из Пусана во Владивосток.",
			Image:       "http://127.0.0.1:9000/test/image_2.svg",
		},
		{
			ID:          4,
			Title:       "Владивосток - Пусан",
			Distance:    943,
			Description: "FESCO Korea Soviet Direct Line перевозит генеральные, опасные и рефрижераторные контейнерные грузы, а также доставляет крупногабаритное и специализированное оборудование из Пусана во Владивосток.",
			Image:       "http://127.0.0.1:9000/test/image_5.svg",
		},
	}
	// обязательно проверяем ошибки, и если они появились - передаем выше, то есть хендлеру
	// тут я снова искусственно обработаю "ошибку" чисто чтобы показать вам как их передавать выше
	if len(requests) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}

	return requests, nil
}
