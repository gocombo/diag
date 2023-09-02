//go:build !release

package testrand

import (
	"time"

	"github.com/jaswdr/faker"
)

var fake = faker.New()

func Faker() faker.Faker {
	return fake
}

func PastTime() time.Time {
	return fake.Time().TimeBetween(
		time.Now().Add(-time.Hour*24*365*50),
		time.Now().Add(-time.Hour*24*365*20),
	)
}
