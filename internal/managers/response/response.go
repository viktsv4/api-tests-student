package response

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func Body(t *testing.T, expectedStatusCode int, httpResponse *http.Response) string {
	t.Helper()
	defer httpResponse.Body.Close()

	require.Equalf(t, expectedStatusCode, httpResponse.StatusCode,
		"HTTP status code должен быть %d", expectedStatusCode)

	body, err := io.ReadAll(httpResponse.Body)
	require.NoError(t, err, "не удалось прочитать тело HTTP-ответа")

	return string(body)
}
