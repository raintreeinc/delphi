// Extensions of the original work are copyright (c) 2016 Raintree Systems Inc.
//
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package scanner implements a scanner for Delphi source text.
// It takes a []byte as source which can then be tokenized
// through repeated calls to the Scan method.
//
package scanner

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/raintreeinc/delphi/token"
)

// An ErrorHandler may be provided to Scanner.Init. If a syntax error is
// encountered and a handler was installed, the handler is called with a
// position and an error message. The position points to the beginning of
// the offending token.
//
type ErrorHandler func(pos token.Position, msg string)

// A Scanner holds the scanner's internal state while processing
// a given text. It can be allocated as part of another data
// structure but must be initialized via Init before use.
//
type Scanner struct {
	// immutable state
	file *token.File  // source file handle
	dir  string       // directory portion of file.Name()
	src  []byte       // source
	err  ErrorHandler // error reporting; or nil
	mode Mode         // scanning mode

	// scanning state
	ch         rune // current character
	offset     int  // character offset
	rdOffset   int  // reading offset (position after current character)
	lineOffset int  // current line offset

	// lookbehind
	lastTok token.Token

	// public state - ok to modify
	ErrorCount int // number of errors encountered
}

const bom = 0xFEFF // byte order mark, only permitted as very first character

// Read the next Unicode char into s.ch.
// s.ch < 0 means end-of-file.
//
func (s *Scanner) next() {
	if s.rdOffset < len(s.src) {
		s.offset = s.rdOffset
		if s.ch == '\n' {
			s.lineOffset = s.offset
			s.file.AddLine(s.offset)
		}
		r := rune(s.src[s.rdOffset])
		s.rdOffset += 1
		s.ch = r
	} else {
		s.offset = len(s.src)
		if s.ch == '\n' {
			s.lineOffset = s.offset
			s.file.AddLine(s.offset)
		}
		s.ch = -1 // eof
	}
}

func (s *Scanner) peek() rune {
	if s.rdOffset >= len(s.src) {
		return -1
	}
	return rune(s.src[s.rdOffset])
}

// A mode value is a set of flags (or 0).
// They control scanner behavior.
//
type Mode uint

const (
	ScanComments Mode = 1 << iota // return comments as COMMENT tokens
)

// Init prepares the scanner s to tokenize the text src by setting the
// scanner at the beginning of src. The scanner uses the file set file
// for position information and it adds line information for each line.
// It is ok to re-use the same file when re-scanning the same file as
// line information which is already present is ignored. Init causes a
// panic if the file size does not match the src size.
//
// Calls to Scan will invoke the error handler err if they encounter a
// syntax error and err is not nil. Also, for each error encountered,
// the Scanner field ErrorCount is incremented by one. The mode parameter
// determines how comments are handled.
//
// Note that Init may call err if there is an error in the first character
// of the file.
//
func (s *Scanner) Init(file *token.File, src []byte, err ErrorHandler, mode Mode) {
	// Explicitly initialize all fields since a scanner may be reused.
	if file.Size() != len(src) {
		panic(fmt.Sprintf("file size (%d) does not match src len (%d)", file.Size(), len(src)))
	}
	s.file = file
	s.dir, _ = filepath.Split(file.Name())
	s.src = src
	s.err = err
	s.mode = mode

	s.ch = ' '
	s.offset = 0
	s.rdOffset = 0
	s.lineOffset = 0
	s.ErrorCount = 0

	s.next()
	if s.ch == bom {
		s.next() // ignore BOM at file beginning
	}
}

func (s *Scanner) error(offs int, msg string) {
	if s.err != nil {
		s.err(s.file.Position(s.file.Pos(offs)), msg)
	}
	s.ErrorCount++
}

func (s *Scanner) scanComment(ch rune) string {
	// There are several possibilities how to get here
	//     ch  s.ch
	//     '{'  any
	//     '('  '*'
	//     '/'  '/'

	offs := s.offset - 1 // position of starting ch

	if ch == '/' && s.ch == '/' {
		//-style comment
		s.next()
		for s.ch != '\n' && s.ch >= 0 {
			s.next()
		}
		goto exit
	}

	// *-style comment
	if ch == '(' && s.ch == '*' {
		s.next()
		for s.ch >= 0 {
			ch := s.ch
			s.next()
			if ch == '*' && s.ch == ')' {
				s.next()
				goto exit
			}
		}
	}

	// {-style comment
	for s.ch >= 0 {
		ch := s.ch
		s.next()
		if ch == '}' {
			goto exit
		}
	}

	s.error(offs, "comment not terminated")

exit:
	lit := s.src[offs:s.offset]
	return string(lit)
}

