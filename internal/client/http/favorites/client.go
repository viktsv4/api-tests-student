package favorites

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"api-tests-template/internal/constants/path"
	apiRunner "api-tests-template/internal/helpers/api-runner"
)

const retryAttempts = 3

type ListOptions struct {
	Page     int
	PageSize int
	Search   string
}

type AddRequest struct {
	AdvertisementID string `json:"advertisement_id"`
}

func HttpGetFavorites(t *testing.T, token string, options ListOptions) *http.Response {
	return retryServiceUnavailable(func() *http.Response {
		request := apiRunner.GetRunner().Auth(token).Create().Get(path.FavoritesPath)
		if options.Page != 0 {
			request.Query("page", strconv.Itoa(options.Page))
		}
		if options.PageSize != 0 {
			request.Query("page_size", strconv.Itoa(options.PageSize))
		}
		if options.Search != "" {
			request.Query("search", options.Search)
		}

		return request.Expect(t).End().Response
	})
}

func HttpPostFavorite(t *testing.T, token, advertisementID string) *http.Response {
	return retryServiceUnavailable(func() *http.Response {
		return apiRunner.GetRunner().Auth(token).Create().Post(path.FavoritesPath).
			ContentType("application/json").
			JSON(AddRequest{AdvertisementID: advertisementID}).
			Expect(t).End().Response
	})
}

func HttpDeleteFavorite(t *testing.T, token, advertisementID string) *http.Response {
	requestPath := fmt.Sprintf("%s/%s", path.FavoritesPath, advertisementID)

	return retryServiceUnavailable(func() *http.Response {
		return apiRunner.GetRunner().Auth(token).Create().Delete(requestPath).
			Expect(t).End().Response
	})
}

func retryServiceUnavailable(request func() *http.Response) *http.Response {
	response := request()
	for attempt := 1; attempt < retryAttempts && response.StatusCode == http.StatusServiceUnavailable; attempt++ {
		response.Body.Close()
		time.Sleep(time.Duration(attempt) * 300 * time.Millisecond)
		response = request()
	}

	return response
}
