package accrual

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/AntonBezemskiy/gophermart/internal/mocks"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Good struct {
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

type Request struct {
	Order string `json:"order"`
	Goods []Good `json:"goods"`
}

type Reward struct {
	Match      string  `json:"match"`
	Reward     float64 `json:"reward"`
	RewardType string  `json:"reward_type"`
}

func TestSender(t *testing.T) {
	// функция для получения свободного порта для запуска приложений
	getFreePort := func() (int, error) {
		// Слушаем на порту 0, чтобы операционная система выбрала свободный порт
		listener, err := net.Listen("tcp", ":0")
		if err != nil {
			return 0, err
		}
		defer listener.Close()

		// Получаем назначенный системой порт
		port := listener.Addr().(*net.TCPAddr).Port
		return port, nil
	}

	// Запускаем программу
	accrualPort, err := getFreePort()
	require.NoError(t, err)
	accrulaAdress := fmt.Sprintf("localhost:%d", accrualPort)
	cmd := exec.Command("./../../cmd/accrual/accrual_linux_amd64", fmt.Sprintf("-a=%s", accrulaAdress))

	// Связываем стандартный вывод и ошибки программы с выводом программы Go
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()

	// Запуск программы
	err = cmd.Start()
	require.NoError(t, err)
	time.Sleep(2 * time.Second) // Ждем 2 секунды для запуска сервиса

	// Запуск тестов------------------------------------
	// устанавливаю адрес доступа к системе accrual
	SetAccrualSystemAddress(fmt.Sprintf("http://%s", accrulaAdress))

	// тест с кодом 204: заказ не зарегистрирован в системе расчёта
	{
		client := resty.New()

		// создаем буферизованный канал для принятия задач в воркер
		jobs := make(chan int64, 1)
		// создаем буферизованный канал для отправки результатов
		results := make(chan Result, 1)

		go Sender(jobs, results, GetAccrualSystemAddress(), client)
		jobs <- 474834550169
		close(jobs)

		result := <-results
		assert.Equal(t, http.StatusNoContent, result.requestStatus)
	}

	// тест с кодом 200
	{
		client := resty.New()

		// добавляю в систему тестовые данные
		request1 := Request{
			Order: "155047814506",
			Goods: []Good{
				{
					Description: "First thing",
					Price:       500,
				},
			},
		}

		// отправляю заказ в систему
		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetResult(request1).
			Post(fmt.Sprintf("http://%s/api/orders", accrulaAdress))
		require.NoError(t, err)
		assert.Equal(t, 202, resp.StatusCode())

		// отправляю тестовый запрос с расчетом получить код 200
		// создаем буферизованный канал для принятия задач в воркер
		jobs := make(chan int64, 1)
		// создаем буферизованный канал для отправки результатов
		results := make(chan Result, 1)

		go Sender(jobs, results, GetAccrualSystemAddress(), client)
		jobs <- 155047814506
		close(jobs)

		result := <-results
		assert.Equal(t, http.StatusOK, result.requestStatus)
	}
	// тест с получением количества баллов
	{
		// создаю нового клиента
		client := resty.New()

		// регистрирую информацию о вознаграждении за заказ
		reward := Reward{
			Match:      "Asus",
			Reward:     5000,
			RewardType: "pt",
		}
		// отправляю заказ в систему
		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetResult(reward).
			Post(fmt.Sprintf("http://%s/api/goods", accrulaAdress))
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())

		// добавляю в систему тестовый заказ
		request1 := Request{
			Order: "707261873236",
			Goods: []Good{
				{
					Description: "Laptop Asus",
					Price:       100500.1,
				},
			},
		}

		// отправляю заказ в систему
		resp, err = client.R().
			SetHeader("Content-Type", "application/json").
			SetResult(request1).
			Post(fmt.Sprintf("http://%s/api/orders", accrulaAdress))
		require.NoError(t, err)
		assert.Equal(t, 202, resp.StatusCode())

		// отправляю тестовый запрос с расчетом получить код 200
		// создаем буферизованный канал для принятия задач в воркер
		jobs := make(chan int64, 1)
		// создаем буферизованный канал для отправки результатов
		results := make(chan Result, 1)

		go Sender(jobs, results, GetAccrualSystemAddress(), client)
		jobs <- 707261873236
		close(jobs)

		result := <-results
		assert.Equal(t, http.StatusOK, result.requestStatus)
		assert.Equal(t, PROCESSED, result.Status)
		// проверяю количество начисленных баллов
		assert.InEpsilon(t, 5000, result.Accrual, 0.001)
	}

	// Останавливаем процесс
	err = cmd.Process.Signal(os.Interrupt) // Посылаем сигнал прерывания
	require.NoError(t, err)
	time.Sleep(time.Second)
}

func TestGenerator(t *testing.T) {
	// функция для получения свободного порта для запуска приложений
	getFreePort := func() (int, error) {
		// Слушаем на порту 0, чтобы операционная система выбрала свободный порт
		listener, err := net.Listen("tcp", ":0")
		if err != nil {
			return 0, err
		}
		defer listener.Close()

		// Получаем назначенный системой порт
		port := listener.Addr().(*net.TCPAddr).Port
		return port, nil
	}

	// Запускаем программу
	accrualPort, err := getFreePort()
	require.NoError(t, err)
	accrulaAdress := fmt.Sprintf(":%d", accrualPort)

	cmd := exec.Command("./../../cmd/accrual/accrual_linux_amd64", fmt.Sprintf("-a=%s", accrulaAdress))

	// Связываем стандартный вывод и ошибки программы с выводом программы Go
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()

	// Запуск программы
	err = cmd.Start()
	require.NoError(t, err)
	time.Sleep(2 * time.Second) // Ждем 2 секунды для запуска сервиса

	// Запуск тестов------------------------------------
	// устанавливаю адрес доступа к системе accrual
	SetAccrualSystemAddress(fmt.Sprintf("http://%s", accrulaAdress))

	// тест с кодом 204: заказ не зарегистрирован в системе расчёта
	{
		// создаю мок для возвращения слайса запросов в generator
		data := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mocks.NewMockOrdersInterface(ctrl)
		m.EXPECT().GetOrdersForAccrual(gomock.Any()).Return(data, nil)
		client := resty.New()
		res, err := Generator(context.Background(), client, m)
		require.NoError(t, err)
		// проверка, что все результаты работы Generator содержат статус 204
		for _, r := range res {
			assert.Equal(t, http.StatusNoContent, r.requestStatus)
		}
	}

	// Останавливаем процесс
	err = cmd.Process.Signal(os.Interrupt) // Посылаем сигнал прерывания
	require.NoError(t, err)
	time.Sleep(time.Second)
}
