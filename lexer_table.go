package goparse

var (
	// Lexical error codes and their strings
	lexErrors = map[string]string{
		"stringne":  "A string cannot be empty",
		"stringesc": `A string escape can must be \\, \t, \n, \', or \"`,
		"rangene":   "A range cannot be empty",
	}

	// Lexical analyzer table, where each row is compressed into a map.
	// Since a rune is actually an int32, use -1 to refer to any other character.
	// If a row does not contain an entry for a given rune, and contains no -1 entry, it is a syntax error.
	lexTable = []map[rune]lexActions{
		// 0 - start
		{
			'\t': {actions: lexSkip | lexAdvance | lexEOFOK, lexType: lexEOF},
			// goiter.RunePositionIter coalesces all EOL sequences into \n
			'\n': {actions: lexSkip | lexAdvance | lexEOFOK, lexType: lexEOF},
			' ':  {actions: lexSkip | lexAdvance | lexEOFOK, lexType: lexEOF},
			'/':  {row: 1},
			'\'': {row: 5},
			'"':  {row: 8},
			'[':  {row: 10},
		},
		// 1
		{
			'/': {actions: lexEOFOK, row: 2, lexType: lexCommentOneLine},
			'*': {row: 3},
		},
		// 2 - comment-one-line
		{
			'\n': {actions: lexUnread | lexDone, lexType: lexCommentOneLine},
			-1:   {actions: lexEOFOK, lexType: lexCommentOneLine, row: 2},
		},
		// 3 - comment-multi-line
		{
			'*': {row: 4},
			-1:  {row: 3},
		},
		// 4
		{
			'*': {row: 4},
			'/': {actions: lexDone, lexType: lexCommentMultiLine},
			-1:  {row: 3},
		},
		// 5 - string: "'" string-sq-chars+ "'"
		{
			'\'': {actions: lexError, errCode: "stringne"},
			'\\': {row: 6},
			-1:   {row: 7},
		},
		// 6
		{
			'\\': {row: 7},
			't':  {row: 7},
			'n':  {row: 7},
			'\'': {row: 7},
			'"':  {row: 7},
			-1:   {actions: lexError, errCode: "stringesc"},
		},
		// 7
		{
			'\'': {actions: lexDone, lexType: lexString},
			'\\': {row: 6},
			-1:   {row: 7},
		},
		// 8 - string: '"' string-qq-chars+ '"'
		{
			'"':  {actions: lexError, errCode: "stringne"},
			'\\': {row: 9},
			-1:   {row: 10},
		},
		// 9
		{
			'\\': {row: 10},
			't':  {row: 10},
			'n':  {row: 10},
			'\'': {row: 10},
			'"':  {row: 10},
		},
		// 10
		{
			'"':  {actions: lexDone, lexType: lexString},
			'\\': {row: 9},
			-1:   {row: 10},
		},
		// 11 - range
		{
			']':  {actions: lexError, errCode: "rangene"},
			'\\': {row: 12},
			-1:   {row: 13},
		},
		// 12
		{
			'\\': {row: 13},
			't':  {row: 13},
			'n':  {row: 13},
			']':  {row: 13},
		},
		// 13
		{
			']':  {actions: lexDone, lexType: lexRange},
			'\\': {row: 12},
			-1:   {row: 13},
		},
	}
)
