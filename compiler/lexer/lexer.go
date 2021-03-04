package lexer

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Lexer struct {
	chunk     string // 源代码
	chunkName string // 源文件名
	line      int    // 当前行号

	// 用于辈份词法分析器的状态
	nextToken     string
	nextTokenKind int
	nextTokenLine int
}

var (
	reNewLine            = regexp.MustCompile("\r\n|\n\r|\n|\r")
	reOpeningLongBracket = regexp.MustCompile(`^\[=*\[`)
	reShortStr           = regexp.MustCompile(`(?s)(^'(\\\\|\\'|\\\n|\\z\s*|[^'\n])*')|(^"(\\\\|\\"|\\\n|\\z\s*|[^"\n])*")`)
	reIdentifier         = regexp.MustCompile(`^[_\d\w]+`)
	reNumber             = regexp.MustCompile(`^0[xX][0-9a-fA-F]*(\.[0-9a-fA-F]*)?([pP][+\-]?[0-9]+)?|^[0-9]*(\.[0-9]*)?([eE][+\-]?[0-9]+)?`)

	reDecEscapeSeq     = regexp.MustCompile(`^\\[0-9]{1,3}`)
	reHexEscapeSeq     = regexp.MustCompile(`^\\x[0-9a-fA-F]{2}`)
	reUnicodeEscapeSeq = regexp.MustCompile(`^\\u\{[0-9a-fA-F]+\}`)
)

func NewLexer(chunk, chunkName string) *Lexer {
	return &Lexer{
		chunk:     chunk,
		chunkName: chunkName,
		line:      1,
	}
}

// NextToken:符号和关键字判断
func (self *Lexer) NextToken() (line, kind int, token string) {
	// 之前如果已经读取过了直接取缓存内容
	if self.nextTokenLine > 0 {
		line = self.nextTokenLine
		kind = self.nextTokenKind
		token = self.nextToken
		self.line = self.nextTokenLine
		self.nextTokenLine = 0
		return
	}

	self.skipWhiteSpaces()
	if len(self.chunk) == 0 {
		return self.line, TOKEN_EOF, "EOF"
	}

	// 符号处理
	switch self.chunk[0] {
	case ';':
		self.next(1)
		return self.line, TOKEN_SEP_SEMI, ";"
	case ',':
		self.next(1)
		return self.line, TOKEN_SEP_COMMA, ","
	case '(':
		self.next(1)
		return self.line, TOKEN_SEP_LPAREN, "("
	case ')':
		self.next(1)
		return self.line, TOKEN_SEP_RPAREN, ")"
	case ']':
		self.next(1)
		return self.line, TOKEN_SEP_RBRACK, "]"
	case '{':
		self.next(1)
		return self.line, TOKEN_SEP_LCURLY, "{"
	case '}':
		self.next(1)
		return self.line, TOKEN_SEP_RCURLY, "}"
	case '+':
		self.next(1)
		return self.line, TOKEN_OP_ADD, "+"
	case '-':
		self.next(1)
		return self.line, TOKEN_OP_MINUS, "-"
	case '*':
		self.next(1)
		return self.line, TOKEN_OP_MUL, "*"
	case '^':
		self.next(1)
		return self.line, TOKEN_OP_POW, "^"
	case '%':
		self.next(1)
		return self.line, TOKEN_OP_MOD, "%"
	case '&':
		self.next(1)
		return self.line, TOKEN_OP_BAND, "&"
	case '|':
		self.next(1)
		return self.line, TOKEN_OP_BOR, "|"
	case '#':
		self.next(1)
		return self.line, TOKEN_OP_LEN, "#"
	case ':':
		if self.test("::") {
			self.next(2)
			return self.line, TOKEN_SEP_LABEL, "::"
		} else {
			self.next(1)
			return self.line, TOKEN_SEP_COLON, ":"
		}
	case '/':
		if self.test("//") {
			self.next(2)
			return self.line, TOKEN_OP_IDIV, "//"
		} else {
			self.next(1)
			return self.line, TOKEN_OP_DIV, "/"
		}
	case '~':
		if self.test("~=") {
			self.next(2)
			return self.line, TOKEN_OP_NE, "~="
		} else {
			self.next(1)
			return self.line, TOKEN_OP_WAVE, "~"
		}
	case '=':
		if self.test("==") {
			self.next(2)
			return self.line, TOKEN_OP_EQ, "=="
		} else {
			self.next(1)
			return self.line, TOKEN_OP_ASSIGN, "="
		}
	case '<':
		if self.test("<<") {
			self.next(2)
			return self.line, TOKEN_OP_SHL, "<<"
		} else if self.test("<=") {
			self.next(2)
			return self.line, TOKEN_OP_LE, "<="
		} else {
			self.next(1)
			return self.line, TOKEN_OP_LT, "<"
		}
	case '>':
		if self.test(">>") {
			self.next(2)
			return self.line, TOKEN_OP_SHR, ">>"
		} else if self.test(">=") {
			self.next(2)
			return self.line, TOKEN_OP_GE, ">="
		} else {
			self.next(1)
			return self.line, TOKEN_OP_GT, ">"
		}
	case '.':
		if self.test("...") {
			self.next(3)
			return self.line, TOKEN_VARARG, "..."
		} else if self.test("..") {
			self.next(2)
			return self.line, TOKEN_OP_CONCAT, ".."
		} else if len(self.chunk) == 1 || !isDigit(self.chunk[1]) {
			self.next(1)
			return self.line, TOKEN_SEP_DOT, "."
		}
	case '[':
		if self.test("[[") || self.test("[=") {
			return self.line, TOKEN_STRING, self.scanLongString()
		} else {
			self.next(1)
			return self.line, TOKEN_SEP_LBRACK, "["
		}
	case '\'', '"':
		return self.line, TOKEN_STRING, self.scanShortString()
	}

	// 数字处理
	c := self.chunk[0]
	if c == '.' || isDigit(c) {
		token := self.scanNumber()
		return self.line, TOKEN_NUMBER, token
	}

	// 关键字和标识符处理
	if c == '_' || isLetter(c) {
		token := self.scanIdentifier()
		if kind, found := keywords[token]; found {
			return self.line, kind, token // keyword
		} else {
			return self.line, TOKEN_IDENTIFIER, token
		}
	}

	self.error("unexpected symbol near %q", c)
	return
}

