package basic

import (
	"fmt"
	"math"
	"strconv"
)

//enumerated type for token type
type tokenType_t int

const (
	INT tokenType_t = iota
	FLOAT
	ADD
	SUB
	MUL
	DIV
	LPAREN
	RPAREN
	EOF
)

// pretty much the only exported function, and it runs all of the code.
// Takes in a string, returns the root node.
// next step is to make it return some sort of EvalResult struct or something.
func Run(txt string, fn string) (*Result_t, error) {
	lex := newLexer(txt, fn)
	tokens, err := lex.makeTokens()
	if err != nil {
		return nil, err
	}

	parser := newParser(tokens)
	ret, err := parser.parse()
	if err != nil {
		return nil, err
	}

	res, err := ret.evaluate()
	if err != nil {
		return nil, err
	}

	return res, nil

}

// struct for the token.
type token_t struct {
	tokenType tokenType_t
	intVal    int32
	floatVal  float64 // GACK! I don't like having to keep 2 different values.
	pos       position_t
}

// gets the string representation of this token
// for example, an integer token with value of 50 would return "INT:50"
// a non-value token (operator token) simply returns it's operator name, like "ADD" or "LPAREN"
func (token token_t) String() string {
	switch token.tokenType {
	case INT:
		return "INT: " + strconv.FormatInt(int64(token.intVal), 10)
	case FLOAT:
		return "FLOAT: " + strconv.FormatFloat(token.floatVal, 'f', -1, 64)
	default:
		return [8]string{"INT", "FLOAT", "ADD", "SUB", "MUL", "DIV", "LPAREN", "RPAREN"}[int(token.tokenType)]
	}
}

// Position struct used to store data about something's position in a file.
type position_t struct {
	index    int
	line     int
	col      int
	filename string
	fileText string
}

// returns a String representation of this position in the form "line 25, column 4 in filename.txt"
func (pos position_t) String() string {
	return fmt.Sprintf("line %d, col %d in file %s", pos.line, pos.col, pos.filename)
}

// constructor for Position objects
func newPosition(name, txt string) position_t {
	return position_t{index: -1, line: 0, col: -1, filename: name, fileText: txt}
}

// Advances this position by incrementing index and col. Wraps over to next line if the current char is a newline.
func (pos *position_t) advance(current byte) {
	pos.index += 1
	pos.col += 1

	if current == '\n' {
		pos.col = 0
		pos.line += 1
	}
}

// Copy constructor for position object.
func (source *position_t) copy() *position_t {
	return &position_t{index: source.index, line: source.line, col: source.col, filename: source.filename, fileText: source.fileText}
}

// Lexer struct. The lexer goes through a string and produces a list of tokens out of it.
type lexer_t struct {
	text        string
	pos         position_t
	currentChar byte
}

// constructor for Lexer object
func newLexer(initStr, filename string) *lexer_t {
	ret := &lexer_t{text: initStr, pos: newPosition(filename, initStr), currentChar: 0}
	ret.advance()
	return ret
}

// Advances the lexer 1 char, updating it's internal position as well.
func (lexer *lexer_t) advance() {
	lexer.pos.advance(lexer.currentChar)
	if lexer.pos.index < len(lexer.text) {
		lexer.currentChar = lexer.text[lexer.pos.index]
	} else {
		lexer.currentChar = 0
	}
}

// makes and returns a list of tokens using the lexer's text.
func (lexer *lexer_t) makeTokens() ([]token_t, error) {
	ret := make([]token_t, 0)

	for {
		if lexer.currentChar == 0 {
			break
		} else if lexer.currentChar == ' ' || lexer.currentChar == '\t' { // skip spaces and tabs
			lexer.advance()
		} else if lexer.currentChar >= '0' && lexer.currentChar <= '9' { // digit, signinfying number literal
			ret = append(ret, lexer.makeNumber())
		} else if lexer.currentChar == '+' {
			ret = append(ret, token_t{tokenType: ADD, pos: *lexer.pos.copy()})
			lexer.advance()
		} else if lexer.currentChar == '-' {
			ret = append(ret, token_t{tokenType: SUB, pos: *lexer.pos.copy()})
			lexer.advance()
		} else if lexer.currentChar == '*' {
			ret = append(ret, token_t{tokenType: MUL, pos: *lexer.pos.copy()})
			lexer.advance()
		} else if lexer.currentChar == '/' {
			ret = append(ret, token_t{tokenType: DIV, pos: *lexer.pos.copy()})
			lexer.advance()
		} else if lexer.currentChar == '(' {
			ret = append(ret, token_t{tokenType: LPAREN, pos: *lexer.pos.copy()})
			lexer.advance()
		} else if lexer.currentChar == ')' {
			ret = append(ret, token_t{tokenType: RPAREN, pos: *lexer.pos.copy()})
			lexer.advance()
		} else { // some other character that isn't implemented
			badChar := lexer.currentChar
			lexer.advance()
			return nil, fmt.Errorf("illegal character '%c' at %s", badChar, lexer.pos)
		}
	}

	ret = append(ret, token_t{tokenType: EOF, pos: *lexer.pos.copy()}) // finish off with an EOF

	return ret, nil
}

