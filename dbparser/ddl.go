package dbparser

import (
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
)

var (
	ErrNotCreateTable = errors.New("not a create table statement")
)

// Token is a struct that represents a lexical token
type Token struct {
	Type  TokenType // The type of the token
	Value string    // The value of the token
}

// Table is a struct that represents a create table statement
type Table struct {
	Name    string // The name of the table
	PkName  string
	Columns []*Column // The slice of columns in the table
	Options []string  // The slice of options for the table
}

// Column is a struct that represents a column in a table
type Column struct {
	Name        string   `json:"name"`        // The name of the column
	Type        string   `json:"type"`        // The type of the column
	Constraints []string `json:"constraints"` // The slice of constraints for the column
}

func (c *Column) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"name":        c.Name,
		"type":        c.Type,
		"constraints": c.Constraints,
	}
}

func (t Table) ToSchema() string {
	sql := fmt.Sprintf("CREATE TABLE %s (\n", t.Name) // start with the table name

	for _, item := range t.Columns {
		sql += fmt.Sprintf("\t%s %s %s,\n", item.Name, item.Type, strings.Join(item.Constraints, " ")) // add each column with its type and extra information
	}
	sql += fmt.Sprintf("\tPRIMARY KEY (%s)\n", t.PkName)
	sql += ");"
	return sql
}

// TokenType is an enum that defines the possible types of tokens
type TokenType int

const (
	// EOF is the end of file token
	EOF TokenType = iota
	// 空白符
	SPACE
	// ILLEGAL is an illegal or unknown token
	ILLEGAL
	// IDENT is an identifier token
	IDENT
	// INT is an integer literal token
	INT
	// STRING is a string literal token
	STRING
	// CREATE is the create keyword token
	CREATE
	// TABLE is the table keyword token
	TABLE
	// IF is the if keyword token
	IF
	// NOT is the not keyword token
	NOT
	// EXISTS is the exists keyword token
	EXISTS
	// LPAREN is the left parenthesis token
	LPAREN
	// RPAREN is the right parenthesis token
	RPAREN
	// COMMA is the comma token
	COMMA
	// ;
	SEMI
)

var keywords = map[string]TokenType{
	"create": CREATE,
	"table":  TABLE,
	"if":     IF,
	"not":    NOT,
	"exists": EXISTS,
}

func ParseMultiSql(allSql string) ([]Table, []string, error) {
	lex := NewLexer(allSql)
	tables, err := NewParser(lex.Tokenize()).Parse()
	if err != nil {
		return nil, nil, err
	}
	return tables, lex.comments, nil
}

func ParserSql(sql string) (Table, error) {
	lex := NewLexer(sql)
	tables, err := NewParser(lex.Tokenize()).Parse()
	if err != nil {
		return Table{}, err
	}
	if len(tables) == 0 {
		return Table{}, errors.New("no tables")
	}
	return tables[0], nil
}

func GetTables(sql string) ([]Table, error) {
	lex := NewLexer(sql)
	tables, err := NewParser(lex.Tokenize()).Parse()
	if err != nil {
		return nil, err
	}
	return tables, nil
}

// GetTableFromSQL extracts the Table structure for a specified table name from SQL statements
// It filters out comments and non-CREATE TABLE statements, and adapts to both SQLite3 and MySQL syntax
func GetTableFromSQL(sql string, tableName string) (Table, error) {
	// Tokenize and parse the SQL
	lex := NewLexer(sql)
	tables, err := NewParser(lex.Tokenize()).Parse()
	if err != nil {
		return Table{}, err
	}

	// Find the table with the specified name (case-insensitive)
	for _, table := range tables {
		if strings.EqualFold(table.Name, tableName) {
			return table, nil
		}
	}

	return Table{}, errors.New("table not found")
}

// LookupIdent returns the token type for an identifier, or ILLEGAL if it's not a keyword
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[strings.ToLower(ident)]; ok {
		return tok
	}
	return IDENT
}

// Lexer is a struct that holds the state of the lexical analysis process
type Lexer struct {
	input    string  // The input string to tokenize
	pos      int     // The current position in the input (points to current char)
	next     int     // The next position in the input (after current char)
	ch       byte    // The current character under examination
	tokens   []Token // The slice of tokens produced by the lexer
	comments []string
}