// skipWhiteSpaces:跳过空白字符，一并跳过注释
func (self *Lexer) skipWhiteSpaces() {
	for len(self.chunk) > 0 {
		if self.test("--") {
			self.skipComment()
		} else if self.test("\r\n") || self.test("\n\r") {
			self.next(2)
			self.line += 1
		} else if isNewLine(self.chunk[0]) {
			self.next(1)
			self.line += 1
		} else if isWhiteSpace(self.chunk[0]) {
			self.next(1)
		} else {
			break
		}
	}
}

// test:判断剩余源代码是否以某种字符串开头
func (self *Lexer) test(s string) bool {
	return strings.HasPrefix(self.chunk, s)
}

// next:跳过n个字符
func (self *Lexer) next(n int) {
	self.chunk = self.chunk[n:]
}

// isWhiteSpace:判断字符是否是空白字符
func isWhiteSpace(c byte) bool {
	switch c {
	case '\t', '\n', '\v', '\f', '\r', ' ':
		return true
	}
	return false
}

// isNewLine:判断字符是否是空白字符
func isNewLine(c byte) bool {
	return c == '\r' || c == '\n'
}

// isDigit:判断字符是否为数字
func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// scanNumber:数字扫描，调用scan
func (self *Lexer) scanNumber() string {
	return self.scan(reNumber)
}

