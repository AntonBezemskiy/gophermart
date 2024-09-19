package pg

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/AntonBezemskiy/gophermart/internal/accrual"
	"github.com/AntonBezemskiy/gophermart/internal/auth"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetOrdersForAccrual(t *testing.T) {
	// тесты с базой данных
	// предварительно необходимо создать тестовую БД и определить параметры сединения host=host user=user password=password dbname=dbname  sslmode=disable
	{
		// инициализация базы данных-------------------------------------------------------------------
		databaseDsn := "host=localhost user=testgophermartpg password=newpassword dbname=testgophermartpg sslmode=disable"

		// создаём соединение с СУБД PostgreSQL
		conn, err := sql.Open("pgx", databaseDsn)
		require.NoError(t, err)
		defer conn.Close()

		// Проверка соединения с БД
		ctx := context.Background()
		err = conn.PingContext(ctx)
		require.NoError(t, err)

		// создаем экземпляр хранилища pg
		stor := NewStore(conn)
		err = stor.Bootstrap(ctx)
		require.NoError(t, err)
		//-------------------------------------------------------------------------------------------------------------

		// проверка успешной выгрузки
		{
			//предварительно загружаю в базу заказы-------------------------
			status, err := stor.Load(ctx, "GetOrdersForAccrual one", "731447180373")
			require.NoError(t, err)
			assert.Equal(t, 202, status)
			//устанавливаю паузу, чтобы заказы имели разное время загрузки
			time.Sleep(100 * time.Millisecond)

			status, err = stor.Load(ctx, "user id one pg", "250788087147")
			require.NoError(t, err)
			assert.Equal(t, 202, status)
			time.Sleep(100 * time.Millisecond)

			status, err = stor.Load(ctx, "user id two pg", "442338022134")
			require.NoError(t, err)
			assert.Equal(t, 202, status)
			time.Sleep(100 * time.Millisecond)

			numbers, err := stor.GetOrdersForAccrual(ctx)
			require.NoError(t, err)
			assert.Equal(t, 3, len(numbers))

			for i, number := range numbers {
				if i == 0 {
					numWant := int64(731447180373)
					assert.Equal(t, numWant, number)
				}
				if i == 1 {
					numWant := int64(250788087147)
					assert.Equal(t, numWant, number)
				}
				if i == 2 {
					numWant := int64(442338022134)
					assert.Equal(t, numWant, number)
				}
			}
		}

		// Удаление данных из тестовых таблиц для выполнения следующих тестов------------------------------------------
		err = stor.Disable(ctx)
		require.NoError(t, err)
	}
}

func TestGetRetryPeriod(t *testing.T) {
	// тесты с базой данных
	// предварительно необходимо создать тестовую БД и определить параметры сединения host=host user=user password=password dbname=dbname  sslmode=disable
	{
		// инициализация базы данных-------------------------------------------------------------------
		databaseDsn := "host=localhost user=testgophermartpg password=newpassword dbname=testgophermartpg sslmode=disable"

		// создаём соединение с СУБД PostgreSQL
		conn, err := sql.Open("pgx", databaseDsn)
		require.NoError(t, err)
		defer conn.Close()

		// Проверка соединения с БД
		ctx := context.Background()
		err = conn.PingContext(ctx)
		require.NoError(t, err)

		// создаем экземпляр хранилища pg
		stor := NewStore(conn)
		err = stor.Bootstrap(ctx)
		require.NoError(t, err)
		//-------------------------------------------------------------------------------------------------------------

		// тестовый случай 1
		tOne := time.Now().UTC().Truncate(time.Second) // Обрезаем наносекунды
		err = stor.AddRetryPeriod(ctx, "accrual", tOne)
		require.NoError(t, err)
		getOne, err := stor.GetRetryPeriod(ctx, "accrual")
		require.NoError(t, err)
		assert.Equal(t, tOne, getOne)

		// тестовый случай 2
		tTwo := time.Now().Add(60 * time.Second).UTC().Truncate(time.Second) // Обрезаем наносекунды
		err = stor.AddRetryPeriod(ctx, "accrual", tTwo)
		require.NoError(t, err)
		getTwo, err := stor.GetRetryPeriod(ctx, "accrual")
		require.NoError(t, err)
		assert.Equal(t, tTwo, getTwo)

		// тестовый случай 3, случай с ошибкой
		_, err = stor.GetRetryPeriod(ctx, "wrong service")
		require.Error(t, err)

		// Удаление данных из тестовых таблиц для выполнения следующих тестов------------------------------------------
		err = stor.Disable(ctx)
		require.NoError(t, err)
	}
}

