package httptst

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/jaswdr/faker"
)

func RandomHttpReq(
	fake faker.Faker,
	ctx context.Context,
) *http.Request {
	userAgent := fake.UserAgent().UserAgent()
	method := fake.Internet().HTTPMethod()
	path := "/" + fake.Internet().Slug()
	req := httptest.NewRequest(method, path, http.NoBody).WithContext(ctx)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("X-H1-"+fake.Lorem().Word(), fake.Internet().Slug())
	req.Header.Add("X-H2-"+fake.Lorem().Word(), fake.Internet().Slug())
	req.Header.Add("X-H3-"+fake.Lorem().Word(), fake.Internet().Slug())
	query := req.URL.Query()
	query.Add("q1-"+fake.Lorem().Word(), fake.Internet().Slug())
	query.Add("q2-"+fake.Lorem().Word(), fake.Internet().Slug())
	query.Add("q3-"+fake.Lorem().Word(), fake.Internet().Slug())
	req.URL.RawQuery = query.Encode()
	return req
}