func (s *Scanner) findLineEnd() bool {
	// initial '/' already consumed

	defer func(offs int) {
		// reset scanner state to where it was upon calling findLineEnd
		s.ch = '/'
		s.offset = offs
		s.rdOffset = offs + 1
		s.next() // consume initial '/' again
	}(s.offset - 1)

	// read ahead until a newline, EOF, or non-comment token is found
	for s.ch == '/' || s.ch == '*' {
		if s.ch == '/' {
			//-style comment always contains a newline
			return true
		}
		/*-style comment: look for newline */
		s.next()
		for s.ch >= 0 {
			ch := s.ch
			if ch == '\n' {
				return true
			}
			s.next()
			if ch == '*' && s.ch == '/' {
				s.next()
				break
			}
		}
		s.skipWhitespace()
		if s.ch < 0 || s.ch == '\n' {
			return true
		}
		if s.ch != '/' {
			// non-comment token
			return false
		}
		s.next() // consume '/'
	}

	return false
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

func isHexDigit(ch rune) bool {
	return ('0' <= ch && ch <= '9') || ('a' <= ch && ch <= 'f') || ('A' <= ch && ch <= 'F')
}

func (s *Scanner) scanIdentifier() string {
	offs := s.offset
	for isLetter(s.ch) || isDigit(s.ch) {
		s.next()
	}
	return string(s.src[offs:s.offset])
}

func digitVal(ch rune) int {
	switch {
	case '0' <= ch && ch <= '9':
		return int(ch - '0')
	case 'a' <= ch && ch <= 'f':
		return int(ch - 'a' + 10)
	case 'A' <= ch && ch <= 'F':
		return int(ch - 'A' + 10)
	}
	return 16 // larger than any legal digit val
}

func (s *Scanner) scanMantissa(base int) {
	for digitVal(s.ch) < base {
		s.next()
	}
}

func (s *Scanner) scanNumber() (token.Token, string) {
	// digitVal(s.ch) < 10
	offs := s.offset
	tok := token.INTEGER

	if s.ch == '$' {
		// hexadecimal int
		s.next()
		s.scanMantissa(16)
		if s.offset-offs <= 1 {
			// only scanned "$"
			s.error(offs, "illegal hexadecimal number")
		}
		goto exit
	}

	// decimal int or float
	s.scanMantissa(10)

	// handle ellipsis, e.g. array[0..14]
	if s.ch == '.' && s.peek() == '.' {
		goto exit
	}

	if s.ch == '.' {
		tok = token.FLOAT
		s.next()
		s.scanMantissa(10)
	}

	if s.ch == 'e' || s.ch == 'E' {
		tok = token.FLOAT
		s.next()
		if s.ch == '-' || s.ch == '+' {
			s.next()
		}
		s.scanMantissa(10)
	}

exit:
	return tok, string(s.src[offs:s.offset])
}

func (s *Scanner) scanString() string {
	// '\'' opening already consumed
	offs := s.offset - 1
	for {
		ch := s.ch
		if ch == '\n' || ch < 0 {
			s.error(offs, "string literal not terminated")
			break
		}
		s.next()
		if ch == '\'' {
			if s.ch == '\'' { // 'hello''s'
				s.next()
			} else {
				break
			}
		}
	}

	return string(s.src[offs:s.offset])
}

func (s *Scanner) scanChar() string {
	// '#' opening already consumed
	offs := s.offset - 1

	hex := s.ch == '$'
	if hex {
		s.next()
	}
	for {
		ch := s.ch
		if ch == '\n' || ch < 0 {
			s.error(offs, "char literal not terminated")
			break
		}

		if !isDigit(ch) {
			if !hex || !isHexDigit(ch) {
				break
			}
		}
		s.next()
	}

	return string(s.src[offs:s.offset])
}

func stripCR(b []byte) []byte {
	c := make([]byte, len(b))
	i := 0
	for _, ch := range b {
		if ch != '\r' {
			c[i] = ch
			i++
		}
	}
	return c[:i]
}

func (s *Scanner) skipWhitespace() {
	for s.ch == ' ' || s.ch == '\t' || s.ch == '\n' || s.ch == '\r' {
		s.next()
	}
}

// Helper functions for scanning multi-byte tokens such as >= <= ..
// Different routines recognize different length tok_i based on matches
// of ch_i. If a token ends in '=', the result is tok1 or tok3
// respectively. Otherwise, the result is tok0 if there was no other
// matching character, or tok2 if the matching character was ch2.

func (s *Scanner) switch2(tok0, tok1 token.Token) token.Token {
	if s.ch == '=' {
		s.next()
		return tok1
	}
	return tok0
}

func (s *Scanner) switch3(tok0, tok1 token.Token, ch2 rune, tok2 token.Token) token.Token {
	if s.ch == '=' {
		s.next()
		return tok1
	}
	if s.ch == ch2 {
		s.next()
		return tok2
	}
	return tok0
}

func (s *Scanner) switch4(tok0, tok1 token.Token, ch2 rune, tok2, tok3 token.Token) token.Token {
	if s.ch == '=' {
		s.next()
		return tok1
	}
	if s.ch == ch2 {
		s.next()
		if s.ch == '=' {
			s.next()
			return tok3
		}
		return tok2
	}
	return tok0
}

// Scan scans the next token and returns the token position, the token,
// and its literal string if applicable. The source end is indicated by
// token.EOF.
//
// If the returned token is a literal (token.IDENT, token.INTEGER, token.FLOAT,
// token.STRING) or token.COMMENT, the literal string has the corresponding
// value.
//
// If the returned token is a keyword, the literal string is the keyword.
//
// If the returned token is token.ILLEGAL, the literal string is the
// offending character.
//
// In all other cases, Scan returns an empty literal string.
//
// For more tolerant parsing, Scan will return a valid token if
// possible even if a syntax error was encountered. Thus, even
// if the resulting token sequence contains no illegal tokens,
// a client may not assume that no error occurred. Instead it
// must check the scanner's ErrorCount or the number of calls
// of the error handler, if there was one installed.
//
// Scan adds line information to the file added to the file
// set with Init. Token positions are relative to that file
// and thus relative to the file set.
//
func (s *Scanner) Scan() (pos token.Pos, tok token.Token, lit string) {
scanAgain:
	s.skipWhitespace()

	// current token start
	pos = s.file.Pos(s.offset)

	// determine token value
	switch ch := s.ch; {
	case isLetter(ch):
		lit = s.scanIdentifier()
		if len(lit) > 1 {
			// keywords are longer than one letter - avoid lookup otherwise
			tok = token.Lookup(lit)
		} else {
			tok = token.IDENT
		}
	case ('0' <= ch && ch <= '9') || ch == '$':
		tok, lit = s.scanNumber()
	default:
		s.next() // always make progress
		switch ch {
		case -1:
			tok = token.EOF
		case '\'':
			tok = token.STRING
			lit = s.scanString()
			if len(lit) == 1 {
				tok = token.CHAR
			}
		case '#':
			tok = token.CHAR
			lit = s.scanChar()
		case '^':
			if s.lastTok == token.IDENT || s.lastTok == token.RPAREN || s.lastTok == token.RBRACK {
				// PChar(P)^, P^
				tok = token.HAT
			} else {
				pch := s.peek()
				// ^TRecord
				if isLetter(s.ch) && (isLetter(pch) || isDigit(pch)) {
					// FIXME: This misclassifies
					//   type
					//     T = record end;
					//     PT = ^T;
					// however it is very rare in practice
					tok = token.HAT
				} else {
					tok = token.CHAR
					lit = "^" + string(s.ch)
					s.next()
				}
			}
		case ':':
			tok = s.switch2(token.COLON, token.ASSIGN)
		case '.':
			if s.ch == '.' {
				s.next()
				tok = token.ELLIPSIS
			} else {
				tok = token.PERIOD
			}
		case ',':
			tok = token.COMMA
		case ';':
			tok = token.SEMICOLON
		case '(':
			if s.ch == '*' {
				// comment
				comment := s.scanComment('(')
				if s.mode&ScanComments == 0 {
					// skip comment
					goto scanAgain
				}
				tok = token.COMMENT
				lit = comment
			} else {
				tok = token.LPAREN
			}
		case ')':
			tok = token.RPAREN
		case '[':
			tok = token.LBRACK
		case ']':
			tok = token.RBRACK
		case '+':
			tok = token.ADD
		case '-':
			tok = token.SUB
		case '*':
			tok = token.MUL
		case '/':
			if s.ch == '/' {
				// comment
				comment := s.scanComment('/')
				if s.mode&ScanComments == 0 {
					// skip comment
					goto scanAgain
				}
				tok = token.COMMENT
				lit = comment
			} else {
				tok = token.FDIV
			}
		case '{':
			tok = token.COMMENT
			if s.ch == '$' {
				lit = s.scanComment('{')
				tok = token.CDIRECTIVE
			} else {
				lit = s.scanComment('{')
				if s.mode&ScanComments == 0 {
					lit = ""
					// skip comment
					goto scanAgain
				}
			}
		case '@':
			tok = token.AT
		case '<':
			tok = s.switch3(token.LSS, token.LEQ, '>', token.NEQ)
		case '>':
			tok = s.switch2(token.GTR, token.GEQ)
		case '=':
			tok = token.EQL
		default:
			// next reports unexpected BOMs - don't repeat
			if ch != bom {
				s.error(s.file.Offset(pos), fmt.Sprintf("illegal character %#U", ch))
			}
			tok = token.ILLEGAL
			lit = string(ch)
		}
	}

	// lookbehind token should ignore comments and compiler directives
	if tok != token.COMMENT && tok != token.CDIRECTIVE {
		s.lastTok = tok
	}
	return
}

var (
	ErrStop = errors.New("stop marker for scan")
)

func Scan(src []byte, mode Mode, fn func(tok token.Token, lit string) error, onerr ErrorHandler) error {
	var s Scanner
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(src))
	s.Init(file, src, onerr, mode)

	var tok token.Token
	var lit string
	for tok != token.EOF {
		_, tok, lit = s.Scan()
		err := fn(tok, lit)
		if err != nil {
			if err == ErrStop {
				return nil
			}
			return err
		}

		if s.ErrorCount > 10 {
			return errors.New("too many errors")
		}
	}

	return nil
}