// NewLexer returns a new instance of a lexer given an input string
func NewLexer(input string) *Lexer {
	if input == "" {
		return &Lexer{}
	}
	input = strings.TrimSpace(input)
	if input[len(input)-1] != ';' {
		input = input + ";"
	}
	l := &Lexer{input: input}
	l.readChar() // Initialize the lexer state by reading the first character
	return l
}

// readChar reads the next character from the input and updates the lexer state
func (l *Lexer) readChar() {
	if l.next >= len(l.input) {
		l.ch = 0 // ASCII code for NUL, signifies end of input
	} else {
		l.ch = l.input[l.next]
	}
	l.pos = l.next // Update the current position to the next position
	l.next++       // Increment the next position by one
}

func (l *Lexer) backChar() {
	l.pos -= 1  // Update the current position to the next position
	l.next -= 1 // Increment the next position by one
	if l.next >= len(l.input) {
		l.ch = 0 // ASCII code for NUL, signifies end of input
	} else {
		l.ch = l.input[l.next]
	}
}

func (l *Lexer) skipChar() {
	l.readChar()
	l.readChar()
}

func (l *Lexer) skipSpace() {
	for l.ch == ' ' {
		l.readChar()
	}
}

func (l *Lexer) skipSpaceAndChar() {
	l.skipSpace()
	l.skipChar()
}

// peekChar returns the next character in the input without consuming it or updating the lexer state
func (l *Lexer) peekChar() byte {
	if l.next >= len(l.input) {
		return 0 // End of input, return NUL byte
	}
	return l.input[l.next] // Return the next character without advancing the lexer state
}

// skipWhitespace skips any whitespace characters in the input and updates the lexer state accordingly
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// skipComment skips any comment in the input and updates the lexer state accordingly
func (l *Lexer) skipComment() {
	if l.ch == '-' && l.peekChar() == '-' { // Single line comment starts with --
		for l.ch != '\n' && l.ch != 0 { // Read until end of line or end of input
			l.readChar()
		}
	}
}

func (l *Lexer) skipAndStoreComment() {
	for l.ch == '-' && l.peekChar() == '-' { // Single line comment starts with --
		comment := []byte{}
		for l.ch != '\n' && l.ch != 0 { // Read until end of line or end of input
			comment = append(comment, l.ch)
			l.readChar()
		}
		if l.ch == '\n' {
			comment = append(comment, l.ch)
			l.readChar()
		}

		l.comments = append(l.comments, strings.TrimSpace(string(comment)[2:]))
	}
}

func (l *Lexer) peekIdent() string {
	start := l.pos
	ch := l.ch                                           // Remember the starting position of the identifier
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' { // An identifier can contain letters, digits and underscores
		l.readChar()
	}
	res := l.input[start:l.pos]
	l.pos = start
	l.next = start + 1
	l.ch = ch
	return res // Return the substring from start to current position as the identifier
}

// readIdent reads an identifier from the input and returns it as a string
func (l *Lexer) readIdent() string {
	start := l.pos                                       // Remember the starting position of the identifier
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' { // An identifier can contain letters, digits and underscores
		l.readChar()
	}
	l.next--
	return l.input[start:l.pos] // Return the substring from start to current position as the identifier
}

// readInt reads an integer literal from the input and returns it as a string
func (l *Lexer) readInt() string {
	start := l.pos      // Remember the starting position of the integer
	for isDigit(l.ch) { // An integer can only contain digits
		l.readChar()
	}
	l.next--
	return l.input[start:l.pos] // Return the substring from start to current position as the integer
}

// readString reads a string literal from the input and returns it as a string
func (l *Lexer) readString() string {
	start := l.pos + 1 // Remember the starting position of the string, skipping the opening quote
	for {
		l.readChar()
		if l.ch == '\\' { // Escape sequence, skip the next character
			l.readChar()
		} else if l.ch == '\'' || l.ch == 0 { // End of string or end of input, break the loop
			break
		}
	}
	return l.input[start:l.pos] // Return the substring from start to current position as the string, excluding the quotes
}

// isLetter returns true if the given byte is an ASCII letter, false otherwise
func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z'
}

// isDigit returns true if the given byte is an ASCII digit, false otherwise
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

// 是否是转义付
func isSpecial(ch byte) bool {
	return ch == '`'
}

