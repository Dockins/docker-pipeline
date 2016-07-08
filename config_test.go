package main

import "testing"

func TestParseCommand(t *testing.T) {

	configs := [][]string{
		{`
stage:
    image: img
    shell: sh
    commands:
    -   run1    
    -   run2    
    -   run3    

`, `
#0 :: 0:stage:Command:img:sh:[run1,run2,run3]
`,
		},
		{`
stage1:
    image: img1
    commands:
    -   run1    

stage2:
    image: img2
    commands:
    -   run2    
`, `
#0 :: 0:stage1:Command:img1::[run1]
#1 :: 1:stage2:Command:img2::[run2]
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
