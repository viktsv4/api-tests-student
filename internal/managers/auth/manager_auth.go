package auth

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"api-tests-template/internal/client/http/auth"
	"api-tests-template/internal/managers/auth/models"
)

// Login выполняет авторизацию пользователя и возвращает токен с данными пользователя, ожидает 200
func Login(t *testing.T, username, password string) models.LoginOkResponse {
	require.NotEmpty(t, username, "Отсутствует логин пользователя")
	require.NotEmpty(t, password, "Отсутствует пароль пользователя")

	request := models.LoginRequest{
		Email:    username,
		Password: password,
	}

	loginRawResponse := auth.HttpPostAuthLogin(t, request)
	defer loginRawResponse.Body.Close()

	respArray, errResponse := io.ReadAll(loginRawResponse.Body)
	require.NoError(t, errResponse)

	loginResponse := models.LoginOkResponse{}
	err := json.Unmarshal(respArray, &loginResponse)
	require.NoError(t, err)

	return loginResponse
}

func Register(t *testing.T, expectedStatusCode int, request models.RegisterRequest) models.LoginOkResponse {
	registerRawResponse := auth.HttpPostAuthRegister(t, request)
	defer registerRawResponse.Body.Close()

	require.Equalf(t, expectedStatusCode, registerRawResponse.StatusCode,
		"HTTP status code должен быть %d", expectedStatusCode)

	responseBytes, err := io.ReadAll(registerRawResponse.Body)
	require.NoError(t, err, "не удалось прочитать ответ регистрации")

	registerResponse := models.LoginOkResponse{}
	err = json.Unmarshal(responseBytes, &registerResponse)
	require.NoError(t, err, "ошибка регистрации")
	require.NotEmpty(t, registerResponse.Token, "в ответе регистрации отсутствует токен")

	return registerResponse
}
