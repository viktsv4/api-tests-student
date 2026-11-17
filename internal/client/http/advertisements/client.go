package advertisements

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"api-tests-template/internal/constants/path"
	apiRunner "api-tests-template/internal/helpers/api-runner"
)

type CreateRequest struct {
	Title       string
	Description string
	Price       float64
	Quantity    int
	PhotoPaths  []string
}

func HttpPostAdvertisement(t *testing.T, token string, request CreateRequest) *http.Response {
	httpRequest := apiRunner.GetRunner().Auth(token).Create().Post(path.AdvertisementPath).
		MultipartFormData("title", request.Title).
		MultipartFormData("description", request.Description).
		MultipartFormData("price", strconv.FormatFloat(request.Price, 'f', -1, 64)).
		MultipartFormData("quantity", strconv.Itoa(request.Quantity))

	if len(request.PhotoPaths) > 0 {
		httpRequest.MultipartFile("photos", request.PhotoPaths...)
	}

	return httpRequest.Expect(t).End().Response
}

func HttpDeleteAdvertisement(t *testing.T, token, advertisementID string) *http.Response {
	return apiRunner.GetRunner().Auth(token).Create().Delete(path.AdvertisementPath).
		Query("id", advertisementID).
		Expect(t).End().Response
}

func HttpGetAdvertisementPhotos(t *testing.T, token, advertisementID string) *http.Response {
	requestPath := fmt.Sprintf("%s/%s/photos", path.AdvertisementsPath, advertisementID)

	return apiRunner.GetRunner().Auth(token).Create().Get(requestPath).
		Expect(t).End().Response
}
