package internal

import (
	"fmt"
	"testing"

	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"
)

var fake = faker.New()

func Test_flattenAndObfuscate(t *testing.T) {
	type args struct {
		values    map[string][]string
		obfuscate []string
	}
	type testCase struct {
		name string
		args args
		want map[string]string
	}
	tests := []func() testCase{
		func() testCase {
			key1 := "key1-" + fake.Lorem().Word()
			val1 := "val1-" + fake.Lorem().Word()
			key2 := "key2-" + fake.Lorem().Word()
			val2 := "val2-" + fake.Lorem().Word()

			values := map[string][]string{
				key1: {val1},
				key2: {val2},
			}
			return testCase{
				name: "single value",
				args: args{values: values},
				want: map[string]string{
					key1: val1,
					key2: val2,
				},
			}
		},
		func() testCase {
			key1 := "key1-" + fake.Lorem().Word()
			val11 := "val11-" + fake.Lorem().Word()
			val12 := "val12-" + fake.Lorem().Word()
			key2 := "key2-" + fake.Lorem().Word()
			val21 := "val21-" + fake.Lorem().Word()
			val22 := "val22" + fake.Lorem().Word()

			values := map[string][]string{
				key1: {val11, val12},
				key2: {val21, val22},
			}
			return testCase{
				name: "multiple values",
				args: args{values: values},
				want: map[string]string{
					key1: val11 + ", " + val12,
					key2: val21 + ", " + val22,
				},
			}
		},
		func() testCase {
			key1 := "key1-" + fake.Lorem().Word()
			val1 := "val1-" + fake.Lorem().Word()
			key2 := "key2-" + fake.Lorem().Word()
			val2 := "val2-" + fake.Lorem().Word()

			values := map[string][]string{
				key1: {val1},
				key2: {val2},
			}
			return testCase{
				name: "obfuscate",
				args: args{values: values, obfuscate: []string{key1}},
				want: map[string]string{
					key1: fmt.Sprint("*obfuscated, length=", len(val1), "*"),
					key2: val2,
				},
			}
		},
		func() testCase {
			val1 := "val1-" + fake.Lorem().Word()

			values := map[string][]string{
				"kEY1": {val1},
			}
			return testCase{
				name: "obfuscate ignore case",
				args: args{values: values, obfuscate: []string{"key1"}},
				want: map[string]string{
					"kEY1": fmt.Sprint("*obfuscated, length=", len(val1), "*"),
				},
			}
		},
	}
	for _, tt := range tests {
		tt := tt()
		t.Run(tt.name, func(t *testing.T) {
			got := FlattenAndObfuscate(tt.args.values, tt.args.obfuscate)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Benchmark_flattenAndObfuscate(b *testing.B) {
	values := map[string][]string{
		"key1": {"val1"},
		"key2": {"val2"},
		"key3": {"val2"},
		"key4": {"val2"},
		"key5": {"val2"},
	}
	keys := []string{"key2", "key4"}

	b.Run("slices as keys", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			FlattenAndObfuscate(values, keys)
		}
	})
}
