package favorites

import (
	"testing"

	favoritesClient "api-tests-template/internal/client/http/favorites"
	"api-tests-template/internal/managers/response"
)

func GetList(t *testing.T, token string, expectedStatusCode int, options favoritesClient.ListOptions) string {
	return response.Body(t, expectedStatusCode, favoritesClient.HttpGetFavorites(t, token, options))
}

func Add(t *testing.T, token string, expectedStatusCode int, advertisementID string) string {
	return response.Body(t, expectedStatusCode, favoritesClient.HttpPostFavorite(t, token, advertisementID))
}

func Delete(t *testing.T, token string, expectedStatusCode int, advertisementID string) string {
	return response.Body(t, expectedStatusCode, favoritesClient.HttpDeleteFavorite(t, token, advertisementID))
}
