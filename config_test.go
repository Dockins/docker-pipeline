package main

import "testing"

func TestParseCommand(t *testing.T) {

	configs := [][]string{
		{`
build:
    image: maven:3.3.3-jdk-8
    commands:
    -   echo hello    
    -   mvn -version    
    `, `#0 :: 0:build:Command:maven:3.3.3-jdk-8::[echo hello,mvn -version]
`,
		},
	}

	for _, c := range configs {
		p, err := Parse([]byte(c[0]))
		if err != nil {
			t.Errorf("Failed to parse yml content %s", err)
			return
		}

		expected := c[1]
		actual := p.String()
		if actual != expected {
			t.Errorf("Parsed config does not match expected: \n\nexpected:\n%s \n\nactual:\n%s", expected, actual)
			return
		}
	}

}
