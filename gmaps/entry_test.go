package gmaps_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/require"

	"github.com/gosom/google-maps-scraper/gmaps"
)

func createGoQueryFromFile(t *testing.T, path string) *goquery.Document {
	t.Helper()

	fd, err := os.Open(path)
	require.NoError(t, err)

	defer fd.Close()

	doc, err := goquery.NewDocumentFromReader(fd)
	require.NoError(t, err)

	return doc
}

func Test_EntryFromJSON(t *testing.T) {
	expected := gmaps.Entry{
		Link:       "https://www.google.com/maps/place/Kipriakon/data=!4m2!3m1!1s0x14e732fd76f0d90d:0xe5415928d6702b47!10m1!1e1",
		Title:      "Kipriakon",
		Category:   "Restaurant",
		Categories: []string{"Restaurant"},
		Address:    "Old port, Limassol 3042",
		OpenHours: map[string][]string{
			"Monday":    {"12:30–10 pm"},
			"Tuesday":   {"12:30–10 pm"},
			"Wednesday": {"12:30–10 pm"},
			"Thursday":  {"12:30–10 pm"},
			"Friday":    {"12:30–10 pm"},
			"Saturday":  {"12:30–10 pm"},
			"Sunday":    {"12:30–10 pm"},
		},
		WebSite:      "",
		Phone:        "25 101555",
		PlusCode:     "M2CR+6X Limassol",
		ReviewCount:  396,
		ReviewRating: 4.2,
		Latitude:     34.670595399999996,
		Longtitude:   33.042456699999995,
		Cid:          "16519582940102929223",
		Status:       "Closed ⋅ Opens 12:30\u202fpm Tue",
		ReviewsLink:  "https://search.google.com/local/reviews?placeid=ChIJDdnwdv0y5xQRRytw1ihZQeU&q=Kipriakon&authuser=0&hl=en&gl=CY",
		Thumbnail:    "https://lh5.googleusercontent.com/p/AF1QipP4Y7A8nYL3KKXznSl69pXSq9p2IXCYUjVvOh0F=w408-h408-k-no",
		Timezone:     "Asia/Nicosia",
		PriceRange:   "€€",
		DataID:       "0x14e732fd76f0d90d:0xe5415928d6702b47",
		PlaceID:      "ChIJDdnwdv0y5xQRRytw1ihZQeU",
		Images: []gmaps.Image{
			{
				Title: "All",
				Image: "https://lh5.googleusercontent.com/p/AF1QipP4Y7A8nYL3KKXznSl69pXSq9p2IXCYUjVvOh0F=w298-h298-k-no",
			},
			{
				Title: "Latest",
				Image: "https://lh5.googleusercontent.com/p/AF1QipNgMqyaQs2MqH1oiGC44eDcvudurxQfNb2RuDsd=w224-h298-k-no",
			},
			{
				Title: "Videos",
				Image: "https://lh5.googleusercontent.com/p/AF1QipPZbq8v8K8RZfvL6gZ_4Dw6qwNJ_MUxxOOfBo7h=w224-h398-k-no",
			},
			{
				Title: "Menu",
				Image: "https://lh5.googleusercontent.com/p/AF1QipNhoFtPcaLCIhdN3GhlJ6sQIvdhaESnRG8nyeC8=w397-h298-k-no",
			},
			{
				Title: "Food & drink",
				Image: "https://lh5.googleusercontent.com/p/AF1QipMbu-iiWkE4DsXx3aI7nGaqyXJKbBYCrBXvzOnu=w298-h298-k-no",
			},
			{
				Title: "Vibe",
				Image: "https://lh5.googleusercontent.com/p/AF1QipOGg_vrD4bzkOre5Ly6CFXuO3YCOGfFxQ-EiEkW=w224-h398-k-no",
			},
			{
				Title: "Fried green tomatoes",
				Image: "https://lh5.googleusercontent.com/p/AF1QipOziHd2hqM1jnK9KfCGf1zVhcOrx8Bj7VdJXj0=w397-h298-k-no",
			},
			{
				Title: "French fries",
				Image: "https://lh5.googleusercontent.com/p/AF1QipNJyq7nAlKtsxxbNy4PHUZOhJ0k7HPP8tTAlwcV=w397-h298-k-no",
			},
			{
				Title: "By owner",
				Image: "https://lh5.googleusercontent.com/p/AF1QipNRE2R5k13zT-0WG4b6XOD_BES9-nMK04hlCMVV=w298-h298-k-no",
			},
			{
				Title: "Street View & 360°",
				Image: "https://lh5.googleusercontent.com/p/AF1QipMwkHP8GmDCSuwnWS7pYVQvtDWdsdk-CUwxtsXL=w224-h298-k-no-pi-23.425545-ya289.20517-ro-8.658787-fo100",
			},
		},
		OrderOnline: []gmaps.LinkSource{
			{
				Link:   "https://foody.com.cy/delivery/lemesos/to-kypriakon?utm_source=google&utm_medium=organic&utm_campaign=google_reserve_place_order_action",
				Source: "foody.com.cy",
			},
			{
				Link:   "https://wolt.com/en/cyp/limassol/restaurant/kypriakon?utm_source=googlemapreserved&utm_campaign=kypriakon",
				Source: "wolt.com",
			},
		},
		Owner: gmaps.Owner{
			ID:   "102769814432182832009",
			Name: "Kipriakon (Owner)",
			Link: "https://www.google.com/maps/contrib/102769814432182832009",
		},
		CompleteAddress: gmaps.Address{
			Borough:    "",
			Street:     "Old port",
			City:       "Limassol",
			PostalCode: "3042",
			State:      "",
			Country:    "CY",
		},
		ReviewsPerRating: map[int]int{
			1: 37,
			2: 16,
			3: 27,
			4: 60,
			5: 256,
		},
	}

	raw, err := os.ReadFile("../testdata/raw.json")
	require.NoError(t, err)
	require.NotEmpty(t, raw)

	entry, err := gmaps.EntryFromJSON(raw)
	require.NoError(t, err)

	require.Len(t, entry.About, 10)

	for _, about := range entry.About {
		require.NotEmpty(t, about.ID)
		require.NotEmpty(t, about.Name)
		require.NotEmpty(t, about.Options)
	}

	entry.About = nil

	require.Len(t, entry.PopularTimes, 7)

	for k, v := range entry.PopularTimes {
		require.Contains(t,
			[]string{
				"Monday",
				"Tuesday",
				"Wednesday",
				"Thursday",
				"Friday",
				"Saturday",
				"Sunday",
			}, k)

		for _, traffic := range v {
			require.GreaterOrEqual(t, traffic, 0)
			require.LessOrEqual(t, traffic, 100)
		}
	}

	monday := entry.PopularTimes["Monday"]
	require.Equal(t, 100, monday[20])

	entry.PopularTimes = nil
	entry.UserReviews = nil
	entry.Media = nil

	require.Equal(t, expected, entry)
}

