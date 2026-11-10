package myAdvertisements

import (
	"net/http"
	"strconv"
	"testing"

	"api-tests-template/internal/constants/path"
	apiRunner "api-tests-template/internal/helpers/api-runner"
)

type AdvertisementListOptions struct {
	Search    string
	PriceSort string
	Quantity  *bool
	Limit     int
	Cursor    string
	SortBy    string
	SortOrder string
}

// HttpGetMyAdvertisements выполняет GET-запрос для получения списка объявлений авторизованного пользователя
func HttpGetMyAdvertisements(t *testing.T, token string, limit int) *http.Response {
	return HttpGetMyAdvertisementsWithOptions(t, token, AdvertisementListOptions{Limit: limit})
}

func HttpGetMyAdvertisementsWithOptions(t *testing.T, token string, options AdvertisementListOptions) *http.Response {
	return getAdvertisements(t, apiRunner.GetRunner().Auth(token), path.MyAdvertisementsPath, options)
}

func HttpGetAdvertisements(t *testing.T, options AdvertisementListOptions) *http.Response {
	return getAdvertisements(t, apiRunner.GetRunner(), path.AdvertisementsPath, options)
}

func HttpGetAdvertisementsWithToken(t *testing.T, token string, options AdvertisementListOptions) *http.Response {
	return getAdvertisements(t, apiRunner.GetRunner().Auth(token), path.AdvertisementsPath, options)
}

func HttpGetAdvertisement(t *testing.T, id string) *http.Response {
	return apiRunner.GetRunner().Create().Get(path.AdvertisementPath).
		Query("id", id).
		Expect(t).
		End().Response
}

func getAdvertisements(t *testing.T, runner *apiRunner.ApiTest, requestPath string, options AdvertisementListOptions) *http.Response {
	request := runner.Create().Get(requestPath)
	if options.Search != "" {
		request.Query("search", options.Search)
	}
	if options.PriceSort != "" {
		request.Query("price_sort", options.PriceSort)
	}
	if options.Quantity != nil {
		request.Query("quantity", strconv.FormatBool(*options.Quantity))
	}
	if options.Limit != 0 {
		request.Query("limit", strconv.Itoa(options.Limit))
	}
	if options.Cursor != "" {
		request.Query("cursor", options.Cursor)
	}
	if options.SortBy != "" {
		request.Query("sort_by", options.SortBy)
	}
	if options.SortOrder != "" {
		request.Query("sort_order", options.SortOrder)
	}

	return request.
		Expect(t).
		End().Response
}
