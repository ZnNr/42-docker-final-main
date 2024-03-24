package main

import (
	"database/sql"
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "./data/tracker.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		err = db.Close()
		if err != nil {
			fmt.Println("Error closing database:", err)
		}
	}() // настройте подключение к БД

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err)
	assert.NotEqual(t, 0, id)

	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	testParcel, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, testParcel.Status, parcel.Status)
	assert.Equal(t, testParcel.Client, parcel.Client)
	assert.Equal(t, testParcel.CreatedAt, parcel.CreatedAt)
	assert.Equal(t, testParcel.Address, parcel.Address)

	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
	err = store.Delete(id)
	require.NoError(t, err)

	_, err = store.Get(id)
	require.Error(t, err)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "./data/tracker.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		err = db.Close()
		if err != nil {
			fmt.Println("Error closing database:", err)
		}
	}() // настройте подключение к БД

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err)
	assert.NotEqual(t, 0, id)

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"

	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	testParcel, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, testParcel.Status, parcel.Status)
	assert.Equal(t, testParcel.Client, parcel.Client)
	assert.Equal(t, testParcel.CreatedAt, parcel.CreatedAt)
	assert.NotEqual(t, testParcel.Address, parcel.Address)
	assert.Equal(t, testParcel.Address, newAddress)

	err = store.Delete(id)
	require.NoError(t, err)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "./data/tracker.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		err = db.Close()
		if err != nil {
			fmt.Println("Error closing database:", err)
		}
	}() // настройте подключение к БД

	store := NewParcelStore(db)
	parcel := getTestParcel() // настройте подключение к БД

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err)
	assert.NotEqual(t, 0, id)

	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	err = store.SetStatus(id, ParcelStatusSent)
	require.NoError(t, err)

	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	testParcel, err := store.Get(id)
	require.NoError(t, err)
	assert.NotEqual(t, testParcel.Status, parcel.Status)
	assert.Equal(t, testParcel.Status, ParcelStatusSent)
	assert.Equal(t, testParcel.Client, parcel.Client)
	assert.Equal(t, testParcel.CreatedAt, parcel.CreatedAt)
	assert.Equal(t, testParcel.Address, parcel.Address)

	err = store.Delete(id)
	require.NoError(t, err)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "./data/tracker.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		err = db.Close()
		if err != nil {
			fmt.Println("Error closing database:", err)
		}
	}() // настройте подключение к БД

	store := NewParcelStore(db)
	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i]) // добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
		require.NoError(t, err)
		assert.NotEqual(t, 0, id)

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client) // получите список посылок по идентификатору клиента, сохранённого в переменной client
	// убедитесь в отсутствии ошибки
	require.NoError(t, err)
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных
	assert.Equal(t, len(parcels), len(storedParcels))

	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		mappedParcel, ok := parcelMap[parcel.Number]
		require.Equal(t, true, ok)
		// убедитесь, что значения полей полученных посылок заполнены верно
		assert.Equal(t, mappedParcel, parcel)

		err = store.Delete(parcel.Number)
		require.NoError(t, err)
	}
}
