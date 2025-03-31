package queryalternatives_test

import (
	"bufio"
	"strings"
	"testing"

	"github.com/kofuk/go-queryalternatives"
	"github.com/stretchr/testify/assert"
)

func Test_ParseString_NoError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected *queryalternatives.Alternatives
	}{
		{
			name: "valid input",
			input: `Name: java
Link: /usr/bin/java
Slaves:
 java.1.gz /usr/share/man/man1/java.1.gz
Status: auto
Best: /usr/lib/jvm/java-21-openjdk-amd64/bin/java
Value: /usr/lib/jvm/java-21-openjdk-amd64/bin/java

Alternative: /usr/lib/jvm/java-21-openjdk-amd64/bin/java
Priority: 2111
Slaves:
 java.1.gz /usr/lib/jvm/java-21-openjdk-amd64/man/man1/java.1.gz
 java.ja.1.gz /usr/lib/jvm/java-21-openjdk-amd64/man/ja/man1/java.1.gz

Alternative: /usr/lib/jvm/java-8-openjdk-amd64/jre/bin/java
Priority: 1081
Slaves:
 java.1.gz /usr/lib/jvm/java-8-openjdk-amd64/jre/man/man1/java.1.gz
`,
			expected: &queryalternatives.Alternatives{
				Name: "java",
				Link: "/usr/bin/java",
				Slaves: map[string]string{
					"java.1.gz": "/usr/share/man/man1/java.1.gz",
				},
				Status: "auto",
				Best:   "/usr/lib/jvm/java-21-openjdk-amd64/bin/java",
				Value:  "/usr/lib/jvm/java-21-openjdk-amd64/bin/java",
				Alternatives: []queryalternatives.Alternative{
					{
						Path:     "/usr/lib/jvm/java-21-openjdk-amd64/bin/java",
						Priority: 2111,
						Slaves: map[string]string{
							"java.1.gz":    "/usr/lib/jvm/java-21-openjdk-amd64/man/man1/java.1.gz",
							"java.ja.1.gz": "/usr/lib/jvm/java-21-openjdk-amd64/man/ja/man1/java.1.gz",
						},
					},
					{
						Path:     "/usr/lib/jvm/java-8-openjdk-amd64/jre/bin/java",
						Priority: 1081,
						Slaves: map[string]string{
							"java.1.gz": "/usr/lib/jvm/java-8-openjdk-amd64/jre/man/man1/java.1.gz",
						},
					},
				},
			},
		},
		{
			name: "valid input with CRLF line endings",
			input: strings.ReplaceAll(`Name: java
Link: /usr/bin/java
Slaves:
 java.1.gz /usr/share/man/man1/java.1.gz
Status: auto
Best: /usr/lib/jvm/java-21-openjdk-amd64/bin/java
Value: /usr/lib/jvm/java-21-openjdk-amd64/bin/java
`, "\n", "\r\n"),
			expected: &queryalternatives.Alternatives{
				Name: "java",
				Link: "/usr/bin/java",
				Slaves: map[string]string{
					"java.1.gz": "/usr/share/man/man1/java.1.gz",
				},
				Status:       "auto",
				Best:         "/usr/lib/jvm/java-21-openjdk-amd64/bin/java",
				Value:        "/usr/lib/jvm/java-21-openjdk-amd64/bin/java",
				Alternatives: []queryalternatives.Alternative{},
			},
		},
		{
			name: "valid input without new line at the end",
			input: `Name: java
Link: /usr/bin/java
Slaves:
 java.1.gz /usr/share/man/man1/java.1.gz
Status: auto
Best: /usr/lib/jvm/java-21-openjdk-amd64/bin/java
Value: /usr/lib/jvm/java-21-openjdk-amd64/bin/java`,
			expected: &queryalternatives.Alternatives{
				Name: "java",
				Link: "/usr/bin/java",
				Slaves: map[string]string{
					"java.1.gz": "/usr/share/man/man1/java.1.gz",
				},
				Status:       "auto",
				Best:         "/usr/lib/jvm/java-21-openjdk-amd64/bin/java",
				Value:        "/usr/lib/jvm/java-21-openjdk-amd64/bin/java",
				Alternatives: []queryalternatives.Alternative{},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			reader := queryalternatives.NewReader(bufio.NewReader(strings.NewReader(test.input)))
			result, err := reader.Read()
			assert.NoError(t, err)
			assert.Equal(t, test.expected, result)
		})
	}
}

func Test_ParseString_Error(t *testing.T) {
	t.Parallel()

	input := `update-alternatives: error: no alternatives for java`
	reader := queryalternatives.NewReader(bufio.NewReader(strings.NewReader(input)))
	result, err := reader.Read()
	assert.Error(t, err, "expected an error")
	assert.Nil(t, result)
}
