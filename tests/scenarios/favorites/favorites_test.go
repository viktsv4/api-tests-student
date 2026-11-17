package favorites

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"

	advertisementsClient "api-tests-template/internal/client/http/advertisements"
	favoritesClient "api-tests-template/internal/client/http/favorites"
	"api-tests-template/internal/managers/advertisements"
	"api-tests-template/internal/managers/auth"
	"api-tests-template/internal/managers/auth/models"
	"api-tests-template/internal/managers/favorites"
	"api-tests-template/internal/utils"
	base "api-tests-template/tests"
)

const (
	temporaryUserPassword = "ApiTests123!"
	temporaryUserName     = "API Test"
	advertisementPrice    = 99.90
	advertisementQuantity = 2
	nonexistentAdID       = "11111111-1111-4111-8111-111111111111"
)

type testSuite struct {
	suite.Suite
	owner     models.LoginOkResponse
	favoriter models.LoginOkResponse
	photoPath string
}

func TestSuiteRun(t *testing.T) {
	suite.Run(t, &testSuite{})
}

func (s *testSuite) SetupSuite() {
	base.SetupSuite()
	base.Precondition("Регистрируем независимых владельца и пользователя избранного")
	s.owner = s.registerTemporaryUser()
	s.favoriter = s.registerTemporaryUser()
	s.photoPath = base.CreatePNGFixture(s.T())
}

func (s *testSuite) TearDownSuite() {
	base.TearDownSuite()
}

func (s *testSuite) TestFavoritesLifecycle() {
	advertisementID := s.createOwnerAdvertisement()
	deleted := false
	s.T().Cleanup(func() {
		if !deleted {
			advertisements.Delete(s.T(), s.owner.Token, http.StatusNoContent, advertisementID)
		}
	})

	s.Run("добавленное объявление появляется в избранном", func() {
		base.Step("Добавляем объявление в избранное")
		body := favorites.Add(s.T(), s.favoriter.Token, http.StatusCreated, advertisementID)
		require.Equal(s.T(), advertisementID, gjson.Get(body, "advertisement.id").String(),
			"в ответе добавления вернулось другое объявление")
		s.requireFavoritePresence(advertisementID, true)
		base.Result("Добавленное объявление отображается в избранном")
	})

	s.Run("удалённое из избранного объявление исчезает из списка", func() {
		base.Step("Удаляем объявление из избранного")
		favorites.Delete(s.T(), s.favoriter.Token, http.StatusNoContent, advertisementID)
		s.requireFavoritePresence(advertisementID, false)
		base.Result("Удалённое объявление отсутствует в избранном")
	})

	s.Run("после удаления объявления владельцем оно исчезает из избранного", func() {
		base.Step("Добавляем объявление в избранное и удаляем его владельцем")
		favorites.Add(s.T(), s.favoriter.Token, http.StatusCreated, advertisementID)
		advertisements.Delete(s.T(), s.owner.Token, http.StatusNoContent, advertisementID)
		deleted = true
		s.requireFavoritePresence(advertisementID, false)
		base.Result("Удалённое владельцем объявление отсутствует в избранном")
	})
}

