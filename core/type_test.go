package core

var typeTestCases = []struct {
	v List
	s string
}{
	{List{}, ""},

	{List{
		{"a", false, nil, nil},
	}, `
a
`},

	{List{
		{"a", true, nil, nil},
	}, `
^a
`},

	{List{
		{"a", false, nil, nil},
		{"b", false, nil, nil},
	}, `
a
b
`},

	{List{
		{"a", false, List{
			{"b", false, List{
				{"c", false, nil, nil},
			}, nil},
			{"d", false, nil, nil},
		}, nil},
		{"e", false, nil, nil},
	}, `
a
	b
		c
	d
e
`},

	{List{
		{"a", false, nil, []string{"a1"}},
	}, `
#a1
a
`},

	{List{
		{"a", false, nil, []string{"a1", "a2"}},
	}, `
#a1
#a2
a
`},

	{List{
		{"a", false, nil, []string{"a1", "a2"}},
		{"b", false, nil, []string{"b1", "b2"}},
	}, `
#a1
#a2
a
#b1
#b2
b
`},
}
