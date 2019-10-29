package main

import (
	"io/ioutil"
	"log"
	"testing"
)

func TestServer_validateHttpSchema(t *testing.T) {
	t.Parallel()

	var cases = []struct {
		input  string
		result bool
	}{
		{
			input:  "http",
			result: true,
		},
		{
			input:  "https",
			result: true,
		},
		{
			input:  "hTTpS",
			result: false,
		},
		{
			input:  "hTTp",
			result: false,
		},
		{
			input:  "foo",
			result: false,
		},
		{
			input:  "",
			result: false,
		},
		{
			input:  "foo bar baz",
			result: false,
		},
	}

	s := NewServer("", "", log.New(ioutil.Discard, "", 0))

	for _, testCase := range cases {
		if s.validateHttpSchema(testCase.input) != testCase.result {
			t.Errorf("For [%s] must returns %+v", testCase.input, testCase.result)
		}
	}
}

func TestServer_buildTargetUri(t *testing.T) {
	t.Parallel()

	var cases = []struct {
		schema, domainAndPath, params string
		result                        string
	}{
		{
			schema:        "https",
			domainAndPath: "google.com",
			params:        "foo=bar",
			result:        "https://google.com?foo=bar",
		},
		{
			schema:        "https",
			domainAndPath: "google.com/some/shit",
			params:        "foo=bar&bar=baz",
			result:        "https://google.com/some/shit?foo=bar&bar=baz",
		},
		{
			schema:        "",
			domainAndPath: "google.com",
			params:        "",
			result:        "http://google.com",
		},
		{
			schema:        "ftp",
			domainAndPath: "google.com",
			params:        "",
			result:        "ftp://google.com",
		},
		{
			schema:        "",
			domainAndPath: "",
			params:        "",
			result:        "",
		},
		{
			schema:        "",
			domainAndPath: "a",
			params:        "",
			result:        "http://a",
		},
	}

	s := NewServer("", "", log.New(ioutil.Discard, "", 0))

	for _, testCase := range cases {
		if s.buildTargetUri(testCase.schema, testCase.domainAndPath, testCase.params) != testCase.result {
			t.Errorf(
				"For [%s, %s, %s] must returns %s",
				testCase.schema,
				testCase.domainAndPath,
				testCase.params,
				testCase.result,
			)
		}
	}
}

func TestServer_uriToSchemaAndPath(t *testing.T) {
	t.Parallel()

	var cases = []struct {
		uri,
		expectedSchema,
		expectedPath string
	}{
		{
			uri:            "https/google.com",
			expectedSchema: "https",
			expectedPath:   "google.com",
		},
		{
			uri:            "http/google.com",
			expectedSchema: "http",
			expectedPath:   "google.com",
		},
		{
			uri:            "hTTps/google.COM",
			expectedSchema: "https",
			expectedPath:   "google.COM",
		},
		{
			uri:            "google.com",
			expectedSchema: "",
			expectedPath:   "google.com",
		},
		{
			uri:            "google.com/foo?bar=baz",
			expectedSchema: "",
			expectedPath:   "google.com/foo?bar=baz",
		},
		{
			uri:            "ftp/google.com/foo?bar=baz",
			expectedSchema: "",
			expectedPath:   "ftp/google.com/foo?bar=baz",
		},
	}

	s := NewServer("", "", log.New(ioutil.Discard, "", 0))

	for _, testCase := range cases {
		gotSchema, gotPath := s.uriToSchemaAndPath(testCase.uri)
		if gotSchema != testCase.expectedSchema || gotPath != testCase.expectedPath {
			t.Errorf(
				"For [%s] must returns schema [%s] and path [%s], but returns [%s, %s]",
				testCase.uri,
				testCase.expectedSchema,
				testCase.expectedPath,
				gotSchema, gotPath,
			)
		}
	}
}
