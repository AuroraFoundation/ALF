// Code generated by "stringer -trimprefix Token -type Token"; DO NOT EDIT.

package lexer

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[TokenError-0]
	_ = x[TokenEOF-1]
	_ = x[TokenNewline-2]
	_ = x[TokenWhitespace-3]
	_ = x[TokenIndent-4]
	_ = x[TokenColon-5]
	_ = x[TokenComment-6]
	_ = x[TokenName-7]
	_ = x[TokenText-8]
	_ = x[TokenList-9]
}

const _Token_name = "ErrorEOFNewlineWhitespaceIndentColonCommentNameTextList"

var _Token_index = [...]uint8{0, 5, 8, 15, 25, 31, 36, 43, 47, 51, 55}

func (i Token) String() string {
	if i < 0 || i >= Token(len(_Token_index)-1) {
		return "Token(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Token_name[_Token_index[i]:_Token_index[i+1]]
}