func Test_EntryFromJSON2(t *testing.T) {
	fnames := []string{
		"../testdata/panic.json",
		"../testdata/panic2.json",
	}
	for _, fname := range fnames {
		raw, err := os.ReadFile(fname)
		require.NoError(t, err)
		require.NotEmpty(t, raw)

		_, err = gmaps.EntryFromJSON(raw)
		require.NoError(t, err)
	}
}

func Test_EntryFromJSONRaw2(t *testing.T) {
	raw, err := os.ReadFile("../testdata/raw2.json")

	require.NoError(t, err)
	require.NotEmpty(t, raw)

	entry, err := gmaps.EntryFromJSON(raw)

	require.NoError(t, err)
	require.Greater(t, len(entry.About), 0)
}

func Test_EntryFromJsonC(t *testing.T) {
	raw, err := os.ReadFile("../testdata/output.json")

	require.NoError(t, err)
	require.NotEmpty(t, raw)

	entries, err := gmaps.ParseSearchResults(raw)

	require.NoError(t, err)

	for _, entry := range entries {
		fmt.Printf("%+v\n", entry)
	}
}

func Test_EntryFromJSON_ExtractMediaWithVideoSources(t *testing.T) {
	darray := make([]any, 52)
	darray[4] = []any{nil, nil, nil, nil, nil, nil, nil, 4.7, 10.0}
	darray[11] = "Test Place"
	darray[13] = []any{"Restaurant"}
	darray[18] = "Test Place, 123 Test St"

	mediaItem := []any{
		"video-token",
		nil,
		nil,
		"Sample Video",
		nil,
		nil,
		[]any{"https://example.com/thumb.jpg", "Sample Place"},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		"Video",
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		[]any{
			6544.0,
			[]any{
				[]any{18.0, 360.0, 640.0, "https://example.com/video-360.mp4", 1.0},
				[]any{37.0, 1080.0, 1920.0, "https://example.com/video-1080.mp4", 1.0},
			},
		},
	}

	darray[51] = []any{[]any{mediaItem}}

	raw, err := json.Marshal([]any{nil, nil, nil, nil, nil, nil, darray})
	require.NoError(t, err)

	entry, err := gmaps.EntryFromJSON(raw)
	require.NoError(t, err)
	require.Len(t, entry.Media, 1)

	media := entry.Media[0]
	require.Equal(t, "video-token", media.ID)
	require.Equal(t, "Video", media.Type)
	require.Equal(t, "Sample Video", media.Title)
	require.Equal(t, "Sample Place", media.Caption)
	require.Equal(t, "https://example.com/thumb.jpg", media.Thumbnail)
	require.Equal(t, "https://example.com/video-1080.mp4", media.URL)
	require.Len(t, media.VideoSources, 2)
	require.Equal(t, "https://example.com/video-360.mp4", media.VideoSources[0].URL)
	require.Equal(t, "https://example.com/video-1080.mp4", media.VideoSources[1].URL)
}

