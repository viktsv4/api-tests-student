package advertisements

import (
	"testing"

	advertisementsClient "api-tests-template/internal/client/http/advertisements"
	"api-tests-template/internal/managers/response"
)

func Create(
	t *testing.T,
	token string,
	expectedStatusCode int,
	request advertisementsClient.CreateRequest,
) string {
	return response.Body(t, expectedStatusCode, advertisementsClient.HttpPostAdvertisement(t, token, request))
}

func Delete(t *testing.T, token string, expectedStatusCode int, advertisementID string) string {
	return response.Body(t, expectedStatusCode,
		advertisementsClient.HttpDeleteAdvertisement(t, token, advertisementID))
}

func GetPhotos(t *testing.T, token string, expectedStatusCode int, advertisementID string) string {
	return response.Body(t, expectedStatusCode,
		advertisementsClient.HttpGetAdvertisementPhotos(t, token, advertisementID))
}