// isLetter:判断字符是否为字母
func isLetter(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

func (self *Lexer) scanIdentifier() string {
	return self.scan(reIdentifier)
}

// skipComment:跳过注释
func (self *Lexer) skipComment() {
	self.next(2)        // 跳过--
	if self.test("[") { // 长注释
		if reOpeningLongBracket.FindString(self.chunk) != "" {
			self.scanLongString()
			return
		}
	}
	// 短注释 跳过一整行
	for len(self.chunk) > 0 && !isNewLine(self.chunk[0]) {
		self.next(1)
	}
}

// scanLongString:长字符串处理
func (self *Lexer) scanLongString() string {
	// 1.寻找左右长方括号
	openingLongBracket := reOpeningLongBracket.FindString(self.chunk)
	if openingLongBracket == "" {
		self.error("invalid long string delimiter near '%s'", self.chunk[0:2])
	}

	closingLongBracket := strings.Replace(openingLongBracket, "[", "]", -1)
	closingLongBracketIdx := strings.Index(self.chunk, closingLongBracket)
	if closingLongBracketIdx < 0 {
		self.error("unfinished long string or comment")
	}

	// 2. 提取左右长方括号内的内容
	str := self.chunk[len(openingLongBracket):closingLongBracketIdx]
	self.next(closingLongBracketIdx + len(closingLongBracket))

	// 3. 把换行符的序列同一转换成换行符\n，从而使长注释内的内容归于一行(即所有的'\r\n'转化为\n)
	str = reNewLine.ReplaceAllString(str, "\n")
	self.line += strings.Count(str, "\n")
	// 如果第一个字符是换行符去掉
	if len(str) > 0 && str[0] == '\n' {
		str = str[1:]
	}
	return str
}

// scanShortString:短字符串处理
func (self *Lexer) scanShortString() string {
	if str := reShortStr.FindString(self.chunk); str != "" {
		self.next(len(str))
		str = str[1 : len(str)-1]
		if strings.Index(str, `\`) >= 0 {
			// 处理短字符串中出现转义符\r或者\n
			self.line += len(reNewLine.FindAllString(str, -1))
			str = self.escape(str)
		}
		return str
	}
	self.error("unfinished string")
	return ""
}

// escape:
func (self *Lexer) escape(str string) string {
	var buf bytes.Buffer

	for len(str) > 0 {
		if str[0] != '\\' {
			buf.WriteByte(str[0])
			str = str[1:]
			continue
		}

		if len(str) == 1 {
			self.error("unfinished string")
		}

		switch str[1] {
		case 'a':
			buf.WriteByte('\a')
			str = str[2:]
			continue
		case 'b':
			buf.WriteByte('\b')
			str = str[2:]
			continue
		case 'f':
			buf.WriteByte('\f')
			str = str[2:]
			continue
		case 'n', '\n':
			buf.WriteByte('\n')
			str = str[2:]
			continue
		case 'r':
			buf.WriteByte('\r')
			str = str[2:]
			continue
		case 't':
			buf.WriteByte('\t')
			str = str[2:]
			continue
		case 'v':
			buf.WriteByte('\v')
			str = str[2:]
			continue
		case '"':
			buf.WriteByte('"')
			str = str[2:]
			continue
		case '\'':
			buf.WriteByte('\'')
			str = str[2:]
			continue
		case '\\':
			buf.WriteByte('\\')
			str = str[2:]
			continue
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9': // \ddd
			if found := reDecEscapeSeq.FindString(str); found != "" {
				d, _ := strconv.ParseInt(found[1:], 10, 32)
				if d <= 0xFF {
					buf.WriteByte(byte(d))
					str = str[len(found):]
					continue
				}
				self.error("decimal escape too large near '%s'", found)
			}
		case 'x': // \xXX
			if found := reHexEscapeSeq.FindString(str); found != "" {
				d, _ := strconv.ParseInt(found[2:], 16, 32)
				buf.WriteByte(byte(d))
				str = str[len(found):]
				continue
			}
		case 'u': // \u{XXX}
			if found := reUnicodeEscapeSeq.FindString(str); found != "" {
				d, err := strconv.ParseInt(found[3:len(found)-1], 16, 32)
				if err == nil && d <= 0x10FFFF {
					buf.WriteRune(rune(d))
					str = str[len(found):]
					continue
				}
				self.error("UTF-8 value too large near '%s'", found)
			}
		case 'z':
			str = str[2:]
			for len(str) > 0 && isWhiteSpace(str[0]) { // todo
				str = str[1:]
			}
			continue
		}
		self.error("invalid escape sequence near '\\%c'", str[1])
	}

	return buf.String()
}

// error:词法错误处理
func (self *Lexer) error(f string, a ...interface{}) {
	err := fmt.Sprintf(f, a)
	err = fmt.Sprintf("%s:%d: %s", self.chunkName, self.line, err)
	panic(err)
}

// scan:针对指定正则进行扫描
func (self *Lexer) scan(re *regexp.Regexp) string {
	if token := re.FindString(self.chunk); token != "" {
		self.next(len(token))
		return token
	}
	panic("unreachable")
}

// LookAhead:用于缓存下一个token的信息
func (self *Lexer) LookAhead() int {
	if self.nextTokenLine > 0 {
		return self.nextTokenKind
	}
	currentLine := self.line
	line, kind, token := self.NextToken()
	self.line = currentLine
	self.nextTokenLine = line
	self.nextTokenKind = kind
	self.nextToken = token
	return kind
}

func (self *Lexer) NextIdentifier() (line int, token string) {
	return self.NextTokenOfKind(TOKEN_IDENTIFIER)
}

func (self *Lexer) NextTokenOfKind(kind int) (line int, token string) {
	line, _kind, token := self.NextToken()
	if kind != _kind {
		self.error("syntax error near '%s'", token)
	}
	return line, token
}

func (self *Lexer) Line() int {
	return self.line
}
