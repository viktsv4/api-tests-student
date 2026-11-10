package advertisement

import (
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"

	myAdvertisementsClient "api-tests-template/internal/client/http/myAdvertisements"
	"api-tests-template/internal/managers/auth"
	"api-tests-template/internal/managers/auth/models"
	"api-tests-template/internal/managers/myAdvertisements"
	base "api-tests-template/tests"
)

type TestSuite struct {
	suite.Suite
	advertisements gjson.Result
	loginData      models.LoginOkResponse
}

const searchAdvertisementTitle = "ertyu"

func TestSuiteRun(t *testing.T) {
	suite.Run(t, &TestSuite{})
}

func (s *TestSuite) SetupSuite() {
	base.SetupSuite()

	base.Precondition("Получаем объявления, которые будут использоваться в проверках каталога")
	body := myAdvertisements.GetAdvertisements(s.T(), http.StatusOK,
		myAdvertisementsClient.AdvertisementListOptions{Limit: 10})
	s.advertisements = gjson.Get(body, "items")
	require.NotEmpty(s.T(), s.advertisements.Array(), "в публичном каталоге нет объявлений для проверки")
}

func (s *TestSuite) TearDownSuite() {
	base.TearDownSuite()
}

func (s *TestSuite) TestGetAdvertisementsPositive() {
	base.Precondition("Запрашиваем первую страницу публичного каталога объявлений")
	body := myAdvertisements.GetAdvertisements(s.T(), http.StatusOK,
		myAdvertisementsClient.AdvertisementListOptions{Limit: 10})

	s.advertisements = gjson.Get(body, "items")
	require.True(s.T(), s.advertisements.IsArray(), "items должен быть массивом")
	require.LessOrEqual(s.T(), len(s.advertisements.Array()), 10, "превышен запрошенный limit")
	require.True(s.T(), gjson.Get(body, "total").Exists(), "в ответе отсутствует total")
	require.True(s.T(), gjson.Get(body, "has_next").Exists(), "в ответе отсутствует has_next")

	for _, item := range s.advertisements.Array() {
		s.assertAdvertisementFields(item)
	}
}

func (s *TestSuite) TestGetAdvertisementByIDPositive() {
	id := gjson.Get(s.advertisements.Array()[0].String(), "id").String()
	body := myAdvertisements.GetAdvertisement(s.T(), http.StatusOK, id)
	advertisement := gjson.Parse(body)

	require.Equal(s.T(), id, gjson.Get(body, "id").String(), "API вернуло другое объявление")
	s.assertAdvertisementFields(advertisement)
}

func (s *TestSuite) TestSearchAdvertisements() {
	s.loginData = auth.Login(s.T(), os.Getenv("TEST_LOGIN"), os.Getenv("TEST_PASSWORD"))

	cases := []struct {
		name  string
		query string
	}{
		{name: "начало", query: "er"},
		{name: "середина", query: "ty"},
		{name: "конец", query: "yu"},
		{name: "нижний_регистр", query: strings.ToLower(searchAdvertisementTitle)},
		{name: "верхний_регистр", query: strings.ToUpper(searchAdvertisementTitle)},
		{name: "смешанный_регистр", query: "eRtYu"},
	}

	for _, testCase := range cases {
		s.Run(testCase.name, func() {
			body := myAdvertisements.GetAdvertisementsWithToken(s.T(), s.loginData.Token, http.StatusOK,
				myAdvertisementsClient.AdvertisementListOptions{Search: testCase.query, Limit: 50})
			items := gjson.Get(body, "items").Array()

			require.NotEmpty(s.T(), items, "поиск не вернул результатов")
			require.True(s.T(), containsAdvertisementTitle(items, searchAdvertisementTitle),
				"в результатах поиска %q нет объявления %q", testCase.query, searchAdvertisementTitle)
		})
	}
}

func (s *TestSuite) TestSortAdvertisementsByPrice() {
	for _, order := range []string{"asc", "desc"} {
		s.Run("price_"+order, func() {
			body := myAdvertisements.GetAdvertisements(s.T(), http.StatusOK,
				myAdvertisementsClient.AdvertisementListOptions{SortBy: "price", SortOrder: order, Limit: 10})
			items := gjson.Get(body, "items").Array()

			for index := 1; index < len(items); index++ {
				previousPrice := gjson.Get(items[index-1].String(), "price").Float()
				currentPrice := gjson.Get(items[index].String(), "price").Float()
				if order == "asc" {
					require.LessOrEqual(s.T(), previousPrice, currentPrice, "цены не отсортированы по возрастанию")
				} else {
					require.GreaterOrEqual(s.T(), previousPrice, currentPrice, "цены не отсортированы по убыванию")
				}
			}
		})
	}
}

func (s *TestSuite) TestGetAdvertisementsInvalidLimit() {
	myAdvertisements.GetAdvertisements(s.T(), http.StatusBadRequest,
		myAdvertisementsClient.AdvertisementListOptions{Limit: 51})
}

func (s *TestSuite) assertAdvertisementFields(item gjson.Result) {
	for _, field := range []string{"id", "title", "description", "user_id", "created_at", "updated_at"} {
		require.NotEmpty(s.T(), gjson.Get(item.String(), field).String(), "%s пустое", field)
	}
	require.NotEmpty(s.T(), gjson.Get(item.String(), "photos").Array(), "у объявления нет фотографий")
}

func containsAdvertisementTitle(items []gjson.Result, title string) bool {
	for _, item := range items {
		if gjson.Get(item.String(), "title").String() == title {
			return true
		}
	}

	return false
}