// parses the number in the string starting at currentChar.
// can parse an int (a sequence of base-10 digits) or a floating point (a sequence of base-10 digits with 1 decimal point)
// any decimal points after the first one are ignored (and signal end of token)
func (lexer *lexer_t) makeNumber() token_t {
	numStr := ""
	decimalPoints := 0
	pos := lexer.pos.copy()
	for {
		if lexer.currentChar != '.' && !(lexer.currentChar >= '0' && lexer.currentChar <= '9') {
			break
		} else if lexer.currentChar == '.' {
			if decimalPoints == 1 {
				break
			}
			numStr += "."
			decimalPoints += 1
		} else {
			numStr += string(lexer.currentChar)
		}
		lexer.advance()
	}

	if decimalPoints == 0 {
		i, _ := strconv.Atoi(numStr)
		return token_t{tokenType: INT, intVal: int32(i), pos: *pos}
	} else {
		f, _ := strconv.ParseFloat(numStr, 64)
		return token_t{tokenType: FLOAT, floatVal: f, pos: *pos}
	}
}

type nodeType_t int

// enum to signal node type
const (
	FACTOR nodeType_t = iota
	TERM
	EXPRESSION
	UNARY_OP
	NODE_ERR
)

// Nodes used to build the Abstract Syntax Tree (AST)
type node_t struct {
	nodeType nodeType_t
	left     *node_t
	tok      token_t
	right    *node_t
}

// Recursively generate a String representation of this node.
func (node *node_t) String() string {
	if node.nodeType == FACTOR {
		return node.tok.String()
	} else if node.nodeType == UNARY_OP {
		return fmt.Sprintf("(%s, %s)", node.tok.String(), node.left.String())
	} else {
		return fmt.Sprintf("(%s, %s, %s)", node.left.String(), node.tok.String(), node.right.String())
	}
}

// parser_t class. This takes a sequence of tokens and builds
// an abstract syntax tree from them.
type parser_t struct {
	tokens       []token_t
	idx          int
	currentToken token_t
}

// constructor
func newParser(toks []token_t) *parser_t {
	ret := parser_t{tokens: toks, idx: -1}
	ret.advance()
	return &ret
}

// consumes a token and sets currentToken to the next available one
func (parser *parser_t) advance() {
	parser.idx += 1
	if parser.idx < len(parser.tokens) {
		parser.currentToken = parser.tokens[parser.idx]
	}
}

// builds and returns a Factor node using the rules laid out in grammar.txt
func (parser *parser_t) factor() (*node_t, error) {

	if parser.currentToken.tokenType == ADD || parser.currentToken.tokenType == SUB { // Unary operation case-- something like -2
		op := parser.currentToken
		parser.advance()
		factor, err := parser.factor()
		if err != nil {
			return nil, err
		}
		ret := node_t{nodeType: UNARY_OP, tok: op, left: factor}
		return &ret, nil

	} else if parser.currentToken.tokenType == LPAREN { // Parentheses signify the expression case--there's an expression in parentheses.
		parser.advance()
		expr, err := parser.expression()
		if err != nil {
			return nil, err
		}
		if parser.currentToken.tokenType == RPAREN {
			parser.advance()
			return expr, nil
		} else {
			return &node_t{nodeType: NODE_ERR}, fmt.Errorf("expected ')' at %s", parser.currentToken.pos.String())
		}
	} else if parser.currentToken.tokenType == INT || parser.currentToken.tokenType == FLOAT { // number literal case
		ret := node_t{nodeType: FACTOR, tok: parser.currentToken}
		parser.advance()
		return &ret, nil
	}
	return &node_t{nodeType: NODE_ERR}, fmt.Errorf("expected factor at %s", parser.currentToken.pos.String())
}

// builds and returns a Term node
func (parser *parser_t) term() (*node_t, error) {
	left, err := parser.factor()
	if err != nil {
		return nil, err
	}

	for { // GACK! my way of writing a while loop--seems wrong.
		if parser.currentToken.tokenType != MUL && parser.currentToken.tokenType != DIV {
			break
		}
		operator := parser.currentToken
		parser.advance()
		right, err := parser.factor()
		if err != nil {
			return nil, err
		}
		left = &node_t{nodeType: TERM, left: left, tok: operator, right: right}
	}

	return left, nil
}

