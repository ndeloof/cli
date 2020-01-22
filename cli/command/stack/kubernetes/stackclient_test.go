package kubernetes

import (
	"bytes"
	"io/ioutil"
	"testing"

	composetypes "github.com/docker/compose-go/types"
	"gotest.tools/assert"
)

func TestFromCompose(t *testing.T) {
	stackClient := &stackV1Beta1{}
	s, err := stackClient.FromCompose(ioutil.Discard, "foo", &composetypes.Config{
		Version:  "3.1",
		Filename: "banana",
		Services: []composetypes.ServiceConfig{
			{
				Name:  "foo",
				Image: "foo",
			},
			{
				Name:  "bar",
				Image: "bar",
			},
		},
	})
	assert.NilError(t, err)
	assert.Equal(t, "foo", s.Name)
	assert.Equal(t, string(`version: "3.5"
services:
  bar:
    image: bar
  foo:
    image: foo
`), s.ComposeFile)
}

func TestFromComposeUnsupportedVersion(t *testing.T) {
	var stderr bytes.Buffer
	stackClient := &stackV1Beta1{}
	_, err := stackClient.FromCompose(&stderr, "foo", &composetypes.Config{
		Version:  "3.6",
		Filename: "banana",
		Services: []composetypes.ServiceConfig{
			{
				Name:  "foo",
				Image: "foo",
				Volumes: []composetypes.ServiceVolumeConfig{
					{
						Type:   "tmpfs",
						Target: "/app",
						Tmpfs: &composetypes.ServiceVolumeTmpfs{
							Size: 10000,
						},
					},
				},
			},
		},
	})
	assert.NilError(t, err)
	assert.Equal(t, stderr.String(), "service \"foo\": tmpfs is not supported\n")
}
