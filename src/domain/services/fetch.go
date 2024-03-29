package services

import (
	"reflect"
	"strconv"
	"sync"

	"github.com/Oleg-Skalozub/testtask/src/domain/entity"
	"github.com/Oleg-Skalozub/testtask/src/domain/repository"
	"github.com/Oleg-Skalozub/testtask/src/infrastructure/client"
	"github.com/Oleg-Skalozub/testtask/src/infrastructure/config"
	"github.com/Oleg-Skalozub/testtask/src/infrastructure/errscan"
)

var wg sync.WaitGroup

// Fetch ...
type Fetcher interface {
	FetchData(day, month int) ([]entity.DataResponse, error)
	GetData(day, month int) ([]entity.DataResponse, error)
	SaveData(eventType, day, month int, events []entity.Event)
}

type fetch struct {
	dataRepository repository.DataRepository
	client         client.ClientInterface
	config         config.Configuration
}

// NewFetch ...
func NewFetch() Fetcher {
	return &fetch{
		dataRepository: repository.NewDataRepository(),
		client:         client.NewClient(),
		config:         config.Config,
	}
}

// FetchData ...
func (f fetch) FetchData(day, month int) ([]entity.DataResponse, error) {

	data, err := f.GetData(day, month)
	if err != nil && err != errscan.EmptyResultError {
		return nil, err
	}
	if data != nil {
		return data, nil
	}

	res, err := f.client.Get(f.config.ApiRoute, day, month)
	if err != nil {
		return nil, err
	}

	ref := reflect.ValueOf(res.Data)

	wg.Add(ref.NumField())

	for i := 0; i < ref.NumField(); i++ {
		dataProcess := ref.Field(i).Interface().([]entity.Event)
		iventType := entity.TaskingTypeNameMap[ref.Type().Field(i).Name]

		go f.SaveData(iventType, day, month, dataProcess)

		data = append(data, entity.DataResponse{
			EventType: entity.TaskingNameTypeMap[iventType],
			Result:    len(dataProcess),
		})
	}
	wg.Wait()

	return data, err
}

// SaveData ...
func (f fetch) SaveData(eventType, day, month int, events []entity.Event) {
	defer wg.Done()
	for _, event := range events {
		t := f.dataRepository.SaveData(day, month, eventType, event.Year, event.Text)
		_ = t // todo add error handling
	}
}

// GetData ...
func (f fetch) GetData(day, month int) ([]entity.DataResponse, error) {
	data, err := f.dataRepository.GetData(day, month)
	if err != nil {
		return nil, err
	}

	cash := false
	for i, dat := range data {
		res, err := strconv.Atoi(dat.EventType)
		if err != nil {
			return nil, err
		}
		data[i].EventType = entity.TaskingNameTypeMap[res]
		if dat.Result != 0 {
			cash = true
		}
	}
	if cash == true {
		return data, nil
	}

	return nil, errscan.EmptyResultError
}
