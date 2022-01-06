package v1

import (
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func TestBadgeHandler_EntityPattern(t *testing.T) {
	type test struct {
		test string
		key  string
		val  string
	}

	pathPrefix := "/compat/shields/v1/current/today/"

	tests := []test{
		{test: pathPrefix + "project:wakapi", key: "project", val: "wakapi"},
		{test: pathPrefix + "os:Linux", key: "os", val: "Linux"},
		{test: pathPrefix + "editor:VSCode", key: "editor", val: "VSCode"},
		{test: pathPrefix + "language:Java", key: "language", val: "Java"},
		{test: pathPrefix + "machine:devmachine", key: "machine", val: "devmachine"},
		{test: pathPrefix + "label:work", key: "label", val: "work"},
		{test: pathPrefix + "foo:bar", key: "", val: ""},                                   // invalid entity
		{test: pathPrefix + "project:01234", key: "project", val: "01234"},                 // digits only
		{test: pathPrefix + "project:anchr-web-ext", key: "project", val: "anchr-web-ext"}, // with dashes
		{test: pathPrefix + "project:wakapi v2", key: "project", val: "wakapi v2"},         // with blank space
		{test: pathPrefix + "project:project", key: "project", val: "project"},
		{test: pathPrefix + "project:Anchr-Android_v2.0", key: "project", val: "Anchr-Android_v2.0"}, // all the way
	}

	sut := regexp.MustCompile(entityFilterPattern)

	for _, tc := range tests {
		var key, val string
		if groups := sut.FindStringSubmatch(tc.test); len(groups) > 2 {
			key, val = groups[1], groups[2]
		}
		assert.Equal(t, tc.key, key)
		assert.Equal(t, tc.val, val)
	}
}
