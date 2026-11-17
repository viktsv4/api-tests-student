package tests

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

var onePixelPNG = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
	0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
	0x89, 0x00, 0x00, 0x00, 0x0d, 0x49, 0x44, 0x41,
	0x54, 0x08, 0xd7, 0x63, 0xf8, 0xcf, 0xc0, 0xf0,
	0x1f, 0x00, 0x05, 0x00, 0x01, 0xff, 0x89, 0x99,
	0x3d, 0x1d, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45,
	0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
}

// CreatePNGFixture создаёт временный PNG-файл для загрузки в объявление.
func CreatePNGFixture(t *testing.T) string {
	t.Helper()

	file, err := os.CreateTemp("", "testboard-photo-*.png")
	require.NoError(t, err, "не удалось создать временную фотографию")

	_, err = file.Write(onePixelPNG)
	require.NoError(t, err, "не удалось записать временную фотографию")
	require.NoError(t, file.Close(), "не удалось закрыть временную фотографию")
	t.Cleanup(func() {
		require.NoError(t, os.Remove(file.Name()), "не удалось удалить временную фотографию")
	})

	return file.Name()
}

// AssertAPIError проверяет код и сообщение ошибки из ответа API.
func AssertAPIError(t *testing.T, body, expectedError, messagePart string) {
	t.Helper()

	require.Equal(t, expectedError, gjson.Get(body, "error").String(), "некорректный код ошибки API")
	message := gjson.Get(body, "message").String()
	require.NotEmpty(t, message, "в ответе об ошибке отсутствует message")
	if messagePart != "" {
		require.Truef(t, strings.Contains(strings.ToLower(message), strings.ToLower(messagePart)),
			"сообщение об ошибке %q не содержит %q", message, messagePart)
	}
}

// ContainsAdvertisementID определяет, есть ли объявление в списке API-ответа.
func ContainsAdvertisementID(items []gjson.Result, advertisementID string) bool {
	for _, item := range items {
		if gjson.Get(item.String(), "id").String() == advertisementID ||
			gjson.Get(item.String(), "advertisement.id").String() == advertisementID {
			return true
		}
	}

	return false
}