func Test_EntryFromJSONRaw2_HasMedia(t *testing.T) {
	raw, err := os.ReadFile("../testdata/raw2.json")
	require.NoError(t, err)
	require.NotEmpty(t, raw)

	entry, err := gmaps.EntryFromJSON(raw)
	require.NoError(t, err)
	require.Greater(t, len(entry.Media), 0)
}

func Test_EntryFromJSON_WithSecurityPrefix(t *testing.T) {
	raw, err := os.ReadFile("../testdata/raw.json")
	require.NoError(t, err)
	require.NotEmpty(t, raw)

	raw = append([]byte(")]}'\n"), raw...)

	entry, err := gmaps.EntryFromJSON(raw)
	require.NoError(t, err)
	require.Equal(t, "Kipriakon", entry.Title)
}

func Test_EntryFromJSON_ReviewLabelFallbackWhenNoText(t *testing.T) {
	darray := make([]any, 176)
	darray[4] = []any{nil, nil, nil, nil, nil, nil, nil, 4.0, 1.0}
	darray[11] = "Test Place"
	darray[13] = []any{"Restaurant"}
	darray[18] = "Test Place, 123 Test St"

	// Build a review payload with no free-text description but with photo labels.
	// This mirrors image-only reviews where dish tags exist in nested metadata.
	review := []any{
		nil,
		[]any{nil, nil, nil, nil, []any{nil, nil, nil, nil, nil, []any{"Reviewer Name"}}},
		[]any{
			[]any{5.0},
			nil,
			[]any{
				[]any{
					nil,
					[]any{
						nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
						nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
						[]any{"Tomyam Pekat", "Ikan Siakap 3 Rasa"},
					},
				},
			},
		},
	}

	darray[175] = []any{
		nil, nil, nil, []any{0.0, 0.0, 0.0, 0.0, 1.0}, nil, nil, nil, nil, nil,
		[]any{[]any{[]any{review}}},
	}

	raw, err := json.Marshal([]any{nil, nil, nil, nil, nil, nil, darray})
	require.NoError(t, err)

	entry, err := gmaps.EntryFromJSON(raw)
	require.NoError(t, err)
	require.Len(t, entry.UserReviews, 1)
	require.Equal(t, "Reviewer Name", entry.UserReviews[0].Name)
	require.Equal(t, "Tomyam Pekat, Ikan Siakap 3 Rasa", entry.UserReviews[0].Description)
}
