package myAdvertisements

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"api-tests-template/internal/client/http/myAdvertisements"
)

// GetMyAdvertisements возвращает список объявлений авторизованного пользователя и проверяет ожидаемый статус-код
func GetMyAdvertisements(t *testing.T, token string, expectedStatusCode int, limit int) string {
	return GetMyAdvertisementsWithOptions(t, token, expectedStatusCode,
		myAdvertisements.AdvertisementListOptions{Limit: limit})
}

func GetMyAdvertisementsWithOptions(
	t *testing.T,
	token string,
	expectedStatusCode int,
	options myAdvertisements.AdvertisementListOptions,
) string {
	return responseBody(t, expectedStatusCode, myAdvertisements.HttpGetMyAdvertisementsWithOptions(t, token, options))
}

func GetAdvertisements(t *testing.T, expectedStatusCode int, options myAdvertisements.AdvertisementListOptions) string {
	return responseBody(t, expectedStatusCode, myAdvertisements.HttpGetAdvertisements(t, options))
}

func GetAdvertisementsWithToken(
	t *testing.T,
	token string,
	expectedStatusCode int,
	options myAdvertisements.AdvertisementListOptions,
) string {
	return responseBody(t, expectedStatusCode, myAdvertisements.HttpGetAdvertisementsWithToken(t, token, options))
}

func GetAdvertisement(t *testing.T, expectedStatusCode int, id string) string {
	return responseBody(t, expectedStatusCode, myAdvertisements.HttpGetAdvertisement(t, id))
}

func responseBody(t *testing.T, expectedStatusCode int, response *http.Response) string {
	defer response.Body.Close()

	require.Equalf(t, expectedStatusCode, response.StatusCode,
		"HTTP status code должен быть %d", expectedStatusCode)

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	return string(body)
}
