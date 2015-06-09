package core

var typeTestCases = []struct {
	v List
	s string
}{
	{List{}, ""},

	{List{
		{"a", nil, nil},
	}, `
a
`},

	{List{
		{"a", nil, nil},
		{"b", nil, nil},
	}, `
a
b
`},

	{List{
		{"a", List{
			{"b", List{
				{"c", nil, nil},
			}, nil},
			{"d", nil, nil},
		}, nil},
		{"e", nil, nil},
	}, `
a
	b
		c
	d
e
`},

	{List{
		{"a", nil, []string{"a1"}},
	}, `
#a1
a
`},

	{List{
		{"a", nil, []string{"a1", "a2"}},
	}, `
#a1
#a2
a
`},

	{List{
		{"a", nil, []string{"a1", "a2"}},
		{"b", nil, []string{"b1", "b2"}},
	}, `
#a1
#a2
a
#b1
#b2
b
`},
}
