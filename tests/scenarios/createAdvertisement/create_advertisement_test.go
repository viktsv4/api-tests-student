package createAdvertisement

import (
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"

	advertisementsClient "api-tests-template/internal/client/http/advertisements"
	myAdvertisementsClient "api-tests-template/internal/client/http/myAdvertisements"
	"api-tests-template/internal/managers/advertisements"
	"api-tests-template/internal/managers/auth"
	"api-tests-template/internal/managers/auth/models"
	"api-tests-template/internal/managers/myAdvertisements"
	"api-tests-template/internal/utils"
	base "api-tests-template/tests"
)

const (
	advertisementDescription = "Объявление создано автоматическим API-тестом"
	advertisementPrice       = 1250.50
	advertisementQuantity    = 2
	invalidToken             = "invalid_token"
)

type testSuite struct {
	suite.Suite
	loginData models.LoginOkResponse
	photoPath string
}

func TestSuiteRun(t *testing.T) {
	suite.Run(t, &testSuite{})
}

func (s *testSuite) SetupSuite() {
	base.SetupSuite()
	base.Precondition("Авторизуем пользователя для создания объявлений")
	s.loginData = auth.Login(s.T(), os.Getenv("TEST_LOGIN"), os.Getenv("TEST_PASSWORD"))
	s.photoPath = base.CreatePNGFixture(s.T())
}

func (s *testSuite) TearDownSuite() {
	base.TearDownSuite()
}

func (s *testSuite) TestCreateAdvertisementWithAllParameters() {
	title := s.uniqueTitle()
	request := s.createRequest(title)

	base.Precondition("Создаём объявление со всеми параметрами и фотографией")
	createBody := advertisements.Create(s.T(), s.loginData.Token, http.StatusCreated, request)
	advertisementID := gjson.Get(createBody, "id").String()
	require.NotEmpty(s.T(), advertisementID, "в ответе создания отсутствует id объявления")
	s.T().Cleanup(func() {
		advertisements.Delete(s.T(), s.loginData.Token, http.StatusNoContent, advertisementID)
	})

	s.Run("объявление доступно по идентификатору с исходными параметрами", func() {
		base.Step("Получаем созданное объявление по идентификатору")
		body := myAdvertisements.GetAdvertisement(s.T(), http.StatusOK, advertisementID)
		s.assertAdvertisement(body, advertisementID, request)
		base.Result("Параметры полученного объявления совпадают с параметрами создания")
	})

	s.Run("объявление находится поиском по полному названию", func() {
		base.Step("Ищем объявление по полному названию")
		body := myAdvertisements.GetAdvertisements(s.T(), http.StatusOK,
			myAdvertisementsClient.AdvertisementListOptions{Search: title, Limit: 50})
		require.True(s.T(), base.ContainsAdvertisementID(gjson.Get(body, "items").Array(), advertisementID),
			"созданное объявление %s отсутствует в результатах поиска по %q", advertisementID, title)
		base.Result("Созданное объявление найдено в результатах поиска")
	})

	s.Run("загруженная фотография доступна через API", func() {
		base.Step("Получаем фотографии созданного объявления")
		body := advertisements.GetPhotos(s.T(), s.loginData.Token, http.StatusOK, advertisementID)
		photos := gjson.Parse(body).Array()
		require.NotEmpty(s.T(), photos, "API не вернуло загруженную фотографию")
		require.NotEmpty(s.T(), gjson.Get(photos[0].String(), "id").String(), "у фотографии отсутствует id")
		require.NotEmpty(s.T(), gjson.Get(photos[0].String(), "url").String(), "у фотографии отсутствует url")
		base.Result("API вернуло загруженную фотографию с идентификатором и URL")
	})
}

func (s *testSuite) TestCreateAdvertisementValidationErrors() {
	longTitle := strings.Repeat("a", 256)
	testCases := []struct {
		name        string
		request     advertisementsClient.CreateRequest
		messagePart string
	}{
		{name: "пустой_title", request: s.createRequest(""), messagePart: "title"},
		{name: "отрицательная_цена", request: s.createRequestWithPrice(s.uniqueTitle(), -1), messagePart: "price"},
		{name: "слишком_длинный_title", request: s.createRequest(longTitle), messagePart: "title"},
		{name: "без_фотографий", request: advertisementsClient.CreateRequest{
			Title: s.uniqueTitle(), Description: advertisementDescription, Price: advertisementPrice, Quantity: advertisementQuantity,
		}, messagePart: "photo"},
	}

	for _, testCase := range testCases {
		s.Run(testCase.name, func() {
			base.Precondition("Отправляем запрос создания с некорректными параметрами: " + testCase.name)
			base.Step("Создаём объявление с невалидными данными")
			body := advertisements.Create(s.T(), s.loginData.Token, http.StatusBadRequest, testCase.request)
			base.AssertAPIError(s.T(), body, "invalid_request", testCase.messagePart)
			base.Result("API вернуло ожидаемую ошибку валидации")
		})
	}
}

func (s *testSuite) TestCreateAdvertisementWithoutAuthorization() {
	for _, testCase := range []struct {
		name        string
		token       string
		messagePart string
	}{
		{name: "неверный_токен", token: invalidToken, messagePart: "invalid or expired token"},
		{name: "отсутствующий_токен", token: "", messagePart: "authorization"},
	} {
		s.Run(testCase.name, func() {
			base.Precondition("Отправляем запрос создания без действующей авторизации: " + testCase.name)
			base.Step("Создаём объявление без действующего токена")
			body := advertisements.Create(s.T(), testCase.token, http.StatusUnauthorized, s.createRequest(s.uniqueTitle()))
			base.AssertAPIError(s.T(), body, "unauthorized", testCase.messagePart)
			base.Result("API вернуло ожидаемую ошибку авторизации")
		})
	}
}

func (s *testSuite) createRequest(title string) advertisementsClient.CreateRequest {
	return s.createRequestWithPrice(title, advertisementPrice)
}

func (s *testSuite) createRequestWithPrice(title string, price float64) advertisementsClient.CreateRequest {
	return advertisementsClient.CreateRequest{
		Title: title, Description: advertisementDescription, Price: price, Quantity: advertisementQuantity, PhotoPaths: []string{s.photoPath},
	}
}

func (s *testSuite) uniqueTitle() string {
	return "API test " + utils.RandomString(16)
}

func (s *testSuite) assertAdvertisement(body, advertisementID string, request advertisementsClient.CreateRequest) {
	require.Equal(s.T(), advertisementID, gjson.Get(body, "id").String(), "API вернуло другое объявление")
	require.Equal(s.T(), request.Title, gjson.Get(body, "title").String(), "title отличается от переданного")
	require.Equal(s.T(), request.Description, gjson.Get(body, "description").String(), "description отличается от переданного")
	require.Equal(s.T(), request.Price, gjson.Get(body, "price").Float(), "price отличается от переданного")
	require.Equal(s.T(), int64(request.Quantity), gjson.Get(body, "quantity").Int(), "quantity отличается от переданного")
	require.Equal(s.T(), s.loginData.User.Id, gjson.Get(body, "user_id").String(), "user_id отличается от владельца")
}
