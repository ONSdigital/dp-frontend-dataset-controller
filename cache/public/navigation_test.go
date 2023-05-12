package public

import (
	"context"
	"testing"

	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-topic-api/models"
	"github.com/ONSdigital/dp-topic-api/sdk"
	apiError "github.com/ONSdigital/dp-topic-api/sdk/errors"
	mockTopicCli "github.com/ONSdigital/dp-topic-api/sdk/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	testNavData = &models.Navigation{
		Description: "This is the top-level description.",
		Links:       testTopicLinks,
		Items:       testTopicNonReferential,
	}

	testNavDataCY = &models.Navigation{
		Description: "This is the WELSH top-level description.",
		Links:       testTopicLinks,
		Items:       testTopicNonReferential,
	}

	testTopicLinks = &models.TopicLinks{
		Self:      testLinkObject,
		Subtopics: testLinkObject,
		Content:   testLinkObject,
	}

	testLinkObject = &models.LinkObject{
		HRef: "https://www.example.com",
		ID:   "1234",
	}

	testTopicNonReferential = &[]models.TopicNonReferential{
		{
			Description: "This is an item description.",
			Label:       "Item label",
			Links:       testTopicLinks,
			Name:        "Item name.",
			Title:       "Item title.",
			URI:         "Item URI.",
		},
	}
)

func TestUpdateNavigationData(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg, err := config.Get()
	if err != nil {
		t.Errorf("failed to get config")
	}
	cfg.EnableNewNavBar = true

	mockedNavigationClient := &mockTopicCli.ClienterMock{
		GetNavigationPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers, options sdk.Options) (*models.Navigation, apiError.Error) {
			if options.Lang == "cy" {
				return testNavDataCY, nil
			}
			return testNavData, nil
		},
	}

	Convey("Given navigation data is being served by the topic API", t, func() {

		Convey("When UpdateNavigationData is called", func() {
			respNavigationCache := UpdateNavigationData(ctx, cfg, "en", mockedNavigationClient)()

			Convey("Then the navigation data is returned", func() {
				So(respNavigationCache, ShouldNotBeNil)

				So(respNavigationCache.Description, ShouldEqual, testNavData.Description)
				So(respNavigationCache.Links, ShouldEqual, testNavData.Links)
				So(respNavigationCache.Items, ShouldEqual, testNavData.Items)
			})
		})

		Convey("When UpdateNavigationData is called with Welsh specified", func() {
			respNavigationCache := UpdateNavigationData(ctx, cfg, "cy", mockedNavigationClient)()

			Convey("Then the navigation data is returned", func() {
				So(respNavigationCache, ShouldNotBeNil)

				So(respNavigationCache.Description, ShouldEqual, testNavDataCY.Description)
			})
		})
	})
}
