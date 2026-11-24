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
	"api-tests-template/internal/utils"
	base "api-tests-template/tests"
)

type TestSuite struct {
	suite.Suite
	advertisementID string
	searchTitle     string
	loginData       models.LoginOkResponse
}

func TestSuiteRun(t *testing.T) {
	suite.Run(t, &TestSuite{})
}

func (s *TestSuite) SetupSuite() {
	base.SetupSuite()

	s.loginData = auth.Login(s.T(), os.Getenv("TEST_LOGIN"), os.Getenv("TEST_PASSWORD"))
	base.Precondition("Создаём объявления для проверки каталога, поиска и сортировки по цене")
	s.searchTitle = "Catalog" + utils.RandomString(18)
	s.advertisementID = base.CreateAdvertisementFixture(s.T(), s.loginData.Token, s.searchTitle, 100)
	base.CreateAdvertisementFixture(s.T(), s.loginData.Token, "Catalog"+utils.RandomString(18), 200)
}

func (s *TestSuite) TearDownSuite() {
	base.TearDownSuite()
}

func (s *TestSuite) TestGetAdvertisementsPositive() {
	base.Precondition("Запрашиваем первую страницу публичного каталога объявлений")
	body := myAdvertisements.GetAdvertisements(s.T(), http.StatusOK,
		myAdvertisementsClient.AdvertisementListOptions{Limit: 10})

	items := gjson.Get(body, "items")
	require.True(s.T(), items.IsArray(), "items должен быть массивом")
	require.LessOrEqual(s.T(), len(items.Array()), 10, "превышен запрошенный limit")
	require.True(s.T(), gjson.Get(body, "total").Exists(), "в ответе отсутствует total")
	require.True(s.T(), gjson.Get(body, "has_next").Exists(), "в ответе отсутствует has_next")

	for _, item := range items.Array() {
		s.assertAdvertisementFields(item)
	}
}

func (s *TestSuite) TestGetAdvertisementByIDPositive() {
	id := s.advertisementID
	body := myAdvertisements.GetAdvertisement(s.T(), http.StatusOK, id)
	advertisement := gjson.Parse(body)

	require.Equal(s.T(), id, gjson.Get(body, "id").String(), "API вернуло другое объявление")
	s.assertAdvertisementFields(advertisement)
}

func (s *TestSuite) TestSearchAdvertisements() {
	title := s.searchTitle

	cases := []struct {
		name  string
		query string
	}{
		{name: "начало", query: title[:13]},
		{name: "середина", query: title[7:17]},
		{name: "конец", query: title[len(title)-10:]},
		{name: "нижний_регистр", query: strings.ToLower(title)},
		{name: "верхний_регистр", query: strings.ToUpper(title)},
		{name: "смешанный_регистр", query: "cAtAlOg" + title[len("Catalog"):]},
	}

	for _, testCase := range cases {
		s.Run(testCase.name, func() {
			body := myAdvertisements.GetAdvertisementsWithToken(s.T(), s.loginData.Token, http.StatusOK,
				myAdvertisementsClient.AdvertisementListOptions{Search: testCase.query, Limit: 50})
			items := gjson.Get(body, "items").Array()

			require.True(s.T(), base.ContainsAdvertisementID(items, s.advertisementID),
				"в результатах поиска %q нет объявления %q", testCase.query, title)
		})
	}
}

func (s *TestSuite) TestSortAdvertisementsByPrice() {
	for _, order := range []string{"asc", "desc"} {
		s.Run("price_"+order, func() {
			body := myAdvertisements.GetAdvertisements(s.T(), http.StatusOK,
				myAdvertisementsClient.AdvertisementListOptions{SortBy: "price", SortOrder: order, Limit: 10})
			items := gjson.Get(body, "items").Array()

			require.GreaterOrEqual(s.T(), len(items), 2, "для проверки сортировки нужны две цены")
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
	photos := gjson.Get(item.String(), "photos")
	require.True(s.T(), photos.IsArray(), "photos должен быть массивом")
}