func (s *testSuite) TestFavoriteNegativeScenarios() {
	base.Precondition("Проверяем ошибки ручек избранного")

	s.Run("несуществующее_объявление_нельзя_добавить", func() {
		base.Step("Добавляем несуществующее объявление в избранное")
		body := favorites.Add(s.T(), s.favoriter.Token, http.StatusNotFound, nonexistentAdID)
		base.AssertAPIError(s.T(), body, "not_found", "advertisement")
		base.Result("API вернуло ожидаемую ошибку отсутствующего объявления")
	})

	s.Run("несуществующее_объявление_нельзя_удалить", func() {
		base.Step("Удаляем несуществующее объявление из избранного")
		body := favorites.Delete(s.T(), s.favoriter.Token, http.StatusNotFound, nonexistentAdID)
		base.AssertAPIError(s.T(), body, "not_found", "favorite")
		base.Result("API вернуло ожидаемую ошибку отсутствующего избранного")
	})

	s.Run("ручки_избранного_требуют_авторизацию", func() {
		for _, testCase := range []struct {
			name string
			call func() string
		}{
			{name: "список", call: func() string {
				return favorites.GetList(s.T(), "", http.StatusUnauthorized, favoritesClient.ListOptions{Page: 1, PageSize: 50})
			}},
			{name: "добавление", call: func() string {
				return favorites.Add(s.T(), "", http.StatusUnauthorized, nonexistentAdID)
			}},
			{name: "удаление", call: func() string {
				return favorites.Delete(s.T(), "", http.StatusUnauthorized, nonexistentAdID)
			}},
		} {
			s.Run(testCase.name, func() {
				base.Step("Вызываем ручку избранного без токена")
				body := testCase.call()
				base.AssertAPIError(s.T(), body, "unauthorized", "authorization")
				base.Result("API вернуло ожидаемую ошибку авторизации")
			})
		}
	})
}

func (s *testSuite) TestFavoriteBusinessRestrictions() {
	base.Precondition("Создаём объявление для проверки ограничений избранного")
	advertisementID := s.createOwnerAdvertisement()
	s.T().Cleanup(func() {
		advertisements.Delete(s.T(), s.owner.Token, http.StatusNoContent, advertisementID)
	})

	s.Run("одно_объявление_нельзя_добавить_дважды", func() {
		base.Step("Повторно добавляем объявление в избранное")
		favorites.Add(s.T(), s.favoriter.Token, http.StatusCreated, advertisementID)
		s.T().Cleanup(func() {
			favorites.Delete(s.T(), s.favoriter.Token, http.StatusNoContent, advertisementID)
		})

		body := favorites.Add(s.T(), s.favoriter.Token, http.StatusConflict, advertisementID)
		base.AssertAPIError(s.T(), body, "already_exists", "already")
		base.Result("API запретило повторное добавление в избранное")
	})

	s.Run("собственное_объявление_нельзя_добавить_в_избранное", func() {
		base.Step("Добавляем собственное объявление в избранное")
		body := favorites.Add(s.T(), s.owner.Token, http.StatusForbidden, advertisementID)
		base.AssertAPIError(s.T(), body, "forbidden", "own")
		base.Result("API запретило добавлять собственное объявление в избранное")
	})
}

func (s *testSuite) registerTemporaryUser() models.LoginOkResponse {
	email := strings.ToLower(utils.RandomString(20)) + "@example.com"

	return auth.Register(s.T(), http.StatusCreated, models.RegisterRequest{
		Email: email, FirstName: temporaryUserName, Password: temporaryUserPassword,
	})
}

func (s *testSuite) createOwnerAdvertisement() string {
	title := "Favorite API test " + utils.RandomString(16)
	base.Precondition("Создаём объявление владельца для проверки избранного: " + title)
	body := advertisements.Create(s.T(), s.owner.Token, http.StatusCreated, advertisementsClient.CreateRequest{
		Title: title, Description: "Объявление для API-теста избранного", Price: advertisementPrice,
		Quantity: advertisementQuantity, PhotoPaths: []string{s.photoPath},
	})
	advertisementID := gjson.Get(body, "id").String()
	require.NotEmpty(s.T(), advertisementID, "в ответе создания отсутствует id объявления")

	return advertisementID
}

func (s *testSuite) requireFavoritePresence(advertisementID string, expectedPresent bool) {
	body := favorites.GetList(s.T(), s.favoriter.Token, http.StatusOK, favoritesClient.ListOptions{Page: 1, PageSize: 50})
	present := base.ContainsAdvertisementID(gjson.Get(body, "items").Array(), advertisementID)
	require.Equalf(s.T(), expectedPresent, present, "неожиданное наличие объявления %s в избранном", advertisementID)
}