func TestGetIdByOrderNumber(t *testing.T) {
	// тесты с базой данных
	// предварительно необходимо создать тестовую БД и определить параметры сединения host=host user=user password=password dbname=dbname  sslmode=disable
	{
		// инициализация базы данных-------------------------------------------------------------------
		databaseDsn := "host=localhost user=testgophermartpg password=newpassword dbname=testgophermartpg sslmode=disable"

		// создаём соединение с СУБД PostgreSQL
		conn, err := sql.Open("pgx", databaseDsn)
		require.NoError(t, err)
		defer conn.Close()

		// Проверка соединения с БД
		ctx := context.Background()
		err = conn.PingContext(ctx)
		require.NoError(t, err)

		// создаем экземпляр хранилища pg
		stor := NewStore(conn)
		err = stor.Bootstrap(ctx)
		require.NoError(t, err)
		//-------------------------------------------------------------------------------------------------------------

		// успешный тест ---------------------------------------------
		// Регистрирую нового пользователя
		ok, token, err := stor.Register(ctx, "new", "user")
		require.NoError(t, err)
		assert.Equal(t, true, ok)
		// получаю id зарегистрированного пользователя
		id, err := auth.GetUserID(token)
		require.NoError(t, err)

		// загружаю в систему новый заказ для теста
		status, err := stor.Load(ctx, id, "555731750165")
		require.NoError(t, err)
		assert.Equal(t, 202, status)

		// получаю id пользователя по номеру заказа
		idDB, err := stor.GetIdByOrderNumber(ctx, "555731750165")
		require.NoError(t, err)
		assert.Equal(t, id, idDB)

		// тест с ошибкой ------------------------------------------------
		// пытаюсь получить id пользователя по номеру заказа, хотя заказ не добавлен в систему
		_, err = stor.GetIdByOrderNumber(ctx, "218233466554")
		require.Error(t, err)

		// Удаление данных из тестовых таблиц для выполнения следующих тестов------------------------------------------
		err = stor.Disable(ctx)
		require.NoError(t, err)
	}
}

func TestUpdateBalance(t *testing.T) {
	// тесты с базой данных
	// предварительно необходимо создать тестовую БД и определить параметры сединения host=host user=user password=password dbname=dbname  sslmode=disable
	{
		// инициализация базы данных-------------------------------------------------------------------
		databaseDsn := "host=localhost user=testgophermartpg password=newpassword dbname=testgophermartpg sslmode=disable"

		// создаём соединение с СУБД PostgreSQL
		conn, err := sql.Open("pgx", databaseDsn)
		require.NoError(t, err)
		defer conn.Close()

		// Проверка соединения с БД
		ctx := context.Background()
		err = conn.PingContext(ctx)
		require.NoError(t, err)

		// создаем экземпляр хранилища pg
		stor := NewStore(conn)
		err = stor.Bootstrap(ctx)
		require.NoError(t, err)
		//-------------------------------------------------------------------------------------------------------------

		// Регистрирую нового пользователя
		ok, token, err := stor.Register(ctx, "new", "user")
		require.NoError(t, err)
		assert.Equal(t, true, ok)
		// получаю id зарегистрированного пользователя
		id, err := auth.GetUserID(token)
		require.NoError(t, err)
		// получаю текующий баланс пользователя
		balance, err := stor.GetBalance(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, 0.0, balance.Current)
		assert.Equal(t, 0.0, balance.Withdrawn)

		// загружаю в систему новый заказ для теста
		status, err := stor.Load(ctx, id, "555731750165")
		require.NoError(t, err)
		assert.Equal(t, 202, status)

		//обновляю баланс пользователя
		err = stor.UpdateBalance(ctx, "555731750165", 391.87)
		require.NoError(t, err)

		// проверяю обновленный баланс
		balance, err = stor.GetBalance(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, 391.87, balance.Current)
		assert.Equal(t, 0.0, balance.Withdrawn)

		// Удаление данных из тестовых таблиц для выполнения следующих тестов------------------------------------------
		err = stor.Disable(ctx)
		require.NoError(t, err)
	}
}