// Tokenize performs the lexical analysis on the input and returns a slice of tokens
func (l *Lexer) Tokenize() []Token {
	for l.ch != 0 { // Loop until end of input
		l.skipWhitespace() // Skip any whitespace
		// l.skipComment()    // Skip any comment
		l.skipAndStoreComment()

		var tok Token // Declare a token variable

		switch l.ch { // Switch on the current character
		// case 0:
		// 	tok = Token{Type: SPACE, Value: " "}
		case '\n', ' ', 0:
			l.readChar()
			continue
		case '(': // Left parenthesis, single-character token
			tok = Token{Type: LPAREN, Value: "("}
		case ')': // Right parenthesis, single-character token
			tok = Token{Type: RPAREN, Value: ")"}
		case ',': // Comma, single-character token
			tok = Token{Type: COMMA, Value: ","}
		case ';':
			tok = Token{Type: SEMI, Value: ";"}
		case '\'': // Single quote, start of a string literal
			tok = Token{Type: STRING, Value: l.readString()} // Read the string literal and set the token value
		default:
			if isLetter(l.ch) { // A letter, start of an identifier or a keyword
				ident := l.readIdent()
				// 1、varchar(20) varchar
				if (ident == "varchar" || ident == "VARCHAR") && l.peekChar() == '(' { // Read the identifier
					l.skipChar()
					length := l.readInt()
					l.readChar()
					ident = fmt.Sprintf("VARCHAR(%s)", length)
				} else if (ident == "not" || ident == "NOT") && l.peekChar() == ' ' { // // 2、not null
					l.skipChar()
					if next := l.peekIdent(); next == "null" || next == "NULL" {
						l.readIdent()
						ident = "NOT NULL"
					} else {
						l.backChar()
					}
				} else if (ident == "DECIMAL" || ident == "decimal") && l.peekChar() == '(' {
					l.skipSpaceAndChar()
					digit := l.readInt()
					l.skipSpaceAndChar()
					fix := l.readInt()
					l.skipSpaceAndChar()
					ident = fmt.Sprintf("DECIMAL(%s,%s)", digit, fix)
				} else if ident == "primary" || ident == "PRIMARY" { // // 3、primary key
					l.skipChar()
					if l.peekIdent() == "key" || l.peekIdent() == "KEY" {
						l.readIdent()
						ident = "PRIMARY KEY"
					} else {
						l.backChar()
					}
				}
				tok = Token{Type: LookupIdent(ident), Value: ident} // Lookup the identifier type and set the token value
			} else if isDigit(l.ch) { // A digit, start of an integer literal
				tok = Token{Type: INT, Value: l.readInt()} // Read the integer literal and set the token value
			} else if isSpecial(l.ch) {
				l.readChar()
				continue
			} else { // Anything else, illegal or unknown token
				tok = Token{Type: ILLEGAL, Value: string(l.ch)}
			}
		}

		l.tokens = append(l.tokens, tok) // Append the token to the slice of tokens

		l.readChar() // Read the next character

	}

	l.tokens = append(l.tokens, Token{Type: EOF}) // Append the end of file token to the slice of tokens

	return l.tokens // Return the slice of tokens

}

// Parser is a struct that holds the state of the syntactic analysis process
type Parser struct {
	tokens []Token // The slice of tokens to parse
	pos    int     // The current position in the tokens (points to current token)
}

// NewParser returns a new instance of a parser given a slice of tokens
func NewParser(tokens []Token) *Parser {
	p := &Parser{tokens: tokens}
	// p.table = Table{} // Initialize an empty table struct
	return p
}

// Parse performs the syntactic analysis on the tokens and returns a table struct or an error
func (p *Parser) Parse() ([]Table, error) {
	tables := []Table{}
	for {
		// Check if we've reached the end
		if p.expect(EOF) {
			break
		}

		if t, err := p.parseTable(); err == nil {
			tables = append(tables, t)
		} else if errors.Is(err, ErrNotCreateTable) {
			// Skip non-CREATE TABLE statements until semicolon
			for !p.expect(SEMI) && !p.expect(EOF) {
				p.next()
			}
			if p.expect(SEMI) {
				p.next()
			}
		} else {
			if err != io.EOF {
				return nil, errors.Wrapf(err, "解析失败")
			}
			break
		}
	}
	return tables, nil
}