// builds and returns an Expression node
func (parser *parser_t) expression() (*node_t, error) {
	left, err := parser.term()
	if err != nil {
		return nil, err
	}

	for { // GACK! my way of writing a while loop--seems wrong.
		if parser.currentToken.tokenType != ADD && parser.currentToken.tokenType != SUB {
			break
		}
		operator := parser.currentToken
		parser.advance()
		right, err := parser.term()
		if err != nil {
			return nil, err
		}
		left = &node_t{nodeType: TERM, left: left, tok: operator, right: right}
	}

	return left, nil
}

// wrapper for parsing expression (kicks off recursion). Also checks for EOF.
func (parser *parser_t) parse() (*node_t, error) {
	ret, err := parser.expression()
	if err == nil && parser.currentToken.tokenType != EOF {
		err = fmt.Errorf("exprected operator at %s", parser.currentToken.pos.String())
	}
	return ret, err
}

type resultType_t int

// enum for result types.
const (
	INTEGER resultType_t = iota
	FLOATING
)

// container for Results.
type Result_t struct {
	ResultType resultType_t
	Ires       int32 // GACK! Any way to just use a single return or something like that?
	Fres       float64
}

// returns a String representation of this result.
func (res *Result_t) String() string {
	if res.ResultType == INTEGER {
		return fmt.Sprintf("Result: %d", res.Ires)
	} else {
		return fmt.Sprintf("Result: %f", res.Fres)
	}
}

// absolute value of an int
func abs(num int32) int32 {
	if num < 0 {
		return -1 * num
	}
	return num
}

// performs the given operation on the given integers and returns the result.
func intop(left, right int32, op tokenType_t) int32 {
	switch op {
	case ADD:
		return left + right
	case SUB:
		return left - right
	case MUL:
		return left * right
	case DIV:
		return left / right // TODO: add div by 0 check
	default:
		return 0
	}
}

// GACK! Literally the exact same as intop, just for floats.
func floatop(left, right float64, op tokenType_t) float64 {
	switch op {
	case ADD:
		return left + right
	case SUB:
		return left - right
	case MUL:
		return left * right
	case DIV:
		return left / right // TODO: add div by 0 check
	default:
		return 0
	}
}

// recursively evaluate a node, returning result struct
func (node *node_t) evaluate() (*Result_t, error) {
	switch node.nodeType {
	case FACTOR: // base case, just return a result with
		if node.tok.tokenType == INT {
			return &Result_t{ResultType: INTEGER, Ires: node.tok.intVal, Fres: float64(node.tok.intVal)}, nil // set the float value too in case we have to upcast to float
		} else {
			return &Result_t{ResultType: FLOATING, Fres: node.tok.floatVal}, nil
		}
	case UNARY_OP: // case of an unary operation, need to evaluate child then apply unary operation
		factorRes, err := node.left.evaluate()
		if err != nil {
			return nil, err
		}
		if node.tok.tokenType == SUB { // negative sign
			if factorRes.ResultType == INTEGER { // GACK! Any way to make this work for both ints and floats?
				return &Result_t{ResultType: INTEGER, Ires: -1 * factorRes.Ires, Fres: -1 * float64(factorRes.Ires)}, nil // set the float value too in case we have to upcast to float
			} else {
				return &Result_t{ResultType: FLOATING, Fres: -1 * factorRes.Fres}, nil
			}
		} else if node.tok.tokenType == ADD { // positive sign
			if factorRes.ResultType == INTEGER { // GACK! Any way to make this work for both ints and floats?
				return &Result_t{ResultType: INTEGER, Ires: abs(factorRes.Ires), Fres: float64(abs(factorRes.Ires))}, nil // set the float value too in case we have to upcast to float
			} else {
				return &Result_t{ResultType: FLOATING, Fres: math.Abs(factorRes.Fres)}, nil
			}
		}
	case TERM, EXPRESSION: // both terms and expressions are binary operations. We need to evaluate both children, then apply the operation
		leftRes, err := node.left.evaluate()
		if err != nil {
			return nil, err
		}
		rightRes, err := node.right.evaluate()
		if err != nil {
			return nil, err
		}
		ret := &Result_t{ResultType: INTEGER} // default to integer
		if leftRes.ResultType == FLOATING || rightRes.ResultType == FLOATING {
			ret.Fres = floatop(leftRes.Fres, rightRes.Fres, node.tok.tokenType)
			ret.ResultType = FLOATING
			return ret, nil
		}
		// GACK! Any way to make this work for both ints and floats?
		ret.Ires = intop(leftRes.Ires, rightRes.Ires, node.tok.tokenType)
		ret.Fres = float64(ret.Ires)
		return ret, nil
	}
	return nil, fmt.Errorf("evaluation error at %s", node.tok.pos.String())
}