func TestUpdateOrder(t *testing.T) {
	// тесты с базой данных
	// предварительно необходимо создать тестовую БД и определить параметры сединения host=host user=user password=password dbname=dbname  sslmode=disable
	{
		// инициализация базы данных-------------------------------------------------------------------
		databaseDsn := "host=localhost user=testgophermartpg password=newpassword dbname=testgophermartpg sslmode=disable"

		// создаём соединение с СУБД PostgreSQL
		conn, err := sql.Open("pgx", databaseDsn)
		require.NoError(t, err)
		defer conn.Close()

		// Проверка соединения с БД
		ctx := context.Background()
		err = conn.PingContext(ctx)
		require.NoError(t, err)

		// создаем экземпляр хранилища pg
		stor := NewStore(conn)
		err = stor.Bootstrap(ctx)
		require.NoError(t, err)
		//-------------------------------------------------------------------------------------------------------------

		// успешный тест--------------------------------------
		// Регистрирую нового пользователя
		ok, token, err := stor.Register(ctx, "new", "user")
		require.NoError(t, err)
		assert.Equal(t, true, ok)
		// получаю id зарегистрированного пользователя
		id1, err := auth.GetUserID(token)
		require.NoError(t, err)
		// получаю текующий баланс пользователя
		balance1, err := stor.GetBalance(ctx, id1)
		require.NoError(t, err)
		assert.Equal(t, 0.0, balance1.Current)
		assert.Equal(t, 0.0, balance1.Withdrawn)

		// предварительно загружаю данные в базу
		stat, err := stor.Load(ctx, id1, "555731750165")
		require.NoError(t, err)
		assert.Equal(t, 202, stat)
		orders, stat, err := stor.GetOrders(ctx, id1)
		require.NoError(t, err)
		assert.Equal(t, 200, stat)
		order := orders[0]
		assert.Equal(t, "NEW", order.Status)
		assert.Equal(t, 0.0, order.Accrual)

		err = stor.UpdateOrder(ctx, "555731750165", "PROCESSED", 899.98)
		require.NoError(t, err)
		orders, stat, err = stor.GetOrders(ctx, id1)
		require.NoError(t, err)
		assert.Equal(t, 200, stat)
		order = orders[0]
		assert.Equal(t, "PROCESSED", order.Status)
		assert.Equal(t, 899.98, order.Accrual)

		// получаю обновленный баланс пользователя
		balance1, err = stor.GetBalance(ctx, id1)
		require.NoError(t, err)
		assert.Equal(t, 899.98, balance1.Current)
		assert.Equal(t, 0.0, balance1.Withdrawn)

		// тест с ошибкой. Пытаюсь обновить несуществующий заказ --------------------------------------
		err = stor.UpdateOrder(ctx, "218233466554", "PROCESSED", 700.0)
		require.Error(t, err)

		// Удаление данных из тестовых таблиц для выполнения следующих тестов------------------------------------------
		err = stor.Disable(ctx)
		require.NoError(t, err)
	}
}