func (p *Parser) parseTable() (Table, error) {
	var table Table

	if p.expect(EOF) {
		return table, io.EOF
	}

	if !p.expect(CREATE) { // Expect a create keyword token
		return table, errors.WithStack(ErrNotCreateTable)
	}
	p.next() // Advance to next token

	if !p.expect(TABLE) { // Expect a table keyword token
		return table, p.error("expected TABLE")
	}
	p.next() // Advance to next token

	if p.accept(IF) { // Accept an if keyword token
		p.next()            // Advance to next token
		if !p.expect(NOT) { // Expect a not keyword token
			return Table{}, p.error("expected NOT")
		}
		p.next()               // Advance to next token
		if !p.expect(EXISTS) { // Expect an exists keyword token
			return Table{}, p.error("expected EXISTS")
		}
		p.next() // Advance to next token
	}

	if !p.expect(IDENT) { // Expect an identifier token for the table name
		return Table{}, p.error("expected table name")
	}
	table.Name = p.current().Value // Set the table name in the table struct
	p.next()                       // Advance to next token

	if !p.expect(LPAREN) { // Expect a left parenthesis token for the column list
		return table, p.error("expected (")
	}
	p.next() // Advance to next token

	for { // Loop until end of column list or end of input
		// Check if it's a table-level PRIMARY KEY constraint
		v := strings.ToUpper(p.current().Value)
		if strings.Contains(v, "PRIMARY KEY") {
			// Handle table-level PRIMARY KEY
			p.next()
			if !p.expect(LPAREN) {
				return table, p.error("expected (")
			}
			p.next()
			if p.expect(IDENT) {
				table.PkName = p.current().Value
				p.next()
			}
			if p.expect(RPAREN) {
				p.next()
			}
			if p.accept(COMMA) {
				p.next()
				continue
			}
			if p.expect(RPAREN) {
				break
			}
			return table, p.error("expected )")
		}

		if !p.expect(IDENT) { // Expect an identifier token for the column name
			return table, p.error("expected column name")
		}
		column := &Column{Name: p.current().Value} // Create a column struct with the column name
		p.next()                                   // Advance to next token

		if !p.expect(IDENT) { // Expect an identifier token for the column type
			return Table{}, p.error("expected column type")
		}
		column.Type = p.current().Value // Set the column type in the column struct
		p.next()                        // Advance to next token

		for p.accept(IDENT) { // Accept any identifier tokens for the column constraints
			v := strings.ToUpper(p.current().Value)
			if strings.Contains(v, "PRIMARY KEY") {
				table.PkName = column.Name
			}
			column.Constraints = append(column.Constraints, p.current().Value) // Append the constraint to the column struct
			p.next()                                                           // Advance to next token
		}

		table.Columns = append(table.Columns, column) // Append the column to the table struct

		if p.accept(COMMA) { // Accept a comma token as a separator for multiple columns
			p.next() // Advance to next token
			continue // Continue with the next column
		}

		if p.expect(RPAREN) { // Expect a right parenthesis token as the end of the column list
			break // Break the loop
		}

		return table, p.error("expected , or )") // Unexpected token, return an error
	}

	p.next() // Advance to next token

	for !p.expect(SEMI) && !p.expect(EOF) {
		if p.accept(IDENT) {
			table.Options = append(table.Options, p.current().Value) // Append the option to the table struct
		}
		p.next() // Advance to next token
	}
	if p.expect(SEMI) {
		p.next()
	}

	// if !p.expect(EOF) { // Expect an end of file token as the end of the statement
	// 	return nil, p.error("expected EOF")
	// }
	return table, nil
}

// current returns the current token in the tokens slice
func (p *Parser) current() Token {
	return p.tokens[p.pos]
}

// next advances the position in the tokens slice by one
func (p *Parser) next() {
	p.pos++
}

// expect returns true if the current token type matches the given token type, false otherwise
func (p *Parser) expect(t TokenType) bool {
	return p.current().Type == t
}

// accept returns true if the current token type matches one of the given token types, false otherwise
func (p *Parser) accept(types ...TokenType) bool {
	for _, t := range types {
		if p.current().Type == t {
			return true
		}
	}
	return false
}

// error returns an error with a formatted message that includes the current token value and type
func (p *Parser) error(msg string) error {
	return fmt.Errorf("%s, got %s", msg, p.current().Value)
}