func TestUpdateOrderTX(t *testing.T) {
	// тесты с базой данных
	// предварительно необходимо создать тестовую БД и определить параметры сединения host=host user=user password=password dbname=dbname  sslmode=disable
	{
		// инициализация базы данных-------------------------------------------------------------------
		databaseDsn := "host=localhost user=testgophermartpg password=newpassword dbname=testgophermartpg sslmode=disable"

		// создаём соединение с СУБД PostgreSQL
		conn, err := sql.Open("pgx", databaseDsn)
		require.NoError(t, err)
		defer conn.Close()

		// Проверка соединения с БД
		ctx := context.Background()
		err = conn.PingContext(ctx)
		require.NoError(t, err)

		// создаем экземпляр хранилища pg
		stor := NewStore(conn)
		err = stor.Bootstrap(ctx)
		require.NoError(t, err)
		//-------------------------------------------------------------------------------------------------------------

		// успешный тест--------------------------------------
		// Регистрирую нового пользователя 1
		ok, token, err := stor.Register(ctx, "new1", "user1")
		require.NoError(t, err)
		assert.Equal(t, true, ok)
		// получаю id зарегистрированного пользователя
		id1, err := auth.GetUserID(token)
		require.NoError(t, err)

		// Регистрирую нового пользователя 2
		ok, token, err = stor.Register(ctx, "new2", "user2")
		require.NoError(t, err)
		assert.Equal(t, true, ok)
		// получаю id зарегистрированного пользователя
		id2, err := auth.GetUserID(token)
		require.NoError(t, err)

		// предварительно загружаю данные в базу
		stat, err := stor.Load(ctx, id1, "555731750165")
		require.NoError(t, err)
		assert.Equal(t, 202, stat)
		orders, stat, err := stor.GetOrders(ctx, id1)
		require.NoError(t, err)
		assert.Equal(t, 200, stat)
		order := orders[0]
		assert.Equal(t, "NEW", order.Status)
		assert.Equal(t, 0.0, order.Accrual)

		stat, err = stor.Load(ctx, id2, "784757004279")
		require.NoError(t, err)
		assert.Equal(t, 202, stat)

		stat, err = stor.Load(ctx, id2, "180326844420")
		require.NoError(t, err)
		assert.Equal(t, 202, stat)

		stat, err = stor.Load(ctx, id2, "218233466554")
		require.NoError(t, err)
		assert.Equal(t, 202, stat)

		data := []accrual.AccrualData{{Order: "555731750165", Status: "PROCESSING"}, {Order: "784757004279", Status: "INVALID"},
			{Order: "180326844420", Status: "PROCESSED", Accrual: 387.12}, {Order: "218233466554", Status: "PROCESSED", Accrual: 556}}

		// выполняю обновление информации в заказе
		err = stor.UpdateOrderTX(ctx, data)
		require.NoError(t, err)
		// проверяю обновленную информацию для пользователя id1
		orders, stat, err = stor.GetOrders(ctx, id1)
		require.NoError(t, err)
		assert.Equal(t, 200, stat)
		order = orders[0]
		assert.Equal(t, "PROCESSING", order.Status)
		assert.Equal(t, 0.0, order.Accrual)

		// получаю обновленный баланс пользователя id1
		balance1, err := stor.GetBalance(ctx, id1)
		require.NoError(t, err)
		assert.Equal(t, 0.0, balance1.Current)
		assert.Equal(t, 0.0, balance1.Withdrawn)

		// проверяю обновленную информацию для пользователя id2
		orders, stat, err = stor.GetOrders(ctx, id2)
		require.NoError(t, err)
		assert.Equal(t, 200, stat)
		order = orders[0]
		assert.Equal(t, "PROCESSED", order.Status)
		assert.Equal(t, 556.0, order.Accrual)

		order = orders[1]
		assert.Equal(t, "PROCESSED", order.Status)
		assert.Equal(t, 387.12, order.Accrual)

		order = orders[2]
		assert.Equal(t, "INVALID", order.Status)
		assert.Equal(t, 0.0, order.Accrual)

		// получаю обновленный баланс пользователя id2
		balance2, err := stor.GetBalance(ctx, id2)
		require.NoError(t, err)
		assert.Equal(t, 943.12, balance2.Current)
		assert.Equal(t, 0.0, balance2.Withdrawn)

		// тест с ошибкой. Пытаюсь обновить несуществующий заказ --------------------------------------
		err = stor.UpdateOrder(ctx, "218233466554", "PROCESSED", 700.0)
		require.NoError(t, err)

		// Удаление данных из тестовых таблиц для выполнения следующих тестов------------------------------------------
		err = stor.Disable(ctx)
		require.NoError(t, err)
	}
}